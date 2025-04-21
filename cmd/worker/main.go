package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"app/internal/features/batch"
	data_enrichment "app/internal/features/data-enrichment"
	"app/internal/features/kilcron"
	"app/internal/features/superscript"
	"app/internal/worker"
	"app/internal/worker/config"
)

func main() {
	// Create worker config
	cfg := config.DefaultConfig().
		WithTaskQueue("default").
		WithNamespace("default")

	// Create worker
	w, err := worker.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}
	defer w.Stop()

	// Register all demo workflows and activities
	registerDemos(w)

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

// registerDemos registers all demo workflows and activities
func registerDemos(w *worker.Worker) {
	// Register batch feature
	registerBatchFeature(w)

	// Register data enrichment feature
	registerDataEnrichmentFeature(w)

	// Register kilcron feature
	registerKilCronFeature(w)

	// Register pgshard feature
	registerPgShardFeature(w)

	// Register superscript feature
	registerSuperScriptFeature(w)
}

// registerBatchFeature registers the batch processing feature
func registerBatchFeature(w *worker.Worker) {
	w.RegisterWorkflow(batch.FeeDeductionWorkflow)
	w.RegisterActivity(&batch.BatchActivities{})
}

// registerDataEnrichmentFeature registers the data enrichment feature
func registerDataEnrichmentFeature(w *worker.Worker) {
	w.RegisterWorkflow(data_enrichment.DataEnrichmentWorkflow)
	w.RegisterWorkflow(data_enrichment.EnrichSingleCustomerWorkflow)
}

// registerKilCronFeature registers the kilcron feature
func registerKilCronFeature(w *worker.Worker) {
	w.RegisterWorkflow(kilcron.PaymentWorkflow)
}

// registerPgShardFeature registers the pgshard feature
func registerPgShardFeature(w *worker.Worker) {
	// TODO: Add pgshard workflow and activities once implemented
}

// registerSuperScriptFeature registers the superscript feature
func registerSuperScriptFeature(w *worker.Worker) {
	w.RegisterWorkflow(superscript.SinglePaymentCollectionWorkflow)
	w.RegisterWorkflow(superscript.OrchestratorWorkflow)
	w.RegisterActivity(&superscript.Activities{})
}
