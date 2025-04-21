package worker

import (
	"context"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"app/internal/worker/config"
)

// Worker represents a Temporal worker instance
type Worker struct {
	client client.Client
	worker worker.Worker
	config *config.WorkerConfig
}

// New creates a new Temporal worker with the given configuration
func New(cfg *config.WorkerConfig) (*Worker, error) {
	// Create Temporal client
	c, err := client.Dial(client.Options(cfg.ClientOptions))
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	// Create worker
	w := worker.New(c, cfg.TaskQueue, cfg.WorkerOptions)

	return &Worker{
		client: c,
		worker: w,
		config: cfg,
	}, nil
}

// RegisterWorkflow registers a workflow with the worker
func (w *Worker) RegisterWorkflow(wf interface{}) {
	w.worker.RegisterWorkflow(wf)
}

// RegisterActivity registers an activity with the worker
func (w *Worker) RegisterActivity(act interface{}) {
	w.worker.RegisterActivity(act)
}

// Start starts the worker
func (w *Worker) Start() error {
	log.Printf("Starting worker for task queue: %s", w.config.TaskQueue)
	return w.worker.Run(worker.InterruptCh())
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.client.Close()
}

// ExecuteWorkflow executes a workflow with the given options
func (w *Worker) ExecuteWorkflow(ctx context.Context, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	opts := w.config.WorkflowOptions
	return w.client.ExecuteWorkflow(ctx, opts, workflow, args...)
}

// GetConfig returns the worker configuration
func (w *Worker) GetConfig() *config.WorkerConfig {
	return w.config
}
