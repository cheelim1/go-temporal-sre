package temporal

import (
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DefaultActivityOptions returns default activity options
func DefaultActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		HeartbeatTimeout:    time.Second * 30,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
}

// DefaultWorkflowOptions returns default workflow options
func DefaultWorkflowOptions() client.StartWorkflowOptions {
	return client.StartWorkflowOptions{
		WorkflowRunTimeout: time.Hour,
		TaskQueue:          "default",
	}
}

// WithActivityOptions sets activity options in the workflow context
func WithActivityOptions(ctx workflow.Context, opts workflow.ActivityOptions) workflow.Context {
	return workflow.WithActivityOptions(ctx, opts)
}

// WithWorkflowOptions sets workflow options
func WithWorkflowOptions(opts client.StartWorkflowOptions, taskQueue string) client.StartWorkflowOptions {
	opts.TaskQueue = taskQueue
	return opts
}
