package data_enrichment

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// Customer represents a customer entity
type Customer struct {
	ID    string
	Name  string
	Email string
}

// Demographics represents demographic information
type Demographics struct {
	Age      int
	Location string
}

// EnrichedCustomer represents a customer with enriched demographic data
type EnrichedCustomer struct {
	Customer     Customer
	Demographics Demographics
}

// DataEnrichmentWorkflowInput represents the input for the data enrichment workflow
type DataEnrichmentWorkflowInput struct {
	DataID string
}

// DataEnrichmentWorkflowResult represents the result of the data enrichment workflow
type DataEnrichmentWorkflowResult struct {
	Success bool
	Message string
}

// DataEnrichmentWorkflow is a workflow that enriches data with additional information
func DataEnrichmentWorkflow(ctx workflow.Context, input DataEnrichmentWorkflowInput) (*DataEnrichmentWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("DataEnrichmentWorkflow started", "DataID", input.DataID)

	// Set activity options
	activityOpts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})

	// Execute enrichment activity
	var result *DataEnrichmentWorkflowResult
	err := workflow.ExecuteActivity(activityOpts, EnrichData, input.DataID).Get(ctx, &result)
	if err != nil {
		logger.Error("DataEnrichmentWorkflow failed", "Error", err)
		return nil, err
	}

	logger.Info("DataEnrichmentWorkflow completed", "Success", result.Success)
	return result, nil
}

// EnrichSingleCustomerWorkflow is a workflow that enriches a single customer's data
func EnrichSingleCustomerWorkflow(ctx workflow.Context, customer Customer) (*EnrichedCustomer, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("EnrichSingleCustomerWorkflow started", "CustomerID", customer.ID)

	// Set activity options
	activityOpts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})

	// Execute activities
	var demographics Demographics
	err := workflow.ExecuteActivity(activityOpts, FetchDemographics, customer.ID).Get(ctx, &demographics)
	if err != nil {
		logger.Error("Failed to fetch demographics", "Error", err)
		return nil, err
	}

	// Create enriched customer
	enriched := &EnrichedCustomer{
		Customer:     customer,
		Demographics: demographics,
	}

	logger.Info("EnrichSingleCustomerWorkflow completed", "CustomerID", customer.ID)
	return enriched, nil
}
