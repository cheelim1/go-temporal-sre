package data_enrichment

import (
	"fmt"
	"go.temporal.io/sdk/workflow"
	"time"
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

// TODO: Use Schedules for a ScheduledDataEnrichmentWorkflow
// this will coordinate the every 6 hours execution and block multiple running ..

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
				WorkflowID: fmt.Sprintf("enrich-" + customer.ID),
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
		},
	)

	dea := DataEnrichmentActivities{}

	// Calls a external data source to have enriched data to work with .
	var demographics Demographics
	if err := workflow.ExecuteActivity(actx, dea.FetchDemographics).Get(ctx, &demographics); err != nil {
		return EnrichedCustomer{}, err
	}

	// Here we can do further local transformation
	// Arguably it might be deterministic if simple enough; no need to use Activities
	var enriched EnrichedCustomer
	if err := workflow.ExecuteActivity(actx, dea.MergeData, customer, demographics).Get(ctx, &enriched); err != nil {
		return EnrichedCustomer{}, err
	}

	// Arguably; this might instead be sending to another workflow which deals with
	// orchestrating global messaging; a reverse ETL
	// This store might be a S3 or a DynamoDB
	if err := workflow.ExecuteActivity(actx, dea.StoreEnrichedData, enriched).Get(ctx, nil); err != nil {
		return EnrichedCustomer{}, err
	}

	// TODO: Send signal to the GlobalNotificationWorkflow that there is a new batch of leads
	// to taek the next step

	return enriched, nil
}
