package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ReplicaRequest represents a request to provision a replica
type ReplicaRequest struct {
	ReplicaName         string
	PrimaryConnString   string
	ConfigOverrides     []string
	ReplicationSlot     string
	SnapshotTimeoutSecs int
}

// ReplicaResult represents the result of provisioning a replica
type ReplicaResult struct {
	Success           bool
	Message           string
	ReplicaName       string
	ReplicaPort       int
	ReplicaConnString string
}

// ProvisionReplicaWorkflow provisions a read replica using pgstream
func ProvisionReplicaWorkflow(ctx workflow.Context, input ReplicaRequest) (*ReplicaResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ProvisionReplicaWorkflow started",
		"workflow_id", workflow.GetInfo(ctx).WorkflowExecution.ID,
		"replica_name", input.ReplicaName,
	)

	// Set defaults
	if input.ReplicationSlot == "" {
		input.ReplicationSlot = fmt.Sprintf("pgstream_%s", input.ReplicaName)
	}
	if input.SnapshotTimeoutSecs == 0 {
		input.SnapshotTimeoutSecs = 300 // 5 minutes default
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

	result := &ReplicaResult{
		ReplicaName: input.ReplicaName,
	}

	// Step 1: Spin up a new pg0 instance
	logger.Info("Step 1: Spinning up new pg0 instance")
	var spinUpOutput SpinUpReplicaOutput
	err := workflow.ExecuteActivity(ctx, "SpinUpReplicaActivity", SpinUpReplicaInput{
		Name:            input.ReplicaName,
		ConfigOverrides: input.ConfigOverrides,
	}).Get(ctx, &spinUpOutput)
	if err != nil {
		logger.Error("Failed to spin up replica", "error", err)
		return nil, err
	}
	logger.Info("replica instance created",
		"name", spinUpOutput.Name,
		"port", spinUpOutput.Port,
		"conn_string", spinUpOutput.ConnString,
	)

	result.ReplicaPort = spinUpOutput.Port
	result.ReplicaConnString = spinUpOutput.ConnString

	// Step 2: Start pgstream producer on primary
	logger.Info("Step 2: Starting pgstream producer on primary")
	var producerOutput StartPgstreamProducerOutput
	err = workflow.ExecuteActivity(ctx, "StartPgstreamProducerActivity", StartPgstreamProducerInput{
		PrimaryConnString: input.PrimaryConnString,
		ReplicationSlot:   input.ReplicationSlot,
	}).Get(ctx, &producerOutput)
	if err != nil {
		logger.Error("Failed to start pgstream producer", "error", err)
		return nil, err
	}
	logger.Info("pgstream producer started", "message", producerOutput.Message)

	// Step 3: Start pgstream consumer on replica
	logger.Info("Step 3: Starting pgstream consumer on replica")
	var consumerOutput StartPgstreamConsumerOutput
	err = workflow.ExecuteActivity(ctx, "StartPgstreamConsumerActivity", StartPgstreamConsumerInput{
		ReplicaConnString: spinUpOutput.ConnString,
		ReplicationSlot:   input.ReplicationSlot,
	}).Get(ctx, &consumerOutput)
	if err != nil {
		logger.Error("Failed to start pgstream consumer", "error", err)
		return nil, err
	}
	logger.Info("pgstream consumer started", "message", consumerOutput.Message)

	// Step 4: Wait for initial snapshot to complete
	logger.Info("Step 4: Waiting for initial snapshot")
	var snapshotOutput WaitForSnapshotOutput
	err = workflow.ExecuteActivity(ctx, "WaitForSnapshotActivity", WaitForSnapshotInput{
		ReplicationSlot: input.ReplicationSlot,
		TimeoutSeconds:  input.SnapshotTimeoutSecs,
	}).Get(ctx, &snapshotOutput)
	if err != nil {
		logger.Error("Failed to complete snapshot", "error", err)
		return nil, err
	}
	logger.Info("snapshot completed", "message", snapshotOutput.Message)

	result.Success = true
	result.Message = "Replica provisioned and synced successfully"

	logger.Info("ProvisionReplicaWorkflow completed successfully")
	return result, nil
}
