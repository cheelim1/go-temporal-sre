package data_enrichment

import (
	"app/internal/worker"
)

// RegisterWorkflows registers all workflows and activities for the data enrichment feature
func RegisterWorkflows(w *worker.Worker) {
	// Register workflows
	w.RegisterWorkflow(DataEnrichmentWorkflow)
	w.RegisterWorkflow(EnrichSingleCustomerWorkflow)

	// Register activities
	w.RegisterActivity(FetchDemographics)
	w.RegisterActivity(MergeData)
	w.RegisterActivity(StoreEnrichedData)
}
