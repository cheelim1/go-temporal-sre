package worker

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
)

// Registry holds all workflow and activity registrations
type Registry struct {
	workflows  map[string]interface{}
	activities map[string]interface{}
	logger     log.Logger
}

// NewRegistry creates a new registry for workflows and activities
func NewRegistry(logger log.Logger) *Registry {
	return &Registry{
		workflows:  make(map[string]interface{}),
		activities: make(map[string]interface{}),
		logger:     logger,
	}
}

// RegisterWorkflow registers a workflow with the registry
func (r *Registry) RegisterWorkflow(name string, workflow interface{}) {
	r.workflows[name] = workflow
	r.logger.Info("Registered workflow", "name", name)
}

// RegisterActivity registers an activity with the registry
func (r *Registry) RegisterActivity(name string, activity interface{}) {
	r.activities[name] = activity
	r.logger.Info("Registered activity", "name", name)
}

// ApplyRegistrations applies all registered workflows and activities to a worker
func (r *Registry) ApplyRegistrations(w worker.Worker) {
	// Register all workflows
	for name, workflow := range r.workflows {
		w.RegisterWorkflow(workflow)
		r.logger.Info("Applied workflow registration", "name", name)
	}

	// Register all activities
	for name, activity := range r.activities {
		w.RegisterActivity(activity)
		r.logger.Info("Applied activity registration", "name", name)
	}
}

// GetRegisteredWorkflows returns the list of registered workflow names
func (r *Registry) GetRegisteredWorkflows() []string {
	var names []string
	for name := range r.workflows {
		names = append(names, name)
	}
	return names
}

// GetRegisteredActivities returns the list of registered activity names
func (r *Registry) GetRegisteredActivities() []string {
	var names []string
	for name := range r.activities {
		names = append(names, name)
	}
	return names
}

// FeatureRegistrar is an interface that features must implement to register their components
type FeatureRegistrar interface {
	RegisterComponents(registry *Registry, config interface{}) error
	GetTaskQueues() []string
	GetFeatureName() string
}

// FeatureManager manages feature registration and lifecycle
type FeatureManager struct {
	features map[string]FeatureRegistrar
	registry *Registry
	logger   log.Logger
}

// NewFeatureManager creates a new feature manager
func NewFeatureManager(registry *Registry, logger log.Logger) *FeatureManager {
	return &FeatureManager{
		features: make(map[string]FeatureRegistrar),
		registry: registry,
		logger:   logger,
	}
}

// RegisterFeature registers a feature with the manager
func (fm *FeatureManager) RegisterFeature(feature FeatureRegistrar) error {
	name := feature.GetFeatureName()
	fm.features[name] = feature
	fm.logger.Info("Registered feature", "name", name)
	return nil
}

// InitializeFeature initializes a specific feature
func (fm *FeatureManager) InitializeFeature(featureName string, config interface{}) error {
	feature, exists := fm.features[featureName]
	if !exists {
		return fmt.Errorf("feature %s not found", featureName)
	}

	if err := feature.RegisterComponents(fm.registry, config); err != nil {
		return fmt.Errorf("failed to register components for feature %s: %w", featureName, err)
	}

	fm.logger.Info("Initialized feature", "name", featureName)
	return nil
}

// GetFeatureTaskQueues returns all task queues for a feature
func (fm *FeatureManager) GetFeatureTaskQueues(featureName string) []string {
	feature, exists := fm.features[featureName]
	if !exists {
		return nil
	}
	return feature.GetTaskQueues()
}

// GetAllTaskQueues returns all task queues from all registered features
func (fm *FeatureManager) GetAllTaskQueues() []string {
	var allQueues []string
	for _, feature := range fm.features {
		queues := feature.GetTaskQueues()
		allQueues = append(allQueues, queues...)
	}
	return allQueues
}

// GetRegisteredFeatures returns the list of registered feature names
func (fm *FeatureManager) GetRegisteredFeatures() []string {
	var names []string
	for name := range fm.features {
		names = append(names, name)
	}
	return names
}
