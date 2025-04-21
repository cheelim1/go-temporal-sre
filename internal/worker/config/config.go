package config

import (
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// WorkerConfig holds configuration for the Temporal worker
type WorkerConfig struct {
	// Temporal client configuration
	ClientOptions client.Options

	// Worker options
	WorkerOptions worker.Options

	// Task queue name
	TaskQueue string

	// Activity configuration
	ActivityOptions workflow.ActivityOptions

	// Workflow configuration
	WorkflowOptions client.StartWorkflowOptions
}

// DefaultConfig returns a default worker configuration
func DefaultConfig() *WorkerConfig {
	return &WorkerConfig{
		ClientOptions: client.Options{
			HostPort:  "localhost:7233",
			Namespace: "default",
		},
		WorkerOptions: worker.Options{
			MaxConcurrentActivityTaskPollers:   2,
			MaxConcurrentWorkflowTaskPollers:   2,
			MaxConcurrentActivityExecutionSize: 100,
		},
		TaskQueue: "default",
		ActivityOptions: workflow.ActivityOptions{
			StartToCloseTimeout: time.Minute * 5,
			HeartbeatTimeout:    time.Second * 30,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    time.Second,
				BackoffCoefficient: 2.0,
				MaximumInterval:    time.Minute,
				MaximumAttempts:    3,
			},
		},
		WorkflowOptions: client.StartWorkflowOptions{
			WorkflowRunTimeout: time.Hour,
		},
	}
}

// WithTaskQueue sets the task queue name
func (c *WorkerConfig) WithTaskQueue(taskQueue string) *WorkerConfig {
	c.TaskQueue = taskQueue
	return c
}

// WithNamespace sets the Temporal namespace
func (c *WorkerConfig) WithNamespace(namespace string) *WorkerConfig {
	c.ClientOptions.Namespace = namespace
	return c
}

// WithHostPort sets the Temporal server address
func (c *WorkerConfig) WithHostPort(hostPort string) *WorkerConfig {
	c.ClientOptions.HostPort = hostPort
	return c
}

// WithActivityTimeout sets the activity timeout
func (c *WorkerConfig) WithActivityTimeout(timeout time.Duration) *WorkerConfig {
	c.ActivityOptions.StartToCloseTimeout = timeout
	return c
}

// WithWorkflowTimeout sets the workflow timeout
func (c *WorkerConfig) WithWorkflowTimeout(timeout time.Duration) *WorkerConfig {
	c.WorkflowOptions.WorkflowRunTimeout = timeout
	return c
}
