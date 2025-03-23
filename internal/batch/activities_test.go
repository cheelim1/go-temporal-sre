package batch

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNonIdempotency(t *testing.T) {
	// Create a new account store and initialize an account with USD 100
	store := NewAccountStore()
	accountID := "ACCT-12345"
	orderID := "ORD-12345"
	initialBalance := 100.0
	store.CreateAccount(accountID, initialBalance)

	// Set up the HTTP handler for deducting fees
	handler := DeductFeeHTTPHandler(store)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer server.Close()

	// Create a WaitGroup to wait for both requests to complete
	var wg sync.WaitGroup
	wg.Add(2)

	// Variables to store results
	var responses [2]*http.Response
	var durations [2]time.Duration
	var errors [2]error
	var responseData [2]FeeDeductionResponse

	// Create the request payload
	payload := FeeDeductionRequest{
		AccountID: accountID,
		Amount:    10.0, // USD 10
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal request payload: %v", err)
	}

	// Function to make a request
	makeRequest := func(index int) {
		defer wg.Done()
		
		start := time.Now()
		resp, err := http.Post(
			server.URL+"/deduct-fee/"+orderID, 
			"application/json", 
			bytes.NewBuffer(payloadBytes),
		)
		duration := time.Since(start)
		
		responses[index] = resp
		durations[index] = duration
		errors[index] = err
		
		// Parse response body if no error
		if err == nil && resp != nil {
			body, readErr := io.ReadAll(resp.Body)
			if readErr == nil {
				json.Unmarshal(body, &responseData[index])
			}
		}
	}

	// Start the overall timer
	overallStart := time.Now()

	// Make the first request
	go makeRequest(0)

	// Wait 100ms before making the second request
	time.Sleep(100 * time.Millisecond)

	// Make the second request
	go makeRequest(1)

	// Wait for both requests to complete
	wg.Wait()
	overallDuration := time.Since(overallStart)

	// Check if there were any errors
	for i, err := range errors {
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
	}

	// Check if both responses are OK (HTTP 200)
	for i, resp := range responses {
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Request %d returned status code %d, expected %d", 
				i+1, resp.StatusCode, http.StatusOK)
		}
		defer resp.Body.Close()
		
		// Verify response data
		if !responseData[i].Success {
			t.Fatalf("Request %d reported failure: %s", i+1, responseData[i].Message)
		}
	}

	// Check the final account balance
	account, err := store.GetAccount(accountID)
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	expectedBalance := initialBalance - 20.0 // Two deductions of USD 10 each
	if account.Balance != expectedBalance {
		t.Errorf("Expected account balance to be %.2f, got %.2f", expectedBalance, account.Balance)
	} else {
		t.Logf("Account balance correctly reduced to %.2f after two non-idempotent calls", account.Balance)
	}

	// Log timing information
	t.Logf("First request took: %v", durations[0])
	t.Logf("Second request took: %v", durations[1])
	t.Logf("Overall operation took: %v", overallDuration)
}
