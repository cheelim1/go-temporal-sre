package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"app/internal/demos/batch"
	"app/internal/worker"
	"app/internal/worker/config"
)

func main() {
	// Create worker config
	cfg := config.DefaultConfig().
		WithTaskQueue("batch-task-queue").
		WithNamespace("default")

	// Create worker
	w, err := worker.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}
	defer w.Stop()

	// Register workflow and activities
	w.RegisterWorkflow(batch.FeeDeductionWorkflow)
	w.RegisterActivity(&batch.BatchActivities{})

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received, stopping worker...")
		w.Stop()
		os.Exit(0)
	}()

	// Start the worker
	if err := w.Start(); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}
}
