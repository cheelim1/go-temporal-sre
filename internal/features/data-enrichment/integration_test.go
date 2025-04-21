package data_enrichment

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

// TestActivities implements test activities for integration tests
type TestActivities struct {
	FetchDemographicsFunc func(string) (Demographics, error)
	MergeDataFunc         func(Customer, Demographics) (EnrichedCustomer, error)
	StoreEnrichedDataFunc func(EnrichedCustomer) error
}

func (a *TestActivities) FetchDemographics(customerID string) (Demographics, error) {
	return a.FetchDemographicsFunc(customerID)
}

func (a *TestActivities) MergeData(customer Customer, demographics Demographics) (EnrichedCustomer, error) {
	return a.MergeDataFunc(customer, demographics)
}

func (a *TestActivities) StoreEnrichedData(enriched EnrichedCustomer) error {
	return a.StoreEnrichedDataFunc(enriched)
}

// TestEnrichSingleCustomerWorkflow tests the single customer enrichment workflow
func TestEnrichSingleCustomerWorkflow(t *testing.T) {
	// Create test suite
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	// Create test activities
	testActivities := &TestActivities{}

	// Configure test activities
	testActivities.FetchDemographicsFunc = func(customerID string) (Demographics, error) {
		return Demographics{
			Age:      30,
			Location: "Test Location",
		}, nil
	}

	testActivities.MergeDataFunc = func(customer Customer, demographics Demographics) (EnrichedCustomer, error) {
		return EnrichedCustomer{
			Customer:     customer,
			Demographics: demographics,
		}, nil
	}

	testActivities.StoreEnrichedDataFunc = func(enriched EnrichedCustomer) error {
		return nil
	}

	// Register workflows
	env.RegisterWorkflow(EnrichSingleCustomerWorkflow)

	// Register activities with their proper names
	env.RegisterActivityWithOptions(testActivities.FetchDemographics, activity.RegisterOptions{
		Name: ActivityFetchDemographics,
	})
	env.RegisterActivityWithOptions(testActivities.MergeData, activity.RegisterOptions{
		Name: ActivityMergeData,
	})
	env.RegisterActivityWithOptions(testActivities.StoreEnrichedData, activity.RegisterOptions{
		Name: ActivityStoreEnrichedData,
	})

	// Test data
	testCustomer := Customer{
		ID:    "test-1",
		Name:  "Test User",
		Email: "test@example.com",
	}

	// Execute workflow
	env.ExecuteWorkflow(EnrichSingleCustomerWorkflow, testCustomer)

	// Verify workflow completed successfully
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// Get workflow result
	var enrichedCustomer EnrichedCustomer
	require.NoError(t, env.GetWorkflowResult(&enrichedCustomer))

	// Verify result
	assert.Equal(t, testCustomer.ID, enrichedCustomer.ID)
	assert.Equal(t, testCustomer.Name, enrichedCustomer.Name)
	assert.Equal(t, testCustomer.Email, enrichedCustomer.Email)
	assert.Equal(t, 30, enrichedCustomer.Age)
	assert.Equal(t, "Test Location", enrichedCustomer.Location)
}

// TestEnrichSingleCustomerWorkflowErrorHandling tests error handling in the workflow
func TestEnrichSingleCustomerWorkflowErrorHandling(t *testing.T) {
	// Create test suite
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	// Create test activities
	testActivities := &TestActivities{}

	// Configure test activities to fail
	testActivities.FetchDemographicsFunc = func(customerID string) (Demographics, error) {
		return Demographics{}, fmt.Errorf("simulated demographics service error")
	}

	testActivities.MergeDataFunc = func(customer Customer, demographics Demographics) (EnrichedCustomer, error) {
		return EnrichedCustomer{}, fmt.Errorf("simulated merge error")
	}

	testActivities.StoreEnrichedDataFunc = func(enriched EnrichedCustomer) error {
		return fmt.Errorf("simulated storage error")
	}

	// Register workflows
	env.RegisterWorkflow(EnrichSingleCustomerWorkflow)

	// Register activities with their proper names
	env.RegisterActivityWithOptions(testActivities.FetchDemographics, activity.RegisterOptions{
		Name: ActivityFetchDemographics,
	})
	env.RegisterActivityWithOptions(testActivities.MergeData, activity.RegisterOptions{
		Name: ActivityMergeData,
	})
	env.RegisterActivityWithOptions(testActivities.StoreEnrichedData, activity.RegisterOptions{
		Name: ActivityStoreEnrichedData,
	})

	// Test data
	testCustomer := Customer{
		ID:    "test-1",
		Name:  "Test User",
		Email: "test@example.com",
	}

	// Execute workflow
	env.ExecuteWorkflow(EnrichSingleCustomerWorkflow, testCustomer)

	// Verify workflow completed with error
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	assert.Contains(t, env.GetWorkflowError().Error(), "simulated demographics service error")
}

// TestEnrichSingleCustomerWorkflowRetry tests retry behavior in the workflow
func TestEnrichSingleCustomerWorkflowRetry(t *testing.T) {
	// Create test suite
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	// Create test activities
	testActivities := &TestActivities{}

	// Track activity calls
	var activityCalled bool
	var fetchAttempts int

	// Configure test activities
	testActivities.FetchDemographicsFunc = func(customerID string) (Demographics, error) {
		activityCalled = true
		fetchAttempts++
		if fetchAttempts < 3 {
			return Demographics{}, temporal.NewApplicationError("temporary error", "TEMPORARY_ERROR")
		}
		return Demographics{
			Age:      30,
			Location: "Test Location",
		}, nil
	}

	testActivities.MergeDataFunc = func(customer Customer, demographics Demographics) (EnrichedCustomer, error) {
		return EnrichedCustomer{
			Customer:     customer,
			Demographics: demographics,
		}, nil
	}

	testActivities.StoreEnrichedDataFunc = func(enriched EnrichedCustomer) error {
		return nil
	}

	// Register workflows and activities
	env.RegisterWorkflowWithOptions(EnrichSingleCustomerWorkflow, workflow.RegisterOptions{
		DisableAlreadyRegisteredCheck: false,
		VersioningBehavior:            workflow.VersioningBehavior(2),
	})

	// Register activities
	env.RegisterActivityWithOptions(testActivities.FetchDemographics, activity.RegisterOptions{
		Name:                          "FetchDemographics",
		DisableAlreadyRegisteredCheck: false,
		SkipInvalidStructFunctions:    false,
	})
	env.RegisterActivity(testActivities.MergeData)
	env.RegisterActivity(testActivities.StoreEnrichedData)

	// Test data
	testCustomer := Customer{
		ID:    "test-1",
		Name:  "Test User",
		Email: "test@example.com",
	}

	// Execute workflow
	env.ExecuteWorkflow(EnrichSingleCustomerWorkflow, testCustomer)

	// Verify workflow completed successfully
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// Verify activity was called and retried
	require.True(t, activityCalled, "FetchDemographics should have been called")
	assert.Equal(t, 3, fetchAttempts, "FetchDemographics should have been retried 3 times")

	// Get workflow result
	var enrichedCustomer EnrichedCustomer
	require.NoError(t, env.GetWorkflowResult(&enrichedCustomer))

	// Verify result
	assert.Equal(t, testCustomer.ID, enrichedCustomer.ID)
	assert.Equal(t, testCustomer.Name, enrichedCustomer.Name)
	assert.Equal(t, testCustomer.Email, enrichedCustomer.Email)
	assert.Equal(t, 30, enrichedCustomer.Age)
	assert.Equal(t, "Test Location", enrichedCustomer.Location)
}
