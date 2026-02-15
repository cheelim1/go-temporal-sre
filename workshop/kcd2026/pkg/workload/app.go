package workload

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// App is the workload generator HTTP API with read/write splitting
type App struct {
	logger *slog.Logger

	// Write pool (primary)
	writeMu   sync.RWMutex
	writePool *pgxpool.Pool

	// Read pool (round-robin across replicas)
	readMu      sync.RWMutex
	readPools   []*pgxpool.Pool
	readPoolIdx atomic.Uint64

	// Maintenance mode
	maintenance atomic.Bool

	server *http.Server
}

// AppConfig configures the workload app
type AppConfig struct {
	PrimaryConnStr  string
	ReplicaConnStrs []string
	ListenAddr      string
	Logger          *slog.Logger
}

// New creates a new workload App
func New(cfg AppConfig) *App {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
	}

	app := &App{
		logger: cfg.Logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /write", app.handleWrite)
	mux.HandleFunc("GET /read", app.handleRead)
	mux.HandleFunc("POST /admin/maintenance", app.handleEnterMaintenance)
	mux.HandleFunc("POST /admin/resume", app.handleResume)
	mux.HandleFunc("GET /health", app.handleHealth)

	app.server = &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: mux,
	}

	return app
}

// Start initializes connection pools and starts the HTTP server
func (a *App) Start(ctx context.Context, cfg AppConfig) error {
	// Connect to primary
	if cfg.PrimaryConnStr != "" {
		pool, err := pgxpool.New(ctx, cfg.PrimaryConnStr)
		if err != nil {
			return fmt.Errorf("failed to connect to primary: %w", err)
		}
		a.writePool = pool
		a.logger.Info("connected to primary", "conn", cfg.PrimaryConnStr)
	}

	// Connect to replicas
	for _, connStr := range cfg.ReplicaConnStrs {
		pool, err := pgxpool.New(ctx, connStr)
		if err != nil {
			a.logger.Warn("failed to connect to replica", "conn", connStr, "error", err)
			continue
		}
		a.readPools = append(a.readPools, pool)
		a.logger.Info("connected to replica", "conn", connStr)
	}

	// Start HTTP server
	go func() {
		a.logger.Info("starting HTTP server", "addr", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("HTTP server error", "error", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the app
func (a *App) Stop(ctx context.Context) error {
	a.logger.Info("shutting down app")

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("failed to shutdown HTTP server", "error", err)
	}

	a.writeMu.Lock()
	if a.writePool != nil {
		a.writePool.Close()
	}
	a.writeMu.Unlock()

	a.readMu.Lock()
	for _, pool := range a.readPools {
		pool.Close()
	}
	a.readMu.Unlock()

	return nil
}

// UpdateWritePool replaces the write pool with a new connection
func (a *App) UpdateWritePool(ctx context.Context, connStr string) error {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to new primary: %w", err)
	}

	a.writeMu.Lock()
	old := a.writePool
	a.writePool = pool
	a.writeMu.Unlock()

	if old != nil {
		old.Close()
	}

	a.logger.Info("write pool updated", "conn", connStr)
	return nil
}

// UpdateReadPools replaces all read pools
func (a *App) UpdateReadPools(ctx context.Context, connStrs []string) error {
	var pools []*pgxpool.Pool
	for _, connStr := range connStrs {
		pool, err := pgxpool.New(ctx, connStr)
		if err != nil {
			// Close already opened pools
			for _, p := range pools {
				p.Close()
			}
			return fmt.Errorf("failed to connect to replica %s: %w", connStr, err)
		}
		pools = append(pools, pool)
	}

	a.readMu.Lock()
	old := a.readPools
	a.readPools = pools
	a.readMu.Unlock()

	for _, p := range old {
		p.Close()
	}

	a.logger.Info("read pools updated", "count", len(pools))
	return nil
}

// SetMaintenance toggles maintenance mode
func (a *App) SetMaintenance(on bool) {
	a.maintenance.Store(on)
	a.logger.Info("maintenance mode changed", "maintenance", on)
}

// getReadPool returns a read pool using round-robin
func (a *App) getReadPool() *pgxpool.Pool {
	a.readMu.RLock()
	defer a.readMu.RUnlock()

	if len(a.readPools) == 0 {
		// Fall back to write pool for reads
		a.writeMu.RLock()
		defer a.writeMu.RUnlock()
		return a.writePool
	}

	idx := a.readPoolIdx.Add(1) - 1
	return a.readPools[idx%uint64(len(a.readPools))]
}

// HTTP Handlers

func (a *App) handleWrite(w http.ResponseWriter, r *http.Request) {
	if a.maintenance.Load() {
		http.Error(w, `{"error":"service in maintenance mode"}`, http.StatusServiceUnavailable)
		return
	}

	a.writeMu.RLock()
	pool := a.writePool
	a.writeMu.RUnlock()

	if pool == nil {
		http.Error(w, `{"error":"no write pool configured"}`, http.StatusInternalServerError)
		return
	}

	// Insert a record with timestamp
	var id int
	err := pool.QueryRow(r.Context(),
		"INSERT INTO orders (user_id, amount, status) VALUES (1, $1, 'pending') RETURNING id",
		time.Now().UnixMilli()%10000,
	).Scan(&id)

	if err != nil {
		a.logger.Error("write failed", "error", err)
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":     id,
		"status": "created",
	})
}

func (a *App) handleRead(w http.ResponseWriter, r *http.Request) {
	pool := a.getReadPool()
	if pool == nil {
		http.Error(w, `{"error":"no read pool configured"}`, http.StatusInternalServerError)
		return
	}

	rows, err := pool.Query(r.Context(),
		"SELECT id, user_id, amount, status, created_at FROM orders ORDER BY id DESC LIMIT 10",
	)
	if err != nil {
		a.logger.Error("read failed", "error", err)
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type order struct {
		ID        int       `json:"id"`
		UserID    int       `json:"user_id"`
		Amount    float64   `json:"amount"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"created_at"`
	}

	var orders []order
	for rows.Next() {
		var o order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Amount, &o.Status, &o.CreatedAt); err != nil {
			continue
		}
		orders = append(orders, o)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"orders": orders,
		"count":  len(orders),
	})
}

func (a *App) handleEnterMaintenance(w http.ResponseWriter, r *http.Request) {
	a.SetMaintenance(true)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "maintenance mode enabled"})
}

func (a *App) handleResume(w http.ResponseWriter, r *http.Request) {
	a.SetMaintenance(false)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "normal mode resumed"})
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := "healthy"
	if a.maintenance.Load() {
		status = "maintenance"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}
