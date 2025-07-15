package jit

import (
	"app/internal/atlas"
	"app/internal/jitaccess"
	"app/internal/worker"
	"app/internal/worker/config"
	"fmt"
)

// Feature represents the JIT (Just-In-Time access) feature
type Feature struct {
	taskQueue string
}

// NewFeature creates a new JIT feature
func NewFeature() *Feature {
	return &Feature{
		taskQueue: "jit_access_task_queue",
	}
}

// RegisterComponents registers JIT workflows and activities
func (f *Feature) RegisterComponents(registry *worker.Registry, cfg interface{}) error {
	// Cast config to get the task queue configuration
	if workerConfig, ok := cfg.(*config.WorkerConfig); ok {
		f.taskQueue = workerConfig.JITTaskQueue
	}

	// Initialize the Atlas client (required for JIT activities)
	if err := atlas.InitAtlasClient(); err != nil {
		return fmt.Errorf("failed to initialize Atlas client for JIT feature: %w", err)
	}

	// Register workflows
	registry.RegisterWorkflow("JITAccessWorkflow", jitaccess.JITAccessWorkflow)

	// Register activities
	registry.RegisterActivity("GetUserRoleActivity", jitaccess.GetUserRoleActivity)
	registry.RegisterActivity("SetUserRoleActivity", jitaccess.SetUserRoleActivity)

	return nil
}

// GetTaskQueues returns the task queues used by this feature
func (f *Feature) GetTaskQueues() []string {
	return []string{f.taskQueue}
}

// GetFeatureName returns the name of this feature
func (f *Feature) GetFeatureName() string {
	return "jit"
}
