package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ScaleRequest represents a request to scale read replicas
type ScaleRequest struct {
	PrimaryConnString   string
	NewReplicaCount     int
	ReplicaNamePrefix   string
	ConfigOverrides     []string
	SnapshotTimeoutSecs int
}

// ScaleResult represents the result of scaling replicas
type ScaleResult struct {
	Success             bool
	Message             string
	NewReplicaConnStrings []string
}

// ScaleReadReplicasWorkflow scales read replicas and updates the app's read pool
func ScaleReadReplicasWorkflow(ctx workflow.Context, input ScaleRequest) (*ScaleResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ScaleReadReplicasWorkflow started",
		"workflow_id", workflow.GetInfo(ctx).WorkflowExecution.ID,
		"new_replica_count", input.NewReplicaCount,
	)

	// Set defaults
	if input.ReplicaNamePrefix == "" {
		input.ReplicaNamePrefix = "replica"
	}
	if input.SnapshotTimeoutSecs == 0 {
		input.SnapshotTimeoutSecs = 300
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

	result := &ScaleResult{
		NewReplicaConnStrings: make([]string, 0, input.NewReplicaCount),
	}

	// Provision each replica
	for i := 0; i < input.NewReplicaCount; i++ {
		replicaName := fmt.Sprintf("%s-%d", input.ReplicaNamePrefix, i+1)
		logger.Info("Provisioning replica", "replica_name", replicaName, "index", i+1, "total", input.NewReplicaCount)

		// Use ProvisionReplicaWorkflow as a child workflow
		childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID: fmt.Sprintf("provision-replica-%s", replicaName),
		})

		var replicaResult ReplicaResult
		err := workflow.ExecuteChildWorkflow(childCtx, ProvisionReplicaWorkflow, ReplicaRequest{
			ReplicaName:         replicaName,
			PrimaryConnString:   input.PrimaryConnString,
			ConfigOverrides:     input.ConfigOverrides,
			ReplicationSlot:     fmt.Sprintf("pgstream_%s", replicaName),
			SnapshotTimeoutSecs: input.SnapshotTimeoutSecs,
		}).Get(ctx, &replicaResult)

		if err != nil {
			logger.Error("Failed to provision replica", "replica_name", replicaName, "error", err)
			return nil, fmt.Errorf("failed to provision replica %s: %w", replicaName, err)
		}

		logger.Info("Replica provisioned successfully",
			"replica_name", replicaName,
			"conn_string", replicaResult.ReplicaConnString,
		)

		result.NewReplicaConnStrings = append(result.NewReplicaConnStrings, replicaResult.ReplicaConnString)
	}

	// Update the app's read pool with all new replicas
	logger.Info("Updating app read pool", "replica_count", len(result.NewReplicaConnStrings))
	var updateOutput UpdateReadPoolOutput
	err := workflow.ExecuteActivity(ctx, "UpdateReadPoolActivity", UpdateReadPoolInput{
		ReplicaConnStrings: result.NewReplicaConnStrings,
	}).Get(ctx, &updateOutput)
	if err != nil {
		logger.Error("Failed to update read pool", "error", err)
		return nil, err
	}

	result.Success = true
	result.Message = fmt.Sprintf("Successfully scaled to %d replicas and updated read pool", input.NewReplicaCount)

	logger.Info("ScaleReadReplicasWorkflow completed successfully")
	return result, nil
}
