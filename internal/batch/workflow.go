package batch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// FeeDeductionWorkflowInput represents the input parameters for the FeeDeductionWorkflow
type FeeDeductionWorkflowInput struct {
	AccountID string  `json:"account_id"`
	OrderID   string  `json:"order_id"`
	Amount    float64 `json:"amount"`
}

// FeeDeductionWorkflowResult represents the result of the FeeDeductionWorkflow
type FeeDeductionWorkflowResult struct {
	NewBalance float64   `json:"new_balance"`
	Success    bool      `json:"success"`
	Message    string    `json:"message,omitempty"`
	OrderID    string    `json:"order_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// ActivityInput represents the input for DeductFeeActivity
type ActivityInput struct {
	AccountID string  `json:"account_id"`
	OrderID   string  `json:"order_id"`
	Amount    float64 `json:"amount"`
}

// ActivityResult represents the result from DeductFeeActivity
type ActivityResult struct {
	NewBalance float64 `json:"new_balance"`
	Success    bool    `json:"success"`
	Error      string  `json:"error,omitempty"`
}

// FeeDeductionWorkflow is a workflow that deducts a fee from an account in an idempotent manner
func FeeDeductionWorkflow(ctx workflow.Context, input FeeDeductionWorkflowInput) (*FeeDeductionWorkflowResult, error) {
	// Logger
	logger := workflow.GetLogger(ctx)
	logger.Info("FeeDeductionWorkflow started", "OrderID", input.OrderID)

	// Activity options
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	// Prepare result
	result := &FeeDeductionWorkflowResult{
		OrderID:   input.OrderID,
		Timestamp: workflow.Now(ctx),
	}

	// Prepare activity input
	activityInput := ActivityInput{
		AccountID: input.AccountID,
		OrderID:   input.OrderID,
		Amount:    input.Amount,
	}

	// Execute activity
	var activityResult ActivityResult
	err := workflow.ExecuteActivity(ctx, DeductFeeActivity, activityInput).Get(ctx, &activityResult)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Activity execution failed: %v", err)
		return result, err
	}

	// Set the result based on activity result
	result.NewBalance = activityResult.NewBalance
	result.Success = activityResult.Success
	if !activityResult.Success {
		result.Message = activityResult.Error
	} else {
		result.Message = "Fee deduction successful"
	}

	logger.Info("FeeDeductionWorkflow completed", "OrderID", input.OrderID, "Success", result.Success)
	return result, nil
}

// DeductFeeActivity is an activity that deducts a fee from an account
// It makes an HTTP call to the fee deduction endpoint to ensure idempotency
func DeductFeeActivity(ctx context.Context, input ActivityInput) (*ActivityResult, error) {
	// In a real implementation, this would use a configuration for the service URL
	// For now, we'll assume the service is running locally
	serviceURL := "http://localhost:8080/deduct-fee"
	
	// Create a client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	// Create the request payload
	payload := FeeDeductionRequest{
		AccountID: input.AccountID,
		Amount:    input.Amount,
	}
	
	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return &ActivityResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to marshal JSON: %v", err),
		}, nil
	}
	
	// Create request with OrderID in path for idempotency
	url := fmt.Sprintf("%s/%s", serviceURL, input.OrderID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return &ActivityResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to create request: %v", err),
		}, nil
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return &ActivityResult{
			Success: false,
			Error:   fmt.Sprintf("HTTP request failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ActivityResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to read response: %v", err),
		}, nil
	}
	
	// Check status code
	if resp.StatusCode != http.StatusOK {
		return &ActivityResult{
			Success: false,
			Error:   fmt.Sprintf("Unexpected status code: %d, body: %s", resp.StatusCode, string(body)),
		}, nil
	}
	
	// Parse response
	var response FeeDeductionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return &ActivityResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse response: %v", err),
		}, nil
	}
	
	// Check if the operation was successful
	if !response.Success {
		return &ActivityResult{
			Success: false,
			Error:   response.Message,
		}, nil
	}
	
	// Return the successful result
	return &ActivityResult{
		NewBalance: response.NewBalance,
		Success:    true,
	}, nil
}
