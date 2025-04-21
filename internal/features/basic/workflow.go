package basic

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// GreetingWorkflow is a simple workflow that demonstrates the basic usage of Temporal
func GreetingWorkflow(ctx workflow.Context, name string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("GreetingWorkflow started", "name", name)

	// Set activity options
	options := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})

	// Execute activity
	var result string
	err := workflow.ExecuteActivity(options, "GreetingActivity", name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed", "error", err)
		return "", err
	}

	logger.Info("GreetingWorkflow completed", "result", result)
	return result, nil
}
