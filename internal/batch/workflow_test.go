package batch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// WorkflowTestSuite is a test suite for the workflow
type WorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment

	// Input
	input ActivityInput
	// Shared store and server for idempotency testing
	testServer *httptest.Server
	testStore  *AccountStore
}

type FakerBatchActivity struct{}

// setupHTTPTestServer sets up a test HTTP server with the fee deduction handler
func setupHTTPTestServer(accountID string, initialBalance float64) (*httptest.Server, *AccountStore) {
	store := NewAccountStore()
	store.CreateAccount(accountID, initialBalance)

	handler := DeductFeeHTTPHandler(store)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))

	return server, store
}

var accountID = "ACCT-45678"
var initialBalance = 200.0
var testServer *httptest.Server
var testStore *AccountStore = NewAccountStore()
var a = FakerBatchActivity{}

func init() {
	testServer, testStore = setupHTTPTestServer(accountID, initialBalance)
	acc, err := testStore.GetAccount(accountID)
	if err != nil {
		panic(err)
	}
	spew.Dump(acc)
}

func (a *FakerBatchActivity) DeductFeeActivity(ctx context.Context, input ActivityInput) (*ActivityResult, error) {
	fmt.Println("Inside FakeDeductFeeActivity.")
	spew.Dump(input)

	// Should just POST to the httptest server .. it has the store there ,.
	// Create the request payload
	payload := FeeDeductionRequest{
		AccountID: input.AccountID,
		Amount:    input.Amount,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	fdr := FeeDeductionResponse{}
	start := time.Now()
	resp, err := http.Post(
		testServer.URL+"/deduct-fee/"+input.OrderID,
		"application/json",
		bytes.NewBuffer(payloadBytes),
	)
	duration := time.Since(start)
	fmt.Println("DeductFeeActivity for ORDID:", input.OrderID, " took", duration)

	// Parse response body if no error
	if err == nil && resp != nil {
		body, readErr := io.ReadAll(resp.Body)
		if readErr == nil {
			json.Unmarshal(body, &fdr)
		}
	}
	// Activity output
	fmt.Println("Message: ", fdr.Message)

	if !fdr.Success {
		return nil, fmt.Errorf("ERR: %s", fdr.Message)
	}
	// All ok if get this far ..
	return &ActivityResult{
		NewBalance: fdr.NewBalance,
		Success:    fdr.Success,
	}, nil
}

func (a *FakerBatchActivity) BasicActivity(ctx context.Context, input ActivityInput) (*ActivityResult, error) {
	fmt.Println("Inside ACT - BasicActivity")
	spew.Dump(input)

	return nil, nil
}

// SetupTest initializes the shared test server and store before each test
func (s *WorkflowTestSuite) SetupTest() {
	fmt.Println("Setting up test workflow")
	spew.Dump(s.input)
	// Create a fresh test environment every test run ..
	env := s.NewTestWorkflowEnvironment()
	// Register workflow and wrapped activity that tracks execution
	env.RegisterWorkflow(FeeDeductionWorkflow)
	//DeductFeeActivity := func(ctx context.Context, input ActivityInput) (*ActivityResult, error) {
	//	fmt.Println("DEDUCTFEE_RUN - OID:", input.OrderID, " AMT:", input.Amount)
	//	// Actually deduct the fee in the first execution
	//	newBalance, xerr := s.testStore.DeductFee(input.AccountID, input.OrderID, input.Amount)
	//	if xerr != nil {
	//		fmt.Println("DEDUCTFEE_RUN - Error:", xerr)
	//		return nil, xerr
	//	}
	//	// All ok ..
	//	return &ActivityResult{
	//		NewBalance: newBalance,
	//		Success:    true,
	//	}, nil
	//}
	env.RegisterActivity(a.DeductFeeActivity)
	env.RegisterActivity(a.BasicActivity)
	// Attach it to be ready for test run ..
	s.env = env
	// Attach the testServer + testStore for use
	s.testServer = testServer
	s.testStore = testStore
	
}

// TearDownTest closes the test server after each test
func (s *WorkflowTestSuite) TearDownTest() {
	//time.Sleep(10 * time.Second)
	fmt.Println("TearDown test server + account for ", accountID)
	// Print out what the AccountID - ACCT-12345 has
	acc, err := testStore.GetAccount(accountID)
	if err != nil {
		s.FailNow("ERR:", err)
	}
	spew.Dump(acc)

	//if s.testServer != nil {
	//	s.testServer.Close()
	//}
}

// Run the test suite with no clashing ...
func TestWorkflowTestSuiteNoClash(t *testing.T) {
	//t.Parallel() // TODO: Does not seem to work ..

	//for i := 1; i < 3; i++ {
	i := 1
	orderID := fmt.Sprintf("ORD-%d", i)
	//t.Run(orderID, func(t *testing.T) {
	// Setup the activity server ..
	fmt.Println("Setup test server + account for ", orderID)
	wts := new(WorkflowTestSuite)
	wts.input = ActivityInput{
		"ACCT-12345",
		fmt.Sprintf("ORD-%d", i),
		10.0,
	}

	//wts.testServer, wts.testStore = setupHTTPTestServer(orderID, 100.0)
	suite.Run(t, wts)
	//})
	//}
}

func (s *WorkflowTestSuite) TestNormalFeeDeduction() {
	fmt.Println("NormalFeeDeduction start")
	// setup input
	//initialBalance := 100.0

	// Setup the workflow input
	input := FeeDeductionWorkflowInput{
		AccountID: s.input.AccountID,
		OrderID:   s.input.OrderID,
		Amount:    s.input.Amount,
	}

	// start workflow
	// Set workflow ID for idempotency using orderID
	workflowID := s.input.OrderID // Using orderID as the workflow ID is key for idempotency
	//s.env.SetStartWorkflowOptions(client.StartWorkflowOptions{
	//	ID: workflowID,
	//})
	//
	//// Execute the workflow with explicit WorkflowID for idempotency
	//s.env.ExecuteWorkflow(FeeDeductionWorkflow, input)
	//
	//// get the results out
	//// Verify workflow completed successfully
	//s.True(s.env.IsWorkflowCompleted())
	//fmt.Println("ERR: ", s.env.GetWorkflowError())
	//// Expected error ..
	//s.Error(s.env.GetWorkflowError())
	// assertions ..
	//
	input.AccountID = "ACCT-45678"
	// Set workflow ID for idempotency using orderID
	workflowID = "ORD-5678" // Using orderID as the workflow ID is key for idempotency
	input.OrderID = "ORD-5678"
	s.env.SetStartWorkflowOptions(client.StartWorkflowOptions{
		ID: workflowID,
	})

	// Execute the workflow with explicit WorkflowID for idempotency
	s.env.ExecuteWorkflow(FeeDeductionWorkflow, input)

	// get the results out
	// Verify workflow completed successfully
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	// assertions ..

	/*
		// Setup the workflow input
		input := FeeDeductionWorkflowInput{
			AccountID: accountID,
			OrderID:   orderID,
			Amount:    amount,
		}

		// Set workflow ID for idempotency using orderID
		workflowID := orderID // Using orderID as the workflow ID is key for idempotency
		env.SetStartWorkflowOptions(client.StartWorkflowOptions{
			ID: workflowID,
		})

		// Execute the workflow with explicit WorkflowID for idempotency
		env.ExecuteWorkflow(FeeDeductionWorkflow, input)

		// Verify workflow completed successfully
		s.True(env.IsWorkflowCompleted())
		s.NoError(env.GetWorkflowError())

		// Get the workflow result
		var result FeeDeductionWorkflowResult
		s.NoError(env.GetWorkflowResult(&result))

		// Verify the result
		s.True(result.Success)
		s.Equal(initialBalance-amount, result.NewBalance)
		s.Equal(orderID, result.OrderID)

		// Verify the activity was called in the first execution
		s.True(firstActivityCalled, "Activity should be called for the first execution")

	*/
}

//// DeductFee implements the fee deduction activity using the AccountStore.DeductFee method
//// This activity is intentionally NOT idempotent - Temporal provides the idempotency
//func (s *WorkflowTestSuite) DeductFee(ctx context.Context, input ActivityInput) (*ActivityResult, error) {
//	// Make sure the account exists
//	if !s.testStore.AccountExists(input.AccountID) {
//		s.testStore.CreateAccount(input.AccountID, 100.0)
//	}
//
//	// Call the AccountStore.DeductFee method, which is NOT idempotent
//	// This means multiple calls with the same OrderID will deduct multiple times
//	// Temporal workflow framework should prevent this by not re-executing activities
//	newBalance, err := s.testStore.DeductFee(input.AccountID, input.OrderID, input.Amount)
//
//	if err != nil {
//		return &ActivityResult{
//			Success: false,
//			Error:   err.Error(),
//		}, nil
//	}
//
//	return &ActivityResult{
//		NewBalance: newBalance,
//		Success:    true,
//	}, nil
//}
//
//// TestIdempotentFeeDeduction tests that the workflow handles idempotent fee deduction
//func (s *WorkflowTestSuite) TestIdempotentFeeDeduction() {
//	// Test parameters
//	accountID := "ACCT-12345"
//	orderID := "ORD-12345"
//	amount := 10.0
//	initialBalance := 100.0
//
//	// Reset shared test store with new account
//	s.testStore.CreateAccount(accountID, initialBalance)
//
//	// Track the activity execution for the first run
//	var firstActivityCalled bool
//
//	// Create a fresh test environment for this test
//	env := s.NewTestWorkflowEnvironment()
//
//	// Register workflow and wrapped activity that tracks execution
//	env.RegisterWorkflow(FeeDeductionWorkflow)
//	env.RegisterActivity(func(ctx context.Context, input ActivityInput) (*ActivityResult, error) {
//		firstActivityCalled = true
//		// Actually deduct the fee in the first execution
//		return s.DeductFee(ctx, input)
//	})
//
//	// We don't need to mock specific activity behavior because our shared test
//	// environment already uses the s.DeductFee method as the activity implementation
//
//	// Setup the workflow input
//	input := FeeDeductionWorkflowInput{
//		AccountID: accountID,
//		OrderID:   orderID,
//		Amount:    amount,
//	}
//
//	// Set workflow ID for idempotency using orderID
//	workflowID := orderID // Using orderID as the workflow ID is key for idempotency
//	env.SetStartWorkflowOptions(client.StartWorkflowOptions{
//		ID: workflowID,
//	})
//
//	// Execute the workflow with explicit WorkflowID for idempotency
//	env.ExecuteWorkflow(FeeDeductionWorkflow, input)
//
//	// Verify workflow completed successfully
//	s.True(env.IsWorkflowCompleted())
//	s.NoError(env.GetWorkflowError())
//
//	// Get the workflow result
//	var result FeeDeductionWorkflowResult
//	s.NoError(env.GetWorkflowResult(&result))
//
//	// Verify the result
//	s.True(result.Success)
//	s.Equal(initialBalance-amount, result.NewBalance)
//	s.Equal(orderID, result.OrderID)
//
//	// Verify the activity was called in the first execution
//	s.True(firstActivityCalled, "Activity should be called for the first execution")
//
//	// Execute the workflow again with the same input to test idempotency
//	// In a real Temporal server, the activity would NOT be executed again with the same workflow ID
//	// The test environment has limitations in simulating Temporal's actual behavior
//	var secondActivityCalled bool
//
//	// Create a new environment for the second execution
//	env2 := s.NewTestWorkflowEnvironment()
//	env2.RegisterWorkflow(FeeDeductionWorkflow)
//	// Register a wrapper around DeductFee that tracks calls for the second execution
//	var activityCalled bool
//	env2.RegisterActivity(func(ctx context.Context, input ActivityInput) (*ActivityResult, error) {
//		secondActivityCalled = true
//		activityCalled = true
//		return s.DeductFee(ctx, input)
//	})
//	// Use variables to avoid unused variable errors
//	_ = secondActivityCalled
//	_ = activityCalled
//
//	// Override activity implementation for the second execution
//	// We don't need to override the activity implementation since the second execution
//	// should reuse the first workflow execution's activity results
//
//	// Now try to execute the same workflow again with the same workflow ID
//	// The previous execution should be reused
//	// Use the same workflowID (orderID) for idempotency
//	env2.SetStartWorkflowOptions(client.StartWorkflowOptions{
//		ID: workflowID, // Using the same workflow ID is key for idempotency
//	})
//
//	// Execute the workflow again with the same input
//	env2.ExecuteWorkflow(FeeDeductionWorkflow, input)
//
//	// Verify workflow completed successfully
//	s.True(env2.IsWorkflowCompleted())
//	s.NoError(env2.GetWorkflowError())
//
//	// Verify the activity was executed in the second run
//	// Note: In a real Temporal service with a real server, the activity would NOT be
//	// executed again with the same workflow ID, but the test framework doesn't simulate this
//	// For our test purposes, we can demonstrate the idempotent behavior by checking the account balance
//	s.True(activityCalled, "Activity execution is tracked for testing purposes")
//
//	// Get the workflow result again
//	var result2 FeeDeductionWorkflowResult
//	s.NoError(env2.GetWorkflowResult(&result2))
//
//	// The results should be identical (proving idempotency)
//	s.Equal(result.NewBalance, result2.NewBalance)
//
//	// Note: In a real Temporal service, the activity would not be called again on a retry
//	// with the same workflowID. The test environment doesn't fully simulate this behavior.
//	s.Equal(result.Success, result2.Success)
//	s.Equal(result.OrderID, result2.OrderID)
//
//	// Important: Verify that the account balance was only deducted once
//	// Even though our DeductFee activity is non-idempotent, the Temporal workflow
//	// framework should have prevented the activity from being called twice
//	account, err := s.testStore.GetAccount(accountID)
//	s.NoError(err)
//	s.Equal(initialBalance-amount, account.Balance, "Balance should only be deducted once")
//}
//
//// TestParallelRequests tests handling multiple requests for the same order
//func (s *WorkflowTestSuite) TestParallelRequests() {
//	// Setup test HTTP server with account store
//	accountID := "ACCT-12345"
//	orderID := "ORD-67890"
//	amount := 10.0
//	initialBalance := 100.0
//
//	// Reset shared test store with new account
//	s.testStore.CreateAccount(accountID, initialBalance)
//
//	// Create a fresh test environment for this test
//	env := s.NewTestWorkflowEnvironment()
//
//	// Register workflow and activity
//	env.RegisterWorkflow(FeeDeductionWorkflow)
//	env.RegisterActivity(s.DeductFee)
//
//	// We don't need to mock specific activity behavior because our shared test
//	// environment already uses the s.DeductFee method as the activity implementation
//
//	// Setup the workflow input
//	input := FeeDeductionWorkflowInput{
//		AccountID: accountID,
//		OrderID:   orderID,
//		Amount:    amount,
//	}
//
//	// Set workflow ID for idempotency
//	env.SetStartWorkflowOptions(client.StartWorkflowOptions{
//		ID: orderID,
//	})
//
//	// Start the workflow for the first time
//	env.ExecuteWorkflow(FeeDeductionWorkflow, input)
//
//	// Verify workflow completed successfully
//	s.True(env.IsWorkflowCompleted())
//	s.NoError(env.GetWorkflowError())
//
//	// Get the workflow result
//	var result1 FeeDeductionWorkflowResult
//	s.NoError(env.GetWorkflowResult(&result1))
//
//	// The account balance should be reduced by 'amount'
//	s.True(result1.Success)
//	s.Equal(initialBalance-amount, result1.NewBalance)
//
//	// Get current account balance after first execution
//	account, err := s.testStore.GetAccount(accountID)
//	s.NoError(err)
//	s.Equal(initialBalance-amount, account.Balance, "Balance should be deducted after first execution")
//
//	// Track if activity was called in the second execution - in a real Temporal deployment with a server,
//	// this activity would NOT be executed when using the same workflow ID, but
//	// the test framework doesn't fully simulate Temporal's caching behavior
//	var secondActivityCalled bool
//
//	// Create a new environment for the second execution
//	env2 := s.NewTestWorkflowEnvironment()
//	env2.RegisterWorkflow(FeeDeductionWorkflow)
//	// Register a wrapper around DeductFee that tracks calls for the second execution
//	var activityCalled bool
//	env2.RegisterActivity(func(ctx context.Context, input ActivityInput) (*ActivityResult, error) {
//		secondActivityCalled = true
//		activityCalled = true
//		return s.DeductFee(ctx, input)
//	})
//	// Use variables to avoid unused variable errors
//	_ = secondActivityCalled
//	_ = activityCalled
//
//	// Override activity implementation for the second execution
//	// We don't need to override the activity implementation since the second execution
//	// should reuse the first workflow execution's activity results
//
//	// Now simulate a second request coming in with the same order ID
//	// If the workflow is truly idempotent, this should not deduct the fee again
//	env2.SetStartWorkflowOptions(client.StartWorkflowOptions{
//		ID: orderID,
//	})
//	env2.ExecuteWorkflow(FeeDeductionWorkflow, input)
//
//	// Verify the activity was executed in the second run when using the test framework
//	// Note that with a real Temporal server, this behavior would depend on whether
//	// the workflow history is still available within the retention period
//
//	// Verify workflow completed successfully
//	s.True(env2.IsWorkflowCompleted())
//	s.NoError(env2.GetWorkflowError())
//
//	// Verify the activity was executed in the second run
//	// Note: In a real Temporal service with a real server, the activity would NOT be
//	// executed again with the same workflow ID, but the test framework doesn't simulate this
//	// For our test purposes, we can demonstrate the idempotent behavior by checking the account balance
//	s.True(activityCalled, "Activity execution is tracked for testing purposes")
//
//	// Get the workflow result
//	var result2 FeeDeductionWorkflowResult
//	s.NoError(env2.GetWorkflowResult(&result2))
//
//	// Verify that the balance remains the same as after the first deduction
//	// This confirms idempotency - the fee was only deducted once
//	s.True(result2.Success)
//	s.Equal(result1.NewBalance, result2.NewBalance)
//	s.Equal(result1.Success, result2.Success)
//
//	// Check account state in the store to confirm idempotency
//	account, err = s.testStore.GetAccount(accountID)
//	s.NoError(err)
//	s.Equal(initialBalance-amount, account.Balance, "Balance should only be deducted once")
//}
//
//// TestWorkflowRetentionPeriod tests that the workflow result can be retrieved even after
//// the workflow has completed, as long as it's within the retention period
//func (s *WorkflowTestSuite) TestWorkflowRetentionPeriod() {
//	// Setup test HTTP server with account store
//	accountID := "ACCT-12345"
//	orderID := "ORD-54321"
//	amount := 10.0
//	initialBalance := 100.0
//
//	// Reset shared test store with new account
//	s.testStore.CreateAccount(accountID, initialBalance)
//
//	// Create a fresh test environment for this test
//	env := s.NewTestWorkflowEnvironment()
//
//	// Register workflow and activity
//	env.RegisterWorkflow(FeeDeductionWorkflow)
//	env.RegisterActivity(s.DeductFee)
//
//	// We don't need to mock specific activity behavior because our shared test
//	// environment already uses the s.DeductFee method as the activity implementation
//
//	// Setup the workflow input
//	input := FeeDeductionWorkflowInput{
//		AccountID: accountID,
//		OrderID:   orderID,
//		Amount:    amount,
//	}
//
//	// Set the workflow ID for idempotency
//	env.SetStartWorkflowOptions(client.StartWorkflowOptions{
//		ID: orderID,
//	})
//
//	// Set workflow ID for idempotency using orderID
//	workflowID := orderID // Using orderID as the workflow ID is key for idempotency
//	env.SetStartWorkflowOptions(client.StartWorkflowOptions{
//		ID: workflowID,
//	})
//
//	// Execute the workflow with explicit WorkflowID for idempotency
//	env.ExecuteWorkflow(FeeDeductionWorkflow, input)
//
//	// Verify workflow completed successfully
//	s.True(env.IsWorkflowCompleted())
//	s.NoError(env.GetWorkflowError())
//
//	// Get the workflow result
//	var result FeeDeductionWorkflowResult
//	s.NoError(env.GetWorkflowResult(&result))
//
//	// Verify the result
//	s.True(result.Success)
//	s.Equal(initialBalance-amount, result.NewBalance)
//
//	// Check account state after first execution
//	account, err := s.testStore.GetAccount(accountID)
//	s.NoError(err)
//	s.Equal(initialBalance-amount, account.Balance, "Balance should be deducted after first execution")
//
//	// Create a new test environment for the second execution
//	env2 := s.NewTestWorkflowEnvironment()
//	env2.RegisterWorkflow(FeeDeductionWorkflow)
//	env2.RegisterActivity(s.DeductFee)
//
//	// Set up the same workflow ID to simulate a call within retention period
//	env2.SetStartWorkflowOptions(client.StartWorkflowOptions{
//		ID: orderID,
//	})
//
//	// For the second workflow execution, we'll track if the activity is called by setting a flag
//	var activityCalled bool
//
//	// Override activity implementation for the second execution - this should NOT be called
//	// if the system properly reuses the workflow history
//	env2.OnActivity(s.DeductFee, mock.Anything, mock.Anything).Return(
//		func(ctx context.Context, input ActivityInput) (*ActivityResult, error) {
//			// Set the flag to indicate this was called
//			activityCalled = true
//
//			// Use the same test HTTP server (which still has the same account state)
//			client := &http.Client{Timeout: 5 * time.Second}
//
//			// Create request payload
//			payload := FeeDeductionRequest{
//				AccountID: input.AccountID,
//				Amount:    input.Amount,
//			}
//
//			// Convert to JSON
//			jsonPayload, err := json.Marshal(payload)
//			if err != nil {
//				return &ActivityResult{
//					Success: false,
//					Error:   "Failed to marshal request: " + err.Error(),
//				}, nil
//			}
//
//			// Create request with OrderID in path for idempotency
//			req, err := http.NewRequestWithContext(
//				ctx,
//				http.MethodPost,
//				s.testServer.URL+"/deduct-fee/"+input.OrderID,
//				bytes.NewBuffer(jsonPayload),
//			)
//			if err != nil {
//				return &ActivityResult{
//					Success: false,
//					Error:   "Failed to create request: " + err.Error(),
//				}, nil
//			}
//
//			// Set headers
//			req.Header.Set("Content-Type", "application/json")
//
//			// Send request
//			resp, err := client.Do(req)
//			if err != nil {
//				return &ActivityResult{
//					Success: false,
//					Error:   "HTTP request failed: " + err.Error(),
//				}, nil
//			}
//			defer resp.Body.Close()
//
//			// Parse response
//			var response FeeDeductionResponse
//			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
//				return &ActivityResult{
//					Success: false,
//					Error:   "Failed to parse response: " + err.Error(),
//				}, nil
//			}
//
//			// Return activity result based on HTTP response
//			return &ActivityResult{
//				NewBalance: response.NewBalance,
//				Success:    response.Success,
//				Error:      response.Message,
//			}, nil
//		},
//	)
//
//	// Try to execute the same workflow again - should get the same result from history
//	env2.ExecuteWorkflow(FeeDeductionWorkflow, input)
//
//	// Verify workflow completed successfully
//	s.True(env2.IsWorkflowCompleted())
//	s.NoError(env2.GetWorkflowError())
//
//	// Verify the activity was executed in the second run
//	// Note: In a real Temporal service with a real server, the activity would NOT be
//	// executed again with the same workflow ID, but the test framework doesn't simulate this
//	// For our test purposes, we can demonstrate the idempotent behavior by checking the account balance
//	s.True(activityCalled, "Activity execution is tracked for testing purposes")
//
//	// Get the workflow result again
//	var result2 FeeDeductionWorkflowResult
//	s.NoError(env2.GetWorkflowResult(&result2))
//
//	// The results should be identical (proving retrieval from history)
//	s.Equal(result.NewBalance, result2.NewBalance)
//	s.Equal(result.Success, result2.Success)
//	s.Equal(result.OrderID, result2.OrderID)
//
//	// Verify the activity wasn't called for the second execution
//	// This is because Temporal should reuse the existing workflow history
//	s.False(activityCalled, "Activity should not be called on the second execution")
//
//	// Check account state after second execution (should be unchanged)
//	account, err = s.testStore.GetAccount(accountID)
//	s.NoError(err)
//	s.Equal(initialBalance-amount, account.Balance, "Balance should not change after second execution")
//}
//
//// TestCompleteIdempotencyImplementation combines multiple scenarios
//func (s *WorkflowTestSuite) TestCompleteIdempotencyImplementation() {
//	// Setup test HTTP server with account store
//	accountID := "ACCT-5678"
//	orderID := "ORD-8765"
//	amount := 25.0
//	initialBalance := 100.0
//
//	// Reset shared test store with new account
//	s.testStore.CreateAccount(accountID, initialBalance)
//
//	// Create a fresh test environment for this test
//	env := s.NewTestWorkflowEnvironment()
//
//	// Register workflow and activity
//	env.RegisterWorkflow(FeeDeductionWorkflow)
//	env.RegisterActivity(s.DeductFee)
//
//	// We don't need to mock specific activity behavior because our shared test
//	// environment already uses the s.DeductFee method as the activity implementation
//
//	// Setup the workflow input
//	input := FeeDeductionWorkflowInput{
//		AccountID: accountID,
//		OrderID:   orderID,
//		Amount:    amount,
//	}
//
//	// Part 1: First execution
//	env.SetStartWorkflowOptions(client.StartWorkflowOptions{
//		ID: orderID,
//	})
//
//	env.ExecuteWorkflow(FeeDeductionWorkflow, input)
//
//	s.True(env.IsWorkflowCompleted())
//	s.NoError(env.GetWorkflowError())
//
//	var firstResult FeeDeductionWorkflowResult
//	s.NoError(env.GetWorkflowResult(&firstResult))
//
//	s.True(firstResult.Success)
//	s.Equal(initialBalance-amount, firstResult.NewBalance)
//
//	// Check account state after first execution
//	account, err := s.testStore.GetAccount(accountID)
//	s.NoError(err)
//	s.Equal(initialBalance-amount, account.Balance, "Balance should be deducted after first execution")
//
//	// Track if activity was called in the second execution - in a real Temporal deployment with a server,
//	// this activity would NOT be executed when using the same workflow ID, but
//	// the test framework doesn't fully simulate Temporal's caching behavior
//	var secondActivityCalled bool
//
//	// Create a new environment for the second execution
//	env2 := s.NewTestWorkflowEnvironment()
//	env2.RegisterWorkflow(FeeDeductionWorkflow)
//	// Register a wrapper around DeductFee that tracks calls for the second execution
//	var activityCalled bool
//	env2.RegisterActivity(func(ctx context.Context, input ActivityInput) (*ActivityResult, error) {
//		secondActivityCalled = true
//		activityCalled = true
//		return s.DeductFee(ctx, input)
//	})
//	// Use variables to avoid unused variable errors
//	_ = secondActivityCalled
//	_ = activityCalled
//
//	// Override activity implementation for the second execution
//	// We don't need to override the activity implementation since the second execution
//	// should reuse the first workflow execution's activity results
//
//	// Part 2: Immediate Retry (simulates retry within a few seconds)
//	// We can't directly set time in the testing env, but we can simulate a retry
//	env2.SetStartWorkflowOptions(client.StartWorkflowOptions{
//		ID: orderID,
//	})
//	env2.ExecuteWorkflow(FeeDeductionWorkflow, input)
//
//	// Verify the activity was executed in the second run when using the test framework
//	// Note that with a real Temporal server, this behavior would depend on whether
//	// the workflow history is still available within the retention period
//
//	var secondResult FeeDeductionWorkflowResult
//	s.NoError(env2.GetWorkflowResult(&secondResult))
//
//	// Should be identical to first result (no double charging)
//	s.Equal(firstResult.NewBalance, secondResult.NewBalance)
//
//	// Check account state after second execution (should be unchanged)
//	account, err = s.testStore.GetAccount(accountID)
//	s.NoError(err)
//	s.Equal(initialBalance-amount, account.Balance, "Balance should not change after second execution")
//
//	// Create a new environment for the third execution
//	env3 := s.NewTestWorkflowEnvironment()
//	env3.RegisterWorkflow(FeeDeductionWorkflow)
//	env3.RegisterActivity(s.DeductFee)
//
//	// Override activity implementation for the third execution
//	env3.OnActivity(s.DeductFee, mock.Anything, mock.Anything).Return(
//		func(ctx context.Context, input ActivityInput) (*ActivityResult, error) {
//			// Use the same test HTTP server (which still has the same account state)
//			client := &http.Client{Timeout: 5 * time.Second}
//
//			// Create request payload
//			payload := FeeDeductionRequest{
//				AccountID: input.AccountID,
//				Amount:    input.Amount,
//			}
//
//			// Convert to JSON
//			jsonPayload, err := json.Marshal(payload)
//			if err != nil {
//				return &ActivityResult{
//					Success: false,
//					Error:   "Failed to marshal request: " + err.Error(),
//				}, nil
//			}
//
//			// Create request with OrderID in path for idempotency
//			req, err := http.NewRequestWithContext(
//				ctx,
//				http.MethodPost,
//				s.testServer.URL+"/deduct-fee/"+input.OrderID,
//				bytes.NewBuffer(jsonPayload),
//			)
//			if err != nil {
//				return &ActivityResult{
//					Success: false,
//					Error:   "Failed to create request: " + err.Error(),
//				}, nil
//			}
//
//			// Set headers
//			req.Header.Set("Content-Type", "application/json")
//
//			// Send request
//			resp, err := client.Do(req)
//			if err != nil {
//				return &ActivityResult{
//					Success: false,
//					Error:   "HTTP request failed: " + err.Error(),
//				}, nil
//			}
//			defer resp.Body.Close()
//
//			// Parse response
//			var response FeeDeductionResponse
//			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
//				return &ActivityResult{
//					Success: false,
//					Error:   "Failed to parse response: " + err.Error(),
//				}, nil
//			}
//
//			// Return activity result based on HTTP response
//			return &ActivityResult{
//				NewBalance: response.NewBalance,
//				Success:    response.Success,
//				Error:      response.Message,
//			}, nil
//		},
//	)
//
//	// Part 3: Retry after some time (but within retention period)
//	// In a real scenario, this would be days later but still within retention
//	env3.SetStartWorkflowOptions(client.StartWorkflowOptions{
//		ID: orderID,
//	})
//	env3.ExecuteWorkflow(FeeDeductionWorkflow, input)
//
//	var thirdResult FeeDeductionWorkflowResult
//	s.NoError(env3.GetWorkflowResult(&thirdResult))
//
//	// Should still be identical (workflow history still available)
//	s.Equal(firstResult.NewBalance, thirdResult.NewBalance)
//
//	// Check account state after third execution (should be unchanged)
//	account, err = s.testStore.GetAccount(accountID)
//	s.NoError(err)
//	s.Equal(initialBalance-amount, account.Balance, "Balance should not change after third execution")
//
//	// Create a new environment for the fourth execution with different order ID
//	env4 := s.NewTestWorkflowEnvironment()
//	env4.RegisterWorkflow(FeeDeductionWorkflow)
//	env4.RegisterActivity(s.DeductFee)
//
//	// Part 4: Different order ID should create a new execution
//	newOrderID := "ORD-9999"
//	newInput := FeeDeductionWorkflowInput{
//		AccountID: accountID,
//		OrderID:   newOrderID,
//		Amount:    amount,
//	}
//
//	// Override activity implementation for the fourth execution
//	env4.OnActivity(s.DeductFee, mock.Anything, mock.Anything).Return(
//		func(ctx context.Context, input ActivityInput) (*ActivityResult, error) {
//			// Use the same test HTTP server (which still has the same account state)
//			client := &http.Client{Timeout: 5 * time.Second}
//
//			// Create request payload
//			payload := FeeDeductionRequest{
//				AccountID: input.AccountID,
//				Amount:    input.Amount,
//			}
//
//			// Convert to JSON
//			jsonPayload, err := json.Marshal(payload)
//			if err != nil {
//				return &ActivityResult{
//					Success: false,
//					Error:   "Failed to marshal request: " + err.Error(),
//				}, nil
//			}
//
//			// Create request with OrderID in path for idempotency
//			req, err := http.NewRequestWithContext(
//				ctx,
//				http.MethodPost,
//				s.testServer.URL+"/deduct-fee/"+input.OrderID,
//				bytes.NewBuffer(jsonPayload),
//			)
//			if err != nil {
//				return &ActivityResult{
//					Success: false,
//					Error:   "Failed to create request: " + err.Error(),
//				}, nil
//			}
//
//			// Set headers
//			req.Header.Set("Content-Type", "application/json")
//
//			// Send request
//			resp, err := client.Do(req)
//			if err != nil {
//				return &ActivityResult{
//					Success: false,
//					Error:   "HTTP request failed: " + err.Error(),
//				}, nil
//			}
//			defer resp.Body.Close()
//
//			// Parse response
//			var response FeeDeductionResponse
//			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
//				return &ActivityResult{
//					Success: false,
//					Error:   "Failed to parse response: " + err.Error(),
//				}, nil
//			}
//
//			// Return activity result based on HTTP response
//			return &ActivityResult{
//				NewBalance: response.NewBalance,
//				Success:    response.Success,
//				Error:      response.Message,
//			}, nil
//		},
//	)
//
//	env4.SetStartWorkflowOptions(client.StartWorkflowOptions{
//		ID: newOrderID,
//	})
//
//	env4.ExecuteWorkflow(FeeDeductionWorkflow, newInput)
//
//	var fourthResult FeeDeductionWorkflowResult
//	s.NoError(env4.GetWorkflowResult(&fourthResult))
//
//	// Should be a different result - balance should be reduced again
//	s.Equal(initialBalance-amount-amount, fourthResult.NewBalance)
//
//	// Check account state after fourth execution (should have a double deduction)
//	account, err = s.testStore.GetAccount(accountID)
//	s.NoError(err)
//	s.Equal(initialBalance-amount-amount, account.Balance, "Balance should be deducted twice after different order ID")
//}
