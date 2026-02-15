package workflows

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// MaintenanceRequest represents a maintenance mode request
type MaintenanceRequest struct {
	AppEndpoint string
}

// MaintenanceResult represents the result of maintenance mode
type MaintenanceResult struct {
	Success         bool
	Message         string
	MaintenanceTime time.Duration
}

const (
	SignalEnterMaintenance = "enter_maintenance"
	SignalResume           = "resume"
)

// MaintenanceModeWorkflow manages maintenance mode with signal-based control.
// It waits for an enter_maintenance signal, then waits for a resume signal.
// Uses simple sequential Receive calls -- no timers or selector loops.
func MaintenanceModeWorkflow(ctx workflow.Context, input MaintenanceRequest) (*MaintenanceResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("MaintenanceModeWorkflow started",
		"workflow_id", workflow.GetInfo(ctx).WorkflowExecution.ID,
	)

	result := &MaintenanceResult{}

	// Step 1: Wait for "enter_maintenance" signal
	enterCh := workflow.GetSignalChannel(ctx, SignalEnterMaintenance)
	logger.Info("Waiting for enter_maintenance signal...")

	var enterSignal any
	enterCh.Receive(ctx, &enterSignal)

	logger.Info("Received enter_maintenance signal")
	logger.Info("App now in maintenance mode - rejecting writes (503)")

	maintenanceStart := workflow.Now(ctx)

	// Step 2: Wait for "resume" signal
	resumeCh := workflow.GetSignalChannel(ctx, SignalResume)
	logger.Info("Waiting for resume signal...")

	var resumeSignal any
	resumeCh.Receive(ctx, &resumeSignal)

	logger.Info("Received resume signal")
	maintenanceDuration := workflow.Now(ctx).Sub(maintenanceStart)
	logger.Info("App resuming normal operation - accepting writes",
		"maintenance_duration", maintenanceDuration,
	)

	result.Success = true
	result.Message = "Maintenance mode completed successfully"
	result.MaintenanceTime = maintenanceDuration

	logger.Info("MaintenanceModeWorkflow completed successfully")
	return result, nil
}
