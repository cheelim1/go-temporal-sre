package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DesiredSchemaInput represents a sqldef schema request
type DesiredSchemaInput struct {
	ConnString string
	SchemaFile string
}

// SchemaResult represents the result of a schema operation
type SchemaResult struct {
	Success    bool
	Message    string
	HasChanges bool
	Changes    string
}

// EnsureSchemaWorkflow ensures the database schema matches the desired state using sqldef
func EnsureSchemaWorkflow(ctx workflow.Context, input DesiredSchemaInput) (*SchemaResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("EnsureSchemaWorkflow started",
		"workflow_id", workflow.GetInfo(ctx).WorkflowExecution.ID,
		"schema_file", input.SchemaFile,
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

	result := &SchemaResult{}

	// Step 1: Dry run to see what changes would be made
	logger.Info("Step 1: Running psqldef dry-run")
	var dryRunOutput PsqldefDryRunOutput
	err := workflow.ExecuteActivity(ctx, "PsqldefDryRunActivity", PsqldefDryRunInput{
		ConnString: input.ConnString,
		SchemaFile: input.SchemaFile,
	}).Get(ctx, &dryRunOutput)
	if err != nil {
		logger.Error("Failed to run dry-run", "error", err)
		return nil, err
	}

	result.HasChanges = dryRunOutput.HasChanges
	result.Changes = dryRunOutput.Changes

	if !dryRunOutput.HasChanges {
		logger.Info("No schema changes needed - schema is already in desired state")
		result.Success = true
		result.Message = "No changes needed - schema already matches desired state (idempotent)"
		return result, nil
	}

	logger.Info("Schema changes detected", "changes", dryRunOutput.Changes)

	// Step 2: Apply the schema
	logger.Info("Step 2: Applying schema changes")
	var applyOutput PsqldefApplyOutput
	err = workflow.ExecuteActivity(ctx, "PsqldefApplyActivity", PsqldefApplyInput{
		ConnString: input.ConnString,
		SchemaFile: input.SchemaFile,
	}).Get(ctx, &applyOutput)
	if err != nil {
		logger.Error("Failed to apply schema", "error", err)
		return nil, err
	}

	result.Success = true
	result.Message = "Schema applied successfully"

	logger.Info("EnsureSchemaWorkflow completed successfully")
	return result, nil
}
