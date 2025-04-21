package template

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"app/internal/shared/temporal"
)

// WorkflowInput represents the input parameters for the workflow
type WorkflowInput struct {
	// Add your input fields here
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

// WorkflowResult represents the result of the workflow
type WorkflowResult struct {
	// Add your result fields here
	ID        string    `json:"id"`
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ActivityInput represents the input for activities
type ActivityInput struct {
	// Add your activity input fields here
	ID string `json:"id"`
}

// ActivityResult represents the result from activities
type ActivityResult struct {
	// Add your activity result fields here
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// TemplateWorkflow is a workflow that demonstrates [purpose]
func TemplateWorkflow(ctx workflow.Context, input WorkflowInput) (*WorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("TemplateWorkflow started", "ID", input.ID)

	// Set activity options
	ctx = temporal.WithActivityOptions(ctx, temporal.DefaultActivityOptions())

	// Prepare result
	result := &WorkflowResult{
		ID:        input.ID,
		Timestamp: workflow.Now(ctx),
	}

	// Execute activities
	var activityResult ActivityResult
	err := workflow.ExecuteActivity(ctx, "TemplateActivity", ActivityInput{
		ID: input.ID,
	}).Get(ctx, &activityResult)

	if err != nil {
		result.Success = false
		result.Message = err.Error()
		return result, err
	}

	result.Success = activityResult.Success
	if !activityResult.Success {
		result.Message = activityResult.Error
	}

	logger.Info("TemplateWorkflow completed", "ID", input.ID, "Success", result.Success)
	return result, nil
}
