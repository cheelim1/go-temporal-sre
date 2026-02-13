package pg0

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInstance(t *testing.T) {
	tests := []struct {
		name    string
		cfg     InstanceConfig
		wantErr bool
	}{
		{
			name: "valid config with auto port",
			cfg: InstanceConfig{
				Name: "test-instance",
			},
			wantErr: false,
		},
		{
			name: "valid config with explicit port",
			cfg: InstanceConfig{
				Name: "test-instance",
				Port: 5555,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			cfg: InstanceConfig{
				Port: 5432,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance, err := NewInstance(tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.cfg.Name, instance.Name)
			assert.NotZero(t, instance.Port)
			assert.Equal(t, "postgres", instance.Username)
			assert.Equal(t, "postgres", instance.Password)
			assert.Equal(t, "postgres", instance.Database)
			assert.Contains(t, instance.ConfigOverrides, "wal_level=logical")
		})
	}
}

func TestAllocatePort(t *testing.T) {
	port1, err := AllocatePort()
	require.NoError(t, err)
	assert.Greater(t, port1, 1024)

	port2, err := AllocatePort()
	require.NoError(t, err)
	assert.Greater(t, port2, 1024)

	// Ports should be different (highly likely)
	assert.NotEqual(t, port1, port2)
}

func TestInstance_StartStop(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	instance, err := NewInstance(InstanceConfig{
		Name:   "test-start-stop",
		Logger: logger,
	})
	require.NoError(t, err)

	// Ensure cleanup
	t.Cleanup(func() {
		dropCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = instance.Drop(dropCtx)
	})

	// Start the instance
	err = instance.Start(ctx)
	require.NoError(t, err)

	// Verify it's running
	assert.True(t, instance.IsRunning(ctx))

	// Verify connection string format
	connStr := instance.ConnString()
	assert.Contains(t, connStr, "postgresql://")
	assert.Contains(t, connStr, "localhost")

	// Stop the instance
	err = instance.Stop(ctx)
	require.NoError(t, err)

	// Give it a moment to fully stop
	time.Sleep(time.Second)

	// Verify it's not running
	assert.False(t, instance.IsRunning(ctx))
}

func TestInstance_PostgresConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	instance, err := NewInstance(InstanceConfig{
		Name:   "test-connectivity",
		Logger: logger,
	})
	require.NoError(t, err)

	// Ensure cleanup
	t.Cleanup(func() {
		dropCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = instance.Drop(dropCtx)
	})

	// Start the instance
	err = instance.Start(ctx)
	require.NoError(t, err)

	// Connect using pgx
	conn, err := pgx.Connect(ctx, instance.ConnString())
	require.NoError(t, err)
	defer conn.Close(ctx)

	// Verify connection with a simple query
	var version string
	err = conn.QueryRow(ctx, "SELECT version()").Scan(&version)
	require.NoError(t, err)
	assert.Contains(t, version, "PostgreSQL")
	t.Logf("Connected to: %s", version)

	// Verify wal_level is set to logical
	var walLevel string
	err = conn.QueryRow(ctx, "SHOW wal_level").Scan(&walLevel)
	require.NoError(t, err)
	assert.Equal(t, "logical", walLevel, "wal_level should be set to logical")
}

func TestInstance_MultipleInstances(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create multiple instances
	instance1, err := NewInstance(InstanceConfig{
		Name:   "test-multi-1",
		Logger: logger,
	})
	require.NoError(t, err)

	instance2, err := NewInstance(InstanceConfig{
		Name:   "test-multi-2",
		Logger: logger,
	})
	require.NoError(t, err)

	instance3, err := NewInstance(InstanceConfig{
		Name:   "test-multi-3",
		Logger: logger,
	})
	require.NoError(t, err)

	// Ensure cleanup
	t.Cleanup(func() {
		dropCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = instance1.Drop(dropCtx)
		_ = instance2.Drop(dropCtx)
		_ = instance3.Drop(dropCtx)
	})

	// Start all instances
	err = instance1.Start(ctx)
	require.NoError(t, err)

	err = instance2.Start(ctx)
	require.NoError(t, err)

	err = instance3.Start(ctx)
	require.NoError(t, err)

	// Verify all have different ports
	assert.NotEqual(t, instance1.Port, instance2.Port)
	assert.NotEqual(t, instance1.Port, instance3.Port)
	assert.NotEqual(t, instance2.Port, instance3.Port)

	// Verify all are running
	assert.True(t, instance1.IsRunning(ctx))
	assert.True(t, instance2.IsRunning(ctx))
	assert.True(t, instance3.IsRunning(ctx))

	// Connect to all instances
	conn1, err := pgx.Connect(ctx, instance1.ConnString())
	require.NoError(t, err)
	defer conn1.Close(ctx)

	conn2, err := pgx.Connect(ctx, instance2.ConnString())
	require.NoError(t, err)
	defer conn2.Close(ctx)

	conn3, err := pgx.Connect(ctx, instance3.ConnString())
	require.NoError(t, err)
	defer conn3.Close(ctx)

	// Verify each connection works independently
	var result int
	err = conn1.QueryRow(ctx, "SELECT 1").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)

	err = conn2.QueryRow(ctx, "SELECT 2").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 2, result)

	err = conn3.QueryRow(ctx, "SELECT 3").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 3, result)

	t.Log("Successfully ran 3 pg0 instances simultaneously without port clashes")
}

func TestInstance_CustomConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	instance, err := NewInstance(InstanceConfig{
		Name:     "test-custom-config",
		Username: "testuser",
		Password: "testpass",
		Database: "testdb",
		ConfigOverrides: []string{
			"max_connections=200",
			"shared_buffers=128MB",
		},
		Logger: logger,
	})
	require.NoError(t, err)

	// Ensure cleanup
	t.Cleanup(func() {
		dropCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = instance.Drop(dropCtx)
	})

	// Start the instance
	err = instance.Start(ctx)
	require.NoError(t, err)

	// Verify custom credentials
	assert.Equal(t, "testuser", instance.Username)
	assert.Equal(t, "testpass", instance.Password)
	assert.Equal(t, "testdb", instance.Database)

	// Verify config overrides are present
	assert.Contains(t, instance.ConfigOverrides, "wal_level=logical")
	assert.Contains(t, instance.ConfigOverrides, "max_connections=200")
	assert.Contains(t, instance.ConfigOverrides, "shared_buffers=128MB")

	// Connect and verify
	conn, err := pgx.Connect(ctx, instance.ConnString())
	require.NoError(t, err)
	defer conn.Close(ctx)

	var maxConns string
	err = conn.QueryRow(ctx, "SHOW max_connections").Scan(&maxConns)
	require.NoError(t, err)
	assert.Equal(t, "200", maxConns)
}
