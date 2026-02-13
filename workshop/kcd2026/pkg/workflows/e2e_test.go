package workflows

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

// TestE2E_MigrateWithPgrollWorkflow tests the pgroll workflow end-to-end
// with a real Temporal DevServer. Activities are still mocked (no real pgroll binary).
func TestE2E_MigrateWithPgrollWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Start Temporal DevServer
	server, err := testsuite.StartDevServer(ctx, testsuite.DevServerOptions{})
	require.NoError(t, err, "Failed to start Temporal DevServer")
	t.Cleanup(func() { _ = server.Stop() })

	c := server.Client()
	t.Cleanup(func() { c.Close() })

	taskQueue := fmt.Sprintf("e2e-test-%s", uuid.New().String())

	// Start worker
	w := worker.New(c, taskQueue, worker.Options{})
	w.RegisterWorkflow(MigrateWithPgrollWorkflow)

	// Register activities with mock exec
	a := &Activities{
		ExecFn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			slog.Info("mock exec", "cmd", name, "args", args)
			return []byte("mock output"), nil
		},
	}
	w.RegisterActivity(a.PgrollInitActivity)
	w.RegisterActivity(a.PgrollStartActivity)
	w.RegisterActivity(a.PgrollVerifyActivity)
	w.RegisterActivity(a.PgrollCompleteActivity)

	err = w.Start()
	require.NoError(t, err)
	t.Cleanup(func() { w.Stop() })

	// Execute workflow
	workflowID := fmt.Sprintf("e2e-pgroll-%s", uuid.New().String())
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: taskQueue,
	}, MigrateWithPgrollWorkflow, SchemaChangeRequest{
		ConnString:    "postgresql://postgres:postgres@localhost:5432/test",
		MigrationFile: "migrations/001_create_users.json",
	})
	require.NoError(t, err)

	// Wait for result
	var result MigrationResult
	err = run.Get(ctx, &result)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.True(t, result.OldSchemaValid)
	require.True(t, result.NewSchemaValid)

	t.Logf("E2E pgroll workflow completed: %+v", result)
}

// TestE2E_EnsureSchemaWorkflow_Idempotency tests that running the schema
// workflow twice produces idempotent results
func TestE2E_EnsureSchemaWorkflow_Idempotency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ctx := context.Background()

	server, err := testsuite.StartDevServer(ctx, testsuite.DevServerOptions{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = server.Stop() })

	c := server.Client()
	t.Cleanup(func() { c.Close() })

	taskQueue := fmt.Sprintf("e2e-test-%s", uuid.New().String())

	callCount := 0
	a := &Activities{
		ExecFn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			callCount++
			slog.Info("mock exec", "cmd", name, "args", args, "call_count", callCount)
			// First call returns changes, subsequent calls return no changes
			if callCount == 1 {
				return []byte("CREATE TABLE users..."), nil
			}
			return []byte(""), nil
		},
	}

	w := worker.New(c, taskQueue, worker.Options{})
	w.RegisterWorkflow(EnsureSchemaWorkflow)
	w.RegisterActivity(a.PsqldefDryRunActivity)
	w.RegisterActivity(a.PsqldefApplyActivity)

	err = w.Start()
	require.NoError(t, err)
	t.Cleanup(func() { w.Stop() })

	input := DesiredSchemaInput{
		ConnString: "postgresql://postgres:postgres@localhost:5432/test",
		SchemaFile: "migrations/desired_schema.sql",
	}

	// First run: should have changes
	run1, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        fmt.Sprintf("e2e-schema-1-%s", uuid.New().String()),
		TaskQueue: taskQueue,
	}, EnsureSchemaWorkflow, input)
	require.NoError(t, err)

	var result1 SchemaResult
	err = run1.Get(ctx, &result1)
	require.NoError(t, err)
	require.True(t, result1.Success)
	require.True(t, result1.HasChanges)

	// Second run: should be idempotent (no changes)
	run2, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        fmt.Sprintf("e2e-schema-2-%s", uuid.New().String()),
		TaskQueue: taskQueue,
	}, EnsureSchemaWorkflow, input)
	require.NoError(t, err)

	var result2 SchemaResult
	err = run2.Get(ctx, &result2)
	require.NoError(t, err)
	require.True(t, result2.Success)
	// Second run detects no changes since dry-run returns empty
	require.False(t, result2.HasChanges)
	require.Contains(t, result2.Message, "idempotent")

	t.Logf("E2E schema idempotency verified: run1=%+v run2=%+v", result1, result2)
}

// TestE2E_MaintenanceWorkflow tests the signal-driven maintenance workflow
// end-to-end with a real Temporal DevServer
func TestE2E_MaintenanceWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ctx := context.Background()

	server, err := testsuite.StartDevServer(ctx, testsuite.DevServerOptions{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = server.Stop() })

	c := server.Client()
	t.Cleanup(func() { c.Close() })

	taskQueue := fmt.Sprintf("e2e-test-%s", uuid.New().String())

	w := worker.New(c, taskQueue, worker.Options{})
	w.RegisterWorkflow(MaintenanceModeWorkflow)

	err = w.Start()
	require.NoError(t, err)
	t.Cleanup(func() { w.Stop() })

	// Start maintenance workflow
	workflowID := fmt.Sprintf("e2e-maintenance-%s", uuid.New().String())
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: taskQueue,
	}, MaintenanceModeWorkflow, MaintenanceRequest{
		AppEndpoint: "http://localhost:8080",
	})
	require.NoError(t, err)

	// Send enter_maintenance signal
	time.Sleep(500 * time.Millisecond)
	err = c.SignalWorkflow(ctx, workflowID, run.GetRunID(), SignalEnterMaintenance, nil)
	require.NoError(t, err)

	// Wait a bit, then send resume signal
	time.Sleep(1 * time.Second)
	err = c.SignalWorkflow(ctx, workflowID, run.GetRunID(), SignalResume, nil)
	require.NoError(t, err)

	// Wait for result
	var result MaintenanceResult
	err = run.Get(ctx, &result)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.Greater(t, result.MaintenanceTime, time.Duration(0))

	t.Logf("E2E maintenance workflow completed: duration=%v", result.MaintenanceTime)
}
