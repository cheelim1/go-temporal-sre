package jitaccess

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// JITAccessRequest defines the input for the JIT access workflow.
type JITAccessRequest struct {
	Username string
	Reason   string
	NewRole  string
	Duration time.Duration
}

// JITAccessWorkflow is the Temporal workflow that performs the JIT access process.
func JITAccessWorkflow(ctx workflow.Context, req JITAccessRequest) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting JITAccessWorkflow", "username", req.Username, "new_role", req.NewRole, "duration", req.Duration)

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    1 * time.Minute,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	var originalRole string
	// Fetch the user's current role.
	if err := workflow.ExecuteActivity(ctx, GetUserRoleActivity, req.Username).Get(ctx, &originalRole); err != nil {
		logger.Error("failed to get user role", "error", err)
		return err
	}
	logger.Info("Fetched current role", "username", req.Username, "current_role", originalRole)

	// Ensure the new role is different
	if originalRole == req.NewRole {
		return temporal.NewNonRetryableApplicationError("new_role cannot be same as current role", "InvalidRole", nil)
	}

	// Update the user's role to the new role.
	if err := workflow.ExecuteActivity(ctx, SetUserRoleActivity, req.Username, req.NewRole).Get(ctx, nil); err != nil {
		logger.Error("failed to set new role", "error", err)
		return err
	}
	logger.Info("User role updated to new role", "username", req.Username, "new_role", req.NewRole)

	// Wait for the duration.
	logger.Info("Sleeping for duration", "duration", req.Duration)
	workflow.Sleep(ctx, req.Duration)

	// Revert the user's role to the original role.
	if err := workflow.ExecuteActivity(ctx, SetUserRoleActivity, req.Username, originalRole).Get(ctx, nil); err != nil {
		logger.Error("failed to revert user role", "error", err)
		return err
	}
	logger.Info("User role reverted to original", "username", req.Username, "original_role", originalRole)
	return nil
}
