package data_enrichment_test

import (
	data_enrichment "app/internal/demos/data-enrichment"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

// TestActivities provides test implementations of the activities
type TestActivities struct {
	FetchDemographicsFunc func(string) (data_enrichment.Demographics, error)
	MergeDataFunc         func(data_enrichment.Customer, data_enrichment.Demographics) (data_enrichment.EnrichedCustomer, error)
	StoreEnrichedDataFunc func(data_enrichment.EnrichedCustomer) error
}

func (a *TestActivities) FetchDemographics(customerID string) (data_enrichment.Demographics, error) {
	return a.FetchDemographicsFunc(customerID)
}

func (a *TestActivities) MergeData(customer data_enrichment.Customer, demographics data_enrichment.Demographics) (data_enrichment.EnrichedCustomer, error) {
	return a.MergeDataFunc(customer, demographics)
}

func (a *TestActivities) StoreEnrichedData(enriched data_enrichment.EnrichedCustomer) error {
	return a.StoreEnrichedDataFunc(enriched)
}

// Note: Integration tests with real Temporal DevServer have been moved to e2e_test.go file.
// This file now contains only mock-based tests using the testsuite.WorkflowTestSuite.

func TestSingleWorkflowHappyPath(t *testing.T) {
	// This test uses the testsuite environment to test a workflow
	// without actually running a real Temporal server
	testSuite := testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Setup test activities
	testActivities := new(TestActivities)
	testActivities.FetchDemographicsFunc = func(customerID string) (data_enrichment.Demographics, error) {
		// No delay in test environment
		return data_enrichment.Demographics{Age: 30, Location: "Test Location"}, nil
		// Below will fail the TDD
		//return data_enrichment.Demographics{}, fmt.Errorf("some error")
	}

	testActivities.MergeDataFunc = func(customer data_enrichment.Customer, demographics data_enrichment.Demographics) (data_enrichment.EnrichedCustomer, error) {
		return data_enrichment.EnrichedCustomer{Customer: customer, Demographics: demographics}, nil
	}

	testActivities.StoreEnrichedDataFunc = func(enriched data_enrichment.EnrichedCustomer) error {
		return nil
	}

	// Register workflows and activities
	env.RegisterWorkflow(data_enrichment.DataEnrichmentWorkflow)
	env.RegisterWorkflow(data_enrichment.EnrichSingleCustomerWorkflow)
	env.RegisterActivity(testActivities.FetchDemographics)
	env.RegisterActivity(testActivities.MergeData)
	env.RegisterActivity(testActivities.StoreEnrichedData)

	// Test customer
	testCustomer := data_enrichment.Customer{ID: "simple-id", Name: "Simple User", Email: "simple@example.com"}

	// Execute the workflow
	env.ExecuteWorkflow(data_enrichment.EnrichSingleCustomerWorkflow, testCustomer)

	// Verify workflow completed successfully
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// Get and verify result
	var result data_enrichment.EnrichedCustomer
	require.NoError(t, env.GetWorkflowResult(&result))

	// Verify result
	assert.Equal(t, testCustomer.ID, result.Customer.ID, "Customer ID should match")
	assert.Equal(t, 30, result.Demographics.Age, "Demographics age should match")
	assert.Equal(t, "Test Location", result.Demographics.Location, "Demographics location should match")
}

func TestSingleWorkflowErrorPath(t *testing.T) {
	// This test uses the testsuite environment to test a workflow error path
	testSuite := testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Setup test activities with an error
	testActivities := new(TestActivities)
	simulatedError := fmt.Errorf("simulated demographics service error")

	// Track if the activity was called
	activityCalled := false
	testActivities.FetchDemographicsFunc = func(customerID string) (data_enrichment.Demographics, error) {
		activityCalled = true
		return data_enrichment.Demographics{}, simulatedError
	}

	// Register workflows and activities
	env.RegisterWorkflow(data_enrichment.DataEnrichmentWorkflow)
	env.RegisterWorkflow(data_enrichment.EnrichSingleCustomerWorkflow)
	env.RegisterActivity(testActivities.FetchDemographics)
	env.RegisterActivity(testActivities.MergeData)
	env.RegisterActivity(testActivities.StoreEnrichedData)

	// Test customer
	testCustomer := data_enrichment.Customer{ID: "error-id", Name: "Error User", Email: "error@example.com"}

	// Execute the workflow
	env.ExecuteWorkflow(data_enrichment.EnrichSingleCustomerWorkflow, testCustomer)

	// Verify workflow completed with an error
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())

	// Verify the activity was called
	require.True(t, activityCalled, "FetchDemographics should have been called")
}

// Additional test cases for workflow scenarios

// Test a simplified version of the workflow with success and error conditions
func TestWorkflowWithSuccessAndError(t *testing.T) {
	// Use the test environment for a simplified test
	testSuite := testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Setup activities for success case
	testActivities := new(TestActivities)

	// In the success case, FetchDemographics returns good demographics
	testActivities.FetchDemographicsFunc = func(customerID string) (data_enrichment.Demographics, error) {
		// Note: The workflow implementation in workflow.go doesn't actually pass the customerID
		// to the FetchDemographics activity (see line 83), so we can't use it to determine success/failure
		// This is a design issue in the workflow that should be fixed
		return data_enrichment.Demographics{Age: 30, Location: "Test Location"}, nil
	}

	// Other activities succeed normally
	testActivities.MergeDataFunc = func(customer data_enrichment.Customer, demographics data_enrichment.Demographics) (data_enrichment.EnrichedCustomer, error) {
		return data_enrichment.EnrichedCustomer{Customer: customer, Demographics: demographics}, nil
	}

	testActivities.StoreEnrichedDataFunc = func(enriched data_enrichment.EnrichedCustomer) error {
		return nil
	}

	// Register activities
	env.RegisterActivity(testActivities.FetchDemographics)
	env.RegisterActivity(testActivities.MergeData)
	env.RegisterActivity(testActivities.StoreEnrichedData)

	// Test with a single customer workflow directly
	testCustomer1 := data_enrichment.Customer{ID: "customer1", Name: "Success User", Email: "success@example.com"}

	// Execute the single customer workflow directly
	env.ExecuteWorkflow(data_enrichment.EnrichSingleCustomerWorkflow, testCustomer1)

	// Verify workflow completed successfully
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// Get and verify result
	var result1 data_enrichment.EnrichedCustomer
	require.NoError(t, env.GetWorkflowResult(&result1))
	assert.Equal(t, "customer1", result1.Customer.ID, "Customer ID should match")

	// Now test the error case with a new test environment
	env = testSuite.NewTestWorkflowEnvironment()

	// For the error case, configure FetchDemographics to return an error
	testActivities = new(TestActivities)
	testActivities.FetchDemographicsFunc = func(customerID string) (data_enrichment.Demographics, error) {
		return data_enrichment.Demographics{}, fmt.Errorf("simulated error in demographics service")
	}

	// Configure the other activities
	testActivities.MergeDataFunc = func(customer data_enrichment.Customer, demographics data_enrichment.Demographics) (data_enrichment.EnrichedCustomer, error) {
		return data_enrichment.EnrichedCustomer{Customer: customer, Demographics: demographics}, nil
	}

	testActivities.StoreEnrichedDataFunc = func(enriched data_enrichment.EnrichedCustomer) error {
		return nil
	}

	// Register activities with the new environment
	env.RegisterActivity(testActivities.FetchDemographics)
	env.RegisterActivity(testActivities.MergeData)
	env.RegisterActivity(testActivities.StoreEnrichedData)

	// Test with a customer that will cause an error
	testCustomer2 := data_enrichment.Customer{ID: "customer2", Name: "Error User", Email: "error@example.com"}

	// Execute the workflow
	env.ExecuteWorkflow(data_enrichment.EnrichSingleCustomerWorkflow, testCustomer2)

	// Verify workflow completed with an error
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
}

// Test the timeout handling scenario
func TestTimeoutHandlingScenario(t *testing.T) {
	// This test uses the testsuite environment to test timeout handling
	testSuite := testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Setup test activities with a simulated timeout
	testActivities := new(TestActivities)

	// Track if the activity was called
	activityCalled := false
	activityCount := 0
	activityCompleted := 0

	// Configure the activity to simulate a timeout
	testActivities.FetchDemographicsFunc = func(customerID string) (data_enrichment.Demographics, error) {
		activityCalled = true
		activityCount++
		fmt.Println("ACT_COUNT:", activityCount)
		if activityCount < 5 {
			time.Sleep(time.Second)
			// Simulate a timeout by returning a timeout error
			return data_enrichment.Demographics{}, fmt.Errorf("activity timeout")
		}
		activityCompleted++
		return data_enrichment.Demographics{Age: 45, Location: "KL Malaysia"}, nil
	}

	// Register workflows and activities
	//env.RegisterWorkflow(data_enrichment.DataEnrichmentWorkflow)
	//env.RegisterWorkflow(data_enrichment.EnrichSingleCustomerWorkflow)
	env.RegisterWorkflowWithOptions(data_enrichment.EnrichSingleCustomerWorkflow, workflow.RegisterOptions{
		//Name:                          "bobo", // Has effect if called with string; no effect if with struct
		DisableAlreadyRegisteredCheck: false,
		VersioningBehavior:            workflow.VersioningBehavior(2),
	})

	//env.RegisterActivity(testActivities.FetchDemographics)
	env.RegisterActivityWithOptions(testActivities.FetchDemographics, activity.RegisterOptions{
		Name:                          "foo", // Does not seem to have effect
		DisableAlreadyRegisteredCheck: false,
		SkipInvalidStructFunctions:    false,
	})
	env.RegisterActivity(testActivities.MergeData)
	env.RegisterActivity(testActivities.StoreEnrichedData)

	// Test customer
	testCustomer := data_enrichment.Customer{ID: "timeout-id", Name: "Timeout User", Email: "timeout@example.com"}

	// Execute the workflow
	env.ExecuteWorkflow("EnrichSingleCustomerWorkflow", testCustomer)

	// Verify workflow completed with an error
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())

	// Verify the activity was called
	require.True(t, activityCalled, "FetchDemographics should have been called")
	assert.Equal(t, 0, activityCompleted, fmt.Sprintf("Activity count should be 1 with rety %d", activityCount))
}

// Note: Integration tests with real Temporal DevServer have been moved to e2e_test.go file.
// This file now contains only mock-based tests using the testsuite.WorkflowTestSuite.
func TestActivityFailureMock(t *testing.T) {
	// This test uses the testsuite environment to test activity failures
	testSuite := testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Setup test activities with an error
	testActivities := new(TestActivities)
	activityError := fmt.Errorf("demographics service unavailable")

	// Track if the activity was called
	fetchCalled := false
	testActivities.FetchDemographicsFunc = func(customerID string) (data_enrichment.Demographics, error) {
		fetchCalled = true
		return data_enrichment.Demographics{}, activityError
	}

	// Register workflows and activities
	env.RegisterWorkflow(data_enrichment.EnrichSingleCustomerWorkflow)
	env.RegisterActivity(testActivities.FetchDemographics)
	env.RegisterActivity(testActivities.MergeData)
	env.RegisterActivity(testActivities.StoreEnrichedData)

	// Test customer
	testCustomer := data_enrichment.Customer{ID: "error-id", Name: "Error User", Email: "error@example.com"}

	// Execute the workflow
	env.ExecuteWorkflow(data_enrichment.EnrichSingleCustomerWorkflow, testCustomer)

	// Verify workflow completed with an error
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())

	// Verify the activity was called
	require.True(t, fetchCalled, "FetchDemographics should have been called")
}

func TestPartialFailureMock(t *testing.T) {
	// This test uses the testsuite environment to test partial batch failures
	testSuite := testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Setup test activities
	testActivities := new(TestActivities)

	// Setup test data with multiple customers
	testCustomers := []data_enrichment.Customer{
		{ID: "success-1", Name: "Success One", Email: "success1@example.com"},
		{ID: "error-1", Name: "Error One", Email: "error1@example.com"},
		{ID: "success-2", Name: "Success Two", Email: "success2@example.com"},
	}

	// Track processed customers
	processedCustomers := make(map[string]bool)
	mergeFailedCustomers := make(map[string]bool)

	// Configure test activities
	testActivities.FetchDemographicsFunc = func(customerID string) (data_enrichment.Demographics, error) {
		processedCustomers[customerID] = true
		return data_enrichment.Demographics{Age: 35, Location: "Austin, TX"}, nil
	}

	testActivities.MergeDataFunc = func(customer data_enrichment.Customer, demographics data_enrichment.Demographics) (data_enrichment.EnrichedCustomer, error) {
		if customer.ID == "error-1" {
			mergeFailedCustomers[customer.ID] = true
			return data_enrichment.EnrichedCustomer{}, fmt.Errorf("data merge failed for customer %s", customer.ID)
		}
		return data_enrichment.EnrichedCustomer{Customer: customer, Demographics: demographics}, nil
	}

	testActivities.StoreEnrichedDataFunc = func(enriched data_enrichment.EnrichedCustomer) error {
		return nil
	}

	// Register workflows and activities
	env.RegisterWorkflow(data_enrichment.DataEnrichmentWorkflow)
	env.RegisterWorkflow(data_enrichment.EnrichSingleCustomerWorkflow)
	env.RegisterActivity(testActivities.FetchDemographics)
	env.RegisterActivity(testActivities.MergeData)
	env.RegisterActivity(testActivities.StoreEnrichedData)

	// Execute the workflow
	env.ExecuteWorkflow(data_enrichment.DataEnrichmentWorkflow, testCustomers)

	// Verify workflow completed
	require.True(t, env.IsWorkflowCompleted())

	// Depending on how the workflow handles errors, we might or might not have a workflow error
	if env.GetWorkflowError() == nil {
		// If no workflow error, we should have partial results
		var results []data_enrichment.EnrichedCustomer
		env.GetWorkflowResult(&results)

		// Should have results for the successful customers only
		expectedSuccessful := 0
		for _, c := range testCustomers {
			if c.ID != "error-1" {
				expectedSuccessful++
			}
		}
		assert.Equal(t, expectedSuccessful, len(results), "Should have results only for successful customers")
	}
}

func TestDataEnrichmentWorkflow(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	// Register activities
	env.RegisterActivity(data_enrichment.EnrichData)

	// Execute workflow
	env.ExecuteWorkflow(data_enrichment.DataEnrichmentWorkflow, data_enrichment.DataEnrichmentWorkflowInput{
		DataID: "test-123",
	})

	// Verify workflow completed
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// Verify workflow result
	var result *data_enrichment.DataEnrichmentWorkflowResult
	require.NoError(t, env.GetWorkflowResult(&result))
	require.True(t, result.Success)
	require.Equal(t, "Data enrichment completed successfully", result.Message)
}
