package iwfsuperscript

import (
	"fmt"
	"log/slog"

	"app/internal/superscript"

	"github.com/indeedeng/iwf-golang-sdk/iwf"
)

const (
	TaskQueueName             = "iwf-superscript-task-queue"
	SinglePaymentWorkflowType = "SinglePaymentWorkflow"
	OrchestratorWorkflowType  = "OrchestratorWorkflow"
)

// RegisterWorkflows registers all workflows and returns the configured WorkerService
// and the Registry used.
// It handles registry creation, activity/workflow instantiation, and registration.
func RegisterWorkflows(workerOptions iwf.WorkerOptions, scriptBasePath string, logger slog.Logger) (iwf.WorkerService, iwf.Registry) {
	registry := iwf.NewRegistry()

	// Instantiate activities
	// Pass the provided scriptBasePath and logger
	activities := superscript.NewActivities(scriptBasePath, logger)

	// Instantiate workflows with dependencies
	singlePaymentWorkflow := NewSinglePaymentWorkflow(activities)
	orchestratorWorkflow := NewOrchestratorWorkflow(activities) // Correctly passes activities

	// Register workflows
	err := registry.AddWorkflow(singlePaymentWorkflow)
	if err != nil {
		// Using panic for critical setup failure, consider logging/returning error
		panic(fmt.Sprintf("failed to register SinglePaymentWorkflow: %v", err))
	}
	err = registry.AddWorkflow(orchestratorWorkflow)
	if err != nil {
		panic(fmt.Sprintf("failed to register OrchestratorWorkflow: %v", err))
	}

	// Create and return the worker service, passing options by pointer
	workerService := iwf.NewWorkerService(registry, &workerOptions)
	return workerService, registry
}
