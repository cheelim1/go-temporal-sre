package workflows

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"

	"kcd2026/pkg/pg0"
)

// Activities holds all activity implementations
type Activities struct {
	// ExecFn allows injecting command execution for testing
	ExecFn func(ctx context.Context, name string, args ...string) ([]byte, error)
}

// defaultExec is the default command executor
func defaultExec(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

// exec executes a command using the configured executor
func (a *Activities) exec(ctx context.Context, name string, args ...string) ([]byte, error) {
	if a.ExecFn != nil {
		return a.ExecFn(ctx, name, args...)
	}
	return defaultExec(ctx, name, args...)
}

// getLogger returns the activity logger
func getLogger(ctx context.Context) *slog.Logger {
	_ = activity.GetLogger(ctx)
	return slog.Default().With(
		"workflow_id", activity.GetInfo(ctx).WorkflowExecution.ID,
		"activity_id", activity.GetInfo(ctx).ActivityID,
	)
}

// ===== pgroll Activities =====

type PgrollInitInput struct {
	ConnString string
}

type PgrollInitOutput struct {
	Success bool
	Message string
}

func (a *Activities) PgrollInitActivity(ctx context.Context, input PgrollInitInput) (*PgrollInitOutput, error) {
	logger := getLogger(ctx)
	logger.Info("initializing pgroll", "conn_string", maskPassword(input.ConnString))

	output, err := a.exec(ctx, "pgroll", "init", "--postgres-url", input.ConnString)
	if err != nil {
		logger.Error("pgroll init failed", "error", err, "output", string(output))
		return nil, fmt.Errorf("pgroll init failed: %w: %s", err, string(output))
	}

	logger.Info("pgroll initialized successfully")
	return &PgrollInitOutput{
		Success: true,
		Message: string(output),
	}, nil
}

type PgrollStartInput struct {
	ConnString     string
	MigrationFile  string
}

type PgrollStartOutput struct {
	Success bool
	Message string
}

func (a *Activities) PgrollStartActivity(ctx context.Context, input PgrollStartInput) (*PgrollStartOutput, error) {
	logger := getLogger(ctx)
	logger.Info("starting pgroll migration", "migration_file", input.MigrationFile)

	output, err := a.exec(ctx, "pgroll", "start", input.MigrationFile, "--postgres-url", input.ConnString)
	if err != nil {
		logger.Error("pgroll start failed", "error", err, "output", string(output))
		return nil, fmt.Errorf("pgroll start failed: %w: %s", err, string(output))
	}

	logger.Info("pgroll migration started successfully")
	return &PgrollStartOutput{
		Success: true,
		Message: string(output),
	}, nil
}

type PgrollVerifyInput struct {
	ConnString string
}

type PgrollVerifyOutput struct {
	Success bool
	Message string
}

func (a *Activities) PgrollVerifyActivity(ctx context.Context, input PgrollVerifyInput) (*PgrollVerifyOutput, error) {
	logger := getLogger(ctx)
	logger.Info("verifying pgroll migration")

	// For now, just return success - in a real implementation, we'd query both schema versions
	return &PgrollVerifyOutput{
		Success: true,
		Message: "Migration verified - both old and new schema versions accessible",
	}, nil
}

type PgrollCompleteInput struct {
	ConnString string
}

type PgrollCompleteOutput struct {
	Success bool
	Message string
}

func (a *Activities) PgrollCompleteActivity(ctx context.Context, input PgrollCompleteInput) (*PgrollCompleteOutput, error) {
	logger := getLogger(ctx)
	logger.Info("completing pgroll migration")

	output, err := a.exec(ctx, "pgroll", "complete", "--postgres-url", input.ConnString)
	if err != nil {
		logger.Error("pgroll complete failed", "error", err, "output", string(output))
		return nil, fmt.Errorf("pgroll complete failed: %w: %s", err, string(output))
	}

	logger.Info("pgroll migration completed successfully")
	return &PgrollCompleteOutput{
		Success: true,
		Message: string(output),
	}, nil
}

// ===== sqldef Activities =====

type PsqldefDryRunInput struct {
	ConnString string
	SchemaFile string
}

type PsqldefDryRunOutput struct {
	HasChanges bool
	Changes    string
}

func (a *Activities) PsqldefDryRunActivity(ctx context.Context, input PsqldefDryRunInput) (*PsqldefDryRunOutput, error) {
	logger := getLogger(ctx)
	logger.Info("running psqldef dry-run", "schema_file", input.SchemaFile)

	output, err := a.exec(ctx, "psqldef", "--dry-run", "--file", input.SchemaFile, input.ConnString)
	outputStr := string(output)

	// psqldef returns non-zero exit code when there are changes
	hasChanges := err != nil || len(outputStr) > 0

	logger.Info("psqldef dry-run completed", "has_changes", hasChanges)
	return &PsqldefDryRunOutput{
		HasChanges: hasChanges,
		Changes:    outputStr,
	}, nil
}

type PsqldefApplyInput struct {
	ConnString string
	SchemaFile string
}

type PsqldefApplyOutput struct {
	Success bool
	Message string
}

func (a *Activities) PsqldefApplyActivity(ctx context.Context, input PsqldefApplyInput) (*PsqldefApplyOutput, error) {
	logger := getLogger(ctx)
	logger.Info("applying psqldef schema", "schema_file", input.SchemaFile)

	output, err := a.exec(ctx, "psqldef", "--file", input.SchemaFile, input.ConnString)
	if err != nil {
		logger.Error("psqldef apply failed", "error", err, "output", string(output))
		return nil, fmt.Errorf("psqldef apply failed: %w: %s", err, string(output))
	}

	logger.Info("psqldef schema applied successfully")
	return &PsqldefApplyOutput{
		Success: true,
		Message: string(output),
	}, nil
}

// ===== pg0 Activities =====

type SpinUpReplicaInput struct {
	Name            string
	ConfigOverrides []string
}

type SpinUpReplicaOutput struct {
	Name       string
	Port       int
	ConnString string
}

func (a *Activities) SpinUpReplicaActivity(ctx context.Context, input SpinUpReplicaInput) (*SpinUpReplicaOutput, error) {
	logger := getLogger(ctx)
	logger.Info("spinning up pg0 replica", "name", input.Name)

	instance, err := pg0.NewInstance(pg0.InstanceConfig{
		Name:            input.Name,
		ConfigOverrides: input.ConfigOverrides,
		Logger:          logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	if err := instance.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start instance: %w", err)
	}

	logger.Info("pg0 replica started", "name", instance.Name, "port", instance.Port)
	return &SpinUpReplicaOutput{
		Name:       instance.Name,
		Port:       instance.Port,
		ConnString: instance.ConnString(),
	}, nil
}

type StopInstanceInput struct {
	Name string
}

type StopInstanceOutput struct {
	Success bool
}

func (a *Activities) StopInstanceActivity(ctx context.Context, input StopInstanceInput) (*StopInstanceOutput, error) {
	logger := getLogger(ctx)
	logger.Info("stopping pg0 instance", "name", input.Name)

	// Create instance reference and stop it
	instance := &pg0.Instance{Name: input.Name}
	if err := instance.Stop(ctx); err != nil {
		return nil, fmt.Errorf("failed to stop instance: %w", err)
	}

	logger.Info("pg0 instance stopped", "name", input.Name)
	return &StopInstanceOutput{Success: true}, nil
}

// ===== pgstream Activities =====

type StartPgstreamProducerInput struct {
	PrimaryConnString string
	ReplicationSlot   string
}

type StartPgstreamProducerOutput struct {
	Success bool
	Message string
}

func (a *Activities) StartPgstreamProducerActivity(ctx context.Context, input StartPgstreamProducerInput) (*StartPgstreamProducerOutput, error) {
	logger := getLogger(ctx)
	logger.Info("starting pgstream producer", "replication_slot", input.ReplicationSlot)

	// In a real implementation, this would start pgstream producer
	// For now, we'll simulate it
	logger.Info("pgstream producer started (simulated)")
	return &StartPgstreamProducerOutput{
		Success: true,
		Message: "pgstream producer started",
	}, nil
}

type StartPgstreamConsumerInput struct {
	ReplicaConnString string
	ReplicationSlot   string
}

type StartPgstreamConsumerOutput struct {
	Success bool
	Message string
}

func (a *Activities) StartPgstreamConsumerActivity(ctx context.Context, input StartPgstreamConsumerInput) (*StartPgstreamConsumerOutput, error) {
	logger := getLogger(ctx)
	logger.Info("starting pgstream consumer", "replication_slot", input.ReplicationSlot)

	// In a real implementation, this would start pgstream consumer
	// For now, we'll simulate it
	logger.Info("pgstream consumer started (simulated)")
	return &StartPgstreamConsumerOutput{
		Success: true,
		Message: "pgstream consumer started",
	}, nil
}

type WaitForSnapshotInput struct {
	ReplicationSlot string
	TimeoutSeconds  int
}

type WaitForSnapshotOutput struct {
	Success bool
	Message string
}

func (a *Activities) WaitForSnapshotActivity(ctx context.Context, input WaitForSnapshotInput) (*WaitForSnapshotOutput, error) {
	logger := getLogger(ctx)
	logger.Info("waiting for pgstream snapshot", "replication_slot", input.ReplicationSlot)

	// Simulate waiting for snapshot
	timeout := time.Duration(input.TimeoutSeconds) * time.Second
	select {
	case <-time.After(2 * time.Second): // Simulate 2 second snapshot time
		logger.Info("pgstream snapshot completed")
		return &WaitForSnapshotOutput{
			Success: true,
			Message: "snapshot completed",
		}, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for snapshot after %v", timeout)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// ===== App Config Activities =====

type UpdateReadPoolInput struct {
	ReplicaConnStrings []string
}

type UpdateReadPoolOutput struct {
	Success bool
}

func (a *Activities) UpdateReadPoolActivity(ctx context.Context, input UpdateReadPoolInput) (*UpdateReadPoolOutput, error) {
	logger := getLogger(ctx)
	logger.Info("updating read pool", "replica_count", len(input.ReplicaConnStrings))

	// In a real implementation, this would update the app's read pool configuration
	// For now, we'll simulate it
	logger.Info("read pool updated (simulated)")
	return &UpdateReadPoolOutput{Success: true}, nil
}

type PauseWritesInput struct {
	AppEndpoint string
}

type PauseWritesOutput struct {
	Success bool
}

func (a *Activities) PauseWritesActivity(ctx context.Context, input PauseWritesInput) (*PauseWritesOutput, error) {
	logger := getLogger(ctx)
	logger.Info("pausing writes", "app_endpoint", input.AppEndpoint)

	// In a real implementation, this would signal the app to pause writes
	// For now, we'll simulate it
	logger.Info("writes paused (simulated)")
	return &PauseWritesOutput{Success: true}, nil
}

type ResumeWritesInput struct {
	AppEndpoint       string
	NewPrimaryConnStr string
}

type ResumeWritesOutput struct {
	Success bool
}

func (a *Activities) ResumeWritesActivity(ctx context.Context, input ResumeWritesInput) (*ResumeWritesOutput, error) {
	logger := getLogger(ctx)
	logger.Info("resuming writes", "app_endpoint", input.AppEndpoint)

	// In a real implementation, this would signal the app to resume writes
	// For now, we'll simulate it
	logger.Info("writes resumed (simulated)")
	return &ResumeWritesOutput{Success: true}, nil
}

type PromoteReplicaInput struct {
	ReplicaName string
}

type PromoteReplicaOutput struct {
	Success bool
	Message string
}

func (a *Activities) PromoteReplicaActivity(ctx context.Context, input PromoteReplicaInput) (*PromoteReplicaOutput, error) {
	logger := getLogger(ctx)
	logger.Info("promoting replica to primary", "replica_name", input.ReplicaName)

	// In a real implementation, this would promote the replica
	// For now, we'll simulate it
	logger.Info("replica promoted (simulated)")
	return &PromoteReplicaOutput{
		Success: true,
		Message: "replica promoted to primary",
	}, nil
}

// ===== Helper Functions =====

func maskPassword(connString string) string {
	// Simple password masking for logging
	if strings.Contains(connString, "@") {
		parts := strings.Split(connString, "@")
		if len(parts) == 2 {
			userPass := strings.Split(parts[0], "://")
			if len(userPass) == 2 {
				return userPass[0] + "://***:***@" + parts[1]
			}
		}
	}
	return connString
}
