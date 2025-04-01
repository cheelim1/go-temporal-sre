package superscript

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// --- Parameter Structs ---

// SinglePaymentWorkflowParams contains the parameters for the SinglePaymentWorkflow
type SinglePaymentWorkflowParams struct {
	OrderID string
}

// OrchestratorWorkflowParams contains the parameters for the OrchestratorWorkflow
type OrchestratorWorkflowParams struct {
	OrderIDs []string
	RunDate  time.Time
	// MaxConcurrent is the maximum number of child workflows to run concurrently.
	// Defaults to 3 if zero or negative.
	MaxConcurrent int
}

// --- Workflows ---

// SinglePaymentCollectionWorkflow executes the payment collection script for a single OrderID
// This workflow wraps a potentially non-idempotent activity call.
func SinglePaymentCollectionWorkflow(ctx workflow.Context, params SinglePaymentWorkflowParams) (*PaymentResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting SinglePaymentCollectionWorkflow", "orderID", params.OrderID)
	startTime := workflow.Now(ctx)

	// Define activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var activityResult PaymentResult // Activity should return this structure or similar
	err := workflow.ExecuteActivity(ctx, "RunPaymentCollectionScript", params.OrderID).Get(ctx, &activityResult)

	// Prepare the workflow result
	result := &PaymentResult{
		OrderID:       params.OrderID,
		Success:       err == nil && activityResult.Success, // Consider activity result's success flag
		ExecutionTime: workflow.Now(ctx).Sub(startTime),
	}

	if err != nil {
		logger.Error("Activity execution failed", "orderID", params.OrderID, "error", err)
		result.Error = err.Error()
		// It's generally better to return the result struct even on activity error,
		// indicating failure within the struct, rather than returning a workflow error,
		// unless the failure prevents the workflow from providing any meaningful result.
		// return result, nil // Indicate handled failure
		return result, fmt.Errorf("activity RunPaymentCollectionScript failed for order %s: %w", params.OrderID, err)
	} else if !activityResult.Success {
		logger.Warn("Activity completed but reported failure", "orderID", params.OrderID, "activityError", activityResult.Error)
		result.Error = activityResult.Error // Propagate error message from activity
	}

	logger.Info("SinglePaymentCollectionWorkflow completed", "orderID", params.OrderID, "success", result.Success)
	return result, nil // Workflow completed successfully, result indicates activity success/failure
}

// OrchestratorWorkflow orchestrates multiple SinglePaymentCollectionWorkflows concurrently
func OrchestratorWorkflow(ctx workflow.Context, params OrchestratorWorkflowParams) (*BatchResult, error) {
	logger := workflow.GetLogger(ctx)

	// Default concurrency
	concurrency := params.MaxConcurrent
	if concurrency <= 0 {
		concurrency = 3 // Default concurrency
	}
	logger.Info("Starting OrchestratorWorkflow", "orderCount", len(params.OrderIDs), "runDate", params.RunDate, "maxConcurrent", concurrency)

	// Initialize the batch result
	batchResult := &BatchResult{
		OrderIDs:     params.OrderIDs,
		Results:      make([]PaymentResult, len(params.OrderIDs)), // Pre-allocate results slice
		TotalCount:   len(params.OrderIDs),
		SuccessCount: 0,
		FailCount:    0,
		StartTime:    workflow.Now(ctx),
	}

	if len(params.OrderIDs) == 0 {
		logger.Info("No OrderIDs to process, completing workflow.")
		batchResult.EndTime = workflow.Now(ctx)
		return batchResult, nil
	}

	selector := workflow.NewSelector(ctx)
	sem := workflow.NewSemaphore(ctx, int64(concurrency))
	numScheduled := 0
	numCompleted := 0
	futuresMap := make(map[workflow.Future]int) // Map future to original index

	logger.Info("Starting concurrent child workflow execution", "concurrency", concurrency)

	for numCompleted < len(params.OrderIDs) {
		// Schedule new workflows if concurrency limit allows
		// Reverting to standard TryAcquire(1) based on conflicting linter feedback
		if numScheduled < len(params.OrderIDs) && sem.TryAcquire(ctx, 1) {
			orderID := params.OrderIDs[numScheduled]
			idx := numScheduled // Capture index for the callback
			numScheduled++      // Increment scheduled count *before* async execution

			logger.Info("Scheduling child workflow", "index", idx, "orderID", orderID)

			workflowID := fmt.Sprintf("%s-%s", SinglePaymentWorkflowType, orderID)
			childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
				WorkflowID:            workflowID,
				WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
				TaskQueue:             SuperscriptTaskQueue,
			})

			exFuture := workflow.ExecuteChildWorkflow(childCtx, SinglePaymentWorkflowType, SinglePaymentWorkflowParams{OrderID: orderID})
			futuresMap[exFuture] = idx // Store mapping

			selector.AddFuture(exFuture, func(f workflow.Future) {
				completedIdx := futuresMap[f]
				completedOrderID := params.OrderIDs[completedIdx]
				completedWorkflowID := fmt.Sprintf("%s-%s", SinglePaymentWorkflowType, completedOrderID)
				var result PaymentResult

				err := f.Get(ctx, &result)
				if err != nil {
					logger.Warn("Child workflow future returned error", "index", completedIdx, "orderID", completedOrderID, "workflowID", completedWorkflowID, "errorType", fmt.Sprintf("%T", err), "error", err)

					var childWorkflowExecutionError *temporal.ChildWorkflowExecutionError
					if errors.As(err, &childWorkflowExecutionError) {
						// Check if this is a ChildWorkflowExecutionAlreadyStartedError
						var childWorkflowExecutionAlreadyStartedError *temporal.ChildWorkflowExecutionAlreadyStartedError
						if errors.As(childWorkflowExecutionError.Unwrap(), &childWorkflowExecutionAlreadyStartedError) {
							// This error occurs when we try to start a workflow with an ID that's already running
							logger.Info("Child workflow already started, recording this as a successful duplicate", "index", completedIdx, "orderID", completedOrderID)

							// In a duplicate workflow scenario with WorkflowIDReusePolicy.REJECT_DUPLICATE,
							// we'll treat this as a successful case since our idempotency mechanism worked
							// The original workflow will complete and handle the operation
							batchResult.Results[completedIdx] = PaymentResult{
								OrderID: completedOrderID,
								Success: true,
								Output: "WorkflowID: " + workflow.GetInfo(ctx).WorkflowExecution.ID +
									" RunID: " + workflow.GetInfo(ctx).WorkflowExecution.RunID +
									fmt.Sprintf(" Attempt: %d", workflow.GetInfo(ctx).Attempt),
								Error:         "Workflow already running - duplicate request handled correctly",
								ExecutionTime: workflow.Now(ctx).Sub(batchResult.StartTime),
							}
							batchResult.SuccessCount++
							// Skip further error handling as we're treating this as a success case
						}
					} else {
						// Non-child workflow execution error (e.g., parent cancelled, workflow task failure)
						logger.Error("Child workflow future Get failed (non-ChildWorkflowExecutionError)", "index", completedIdx, "error", err)
						batchResult.Results[completedIdx] = PaymentResult{OrderID: completedOrderID, Success: false, Error: err.Error()}
						batchResult.FailCount++
					}
				} else {
					// Success from f.Get()
					logger.Info("Child workflow completed successfully", "index", completedIdx, "success", result.Success)
					batchResult.Results[completedIdx] = result
					if result.Success {
						batchResult.SuccessCount++
					} else {
						batchResult.FailCount++
					}
				}

				// Reverting to standard Release(1) based on conflicting linter feedback
				sem.Release(1) // Release semaphore
				numCompleted++ // Increment completed count *after* processing
			})
		} else {
			// Wait for a workflow to complete if we can't schedule more
			if numScheduled > numCompleted {
				selector.Select(ctx)
			} else if numScheduled == len(params.OrderIDs) {
				// All scheduled, but not all completed yet. Keep selecting.
				selector.Select(ctx)
			} else {
				// Should not happen unless OrderIDs is empty (handled above)
				// or semaphore starts at 0. Break loop as a safeguard.
				logger.Error("Unexpected state in concurrency loop", "numScheduled", numScheduled, "numCompleted", numCompleted, "totalOrders", len(params.OrderIDs))
				break
			}
		}
	}

	batchResult.EndTime = workflow.Now(ctx)
	logger.Info("Orchestrator workflow completed",
		"totalCount", batchResult.TotalCount,
		"successCount", batchResult.SuccessCount,
		"failCount", batchResult.FailCount,
		"successRate", fmt.Sprintf("%d%%", batchResult.GetSuccessRate()),
		"duration", batchResult.EndTime.Sub(batchResult.StartTime),
	)
	return batchResult, nil
}
