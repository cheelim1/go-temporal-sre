package worker

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"app/internal/worker/config"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
)

// CentralizedWorker manages a centralized Temporal worker with multiple features
type CentralizedWorker struct {
	config         *config.WorkerConfig
	client         client.Client
	workers        map[string]worker.Worker
	registry       *Registry
	featureManager *FeatureManager
	logger         log.Logger
	isRunning      bool
	shutdown       chan struct{}
	wg             sync.WaitGroup
}

// NewCentralizedWorker creates a new centralized worker
func NewCentralizedWorker(cfg *config.WorkerConfig, logger log.Logger) (*CentralizedWorker, error) {
	// Create Temporal client
	temporalClient, err := client.Dial(client.Options{
		HostPort:  cfg.TemporalHost,
		Namespace: cfg.TemporalNamespace,
		Logger:    logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	// Create registry and feature manager
	registry := NewRegistry(logger)
	featureManager := NewFeatureManager(registry, logger)

	return &CentralizedWorker{
		config:         cfg,
		client:         temporalClient,
		workers:        make(map[string]worker.Worker),
		registry:       registry,
		featureManager: featureManager,
		logger:         logger,
		shutdown:       make(chan struct{}),
	}, nil
}

// RegisterFeature registers a feature with the centralized worker
func (cw *CentralizedWorker) RegisterFeature(feature FeatureRegistrar) error {
	return cw.featureManager.RegisterFeature(feature)
}

// InitializeFeatures initializes all enabled features
func (cw *CentralizedWorker) InitializeFeatures() error {
	cw.logger.Info("Initializing features", "enabled", cw.config.EnabledFeatures)

	for _, featureName := range cw.config.EnabledFeatures {
		if err := cw.featureManager.InitializeFeature(featureName, cw.config); err != nil {
			cw.logger.Error("Failed to initialize feature", "name", featureName, "error", err)
			return fmt.Errorf("failed to initialize feature %s: %w", featureName, err)
		}
	}

	return nil
}

// CreateWorkers creates Temporal workers for all required task queues
func (cw *CentralizedWorker) CreateWorkers() error {
	// Get all task queues from enabled features
	taskQueues := cw.featureManager.GetAllTaskQueues()
	if len(taskQueues) == 0 {
		cw.logger.Warn("No task queues found, creating a default worker")
		taskQueues = []string{"default"}
	}

	// Create workers for each task queue
	for _, taskQueue := range taskQueues {
		w := worker.New(cw.client, taskQueue, worker.Options{
			MaxConcurrentActivityExecutionSize:     cw.config.MaxConcurrentActivities,
			MaxConcurrentWorkflowTaskExecutionSize: cw.config.MaxConcurrentWorkflows,
		})

		// Apply all registrations to this worker
		cw.registry.ApplyRegistrations(w)

		cw.workers[taskQueue] = w
		cw.logger.Info("Created worker for task queue", "taskQueue", taskQueue)
	}

	return nil
}

// Start starts the centralized worker
func (cw *CentralizedWorker) Start() error {
	if cw.isRunning {
		return fmt.Errorf("worker is already running")
	}

	cw.logger.Info("Starting centralized Temporal worker")

	// Initialize features
	if err := cw.InitializeFeatures(); err != nil {
		return fmt.Errorf("failed to initialize features: %w", err)
	}

	// Create workers
	if err := cw.CreateWorkers(); err != nil {
		return fmt.Errorf("failed to create workers: %w", err)
	}

	// Start all workers
	for taskQueue, w := range cw.workers {
		cw.wg.Add(1)
		go func(tq string, w worker.Worker) {
			defer cw.wg.Done()
			cw.logger.Info("Starting worker", "taskQueue", tq)
			if err := w.Run(worker.InterruptCh()); err != nil {
				cw.logger.Error("Worker failed", "taskQueue", tq, "error", err)
			}
		}(taskQueue, w)
	}

	cw.isRunning = true
	cw.logger.Info("Centralized worker started successfully",
		"features", cw.config.EnabledFeatures,
		"taskQueues", len(cw.workers))

	return nil
}

// Stop gracefully stops the centralized worker
func (cw *CentralizedWorker) Stop() {
	if !cw.isRunning {
		return
	}

	cw.logger.Info("Stopping centralized worker")

	// Signal shutdown to any waiting goroutines
	close(cw.shutdown)

	// Stop all workers
	for taskQueue, w := range cw.workers {
		cw.logger.Info("Stopping worker", "taskQueue", taskQueue)
		w.Stop()
	}

	// Wait for all workers to stop
	cw.wg.Wait()

	// Close the Temporal client
	cw.client.Close()

	cw.isRunning = false
	cw.logger.Info("Centralized worker stopped")
}

// WaitForShutdown waits for shutdown signal
func (cw *CentralizedWorker) WaitForShutdown() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		cw.logger.Info("Received shutdown signal", "signal", sig)
	case <-cw.shutdown:
		cw.logger.Info("Shutdown requested")
	}

	cw.Stop()
}

// GetClient returns the Temporal client
func (cw *CentralizedWorker) GetClient() client.Client {
	return cw.client
}

// GetRegistry returns the registry
func (cw *CentralizedWorker) GetRegistry() *Registry {
	return cw.registry
}

// GetFeatureManager returns the feature manager
func (cw *CentralizedWorker) GetFeatureManager() *FeatureManager {
	return cw.featureManager
}

// GetStatus returns the current status of the worker
func (cw *CentralizedWorker) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"isRunning":            cw.isRunning,
		"enabledFeatures":      cw.config.EnabledFeatures,
		"taskQueues":           len(cw.workers),
		"registeredWorkflows":  len(cw.registry.GetRegisteredWorkflows()),
		"registeredActivities": len(cw.registry.GetRegisteredActivities()),
	}
}
