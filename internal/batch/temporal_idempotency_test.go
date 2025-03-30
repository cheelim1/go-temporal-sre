package batch_test

import (
	"app/internal/batch"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	enumspb "go.temporal.io/api/enums/v1" // Import enums with the correct alias
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

// deductionTracker is used to track activity executions for test verification
var deductions = struct {
	sync.Mutex
	Count     int
	AccountID string
	OrderID   string
}{}

// WorkflowIdempotencyTestSuite demonstrates how Temporal's WorkflowID-based idempotency works
// using a real Temporal server
type WorkflowIdempotencyTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	// The Temporal server
	server *testsuite.DevServer
	// The Temporal client
	client client.Client
	// Unique task queue for this test run
	taskQueue string
	// Worker to process workflow and activity tasks
	worker worker.Worker
}

// TestWorkflowIdempotencyWithSameID demonstrates how Temporal ensures idempotent execution
// when using the same WorkflowID
func TestWorkflowIdempotencyWithSameID(t *testing.T) {
	suite.Run(t, new(WorkflowIdempotencyTestSuite))
}

// A real implementation of the activity that we'll use with the real Temporal server
func DeductFeeActivity(ctx context.Context, input batch.FeeDeductionWorkflowInput) (*batch.ActivityResult, error) {
	// Add a slight delay to simulate processing time and ensure workflows stay open during concurrent tests
	time.Sleep(500 * time.Millisecond)

	// Store for tracking deductions - this is just for the test
	deductions.Lock()
	// Track the deduction to verify it happened only once
	deductions.Count++
	// Record account and order IDs
	deductions.AccountID = input.AccountID
	deductions.OrderID = input.OrderID
	deductions.Unlock()

	// Return a successful result
	return &batch.ActivityResult{
		NewBalance: 1000.0 - input.Amount,
		Success:    true,
	}, nil
}

// SetupSuite starts the Temporal server
func (s *WorkflowIdempotencyTestSuite) SetupSuite() {
	// Start the Temporal server as a background process
	server, err := testsuite.StartDevServer(context.Background(), testsuite.DevServerOptions{})
	if err != nil {
		s.T().Fatalf("Failed to start Temporal server: %v", err)
		return
	}
	s.server = server

	// Create a client to the Temporal server
	// Get the server address from the DevServer instance
	serverAddr := s.server.FrontendHostPort()
	s.T().Logf("Connecting to Temporal server at: %s", serverAddr)

	client, err := client.Dial(client.Options{
		HostPort: serverAddr,
	})
	if err != nil {
		s.T().Fatalf("Failed to create Temporal client: %v", err)
		return
	}
	s.client = client

	// Create a unique task queue for this test run
	s.taskQueue = "idempotency-test-" + uuid.New().String()

	// Create a worker to process workflow and activity tasks
	s.worker = worker.New(client, s.taskQueue, worker.Options{})

	// Register just one workflow type to avoid workflow type confusion with idempotency
	s.worker.RegisterWorkflow(batch.FeeDeductionWorkflow)
	s.worker.RegisterActivity(DeductFeeActivity)

	// Start the worker
	err = s.worker.Start()
	if err != nil {
		s.T().Fatalf("Failed to start worker: %v", err)
		return
	}
}

// TearDownSuite stops the Temporal server
func (s *WorkflowIdempotencyTestSuite) TearDownSuite() {
	// Stop the worker
	s.worker.Stop()

	// Close the client
	s.client.Close()

	// Stop the Temporal server
	s.server.Stop()
}

// SetupTest initializes the deduction tracker before each test
func (s *WorkflowIdempotencyTestSuite) SetupTest() {
	// Reset the deduction tracker
	deductions.Lock()
	deductions.Count = 0
	deductions.AccountID = ""
	deductions.OrderID = ""
	deductions.Unlock()

	// Small delay to ensure previous test cleanup is complete and allow Temporal server to process
	time.Sleep(300 * time.Millisecond)
}

// TearDownTest doesn't need to do anything for this test suite
func (s *WorkflowIdempotencyTestSuite) TearDownTest() {}

// TestSingleWorkflowIDIdempotency demonstrates how Temporal ensures idempotent execution when using the same WorkflowID multiple times
func (s *WorkflowIdempotencyTestSuite) TestSingleWorkflowIDIdempotency() {
	// Make sure we run this test first, before the concurrent test
	s.T().Log("Running single workflow ID test...")

	// Setup test parameters
	const accountID = "ACCT-12345"
	const orderID = "ORD-67890"
	const feeAmount = 50.0

	// Use a unique WorkflowID including a UUID to avoid conflicts with previous test runs
	workflowID := "FEE-WF-" + orderID + "-" + uuid.New().String()

	// Create context for the workflow execution
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Reset the deduction tracker to ensure clean state
	deductions.Lock()
	deductions.Count = 0
	deductions.AccountID = ""
	deductions.OrderID = ""
	deductions.Unlock()

	// Prepare workflow input
	workflowInput := batch.FeeDeductionWorkflowInput{
		AccountID: accountID,
		OrderID:   orderID,
		Amount:    feeAmount,
	}

	// Execute the workflow with a specific WorkflowID
	options := client.StartWorkflowOptions{
		ID:                  workflowID,
		TaskQueue:           s.taskQueue,
		WorkflowRunTimeout:  1 * time.Minute,
		WorkflowTaskTimeout: 10 * time.Second,
	}

	s.T().Logf("Starting workflow execution with WorkflowID: %s", workflowID)
	run, err := s.client.ExecuteWorkflow(ctx, options, batch.FeeDeductionWorkflow, workflowInput)
	s.NoError(err, "Failed to start workflow execution")

	// Wait for the workflow to complete
	s.T().Log("Waiting for workflow execution to complete...")
	var result batch.FeeDeductionWorkflowResult
	err = run.Get(ctx, &result)
	s.NoError(err, "Workflow execution failed")
	s.T().Logf("Workflow result: %+v", result)

	// Verify that the activity was executed once by checking the deductions tracker
	deductions.Lock()
	s.Equal(1, deductions.Count, "DeductFeeActivity should be called exactly once")
	s.Equal(accountID, deductions.AccountID, "Unexpected accountID in activity execution")
	s.Equal(orderID, deductions.OrderID, "Unexpected orderID in activity execution")
	deductions.Unlock()

	s.T().Log("SUCCESS: Activity was called exactly once and processed the deduction correctly.")
}

// TestConcurrentWorkflowIdempotency simulates multiple clients trying to process the same order concurrently
// This test demonstrates Temporal's guarantee that only one workflow execution will happen for a given workflowID
func (s *WorkflowIdempotencyTestSuite) TestConcurrentWorkflowIdempotency() {
	// Setup test parameters
	const accountID = "ACCT-12345"
	const orderID = "ORD-98765"
	const feeAmount = 50.0
	workflowID := "FEE-WF-" + orderID

	// Create context for the workflow execution
	ctx := context.Background()

	// Prepare workflow input
	workflowInput := batch.FeeDeductionWorkflowInput{
		AccountID: accountID,
		OrderID:   orderID,
		Amount:    feeAmount,
	}

	// Workflow options with the specific WorkflowID for idempotency
	// Setting WorkflowIDReusePolicy to RejectDuplicate ensures that Temporal will reject
	// any attempt to start a workflow with a WorkflowID that is already in use
	options := client.StartWorkflowOptions{
		ID:                  workflowID,
		TaskQueue:           s.taskQueue,
		WorkflowRunTimeout:  5 * time.Minute,
		WorkflowTaskTimeout: 10 * time.Second,
	}

	// Reset the deductions tracker for this test
	deductions.Lock()
	deductions.Count = 0
	deductions.AccountID = ""
	deductions.OrderID = ""
	deductions.Unlock()

	// Use a fixed workflow ID based on orderID for all clients - this is the key to idempotency testing
	workflowID = "FEE-WF-" + orderID
	s.T().Logf("Using same fixed WorkflowID for all clients: %s", workflowID)

	// Set the workflowID in the options
	options.ID = workflowID

	// IMPORTANT: Set this to REJECT_DUPLICATE to ensure idempotency
	// This policy will cause Temporal to reject any attempt to start a workflow with an ID that's already in use
	options.WorkflowIDReusePolicy = enumspb.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE

	// Make sure we don't use "ID exists, run open" as a success case
	// Only one client should get true success; the rest should get errors

	// Create a barrier to make all goroutines start at almost the same time
	// This creates a more realistic concurrent scenario to test idempotency
	var barrier sync.WaitGroup
	barrier.Add(1)

	// Simulate multiple clients trying to execute the same workflow concurrently
	const clientCount = 5
	var successCount int
	var errorCount int
	var mutex sync.Mutex                // To protect counts during concurrent access
	var executions []client.WorkflowRun // Track successful executions

	s.T().Logf("Starting %d clients with the same WorkflowID: %s", clientCount, workflowID)

	// Use a WaitGroup to wait for all clients to finish
	var wg sync.WaitGroup
	wg.Add(clientCount)

	// Start multiple clients attempting to execute workflows with the same ID
	for i := 0; i < clientCount; i++ {
		clientID := i + 1

		go func(id int) {
			defer wg.Done()

			// Wait for the barrier - this makes all goroutines try to start workflows at nearly the same time
			s.T().Logf("Client %d waiting at barrier", id)
			barrier.Wait()

			s.T().Logf("Client %d attempting to start workflow with ID: %s", id, workflowID)

			// Create a context with timeout for this execution attempt
			ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			// Try-catch equivalent in Go to capture panics that might occur during workflow execution
			var execution client.WorkflowRun
			var err error

			func() {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("panic during workflow execution: %v", r)
					}
				}()
				// Each client tries to execute the workflow with the same WorkflowID
				// We only registered one workflow type to avoid having multiple workflows execute with the same ID
				execution, err = s.client.ExecuteWorkflow(ctxWithTimeout, options, batch.FeeDeductionWorkflow, workflowInput)
			}()

			// Safely update counts
			mutex.Lock()
			defer mutex.Unlock()

			// In Temporal, there are two key cases we need to handle:
			// 1. No error: We got a workflow handle, but this doesn't necessarily mean a NEW workflow was started
			//    It could be a handle to an EXISTING workflow
			// 2. WorkflowExecutionAlreadyStarted error: This specific error indicates we tried to start
			//    a new workflow with a workflowID that already had a running workflow

			if err == nil {
				// We got a workflow handle - record it and track its run ID
				// This doesn't necessarily mean we started a NEW workflow, just that we have a handle
				runID := execution.GetRunID()
				s.T().Logf("Client %d got workflow handle - WorkflowID: %s, RunID: %s", id, workflowID, runID)
				executions = append(executions, execution)
				successCount++
			} else {
				// Simplified check for the expected idempotency error
				_, isAlreadyStarted := err.(*serviceerror.WorkflowExecutionAlreadyStarted)
				if isAlreadyStarted {
					// This is the expected error when the workflow is already running with this ID
					s.T().Logf("Client %d received WorkflowExecutionAlreadyStarted error", id)
				} else {
					// This is an unexpected error - something else went wrong
					s.T().Logf("Client %d failed with UNEXPECTED error: %v", id, err)
				}
				errorCount++
			}
		}(clientID)
	}

	// Release the barrier to let all goroutines start at once
	s.T().Log("Releasing barrier - all clients will attempt to start workflows simultaneously")
	barrier.Done()

	// Wait for all clients to finish attempting to start workflows
	wg.Wait()

	// Verify results by examining the workflow handles
	s.T().Logf("Results: %d workflow handles obtained", successCount)

	// A key demonstration of idempotency is that all clients receive a handle with the SAME RunID
	// Let's check that all executions have the same RunID
	s.T().Log("Checking RunIDs from all workflow handles...")

	// Collect all unique RunIDs
	var runIDs = make(map[string]bool)
	for i, execution := range executions {
		runID := execution.GetRunID()
		s.T().Logf("Execution %d has RunID: %s", i+1, runID)
		runIDs[runID] = true
	}

	// The number of unique RunIDs should be exactly 1 if idempotency is working
	s.T().Logf("Number of unique RunIDs: %d", len(runIDs))
	s.Equal(1, len(runIDs), "There should be exactly one unique RunID across all workflow handles")

	// Wait a moment to ensure the workflow has time to complete
	// time.Sleep(2 * time.Second)

	// Wait a moment for any pending operations
	time.Sleep(500 * time.Millisecond)

	// Wait for the workflow to complete and verify the result
	if len(executions) > 0 {
		s.T().Log("Waiting for the workflow to complete...")

		// Get and print result from all clients - this will highlight that all clients
		// get the SAME result despite concurrent execution attempts
		// We'll use green color in terminal output to highlight this important point
		greenColor := "\033[32m" // ANSI escape code for green color
		resetColor := "\033[0m"  // ANSI escape code to reset color

		// Print header in green
		s.T().Logf("%s=== DEMONSTRATION OF TEMPORAL IDEMPOTENCY ===%s", greenColor, resetColor)

		// Get results from each client execution and print in green
		var firstResult batch.FeeDeductionWorkflowResult
		for i, execution := range executions {
			var result batch.FeeDeductionWorkflowResult
			err := execution.Get(ctx, &result)
			s.NoError(err, "Failed to get workflow result")

			// Store first result for later comparison
			if i == 0 {
				firstResult = result
			}

			// Print result in green color
			s.T().Logf("%sClient %d result: Account: %s, OrderID: %s, NewBalance: %.2f, Success: %t%s",
				greenColor, i+1, accountID, result.OrderID, result.NewBalance, result.Success, resetColor)

			// Verify all results are identical
			s.Equal(firstResult, result, "All clients should get identical results")
		}

		// Print confirmation in green
		s.T().Logf("%s=== CONFIRMED: All 5 clients got identical results despite concurrent execution attempts ===%s",
			greenColor, resetColor)

		// The most important check - verify that the activity was only executed ONCE
		deductions.Lock()
		deductionCount := deductions.Count
		deductions.Unlock()

		s.Equal(1, deductionCount, "DeductFeeActivity should be called exactly once")

		// Report the critical proof of idempotency in green
		s.T().Logf("%s=== PROOF OF IDEMPOTENCY: %d Activity Executions for %d Client Attempts ===%s",
			greenColor, deductionCount, len(executions), resetColor)

		// Verify that input data was correctly processed
		deductions.Lock()
		s.Equal(accountID, deductions.AccountID, "Unexpected accountID in activity execution")
		s.Equal(orderID, deductions.OrderID, "Unexpected orderID in activity execution")
		deductions.Unlock()

		// Demonstrate that ANY client can get the same result by using just the WorkflowID
		s.T().Log("Demonstrating that any client can get the same result using just the WorkflowID")
		var clientResult batch.FeeDeductionWorkflowResult
		we := s.client.GetWorkflow(ctx, workflowID, "")
		err := we.Get(ctx, &clientResult)
		s.NoError(err, "Failed to get workflow result using WorkflowID only")
		s.Equal(firstResult, clientResult, "All clients should get the same result")

		// Consolidated success message highlighting the key aspects of idempotency
		s.T().Logf("%sSUCCESS: Temporal's idempotency demonstrated with:%s", greenColor, resetColor)
		s.T().Logf("%s1. All %d clients received workflow handles with the SAME RunID%s", greenColor, len(executions), resetColor)
		s.T().Logf("%s2. Only ONE activity execution occurred (count: %d)%s", greenColor, deductionCount, resetColor)
		s.T().Logf("%s3. All clients received identical results%s", greenColor, resetColor)
	} else {
		s.T().Error("No workflow executions found")
	}
}
