package jitaccess

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cheelim1/go-temporal-sre/demo/jit/demo-be/internal/atlas"
)

// GetUserRoleActivity is an activity that fetches the current role for a user from Atlas.
func GetUserRoleActivity(ctx context.Context, username string) (string, error) {
	slog.Info("GetUserRoleActivity started", "username", username)
	role, err := atlas.GetUserRole(ctx, username)
	if err != nil {
		slog.Error("GetUserRoleActivity failed", "username", username, "error", err)
		return "", err
	}
	slog.Info("GetUserRoleActivity completed", "username", username, "role", role)
	return role, nil
}

// SetUserRoleActivity is an activity that updates the user's role in Atlas.
func SetUserRoleActivity(ctx context.Context, username string, role string) error {
	slog.Info("SetUserRoleActivity started", "username", username, "role", role)
	if err := atlas.SetUserRole(ctx, username, role); err != nil {
		slog.Error("SetUserRoleActivity failed", "username", username, "role", role, "error", err)
		return fmt.Errorf("failed to set role: %w", err)
	}
	slog.Info("SetUserRoleActivity completed", "username", username, "role", role)
	return nil
}
