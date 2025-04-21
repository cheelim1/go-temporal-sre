package data_enrichment

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Activity names
const (
	ActivityFetchDemographics = "FetchDemographics"
	ActivityMergeData         = "MergeData"
	ActivityStoreEnrichedData = "StoreEnrichedData"
)

type Customer struct {
	ID    string
	Name  string
	Email string
}

type Demographics struct {
	Age      int
	Location string
}

type EnrichedCustomer struct {
	Customer
	Demographics
}

type TaskResult struct {
	Status  string
	Message string
}

var TQ = "data-enrichment-demo"

// DataEnrichmentWorkflow is a workflow that gets kicked off every 6 hours
func DataEnrichmentWorkflow(
	ctx workflow.Context,
	customers []Customer,
) ([]EnrichedCustomer, error) {
	var enrichedCustomers []EnrichedCustomer
	var futures []workflow.Future

	for _, customer := range customers {
		childCtx := workflow.WithChildOptions(
			ctx,
			workflow.ChildWorkflowOptions{
				WorkflowID: fmt.Sprintf("enrich-%s", customer.ID),
				TaskQueue:  TQ,
			},
		)

		future := workflow.ExecuteChildWorkflow(childCtx, EnrichSingleCustomerWorkflow, customer)
		futures = append(futures, future)
	}

	for _, future := range futures {
		var enriched EnrichedCustomer
		if err := future.Get(ctx, &enriched); err != nil {
			fmt.Println("enrich failed - ID:", workflow.GetInfo(ctx).WorkflowExecution.ID, " ERR: ", err)
			continue
		}
		enrichedCustomers = append(enrichedCustomers, enriched)
	}

	return enrichedCustomers, nil
}

func EnrichSingleCustomerWorkflow(
	ctx workflow.Context,
	customer Customer,
) (EnrichedCustomer, error) {
	actx := workflow.WithActivityOptions(
		ctx,
		workflow.ActivityOptions{
			TaskQueue:           TQ,
			StartToCloseTimeout: 5 * time.Second,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 2,
			},
		},
	)

	var demographics Demographics
	if err := workflow.ExecuteActivity(actx, ActivityFetchDemographics, customer.ID).Get(ctx, &demographics); err != nil {
		return EnrichedCustomer{}, err
	}

	var enriched EnrichedCustomer
	if err := workflow.ExecuteActivity(actx, ActivityMergeData, customer, demographics).Get(ctx, &enriched); err != nil {
		return EnrichedCustomer{}, err
	}

	if err := workflow.ExecuteActivity(actx, ActivityStoreEnrichedData, enriched).Get(ctx, nil); err != nil {
		return EnrichedCustomer{}, err
	}

	return enriched, nil
}
