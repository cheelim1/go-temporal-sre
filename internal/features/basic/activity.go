package basic

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
)

// GreetingActivity is a simple activity that returns a greeting
func GreetingActivity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("GreetingActivity started", "name", name)

	// Simulate some work
	greeting := fmt.Sprintf("Hello, %s!", name)

	logger.Info("GreetingActivity completed", "greeting", greeting)
	return greeting, nil
}
