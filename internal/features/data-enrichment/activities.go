package data_enrichment

import (
	"context"
	"math/rand"
	"time"
)

func FetchDemographics(ctx context.Context, customerID string) (Demographics, error) {
	// Simulate API call delay
	time.Sleep(100 * time.Millisecond)

	// Simulate random demographics data
	rand.Seed(time.Now().UnixNano())
	age := rand.Intn(50) + 18 // Random age between 18-67
	locations := []string{"New York", "London", "Tokyo", "Sydney", "Berlin"}
	location := locations[rand.Intn(len(locations))]

	return Demographics{
		Age:      age,
		Location: location,
	}, nil
}

func MergeData(ctx context.Context, customer Customer, demographics Demographics) (EnrichedCustomer, error) {
	return EnrichedCustomer{
		Customer:     customer,
		Demographics: demographics,
	}, nil
}

func StoreEnrichedData(ctx context.Context, enriched EnrichedCustomer) error {
	// Simulate database write delay
	time.Sleep(50 * time.Millisecond)
	return nil
}
