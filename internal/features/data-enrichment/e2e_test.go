package data_enrichment

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

// E2ETestActivities implements test activities for E2E tests with various behaviors
type E2ETestActivities struct {
	// For tracking activity executions
	FetchCount   int
	MergeCount   int
	StoreCount   int
	ProcessedIDs map[string]bool

	// For controlling test behavior
	ShouldFailFetch  bool
	FailOnCustomerID string
	SimulateDelay    bool
}

// NewE2ETestActivities creates a new instance of E2E test activities
func NewE2ETestActivities() *E2ETestActivities {
	return &E2ETestActivities{
		ProcessedIDs: make(map[string]bool),
	}
}

// FetchDemographics returns demographics data for the customer
func (a *E2ETestActivities) FetchDemographics(customerID string) (Demographics, error) {
	a.FetchCount++
	fmt.Println("FetchDemographics activity called with ID:", customerID)

	// Mark this customer ID as processed
	a.ProcessedIDs[customerID] = true

	// Simulate a delay if configured to do so
	if a.SimulateDelay {
		time.Sleep(100 * time.Millisecond)
	}

	// If configured to fail for all fetches
	if a.ShouldFailFetch {
		return Demographics{}, fmt.Errorf("simulated demographics service error")
	}

	// If configured to fail for a specific customer ID
	if customerID == a.FailOnCustomerID {
		return Demographics{}, fmt.Errorf("simulated error for customer: %s", customerID)
	}

	// Return test data
	return Demographics{Age: 30, Location: "Test Location"}, nil
}

// MergeData combines customer and demographics data
func (a *E2ETestActivities) MergeData(customer Customer, demographics Demographics) (EnrichedCustomer, error) {
	a.MergeCount++
	fmt.Println("MergeData activity called with customer ID:", customer.ID)

	// Simulate a delay if configured to do so
	if a.SimulateDelay {
		time.Sleep(100 * time.Millisecond)
	}

	// If configured to fail for a specific customer ID
	if a.FailOnCustomerID != "" && customer.ID == a.FailOnCustomerID {
		fmt.Printf("Simulating MergeData failure for customer: %s\n", customer.ID)
		return EnrichedCustomer{}, fmt.Errorf("simulated merge error for customer: %s", customer.ID)
	}

	// Create the enriched customer
	return EnrichedCustomer{
		Customer:     customer,
		Demographics: demographics,
	}, nil
}

// StoreEnrichedData stores the enriched customer data
func (a *E2ETestActivities) StoreEnrichedData(enriched EnrichedCustomer) error {
	a.StoreCount++
	fmt.Println("StoreEnrichedData activity called with customer ID:", enriched.ID)

	// Simulate a delay if configured to do so
	if a.SimulateDelay {
		time.Sleep(100 * time.Millisecond)
	}

	// If configured to fail for a specific customer ID
	if a.FailOnCustomerID != "" && enriched.ID == a.FailOnCustomerID {
		fmt.Printf("Simulating StoreEnrichedData failure for customer: %s\n", enriched.ID)
		return fmt.Errorf("simulated storage error for customer: %s", enriched.ID)
	}

	return nil
}

// setupE2ETestEnvironment creates a test environment with Temporal DevServer
func setupE2ETestEnvironment(t *testing.T) (*testsuite.DevServer, client.Client, string, *E2ETestActivities, func()) {
	// Start the Temporal dev server
	server, err := testsuite.StartDevServer(context.Background(), testsuite.DevServerOptions{})
	require.NoError(t, err, "Failed to start Temporal dev server")

	// Create client
	tempClient := server.Client()

	// Create a unique task queue for this test
	taskQueue := fmt.Sprintf("e2e-test-queue-%s", uuid.New().String())
	fmt.Println("Using task queue:", taskQueue)

	// Create the test activities
	testActivities := NewE2ETestActivities()

	// Create worker
	w := worker.New(tempClient, taskQueue, worker.Options{})

	// Register workflows and activities
	w.RegisterWorkflow(DataEnrichmentWorkflow)
	w.RegisterWorkflow(EnrichSingleCustomerWorkflow)
	w.RegisterActivity(testActivities.FetchDemographics)
	w.RegisterActivity(testActivities.MergeData)
	w.RegisterActivity(testActivities.StoreEnrichedData)

	// Start the worker
	go func() {
		if err := w.Run(worker.InterruptCh()); err != nil {
			t.Logf("Worker error: %v", err)
		}
	}()

	// Return a cleanup function
	cleanup := func() {
		w.Stop()
		tempClient.Close()
		server.Stop()
	}

	return server, tempClient, taskQueue, testActivities, cleanup
}

// TestE2EDataEnrichmentSingleCustomer runs a simple end-to-end test for the single customer workflow
func TestE2EDataEnrichmentSingleCustomer(t *testing.T) {
	fmt.Println("Starting E2E test for single customer enrichment...")

	// Setup test environment
	_, tempClient, taskQueue, testActivities, cleanup := setupE2ETestEnvironment(t)
	defer cleanup()

	// Test data
	testCustomer := Customer{
		ID:    "test-1",
		Name:  "Test User",
		Email: "test@example.com",
	}

	// Create a context with sufficient timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set up workflow options with sufficient timeouts
	workflowOptions := client.StartWorkflowOptions{
		ID:                  "e2e-single-" + uuid.New().String(),
		TaskQueue:           taskQueue,
		WorkflowRunTimeout:  20 * time.Second,
		WorkflowTaskTimeout: 5 * time.Second,
	}

	// Execute the workflow
	we, err := tempClient.ExecuteWorkflow(ctx, workflowOptions, EnrichSingleCustomerWorkflow, testCustomer)
	require.NoError(t, err, "Failed to start workflow")
	fmt.Println("Started workflow execution. WorkflowID:", we.GetID())

	// Wait for workflow completion
	var result EnrichedCustomer
	err = we.Get(ctx, &result)
	require.NoError(t, err, "Workflow execution failed")

	// Verify the result
	assert.Equal(t, testCustomer.ID, result.ID, "Customer ID should match")
	assert.Equal(t, testCustomer.Name, result.Name, "Customer name should match")
	assert.Equal(t, testCustomer.Email, result.Email, "Customer email should match")
	assert.Equal(t, 30, result.Age, "Demographics age should be set")
	assert.Equal(t, "Test Location", result.Location, "Demographics location should be set")

	// Verify activity executions
	assert.Equal(t, 1, testActivities.FetchCount, "FetchDemographics should be called once")
	assert.Equal(t, 1, testActivities.MergeCount, "MergeData should be called once")
	assert.Equal(t, 1, testActivities.StoreCount, "StoreEnrichedData should be called once")

	fmt.Println("✅ Single customer enrichment test passed!")
}

// TestE2EDataEnrichmentErrorHandling tests error handling in the workflow
func TestE2EDataEnrichmentErrorHandling(t *testing.T) {
	fmt.Println("Starting E2E test for error handling...")

	// Setup test environment
	_, tempClient, taskQueue, testActivities, cleanup := setupE2ETestEnvironment(t)
	defer cleanup()

	// For error handling, use a single customer that will fail during FetchDemographics
	customer := Customer{
		ID:    "error-customer",
		Name:  "Error Test",
		Email: "error@example.com",
	}

	// Configure activity to fail for all fetch operations
	testActivities.ShouldFailFetch = true

	// Create a context with sufficient timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set up workflow options
	workflowOptions := client.StartWorkflowOptions{
		ID:                  "e2e-error-" + uuid.New().String(),
		TaskQueue:           taskQueue,
		WorkflowRunTimeout:  20 * time.Second,
		WorkflowTaskTimeout: 5 * time.Second,
	}

	// Execute workflow expected to fail
	we, err := tempClient.ExecuteWorkflow(ctx, workflowOptions, EnrichSingleCustomerWorkflow, customer)
	require.NoError(t, err, "Failed to start workflow")
	fmt.Println("Started error test workflow. WorkflowID:", we.GetID())

	// Wait for workflow completion - we expect an error
	var result EnrichedCustomer
	err = we.Get(ctx, &result)

	// Verify we got an error from the activity
	require.Error(t, err, "Expected an error from the workflow due to simulated activity failure")
	assert.Contains(t, err.Error(), "simulated demographics service error", "Error should contain reference to the simulated error")

	// Verify the activity was called with appropriate retry attempts
	// The exact count depends on the retry policy, but should be greater than 1 due to retries
	assert.GreaterOrEqual(t, testActivities.FetchCount, 1, "FetchDemographics should be called at least once")
	assert.Equal(t, 0, testActivities.MergeCount, "MergeData should not be called after FetchDemographics fails")
	assert.Equal(t, 0, testActivities.StoreCount, "StoreEnrichedData should not be called after FetchDemographics fails")

	fmt.Println("✅ Error handling test passed with expected failure: ", err)
}

// TestE2EDataEnrichmentIdempotency verifies workflow idempotency behavior
func TestE2EDataEnrichmentIdempotency(t *testing.T) {
	fmt.Println("Starting E2E test for workflow idempotency...")

	// Setup test environment
	_, tempClient, taskQueue, testActivities, cleanup := setupE2ETestEnvironment(t)
	defer cleanup()

	// Test data
	testCustomer := Customer{
		ID:    "idempotent-test",
		Name:  "Idempotent User",
		Email: "idempotent@example.com",
	}

	// Create a context with sufficient timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a consistent workflow ID to test idempotency
	workflowID := "e2e-idempotent-" + uuid.New().String()

	// Set up workflow options with idempotency policy
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID, // Same ID for all executions
		TaskQueue:             taskQueue,
		WorkflowRunTimeout:    20 * time.Second,
		WorkflowTaskTimeout:   5 * time.Second,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	// Execute the workflow for the first time
	we1, err := tempClient.ExecuteWorkflow(ctx, workflowOptions, EnrichSingleCustomerWorkflow, testCustomer)
	require.NoError(t, err, "Failed to start first workflow")
	fmt.Println("Started first workflow. WorkflowID:", workflowID, "RunID:", we1.GetRunID())

	// Store the run ID for comparison
	firstRunID := we1.GetRunID()

	// Wait for the first workflow to complete
	var result1 EnrichedCustomer
	err = we1.Get(ctx, &result1)
	require.NoError(t, err, "First workflow execution failed")

	// Record the activity counts after first execution
	firstFetchCount := testActivities.FetchCount
	firstMergeCount := testActivities.MergeCount
	firstStoreCount := testActivities.StoreCount

	// Now try to execute the workflow with the same ID again
	we2, err := tempClient.ExecuteWorkflow(ctx, workflowOptions, EnrichSingleCustomerWorkflow, testCustomer)

	// We expect either:
	// 1. An error because the workflow ID already exists with REJECT_DUPLICATE policy
	// 2. OR a successful execution with the SAME run ID as the first workflow
	if err != nil {
		// This is the expected behavior with WorkflowIDReusePolicyRejectDuplicate
		fmt.Println("✅ Second workflow execution correctly rejected with error:", err)
	} else {
		// If we didn't get an error, we should have the same run ID
		secondRunID := we2.GetRunID()
		assert.Equal(t, firstRunID, secondRunID, "Run IDs should match for idempotent executions")

		// Get the result from the second execution
		var result2 EnrichedCustomer
		err = we2.Get(ctx, &result2)
		require.NoError(t, err, "Second workflow execution failed")

		// Verify both results are identical
		assert.Equal(t, result1, result2, "Results should be identical for idempotent executions")
	}

	// Verify activity counts haven't increased (no new activities were executed)
	assert.Equal(t, firstFetchCount, testActivities.FetchCount, "FetchDemographics should not be called again")
	assert.Equal(t, firstMergeCount, testActivities.MergeCount, "MergeData should not be called again")
	assert.Equal(t, firstStoreCount, testActivities.StoreCount, "StoreEnrichedData should not be called again")

	fmt.Println("✅ Idempotency test passed!")
}

// TestE2E aliases the single customer test to maintain backward compatibility
func TestE2E(t *testing.T) {
	TestE2EDataEnrichmentSingleCustomer(t)
}

// TestAllE2E runs all E2E tests in a specific order
func TestAllE2E(t *testing.T) {
	t.Run("Single", TestE2EDataEnrichmentSingleCustomer)
	t.Run("ErrorHandling", TestE2EDataEnrichmentErrorHandling)
	t.Run("Idempotency", TestE2EDataEnrichmentIdempotency)
}

// TestDataEnrichmentE2E tests the data enrichment workflow
func TestDataEnrichmentE2E(t *testing.T) {
	// Setup test environment
	_, tempClient, taskQueue, testActivities, cleanup := setupE2ETestEnvironment(t)
	defer cleanup()

	// Test data
	testCustomer := Customer{
		ID:    "test-1",
		Name:  "Test User",
		Email: "test@example.com",
	}

	// Create a context with sufficient timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set up workflow options with sufficient timeouts
	workflowOptions := client.StartWorkflowOptions{
		ID:                  "e2e-single-" + uuid.New().String(),
		TaskQueue:           taskQueue,
		WorkflowRunTimeout:  20 * time.Second,
		WorkflowTaskTimeout: 5 * time.Second,
	}

	// Execute workflow
	we, err := tempClient.ExecuteWorkflow(ctx, workflowOptions, EnrichSingleCustomerWorkflow, testCustomer)
	require.NoError(t, err, "Failed to start workflow")
	fmt.Println("Started workflow execution. WorkflowID:", we.GetID())

	// Wait for workflow completion
	var result EnrichedCustomer
	err = we.Get(ctx, &result)
	require.NoError(t, err, "Workflow execution failed")

	// Verify result
	require.Equal(t, testCustomer.ID, result.ID)
	require.Equal(t, testCustomer.Name, result.Name)
	require.Equal(t, testCustomer.Email, result.Email)
	require.Equal(t, 30, result.Age)
	require.Equal(t, "Test Location", result.Location)

	// Verify activity executions
	assert.Equal(t, 1, testActivities.FetchCount, "FetchDemographics should be called once")
	assert.Equal(t, 1, testActivities.MergeCount, "MergeData should be called once")
	assert.Equal(t, 1, testActivities.StoreCount, "StoreEnrichedData should be called once")
}

// TestE2EDataEnrichmentWorkflow tests the data enrichment workflow end-to-end
func TestE2EDataEnrichmentWorkflow(t *testing.T) {
	// Setup test environment
	_, tempClient, taskQueue, testActivities, cleanup := setupE2ETestEnvironment(t)
	defer cleanup()

	// Test data
	testCustomers := []Customer{
		{
			ID:    "test-1",
			Name:  "Test User",
			Email: "test@example.com",
		},
	}

	// Create a context with sufficient timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set up workflow options with sufficient timeouts
	workflowOptions := client.StartWorkflowOptions{
		ID:                  "e2e-batch-" + uuid.New().String(),
		TaskQueue:           taskQueue,
		WorkflowRunTimeout:  20 * time.Second,
		WorkflowTaskTimeout: 5 * time.Second,
	}

	// Execute workflow
	we, err := tempClient.ExecuteWorkflow(ctx, workflowOptions, DataEnrichmentWorkflow, testCustomers)
	require.NoError(t, err, "Failed to start workflow")
	fmt.Println("Started workflow execution. WorkflowID:", we.GetID())

	// Wait for workflow completion
	var result []EnrichedCustomer
	err = we.Get(ctx, &result)
	require.NoError(t, err, "Workflow execution failed")

	// Verify workflow result
	require.Len(t, result, 1)
	require.Equal(t, testCustomers[0].ID, result[0].ID)
	require.Equal(t, 30, result[0].Age)
	require.Equal(t, "Test Location", result[0].Location)

	// Verify activity executions
	assert.Equal(t, 1, testActivities.FetchCount, "FetchDemographics should be called once")
	assert.Equal(t, 1, testActivities.MergeCount, "MergeData should be called once")
	assert.Equal(t, 1, testActivities.StoreCount, "StoreEnrichedData should be called once")
}
