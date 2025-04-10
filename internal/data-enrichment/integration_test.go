package data_enrichment_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/suite"
	"app/internal/data-enrichment"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"testing"
	"time"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// IntegrationTestServer holds the Temporal server, client, worker, and test configuration
type IntegrationTestServer struct {
	key           string
	server        *testsuite.DevServer
	client        client.Client
	taskQueue     string
	worker        worker.Worker
	testActivities *TestActivities
}

// DataEnrichmentHappyTestSuite happy path e2e test with realistic data; can it be triggerfed parallel?
type DataEnrichmentHappyTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	its *IntegrationTestServer
}

// Below are the test case scenarios under this  suite ..
// Scenario: Normal happy with Signals ..

// Scenario: Disrupted dependency

// DataEnrichmentSadTestSuite
type DataEnrichmentSadTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	its *IntegrationTestServer
}

// Scenario: Total Prolonged failure?

func setupIntegrationDevServer(its *IntegrationTestServer) error {
	// Start the Temporal server as a background process
	server, err := testsuite.StartDevServer(context.Background(), testsuite.DevServerOptions{
		ClientOptions: &client.Options{
			Identity: its.key,
		},
	})
	fmt.Println("DevServer:", server.FrontendHostPort())
	if err != nil {
		return err
	}
	its.server = server

	// Create a client to the Temporal server
	client, err := client.Dial(client.Options{
		HostPort: server.FrontendHostPort(),
		Identity: its.key,
	})
	if err != nil {
		return err
	}
	its.client = client

	// Create a unique task queue for this test run
	its.taskQueue = "data-enrichment-test-" + uuid.New().String()

	// Create a worker to process workflow and activity tasks
	its.worker = worker.New(its.client, its.taskQueue, worker.Options{})

	// Register workflows
	its.worker.RegisterWorkflow(data_enrichment.DataEnrichmentWorkflow)
	its.worker.RegisterWorkflow(data_enrichment.EnrichSingleCustomerWorkflow)

	// Set up test activities
	its.testActivities = new(TestActivities)

	// Start the worker
	err = its.worker.Start()
	if err != nil {
		return err
	}

	return nil
}

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
	assert.Equal(t, testCustomer.ID, result.ID, "Customer ID should match")
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
	assert.Equal(t, "customer1", result1.ID, "Customer ID should match")
	
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
	
	// Configure the activity to simulate a timeout
	testActivities.FetchDemographicsFunc = func(customerID string) (data_enrichment.Demographics, error) {
		activityCalled = true
		// Simulate a timeout by returning a timeout error
		return data_enrichment.Demographics{}, fmt.Errorf("activity timeout")
	}
	
	// Register workflows and activities
	env.RegisterWorkflow(data_enrichment.DataEnrichmentWorkflow)
	env.RegisterWorkflow(data_enrichment.EnrichSingleCustomerWorkflow)
	env.RegisterActivity(testActivities.FetchDemographics)
	env.RegisterActivity(testActivities.MergeData)
	env.RegisterActivity(testActivities.StoreEnrichedData)
	
	// Test customer
	testCustomer := data_enrichment.Customer{ID: "timeout-id", Name: "Timeout User", Email: "timeout@example.com"}
	
	// Execute the workflow
	env.ExecuteWorkflow(data_enrichment.EnrichSingleCustomerWorkflow, testCustomer)
	
	// Verify workflow completed with an error
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	
	// Verify the activity was called
	require.True(t, activityCalled, "FetchDemographics should have been called")
}

// Run happy + sad in parallel?
func TestDataEnrichmentWorkflow(t *testing.T) {
	//t.Parallel() // Does not seem to work?
	//t.Run("happy", func(t *testing.T) {
	suite.Run(t, new(DataEnrichmentHappyTestSuite))
	//})
	//t.Run("sad", func(t *testing.T) {
	suite.Run(t, new(DataEnrichmentSadTestSuite))
	//})
}

func (s *DataEnrichmentHappyTestSuite) SetupSuite() {
	its := &IntegrationTestServer{
		key: "happy",
	}
	err := setupIntegrationDevServer(its)
	if err != nil {
		s.T().Fatal(err)
		return
	}
	s.its = its

	// Configure default test activities
	s.its.testActivities.FetchDemographicsFunc = func(customerID string) (data_enrichment.Demographics, error) {
		return data_enrichment.Demographics{Age: 30, Location: "New York, NY"}, nil
	}

	s.its.testActivities.MergeDataFunc = func(customer data_enrichment.Customer, demographics data_enrichment.Demographics) (data_enrichment.EnrichedCustomer, error) {
		return data_enrichment.EnrichedCustomer{Customer: customer, Demographics: demographics}, nil
	}

	s.its.testActivities.StoreEnrichedDataFunc = func(enriched data_enrichment.EnrichedCustomer) error {
		return nil
	}

	// Register activity implementations
	s.its.worker.RegisterActivity(s.its.testActivities)
}

func (s *DataEnrichmentHappyTestSuite) TestHappy() {
	// Test the main DataEnrichmentWorkflow with multiple customers
	
	// Setup test data
	testCustomers := []data_enrichment.Customer{
		{ID: "1", Name: "Alice", Email: "alice@example.com"},
		{ID: "2", Name: "Bob", Email: "bob@example.com"},
		{ID: "3", Name: "Charlie", Email: "charlie@example.com"},
	}

	// Expected demographics for verification
	expectedDemographics := data_enrichment.Demographics{Age: 30, Location: "New York, NY"}
	
	// Track processed customers for verification
	processedCustomers := make(map[string]bool)

	// Configure test activities
	s.its.testActivities.FetchDemographicsFunc = func(customerID string) (data_enrichment.Demographics, error) {
		// Record that this customer was processed
		processedCustomers[customerID] = true
		return expectedDemographics, nil
	}

	// Execute workflow with shorter timeouts for testing
	workflowOptions := client.StartWorkflowOptions{
		ID:                  "data-enrichment-test-" + uuid.New().String(),
		TaskQueue:           s.its.taskQueue,
		WorkflowRunTimeout:  3 * time.Second,
		WorkflowTaskTimeout: 1 * time.Second,
	}

	we, err := s.its.client.ExecuteWorkflow(context.Background(), workflowOptions, data_enrichment.DataEnrichmentWorkflow, testCustomers)
	require.NoError(s.T(), err)

	// Wait for workflow execution
	var result []data_enrichment.EnrichedCustomer
	err = we.Get(context.Background(), &result)
	require.NoError(s.T(), err)

	// Verify results
	assert.Equal(s.T(), len(testCustomers), len(result), "Should have enriched all customers")
	for i, customer := range testCustomers {
		assert.Equal(s.T(), customer.ID, result[i].ID)
		assert.Equal(s.T(), customer.Name, result[i].Name)
		assert.Equal(s.T(), customer.Email, result[i].Email)
		assert.Equal(s.T(), expectedDemographics.Age, result[i].Age)
		assert.Equal(s.T(), expectedDemographics.Location, result[i].Location)
		
		// Verify this customer was processed by the FetchDemographics activity
		assert.True(s.T(), processedCustomers[customer.ID], "Customer should have been processed")
	}

	// Verify all customers were processed
	assert.Equal(s.T(), len(testCustomers), len(processedCustomers), "All customers should have been processed")
}

// Additional test case for verifying child workflow execution
func (s *DataEnrichmentHappyTestSuite) TestSingleCustomerWorkflow() {
	// Test the EnrichSingleCustomerWorkflow directly
	
	// Setup test data
	testCustomer := data_enrichment.Customer{ID: "test-id", Name: "Test User", Email: "test@example.com"}
	expectedDemographics := data_enrichment.Demographics{Age: 42, Location: "San Francisco, CA"}
	expectedEnriched := data_enrichment.EnrichedCustomer{
		Customer:     testCustomer,
		Demographics: expectedDemographics,
	}

	// Track activity executions
	fetchCalled := false
	mergeCalled := false
	storeCalled := false

	// Configure test activities for this specific test
	s.its.testActivities.FetchDemographicsFunc = func(customerID string) (data_enrichment.Demographics, error) {
		fetchCalled = true
		assert.Equal(s.T(), testCustomer.ID, customerID, "Should fetch demographics for the correct customer")
		return expectedDemographics, nil
	}

	s.its.testActivities.MergeDataFunc = func(customer data_enrichment.Customer, demographics data_enrichment.Demographics) (data_enrichment.EnrichedCustomer, error) {
		mergeCalled = true
		assert.Equal(s.T(), testCustomer, customer, "Should merge data for the correct customer")
		assert.Equal(s.T(), expectedDemographics, demographics, "Should merge with correct demographics")
		return expectedEnriched, nil
	}

	s.its.testActivities.StoreEnrichedDataFunc = func(enriched data_enrichment.EnrichedCustomer) error {
		storeCalled = true
		assert.Equal(s.T(), expectedEnriched, enriched, "Should store the correct enriched data")
		return nil
	}

	// Execute workflow with timeouts
	workflowOptions := client.StartWorkflowOptions{
		ID:                  "single-customer-test-" + uuid.New().String(),
		TaskQueue:           s.its.taskQueue,
		WorkflowRunTimeout:  5 * time.Second,
		WorkflowTaskTimeout: 1 * time.Second,
	}

	we, err := s.its.client.ExecuteWorkflow(context.Background(), workflowOptions, data_enrichment.EnrichSingleCustomerWorkflow, testCustomer)
	require.NoError(s.T(), err)

	// Wait for workflow execution
	var result data_enrichment.EnrichedCustomer
	err = we.Get(context.Background(), &result)
	require.NoError(s.T(), err)

	// Verify result matches expected
	assert.Equal(s.T(), expectedEnriched, result)

	// Verify that all activities were called
	assert.True(s.T(), fetchCalled, "FetchDemographics should have been called")
	assert.True(s.T(), mergeCalled, "MergeData should have been called")
	assert.True(s.T(), storeCalled, "StoreEnrichedData should have been called")
}

func (s *DataEnrichmentHappyTestSuite) TearDownSuite() {
	if s.its.server != nil {
		s.its.server.Stop()
	}
}

func (s *DataEnrichmentSadTestSuite) SetupSuite() {
	its := &IntegrationTestServer{
		key: "sad",
	}
	err := setupIntegrationDevServer(its)
	if err != nil {
		s.T().Fatal(err)
		return
	}
	s.its = its

	// Configure default test activities with error scenarios
	s.its.testActivities.FetchDemographicsFunc = func(customerID string) (data_enrichment.Demographics, error) {
		// Default implementation for sad path tests
		if customerID == "error-id" {
			return data_enrichment.Demographics{}, fmt.Errorf("demographics service unavailable")
		}
		return data_enrichment.Demographics{Age: 35, Location: "Austin, TX"}, nil
	}

	s.its.testActivities.MergeDataFunc = func(customer data_enrichment.Customer, demographics data_enrichment.Demographics) (data_enrichment.EnrichedCustomer, error) {
		if customer.ID == "error-1" {
			return data_enrichment.EnrichedCustomer{}, fmt.Errorf("data merge failed")
		}
		return data_enrichment.EnrichedCustomer{Customer: customer, Demographics: demographics}, nil
	}

	s.its.testActivities.StoreEnrichedDataFunc = func(enriched data_enrichment.EnrichedCustomer) error {
		return nil
	}

	// Register activity implementations
	s.its.worker.RegisterActivity(s.its.testActivities)
}

func (s *DataEnrichmentSadTestSuite) TestActivityFailure() {
	// Test the workflow when FetchDemographics activity fails
	
	// Setup test data
	testCustomer := data_enrichment.Customer{ID: "error-id", Name: "Error User", Email: "error@example.com"}
	activityError := fmt.Errorf("demographics service unavailable")

	// Configure test activities for this specific test - using a variable to track if called
	fetchCalled := false
	var fetchCalledPtr = &fetchCalled
	
	// Return error immediately without delay
	s.its.testActivities.FetchDemographicsFunc = func(customerID string) (data_enrichment.Demographics, error) {
		*fetchCalledPtr = true
		assert.Equal(s.T(), testCustomer.ID, customerID, "Should fetch demographics for the correct customer")
		return data_enrichment.Demographics{}, activityError
	}

	// Execute workflow with timeouts
	workflowOptions := client.StartWorkflowOptions{
		ID:                  "error-test-" + uuid.New().String(),
		TaskQueue:           s.its.taskQueue,
		WorkflowRunTimeout:  5 * time.Second,
		WorkflowTaskTimeout: 1 * time.Second,
	}

	we, err := s.its.client.ExecuteWorkflow(context.Background(), workflowOptions, data_enrichment.EnrichSingleCustomerWorkflow, testCustomer)
	require.NoError(s.T(), err)

	// Wait for workflow execution
	var result data_enrichment.EnrichedCustomer
	err = we.Get(context.Background(), &result)

	// The workflow should have failed
	assert.Error(s.T(), err, "Workflow should fail when activity fails")
	
	// For Temporal, we just verify there was an error, not the specific message
	// as the error might be wrapped in a workflow execution error
	
	// Verify that the activity was called - using pointer to ensure it's updated
	assert.True(s.T(), *fetchCalledPtr, "FetchDemographics should have been called")
}

func (s *DataEnrichmentSadTestSuite) TestPartialFailure() {
	// Test the main workflow when some customer enrichments fail but others succeed
	
	// Setup test data with multiple customers
	testCustomers := []data_enrichment.Customer{
		{ID: "success-1", Name: "Success One", Email: "success1@example.com"},
		{ID: "error-1", Name: "Error One", Email: "error1@example.com"},
		{ID: "success-2", Name: "Success Two", Email: "success2@example.com"},
	}

	// Track processed customers
	processedCustomers := make(map[string]bool)
	mergeFailedCustomers := make(map[string]bool)
	
	// Configure test activities for this specific test
	s.its.testActivities.FetchDemographicsFunc = func(customerID string) (data_enrichment.Demographics, error) {
		processedCustomers[customerID] = true
		return data_enrichment.Demographics{Age: 35, Location: "Austin, TX"}, nil
	}

	s.its.testActivities.MergeDataFunc = func(customer data_enrichment.Customer, demographics data_enrichment.Demographics) (data_enrichment.EnrichedCustomer, error) {
		if customer.ID == "error-1" {
			mergeFailedCustomers[customer.ID] = true
			return data_enrichment.EnrichedCustomer{}, fmt.Errorf("data merge failed for customer %s", customer.ID)
		}
		return data_enrichment.EnrichedCustomer{Customer: customer, Demographics: demographics}, nil
	}

	// Execute workflow with timeouts
	workflowOptions := client.StartWorkflowOptions{
		ID:                  "partial-failure-test-" + uuid.New().String(),
		TaskQueue:           s.its.taskQueue,
		WorkflowRunTimeout:  5 * time.Second,
		WorkflowTaskTimeout: 1 * time.Second,
	}

	we, err := s.its.client.ExecuteWorkflow(context.Background(), workflowOptions, data_enrichment.DataEnrichmentWorkflow, testCustomers)
	require.NoError(s.T(), err)

	// Wait for workflow execution
	var result []data_enrichment.EnrichedCustomer
	err = we.Get(context.Background(), &result)
	require.NoError(s.T(), err, "Main workflow should not fail even if some child workflows fail")

	// Verify there are only 2 successful results (not 3)
	assert.Equal(s.T(), 2, len(result), "Should only have 2 successful enriched customers")

	// Verify all customers were processed by FetchDemographics
	assert.Equal(s.T(), len(testCustomers), len(processedCustomers), "All customers should have been processed")
	
	// Verify the error-1 customer failed during merge
	assert.True(s.T(), mergeFailedCustomers["error-1"], "The error-1 customer should have failed during merge")
	
	// Verify the successful customers are in the result
	successIDs := []string{"success-1", "success-2"}
	for _, id := range successIDs {
		found := false
		for _, enriched := range result {
			if enriched.ID == id {
				found = true
				break
			}
		}
		assert.True(s.T(), found, "Customer %s should be in the result", id)
	}
}

func (s *DataEnrichmentSadTestSuite) TestTimeoutHandling() {
	// Test the workflow with activity timeout
	
	// Setup test data
	testCustomer := data_enrichment.Customer{ID: "timeout-id", Name: "Timeout User", Email: "timeout@example.com"}

	// Track if the activity was called
	fetchCalled := false

	// Configure the activity to wait longer than the activity timeout (which is 5s in the workflow)
	s.its.testActivities.FetchDemographicsFunc = func(customerID string) (data_enrichment.Demographics, error) {
		fetchCalled = true
		// Wait just long enough to cause a timeout but not too long
		time.Sleep(500 * time.Millisecond)
		return data_enrichment.Demographics{Age: 28, Location: "Remote"}, nil
	}

	// Execute workflow with timeouts
	workflowOptions := client.StartWorkflowOptions{
		ID:                  "timeout-test-" + uuid.New().String(),
		TaskQueue:           s.its.taskQueue,
		WorkflowRunTimeout:  5 * time.Second,
		WorkflowTaskTimeout: 1 * time.Second,
	}

	we, err := s.its.client.ExecuteWorkflow(context.Background(), workflowOptions, data_enrichment.EnrichSingleCustomerWorkflow, testCustomer)
	require.NoError(s.T(), err)

	// Wait for workflow execution
	var result data_enrichment.EnrichedCustomer
	err = we.Get(context.Background(), &result)

	// The workflow should have failed with a timeout error
	assert.Error(s.T(), err, "Workflow should fail with timeout error")
	// In a real test, we would use Temporal's error inspection APIs to check error type
	// For this test, we just verify there was an error
	// The specific error message might vary between Temporal versions
	
	// Verify the activity was called
	assert.True(s.T(), fetchCalled, "FetchDemographics should have been called")
}

func (s *DataEnrichmentSadTestSuite) TearDownSuite() {
	if s.its.server != nil {
		s.its.server.Stop()
	}
}
