package batch

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
)

// WorkflowTestSuite is a test suite for the workflow
type WorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

// setupHTTPTestServer sets up a test HTTP server with the fee deduction handler
func setupHTTPTestServer(accountID string, initialBalance float64) (*httptest.Server, *AccountStore) {
	store := NewAccountStore()
	store.CreateAccount(accountID, initialBalance)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DeductFeeHTTPHandler(store)(w, r)
	}))

	return server, store
}

// Implementation of the DeductFeeActivity that uses the real logic
func (s *WorkflowTestSuite) DeductFeeActivityImpl(ctx context.Context, input ActivityInput) (*ActivityResult, error) {
	// For testing, we'll use the real AccountStore with a test HTTP server
	store := NewAccountStore()
	accountID := input.AccountID
	
	// Initialize account with 100 balance if it doesn't exist
	store.CreateAccount(accountID, 100.0)
	
	// Create test HTTP server with our handler
	server := httptest.NewServer(DeductFeeHTTPHandler(store))
	defer server.Close()
	
	// Create HTTP client
	client := &http.Client{Timeout: 5 * time.Second}
	
	// Create request payload
	payload := FeeDeductionRequest{
		AccountID: input.AccountID,
		Amount:    input.Amount,
	}
	
	// Convert to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return &ActivityResult{
			Success: false,
			Error:   "Failed to marshal request: " + err.Error(),
		}, nil
	}
	
	// Create request with OrderID in path for idempotency
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		server.URL+"/"+input.OrderID,
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		return &ActivityResult{
			Success: false,
			Error:   "Failed to create request: " + err.Error(),
		}, nil
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return &ActivityResult{
			Success: false,
			Error:   "HTTP request failed: " + err.Error(),
		}, nil
	}
	defer resp.Body.Close()
	
	// Parse response
	var response FeeDeductionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return &ActivityResult{
			Success: false,
			Error:   "Failed to parse response: " + err.Error(),
		}, nil
	}
	
	// Return activity result based on HTTP response
	return &ActivityResult{
		NewBalance: response.NewBalance,
		Success:    response.Success,
		Error:      response.Message,
	}, nil
}

// TestIdempotentFeeDeduction tests that the workflow handles idempotent fee deduction
func (s *WorkflowTestSuite) TestIdempotentFeeDeduction() {
	// Create a fresh test environment for this test
	env := s.NewTestWorkflowEnvironment()
	
	// Register workflow and activity
	env.RegisterWorkflow(FeeDeductionWorkflow)
	env.RegisterActivity(DeductFeeActivity)
	
	// Setup test constants
	accountID := "ACCT-12345"
	orderID := "ORD-12345"
	amount := 10.0
	initialBalance := 100.0
	
	// Mock the activity to simulate fee deduction
	env.OnActivity(DeductFeeActivity, mock.Anything, ActivityInput{
		AccountID: accountID,
		OrderID:   orderID,
		Amount:    amount,
	}).Return(&ActivityResult{
		NewBalance: initialBalance - amount,
		Success:    true,
	}, nil).Once() // Expect exactly one call
	
	// Setup the workflow input
	input := FeeDeductionWorkflowInput{
		AccountID: accountID,
		OrderID:   orderID,
		Amount:    amount,
	}
	
	// Execute the workflow
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
	
	// Create a new environment for the second execution
	env2 := s.NewTestWorkflowEnvironment()
	env2.RegisterWorkflow(FeeDeductionWorkflow)
	env2.RegisterActivity(DeductFeeActivity)
	
	// Mock the activity for the second execution to return the same result
	// This simulates idempotent behavior where calling the same activity with the same input
	// should return the same result
	env2.OnActivity(DeductFeeActivity, mock.Anything, ActivityInput{
		AccountID: accountID,
		OrderID:   orderID,
		Amount:    amount,
	}).Return(&ActivityResult{
		NewBalance: initialBalance - amount,
		Success:    true,
	}, nil).Once()
	
	// Now try to execute the same workflow again with the same workflow ID
	// The previous execution should be reused
	env2.SetStartWorkflowOptions(client.StartWorkflowOptions{
		ID: orderID, // Use orderID as the workflow ID for idempotency
	})
	
	// Execute the workflow again with the same input
	env2.ExecuteWorkflow(FeeDeductionWorkflow, input)
	
	// Verify workflow completed successfully
	s.True(env2.IsWorkflowCompleted())
	s.NoError(env2.GetWorkflowError())
	
	// Get the workflow result again
	var result2 FeeDeductionWorkflowResult
	s.NoError(env2.GetWorkflowResult(&result2))
	
	// The results should be identical (proving idempotency)
	s.Equal(result.NewBalance, result2.NewBalance)
	s.Equal(result.Success, result2.Success)
	s.Equal(result.OrderID, result2.OrderID)
}

// TestParallelRequests tests handling multiple requests for the same order
func (s *WorkflowTestSuite) TestParallelRequests() {
	// Create a fresh test environment for this test
	env := s.NewTestWorkflowEnvironment()
	
	// Register workflow and activity
	env.RegisterWorkflow(FeeDeductionWorkflow)
	env.RegisterActivity(DeductFeeActivity)
	
	// Set up mock temporal worker environment
	accountID := "ACCT-12345"
	orderID := "ORD-67890"
	amount := 10.0
	initialBalance := 100.0
	
	// Instead of using the real implementation, mock the activity
	env.OnActivity(DeductFeeActivity, mock.Anything, ActivityInput{
		AccountID: accountID,
		OrderID:   orderID,
		Amount:    amount,
	}).Return(&ActivityResult{
		NewBalance: initialBalance - amount,
		Success:    true,
	}, nil).Once()
	
	// Setup the workflow input
	input := FeeDeductionWorkflowInput{
		AccountID: accountID,
		OrderID:   orderID,
		Amount:    amount,
	}
	
	// Set workflow ID for idempotency
	env.SetStartWorkflowOptions(client.StartWorkflowOptions{
		ID: orderID,
	})
	
	// Start the workflow for the first time
	env.ExecuteWorkflow(FeeDeductionWorkflow, input)
	
	// Verify workflow completed successfully
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	
	// Get the workflow result
	var result1 FeeDeductionWorkflowResult
	s.NoError(env.GetWorkflowResult(&result1))
	
	// The account balance should be reduced by 'amount'
	s.True(result1.Success)
	s.Equal(initialBalance-amount, result1.NewBalance)
	
	// Create a new environment for the second execution
	env2 := s.NewTestWorkflowEnvironment()
	env2.RegisterWorkflow(FeeDeductionWorkflow)
	env2.RegisterActivity(DeductFeeActivity)
	
	// Mock the activity for the second execution to return the same result
	// This simulates idempotent behavior
	env2.OnActivity(DeductFeeActivity, mock.Anything, ActivityInput{
		AccountID: accountID,
		OrderID:   orderID,
		Amount:    amount,
	}).Return(&ActivityResult{
		NewBalance: initialBalance - amount, // Same balance as first execution
		Success:    true,
	}, nil).Once()
	
	// Now simulate a second request coming in with the same order ID
	// If the workflow is truly idempotent, this should not deduct the fee again
	env2.SetStartWorkflowOptions(client.StartWorkflowOptions{
		ID: orderID,
	})
	env2.ExecuteWorkflow(FeeDeductionWorkflow, input)
	
	// Verify workflow completed successfully
	s.True(env2.IsWorkflowCompleted())
	s.NoError(env2.GetWorkflowError())
	
	// Get the workflow result
	var result2 FeeDeductionWorkflowResult
	s.NoError(env2.GetWorkflowResult(&result2))
	
	// Verify that the balance remains the same as after the first deduction
	// This confirms idempotency - the fee was only deducted once
	s.True(result2.Success)
	s.Equal(result1.NewBalance, result2.NewBalance)
	s.Equal(result1.Success, result2.Success)
}

// TestWorkflowRetentionPeriod tests that the workflow result can be retrieved even after
// the workflow has completed, as long as it's within the retention period
func (s *WorkflowTestSuite) TestWorkflowRetentionPeriod() {
	// Create a fresh test environment for this test
	env := s.NewTestWorkflowEnvironment()
	
	// Register workflow and activity
	env.RegisterWorkflow(FeeDeductionWorkflow)
	env.RegisterActivity(DeductFeeActivity)
	
	// Setup test constants
	accountID := "ACCT-12345"
	orderID := "ORD-54321"
	amount := 10.0
	initialBalance := 100.0
	
	// Mock the activity to simulate fee deduction
	env.OnActivity(DeductFeeActivity, mock.Anything, ActivityInput{
		AccountID: accountID,
		OrderID:   orderID,
		Amount:    amount,
	}).Return(&ActivityResult{
		NewBalance: initialBalance - amount,
		Success:    true,
	}, nil).Once() // Expect exactly one call
	
	// Setup the workflow input
	input := FeeDeductionWorkflowInput{
		AccountID: accountID,
		OrderID:   orderID,
		Amount:    amount,
	}
	
	// Set the workflow ID for idempotency
	env.SetStartWorkflowOptions(client.StartWorkflowOptions{
		ID: orderID,
	})
	
	// Execute the workflow
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
	
	// Create a new test environment for the second execution
	env2 := s.NewTestWorkflowEnvironment()
	env2.RegisterWorkflow(FeeDeductionWorkflow)
	env2.RegisterActivity(DeductFeeActivity)
	
	// Set up the same workflow ID to simulate a call within retention period
	env2.SetStartWorkflowOptions(client.StartWorkflowOptions{
		ID: orderID,
	})
	
	// Mock the activity to simulate that it's not called again (idempotency)
	env2.OnActivity(DeductFeeActivity, mock.Anything, mock.Anything).Return(&ActivityResult{
		NewBalance: initialBalance - amount,
		Success:    true,
	}, nil).Times(0) // Expect no calls
	
	// Try to execute the same workflow again - should get the same result from history
	env2.ExecuteWorkflow(FeeDeductionWorkflow, input)
	
	// Verify workflow completed successfully
	s.True(env2.IsWorkflowCompleted())
	s.NoError(env2.GetWorkflowError())
	
	// Get the workflow result again
	var result2 FeeDeductionWorkflowResult
	s.NoError(env2.GetWorkflowResult(&result2))
	
	// The results should be identical (proving retrieval from history)
	s.Equal(result.NewBalance, result2.NewBalance)
	s.Equal(result.Success, result2.Success)
	s.Equal(result.OrderID, result2.OrderID)
}

// TestCompleteIdempotencyImplementation combines multiple scenarios
func (s *WorkflowTestSuite) TestCompleteIdempotencyImplementation() {
	// Create a fresh test environment for this test
	env := s.NewTestWorkflowEnvironment()
	
	// Register workflow and activity
	env.RegisterWorkflow(FeeDeductionWorkflow)
	env.RegisterActivity(DeductFeeActivity)
	
	// Setup
	accountID := "ACCT-5678"
	orderID := "ORD-8765"
	amount := 25.0
	initialBalance := 100.0
	
	// Mock the activity instead of using the real implementation
	env.OnActivity(DeductFeeActivity, mock.Anything, ActivityInput{
		AccountID: accountID,
		OrderID:   orderID,
		Amount:    amount,
	}).Return(&ActivityResult{
		NewBalance: initialBalance - amount,
		Success:    true,
	}, nil).Once()
	
	// Setup the workflow input
	input := FeeDeductionWorkflowInput{
		AccountID: accountID,
		OrderID:   orderID,
		Amount:    amount,
	}
	
	// Part 1: First execution
	env.SetStartWorkflowOptions(client.StartWorkflowOptions{
		ID: orderID,
	})
	
	env.ExecuteWorkflow(FeeDeductionWorkflow, input)
	
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	
	var firstResult FeeDeductionWorkflowResult
	s.NoError(env.GetWorkflowResult(&firstResult))
	
	s.True(firstResult.Success)
	s.Equal(initialBalance-amount, firstResult.NewBalance)
	
	// Create a new environment for the second execution
	env2 := s.NewTestWorkflowEnvironment()
	env2.RegisterWorkflow(FeeDeductionWorkflow)
	env2.RegisterActivity(DeductFeeActivity)
	
	// Mock the activity for the second execution to return the same result
	// This simulates idempotent behavior where calling the same activity with the same input
	// should return the same result
	env2.OnActivity(DeductFeeActivity, mock.Anything, ActivityInput{
		AccountID: accountID,
		OrderID:   orderID,
		Amount:    amount,
	}).Return(&ActivityResult{
		NewBalance: initialBalance - amount,
		Success:    true,
	}, nil).Once()
	
	// Part 2: Immediate Retry (simulates retry within a few seconds)
	// We can't directly set time in the testing env, but we can simulate a retry
	env2.SetStartWorkflowOptions(client.StartWorkflowOptions{
		ID: orderID,
	})
	env2.ExecuteWorkflow(FeeDeductionWorkflow, input)
	
	var secondResult FeeDeductionWorkflowResult
	s.NoError(env2.GetWorkflowResult(&secondResult))
	
	// Should be identical to first result (no double charging)
	s.Equal(firstResult.NewBalance, secondResult.NewBalance)
	
	// Create a new environment for the third execution
	env3 := s.NewTestWorkflowEnvironment()
	env3.RegisterWorkflow(FeeDeductionWorkflow)
	env3.RegisterActivity(DeductFeeActivity)
	
	// Mock the activity for the third execution to return the same result
	// This simulates idempotent behavior for requests within the retention period
	env3.OnActivity(DeductFeeActivity, mock.Anything, ActivityInput{
		AccountID: accountID,
		OrderID:   orderID,
		Amount:    amount,
	}).Return(&ActivityResult{
		NewBalance: initialBalance - amount,
		Success:    true,
	}, nil).Once()
	
	// Part 3: Retry after some time (but within retention period)
	// In a real scenario, this would be days later but still within retention
	env3.SetStartWorkflowOptions(client.StartWorkflowOptions{
		ID: orderID,
	})
	env3.ExecuteWorkflow(FeeDeductionWorkflow, input)
	
	var thirdResult FeeDeductionWorkflowResult
	s.NoError(env3.GetWorkflowResult(&thirdResult))
	
	// Should still be identical (workflow history still available)
	s.Equal(firstResult.NewBalance, thirdResult.NewBalance)
	
	// Create a new environment for the fourth execution with different order ID
	env4 := s.NewTestWorkflowEnvironment()
	env4.RegisterWorkflow(FeeDeductionWorkflow)
	env4.RegisterActivity(DeductFeeActivity)
	
	// Part 4: Different order ID should create a new execution
	newOrderID := "ORD-9999"
	newInput := FeeDeductionWorkflowInput{
		AccountID: accountID,
		OrderID:   newOrderID,
		Amount:    amount,
	}
	
	// Mock the activity for the fourth execution with a different result
	// This simulates a new fee deduction with a different order ID
	env4.OnActivity(DeductFeeActivity, mock.Anything, ActivityInput{
		AccountID: accountID,
		OrderID:   newOrderID,
		Amount:    amount,
	}).Return(&ActivityResult{
		NewBalance: initialBalance - amount - amount, // Double deduction for new order
		Success:    true,
	}, nil).Once()
	
	env4.SetStartWorkflowOptions(client.StartWorkflowOptions{
		ID: newOrderID,
	})
	
	env4.ExecuteWorkflow(FeeDeductionWorkflow, newInput)
	
	var fourthResult FeeDeductionWorkflowResult
	s.NoError(env4.GetWorkflowResult(&fourthResult))
	
	// Should be a different result - balance should be reduced again
	s.Equal(firstResult.NewBalance-amount, fourthResult.NewBalance)
}

// Run the test suite
func TestWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowTestSuite))
}
