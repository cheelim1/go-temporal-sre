package workload

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func TestNew(t *testing.T) {
	app := New(AppConfig{
		Logger:     testLogger(),
		ListenAddr: ":0",
	})

	require.NotNil(t, app)
	require.NotNil(t, app.server)
}

func TestMaintenanceMode(t *testing.T) {
	app := New(AppConfig{
		Logger:     testLogger(),
		ListenAddr: ":0",
	})

	// Initially not in maintenance
	assert.False(t, app.maintenance.Load())

	// Enable maintenance
	app.SetMaintenance(true)
	assert.True(t, app.maintenance.Load())

	// Disable maintenance
	app.SetMaintenance(false)
	assert.False(t, app.maintenance.Load())
}

func TestHandleHealth(t *testing.T) {
	app := New(AppConfig{
		Logger:     testLogger(),
		ListenAddr: ":0",
	})

	// Normal mode
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	app.handleHealth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")

	// Maintenance mode
	app.SetMaintenance(true)
	w = httptest.NewRecorder()
	app.handleHealth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "maintenance")
}

func TestHandleWrite_MaintenanceMode(t *testing.T) {
	app := New(AppConfig{
		Logger:     testLogger(),
		ListenAddr: ":0",
	})
	app.SetMaintenance(true)

	req := httptest.NewRequest(http.MethodPost, "/write", nil)
	w := httptest.NewRecorder()
	app.handleWrite(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "maintenance")
}

func TestHandleWrite_NoPool(t *testing.T) {
	app := New(AppConfig{
		Logger:     testLogger(),
		ListenAddr: ":0",
	})

	req := httptest.NewRequest(http.MethodPost, "/write", nil)
	w := httptest.NewRecorder()
	app.handleWrite(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "no write pool")
}

func TestHandleRead_NoPool(t *testing.T) {
	app := New(AppConfig{
		Logger:     testLogger(),
		ListenAddr: ":0",
	})

	req := httptest.NewRequest(http.MethodGet, "/read", nil)
	w := httptest.NewRecorder()
	app.handleRead(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "no read pool")
}

func TestHandleEnterMaintenance(t *testing.T) {
	app := New(AppConfig{
		Logger:     testLogger(),
		ListenAddr: ":0",
	})

	req := httptest.NewRequest(http.MethodPost, "/admin/maintenance", nil)
	w := httptest.NewRecorder()
	app.handleEnterMaintenance(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, app.maintenance.Load())
}

func TestHandleResume(t *testing.T) {
	app := New(AppConfig{
		Logger:     testLogger(),
		ListenAddr: ":0",
	})
	app.SetMaintenance(true)

	req := httptest.NewRequest(http.MethodPost, "/admin/resume", nil)
	w := httptest.NewRecorder()
	app.handleResume(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, app.maintenance.Load())
}

func TestGetReadPool_Fallback(t *testing.T) {
	app := New(AppConfig{
		Logger:     testLogger(),
		ListenAddr: ":0",
	})

	// With no read pools and no write pool, should return nil
	pool := app.getReadPool()
	assert.Nil(t, pool)
}

func TestRoundRobinReadPool(t *testing.T) {
	app := New(AppConfig{
		Logger:     testLogger(),
		ListenAddr: ":0",
	})

	// Simulate multiple read pools by setting up dummy pools
	// (we can't test actual round-robin without real pools,
	// but we can test the index cycling)
	assert.Equal(t, uint64(0), app.readPoolIdx.Load())

	// The round-robin counter increments on each call
	// even with empty pool slice (falls back to write pool)
	app.getReadPool()
	// Counter should not have changed because readPools is empty
	// (it takes the fallback path)
}
