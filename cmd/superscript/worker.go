package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"app/internal/superscript"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
)

// Worker represents a Temporal worker process
type Worker struct {
	Client     client.Client
	Worker     worker.Worker
	Logger     log.Logger
	Activities *superscript.Activities
	isRunning  bool
}

// NewWorker creates a new Temporal worker
func NewWorker(c client.Client, logger log.Logger) *Worker {
	// Create activities with the proper script base path
	activities := superscript.NewActivities("./internal/superscript/", logger)

	// Create the worker
	w := worker.New(c, superscript.SuperscriptTaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize: 5,
	})

	// Register workflows
	w.RegisterWorkflow(superscript.SinglePaymentCollectionWorkflow)
	w.RegisterWorkflow(superscript.OrchestratorWorkflow)

	// Register activities
	w.RegisterActivity(activities.RunPaymentCollectionScript)

	return &Worker{
		Client:     c,
		Worker:     w,
		Logger:     logger,
		Activities: activities,
	}
}

// Start begins the worker process
func (w *Worker) Start() error {
	if w.isRunning {
		return fmt.Errorf("worker is already running")
	}

	err := w.Worker.Start()
	if err != nil {
		return fmt.Errorf("failed to start worker: %w", err)
	}

	w.isRunning = true
	w.Logger.Info("Worker started", "taskQueue", superscript.SuperscriptTaskQueue)
	return nil
}

// Stop gracefully stops the worker
func (w *Worker) Stop() {
	if !w.isRunning {
		return
	}

	w.Worker.Stop()
	w.isRunning = false
	w.Logger.Info("Worker stopped")
}

// WaitForInterrupt blocks until an interrupt signal is received
func WaitForInterrupt() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
}
