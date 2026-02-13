package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// SchemaChangeRequest represents a pgroll migration request
type SchemaChangeRequest struct {
	ConnString    string
	MigrationFile string
}

// MigrationResult represents the result of a migration
type MigrationResult struct {
	Success        bool
	Message        string
	OldSchemaValid bool
	NewSchemaValid bool
}

// MigrateWithPgrollWorkflow orchestrates a zero-downtime schema migration using pgroll
func MigrateWithPgrollWorkflow(ctx workflow.Context, input SchemaChangeRequest) (*MigrationResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("MigrateWithPgrollWorkflow started",
		"workflow_id", workflow.GetInfo(ctx).WorkflowExecution.ID,
		"migration_file", input.MigrationFile,
	)

	// Activity options
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	result := &MigrationResult{}

	// Step 1: Initialize pgroll
	logger.Info("Step 1: Initializing pgroll")
	var initOutput PgrollInitOutput
	err := workflow.ExecuteActivity(ctx, "PgrollInitActivity", PgrollInitInput{
		ConnString: input.ConnString,
	}).Get(ctx, &initOutput)
	if err != nil {
		logger.Error("Failed to initialize pgroll", "error", err)
		return nil, err
	}
	logger.Info("pgroll initialized", "message", initOutput.Message)

	// Step 2: Start migration
	logger.Info("Step 2: Starting migration")
	var startOutput PgrollStartOutput
	err = workflow.ExecuteActivity(ctx, "PgrollStartActivity", PgrollStartInput{
		ConnString:    input.ConnString,
		MigrationFile: input.MigrationFile,
	}).Get(ctx, &startOutput)
	if err != nil {
		logger.Error("Failed to start migration", "error", err)
		return nil, err
	}
	logger.Info("migration started", "message", startOutput.Message)

	// Step 3: Verify dual schema access
	logger.Info("Step 3: Verifying dual schema access")
	var verifyOutput PgrollVerifyOutput
	err = workflow.ExecuteActivity(ctx, "PgrollVerifyActivity", PgrollVerifyInput{
		ConnString: input.ConnString,
	}).Get(ctx, &verifyOutput)
	if err != nil {
		logger.Error("Failed to verify migration", "error", err)
		return nil, err
	}
	logger.Info("migration verified", "message", verifyOutput.Message)
	result.OldSchemaValid = true
	result.NewSchemaValid = true

	// Step 4: Complete migration
	logger.Info("Step 4: Completing migration")
	var completeOutput PgrollCompleteOutput
	err = workflow.ExecuteActivity(ctx, "PgrollCompleteActivity", PgrollCompleteInput{
		ConnString: input.ConnString,
	}).Get(ctx, &completeOutput)
	if err != nil {
		logger.Error("Failed to complete migration", "error", err)
		return nil, err
	}
	logger.Info("migration completed", "message", completeOutput.Message)

	result.Success = true
	result.Message = "Zero-downtime migration completed successfully"

	logger.Info("MigrateWithPgrollWorkflow completed successfully")
	return result, nil
}
