package template

import (
	"context"
	"log"
	"time"
)

// TemplateActivities contains all activities for the template
type TemplateActivities struct{}

// TemplateActivity is a sample activity
func (a *TemplateActivities) TemplateActivity(ctx context.Context, input ActivityInput) (ActivityResult, error) {
	log.Printf("TemplateActivity started for ID: %s", input.ID)

	// Simulate work
	time.Sleep(2 * time.Second)

	// In a real implementation, this would perform the actual work
	return ActivityResult{
		Success: true,
	}, nil
}
