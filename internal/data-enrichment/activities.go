package data_enrichment

import "time"

type DataEnrichmentActivities struct{}

func (a *DataEnrichmentActivities) FetchDemographics(customerID string) (Demographics, error) {
	// Simulate real system delay
	time.Sleep(time.Second)
	return Demographics{Age: 30, Location: "New York, NY"}, nil
}

func (a *DataEnrichmentActivities) MergeData(customer Customer, demographics Demographics) (EnrichedCustomer, error) {
	// Simulate complex setup
	time.Sleep(30 * time.Second)
	return EnrichedCustomer{customer, demographics}, nil
}

func (a *DataEnrichmentActivities) StoreEnrichedData(enriched EnrichedCustomer) error {
	// Simulate real system delay
	time.Sleep(time.Second)
	return nil
}

type CustomerFetchActivities struct{}

func (a *CustomerFetchActivities) FetchCustomersNeedingEnrichment() ([]Customer, error) {
	return []Customer{
		{ID: "1", Name: "Alice", Email: "alice@example.com"},
		{ID: "2", Name: "Bob", Email: "bob@example.com"},
	}, nil
}
