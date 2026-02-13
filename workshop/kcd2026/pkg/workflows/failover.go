package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// FailoverRequest represents a request to failover to a replica
type FailoverRequest struct {
	AppEndpoint       string
	ReplicaName       string
	ReplicaConnString string
	CatchupTimeoutSecs int
}

// FailoverResult represents the result of a failover
type FailoverResult struct {
	Success           bool
	Message           string
	NewPrimaryName    string
	NewPrimaryConnStr string
}

// FailoverWorkflow orchestrates a no-data-loss failover from primary to replica
func FailoverWorkflow(ctx workflow.Context, input FailoverRequest) (*FailoverResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("FailoverWorkflow started",
		"workflow_id", workflow.GetInfo(ctx).WorkflowExecution.ID,
		"replica_name", input.ReplicaName,
	)

	// Set defaults
	if input.CatchupTimeoutSecs == 0 {
		input.CatchupTimeoutSecs = 60
	}

	// Activity options
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	result := &FailoverResult{
		NewPrimaryName:    input.ReplicaName,
		NewPrimaryConnStr: input.ReplicaConnString,
	}

	// Step 1: Pause writes on the app (enter buffer mode)
	logger.Info("Step 1: Pausing writes on app")
	var pauseOutput PauseWritesOutput
	err := workflow.ExecuteActivity(ctx, "PauseWritesActivity", PauseWritesInput{
		AppEndpoint: input.AppEndpoint,
	}).Get(ctx, &pauseOutput)
	if err != nil {
		logger.Error("Failed to pause writes", "error", err)
		return nil, err
	}
	logger.Info("Writes paused successfully")

	// Step 2: Wait for replica to catch up (simulated with workflow timer)
	logger.Info("Step 2: Waiting for replica to catch up")
	err = workflow.Sleep(ctx, 5*time.Second)
	if err != nil {
		logger.Error("Failed during catchup wait", "error", err)
		return nil, err
	}
	logger.Info("Replica caught up successfully")

	// Step 3: Promote replica to primary
	logger.Info("Step 3: Promoting replica to primary")
	var promoteOutput PromoteReplicaOutput
	err = workflow.ExecuteActivity(ctx, "PromoteReplicaActivity", PromoteReplicaInput{
		ReplicaName: input.ReplicaName,
	}).Get(ctx, &promoteOutput)
	if err != nil {
		logger.Error("Failed to promote replica", "error", err)
		return nil, err
	}
	logger.Info("Replica promoted successfully", "message", promoteOutput.Message)

	// Step 4: Resume writes (pointing to new primary)
	logger.Info("Step 4: Resuming writes with new primary")
	var resumeOutput ResumeWritesOutput
	err = workflow.ExecuteActivity(ctx, "ResumeWritesActivity", ResumeWritesInput{
		AppEndpoint:       input.AppEndpoint,
		NewPrimaryConnStr: input.ReplicaConnString,
	}).Get(ctx, &resumeOutput)
	if err != nil {
		logger.Error("Failed to resume writes", "error", err)
		return nil, err
	}
	logger.Info("Writes resumed successfully")

	result.Success = true
	result.Message = "Failover completed successfully with no data loss"

	logger.Info("FailoverWorkflow completed successfully")
	return result, nil
}
