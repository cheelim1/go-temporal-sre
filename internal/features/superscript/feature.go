package superscript

import (
	"app/internal/superscript"
	"app/internal/worker"
	"app/internal/worker/config"
	"log/slog"

	"go.temporal.io/sdk/log"
)

// Feature represents the superscript feature
type Feature struct {
	taskQueue  string
	activities *superscript.Activities
	logger     log.Logger
}

// NewFeature creates a new superscript feature
func NewFeature(logger log.Logger) *Feature {
	return &Feature{
		taskQueue: "superscript_task_queue",
		logger:    logger,
	}
}

// RegisterComponents registers superscript workflows and activities
func (f *Feature) RegisterComponents(registry *worker.Registry, cfg interface{}) error {
	// Cast config to get the script base path
	scriptBasePath := "./internal/superscript/scripts/"
	if workerConfig, ok := cfg.(*config.WorkerConfig); ok {
		f.taskQueue = superscript.SuperscriptTaskQueue
		scriptBasePath = workerConfig.SuperscriptBasePath
	}

	// Create activities using the proper constructor
	// Convert temporal logger to slog.Logger (simplified approach)
	slogLogger := slog.Default()
	f.activities = superscript.NewActivities(scriptBasePath, *slogLogger)

	// Register workflows
	registry.RegisterWorkflow("SinglePaymentCollectionWorkflow", superscript.SinglePaymentCollectionWorkflow)
	registry.RegisterWorkflow("OrchestratorWorkflow", superscript.OrchestratorWorkflow)

	// Register activities
	registry.RegisterActivity("RunPaymentCollectionScript", f.activities.RunPaymentCollectionScript)

	return nil
}

// GetTaskQueues returns the task queues used by this feature
func (f *Feature) GetTaskQueues() []string {
	return []string{f.taskQueue}
}

// GetFeatureName returns the name of this feature
func (f *Feature) GetFeatureName() string {
	return "superscript"
}
