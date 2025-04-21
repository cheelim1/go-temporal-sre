package data_enrichment

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
)

// EnrichData is an activity that enriches data with additional information
func EnrichData(ctx context.Context, dataID string) (*DataEnrichmentWorkflowResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("EnrichData activity started", "DataID", dataID)

	// Simulate some work
	time.Sleep(2 * time.Second)

	// Simulate successful enrichment
	result := &DataEnrichmentWorkflowResult{
		Success: true,
		Message: "Data enriched successfully",
	}

	logger.Info("EnrichData activity completed", "Success", result.Success)
	return result, nil
}

// FetchDemographics is an activity that fetches demographic data for a customer
func FetchDemographics(ctx context.Context, customerID string) (Demographics, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("FetchDemographics activity started", "CustomerID", customerID)

	// Simulate some work
	time.Sleep(1 * time.Second)

	// Simulate successful fetch
	demographics := Demographics{
		Age:      30,
		Location: "New York",
	}

	logger.Info("FetchDemographics activity completed", "CustomerID", customerID)
	return demographics, nil
}
