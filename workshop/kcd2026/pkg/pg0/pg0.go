package pg0

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Instance represents a pg0 PostgreSQL instance
type Instance struct {
	Name            string
	Port            int
	DataDir         string
	Username        string
	Password        string
	Database        string
	ConfigOverrides []string
	logger          *slog.Logger
}

// InstanceConfig holds configuration for creating a new Instance
type InstanceConfig struct {
	Name            string
	Port            int // 0 for auto-allocation
	DataDir         string
	Username        string
	Password        string
	Database        string
	ConfigOverrides []string
	Logger          *slog.Logger
}

// NewInstance creates a new pg0 Instance with the given configuration
func NewInstance(cfg InstanceConfig) (*Instance, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("instance name is required")
	}

	// Auto-allocate port if not specified
	port := cfg.Port
	if port == 0 {
		allocatedPort, err := AllocatePort()
		if err != nil {
			return nil, fmt.Errorf("failed to allocate port: %w", err)
		}
		port = allocatedPort
	}

	// Set defaults
	if cfg.Username == "" {
		cfg.Username = "postgres"
	}
	if cfg.Password == "" {
		cfg.Password = "postgres"
	}
	if cfg.Database == "" {
		cfg.Database = "postgres"
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	// Always include wal_level=logical for pgroll and pgstream
	configOverrides := append([]string{"wal_level=logical"}, cfg.ConfigOverrides...)

	return &Instance{
		Name:            cfg.Name,
		Port:            port,
		DataDir:         cfg.DataDir,
		Username:        cfg.Username,
		Password:        cfg.Password,
		Database:        cfg.Database,
		ConfigOverrides: configOverrides,
		logger:          cfg.Logger,
	}, nil
}

// Start starts the pg0 instance
func (i *Instance) Start(ctx context.Context) error {
	i.logger.Info("starting pg0 instance",
		"name", i.Name,
		"port", i.Port,
	)

	args := []string{
		"start",
		"--name", i.Name,
		"--port", strconv.Itoa(i.Port),
		"--username", i.Username,
		"--password", i.Password,
		"--database", i.Database,
	}

	if i.DataDir != "" {
		args = append(args, "--data-dir", i.DataDir)
	}

	// Add config overrides
	for _, override := range i.ConfigOverrides {
		args = append(args, "-c", override)
	}

	cmd := exec.CommandContext(ctx, "pg0", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start pg0: %w", err)
	}

	// Wait for PostgreSQL to be ready
	if err := i.waitForReady(ctx); err != nil {
		return fmt.Errorf("pg0 started but not ready: %w", err)
	}

	i.logger.Info("pg0 instance started successfully",
		"name", i.Name,
		"port", i.Port,
		"connString", i.ConnString(),
	)

	return nil
}

// Stop stops the pg0 instance
func (i *Instance) Stop(ctx context.Context) error {
	i.logger.Info("stopping pg0 instance", "name", i.Name)

	cmd := exec.CommandContext(ctx, "pg0", "stop", "--name", i.Name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop pg0: %w", err)
	}

	i.logger.Info("pg0 instance stopped", "name", i.Name)
	return nil
}

// Drop stops and permanently deletes the instance
func (i *Instance) Drop(ctx context.Context) error {
	i.logger.Info("dropping pg0 instance", "name", i.Name)

	cmd := exec.CommandContext(ctx, "pg0", "drop", "--name", i.Name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to drop pg0: %w", err)
	}

	i.logger.Info("pg0 instance dropped", "name", i.Name)
	return nil
}

// ConnString returns the PostgreSQL connection string
func (i *Instance) ConnString() string {
	return fmt.Sprintf("postgresql://%s:%s@localhost:%d/%s",
		i.Username, i.Password, i.Port, i.Database)
}

// IsRunning checks if the instance is running
func (i *Instance) IsRunning(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "pg0", "info", "--name", i.Name, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Simple check: if output contains "running": true
	return strings.Contains(string(output), `"running":true`) ||
		strings.Contains(string(output), `"running": true`)
}

// waitForReady waits for PostgreSQL to be ready to accept connections
func (i *Instance) waitForReady(ctx context.Context) error {
	timeout := 30 * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Try to connect to the port
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", i.Port), time.Second)
		if err == nil {
			conn.Close()
			// Give it a bit more time to fully initialize
			time.Sleep(500 * time.Millisecond)
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for pg0 to be ready after %v", timeout)
}

// AllocatePort finds and returns an available ephemeral port
func AllocatePort() (int, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, fmt.Errorf("failed to allocate port: %w", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}
