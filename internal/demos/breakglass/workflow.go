package breakglass

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"app/internal/shared/temporal"
)

// BreakglassWorkflowInput represents the input for the breakglass workflow
type BreakglassWorkflowInput struct {
	ServiceID   string            `json:"service_id"`
	Action      string            `json:"action"` // e.g., "restart", "scale", "rollback"
	Parameters  map[string]string `json:"parameters"`
	RequestedBy string            `json:"requested_by"`
}

// BreakglassWorkflowResult represents the result of the breakglass workflow
type BreakglassWorkflowResult struct {
	ServiceID   string    `json:"service_id"`
	Action      string    `json:"action"`
	Success     bool      `json:"success"`
	Message     string    `json:"message,omitempty"`
	CompletedAt time.Time `json:"completed_at"`
}

// BreakglassWorkflow is a workflow that handles emergency breakglass scenarios
func BreakglassWorkflow(ctx workflow.Context, input BreakglassWorkflowInput) (*BreakglassWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("BreakglassWorkflow started", "ServiceID", input.ServiceID, "Action", input.Action)

	// Set activity options
	ctx = temporal.WithActivityOptions(ctx, temporal.DefaultActivityOptions())

	// Prepare result
	result := &BreakglassWorkflowResult{
		ServiceID:   input.ServiceID,
		Action:      input.Action,
		CompletedAt: workflow.Now(ctx),
	}

	// Execute the appropriate action based on input
	var activityResult bool
	var err error

	switch input.Action {
	case "restart":
		err = workflow.ExecuteActivity(ctx, "RestartServiceActivity", input).Get(ctx, &activityResult)
	case "scale":
		err = workflow.ExecuteActivity(ctx, "ScaleServiceActivity", input).Get(ctx, &activityResult)
	case "rollback":
		err = workflow.ExecuteActivity(ctx, "RollbackServiceActivity", input).Get(ctx, &activityResult)
	default:
		result.Success = false
		result.Message = "Unsupported action"
		return result, nil
	}

	if err != nil {
		result.Success = false
		result.Message = err.Error()
		return result, err
	}

	result.Success = activityResult
	if activityResult {
		result.Message = "Breakglass action completed successfully"
	} else {
		result.Message = "Breakglass action failed"
	}

	logger.Info("BreakglassWorkflow completed", "ServiceID", input.ServiceID, "Success", result.Success)
	return result, nil
}
