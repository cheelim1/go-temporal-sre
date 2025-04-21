package batch

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// FeeDeductionWorkflowInput represents the input parameters for the FeeDeductionWorkflow
type FeeDeductionWorkflowInput struct {
	AccountID string  `json:"account_id"`
	OrderID   string  `json:"order_id"`
	Amount    float64 `json:"amount"`
}

// FeeDeductionWorkflowResult represents the result of the FeeDeductionWorkflow
type FeeDeductionWorkflowResult struct {
	NewBalance float64   `json:"new_balance"`
	Success    bool      `json:"success"`
	Message    string    `json:"message,omitempty"`
	OrderID    string    `json:"order_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// ActivityInput represents the input for fee deduction activity
type ActivityInput struct {
	AccountID string  `json:"account_id"`
	OrderID   string  `json:"order_id"`
	Amount    float64 `json:"amount"`
}

// ActivityResult represents the result from fee deduction activity
type ActivityResult struct {
	NewBalance float64 `json:"new_balance"`
	Success    bool    `json:"success"`
	Error      string  `json:"error,omitempty"`
}

// FeeDeductionWorkflow is a workflow that deducts a fee from an account in an idempotent manner
func FeeDeductionWorkflow(ctx workflow.Context, input FeeDeductionWorkflowInput) (*FeeDeductionWorkflowResult, error) {
	// Logger
	logger := workflow.GetLogger(ctx)
	logger.Info("FeeDeductionWorkflow started", "OrderID", input.OrderID, " WFID: ", workflow.GetInfo(ctx).WorkflowExecution.ID)

	// Activity options - 1 time only as script is not idempotent
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 3 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    50, // Will never reach here as StartToClose will trogger ..
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	// Prepare result
	result := &FeeDeductionWorkflowResult{
		OrderID:   input.OrderID,
		Timestamp: workflow.Now(ctx),
	}

	// Prepare activity input
	//activityInput := ActivityInput{
	//	AccountID: input.AccountID,
	//	OrderID:   input.OrderID,
	//	Amount:    input.Amount,
	//}

	// Execute activity
	var activityResult ActivityResult

	// In Temporal, the workflow ID serves as the deduplication key
	// When the workflow is executed with the same workflow ID (which is set to OrderID in tests),
	// Temporal will automatically deduplicate the workflow execution.
	// The underlying activity is intentionally non-idempotent to show how Temporal handles this.
	logger.Info("Executing DeductFee activity", "OrderID", input.OrderID)

	// Debug using Basic
	//workflow.ExecuteActivity(ctx, "BasicActivity")

	// Execute the fee deduction activity; block until done ..
	err := workflow.ExecuteActivity(ctx, "DeductFeeActivity", input).Get(ctx, &activityResult)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Activity execution failed: %v", err)
		return result, err
	}

	// Set the result based on activity result
	result.NewBalance = activityResult.NewBalance
	result.Success = activityResult.Success
	if !activityResult.Success {
		result.Message = activityResult.Error
	} else {
		result.Message = fmt.Sprintf("Fee deduction successful. NewBalance: %f", result.NewBalance)
	}

	logger.Info("FeeDeductionWorkflow completed", "OrderID", input.OrderID, "Success", result.Success)
	return result, nil
}
