package kilcron

import (
	"app/internal/kilcron"
	"app/internal/worker"
	"app/internal/worker/config"
)

// Feature represents the kilcron feature
type Feature struct {
	taskQueue string
}

// NewFeature creates a new kilcron feature
func NewFeature() *Feature {
	return &Feature{
		taskQueue: "kilcron_task_queue",
	}
}

// RegisterComponents registers kilcron workflows and activities
func (f *Feature) RegisterComponents(registry *worker.Registry, cfg interface{}) error {
	// Cast config to our expected type
	workerConfig, ok := cfg.(*config.WorkerConfig)
	if ok {
		f.taskQueue = workerConfig.KilcronTaskQueue
	}

	// Register workflows
	registry.RegisterWorkflow("PaymentWorkflow", kilcron.PaymentWorkflow)

	// Register activities
	registry.RegisterActivity("MakePayment", kilcron.MakePayment)

	return nil
}

// GetTaskQueues returns the task queues used by this feature
func (f *Feature) GetTaskQueues() []string {
	return []string{f.taskQueue}
}

// GetFeatureName returns the name of this feature
func (f *Feature) GetFeatureName() string {
	return "kilcron"
}
