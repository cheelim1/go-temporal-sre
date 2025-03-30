package superscript

import (
	"fmt"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// SinglePaymentWorkflowParams contains the parameters for the SinglePaymentWorkflow
type SinglePaymentWorkflowParams struct {
	OrderID string
}

// OrchestratorWorkflowParams contains the parameters for the OrchestratorWorkflow
type OrchestratorWorkflowParams struct {
	OrderIDs []string
	RunDate  time.Time
}

// SinglePaymentCollectionWorkflow executes the payment collection script for a single OrderID
// This workflow wraps a non-idempotent script in a way that makes it idempotent through Temporal
func SinglePaymentCollectionWorkflow(ctx workflow.Context, params SinglePaymentWorkflowParams) (*PaymentResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting SinglePaymentCollectionWorkflow", "orderID", params.OrderID)

	// Define activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	var result PaymentResult
	// Execute the activity to run the script
	err := workflow.ExecuteActivity(ctx, "RunPaymentCollectionScript", params.OrderID).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity execution failed", "error", err)
		return nil, err
	}

	logger.Info("Payment collection completed", "success", result.Success, "executionTime", result.ExecutionTime)

	return &result, nil
}

// OrchestratorWorkflow orchestrates multiple SinglePaymentCollectionWorkflows
// It replicates the behavior of the traditional_payment_collection.sh script
func OrchestratorWorkflow(ctx workflow.Context, params OrchestratorWorkflowParams) (*BatchResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting OrchestratorWorkflow", "orderCount", len(params.OrderIDs), "runDate", params.RunDate)

	// Initialize the batch result
	batchResult := &BatchResult{
		OrderIDs:     params.OrderIDs,
		Results:      make([]PaymentResult, 0, len(params.OrderIDs)),
		TotalCount:   len(params.OrderIDs),
		SuccessCount: 0,
		FailCount:    0,
		StartTime:    workflow.Now(ctx),
	}

	// Create child workflow options
	cwo := workflow.ChildWorkflowOptions{
		// Allow duplicate executions in the parent workflow
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
		TaskQueue:             SuperscriptTaskQueue,
	}

	ctx = workflow.WithChildOptions(ctx, cwo)

	// Process each OrderID by starting a child workflow
	for i, orderID := range params.OrderIDs {
		logger.Info(fmt.Sprintf("Processing OrderID: %s (%d/%d)", orderID, i+1, len(params.OrderIDs)))

		// Generate a deterministic workflow ID from the orderID
		workflowID := fmt.Sprintf("%s-%s", SinglePaymentWorkflowType, orderID)

		var result PaymentResult
		// Configure the child workflow options with the specific workflow ID
		child_ctx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID: workflowID,
			// Reject duplicate ensures only one execution of the same workflow ID can be running
			WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		})
		
		// Execute the child workflow to process this order
		errFuture := workflow.ExecuteChildWorkflow(
			child_ctx, 
			"SinglePaymentCollectionWorkflow", 
			SinglePaymentWorkflowParams{OrderID: orderID},
		)
		err := errFuture.Get(ctx, &result)

		if err != nil {
			logger.Error("Child workflow execution failed", "orderID", orderID, "error", err)
			batchResult.FailCount++
			continue
		}

		// Add the result to our batch
		batchResult.Results = append(batchResult.Results, result)
		if result.Success {
			batchResult.SuccessCount++
		} else {
			batchResult.FailCount++
		}

		logger.Info(fmt.Sprintf("Completed OrderID: %s - Success: %t", orderID, result.Success))
	}

	// Set the end time
	batchResult.EndTime = workflow.Now(ctx)

	// Log summary
	logger.Info("Orchestrator workflow completed", 
		"totalCount", batchResult.TotalCount,
		"successCount", batchResult.SuccessCount,
		"failCount", batchResult.FailCount,
		"successRate", batchResult.GetSuccessRate(),
	)

	return batchResult, nil
}
