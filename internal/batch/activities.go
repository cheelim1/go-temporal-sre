package batch

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// Account represents a user account with a balance
type Account struct {
	ID      string
	Balance float64
}

// AccountStore is a simple in-memory store for accounts
type AccountStore struct {
	mu       sync.Mutex
	accounts map[string]*Account
}

// NewAccountStore creates a new account store
func NewAccountStore() *AccountStore {
	return &AccountStore{
		accounts: make(map[string]*Account),
	}
}

// GetAccount retrieves an account by ID
func (s *AccountStore) GetAccount(id string) (*Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, exists := s.accounts[id]
	if !exists {
		return nil, fmt.Errorf("account %s not found", id)
	}
	return account, nil
}

// CreateAccount creates a new account with the given ID and initial balance
func (s *AccountStore) CreateAccount(id string, initialBalance float64) *Account {
	s.mu.Lock()
	defer s.mu.Unlock()

	account := &Account{
		ID:      id,
		Balance: initialBalance,
	}
	s.accounts[id] = account
	return account
}

// DeductFee deducts a fee from an account after a random delay
// This function is NOT idempotent, causing double-deduction issues
func (s *AccountStore) DeductFee(accountID, orderID string, amount float64) (float64, error) {
	// Introduce a random delay between 200ms and 2s
	delayMs := 200 + rand.Intn(1800)
	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	// Lock the store for the entire operation to prevent race conditions
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the account directly from the map (we're already locked)
	account, exists := s.accounts[accountID]
	if !exists {
		return 0, fmt.Errorf("account %s not found", accountID)
	}

	// Deduct the fee
	if account.Balance < amount {
		return account.Balance, fmt.Errorf("insufficient funds: balance %.2f, fee %.2f", account.Balance, amount)
	}

	account.Balance -= amount
	return account.Balance, nil
}

// FeeDeductionRequest represents the JSON payload for fee deduction
type FeeDeductionRequest struct {
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
}

// FeeDeductionResponse represents the JSON response after fee deduction
type FeeDeductionResponse struct {
	NewBalance float64 `json:"new_balance"`
	Success    bool    `json:"success"`
	Message    string  `json:"message,omitempty"`
}

// HTTPHandler returns a handler for the DeductFee endpoint
func DeductFeeHTTPHandler(store *AccountStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set content type for all responses
		w.Header().Set("Content-Type", "application/json")

		// Only allow POST method
		if r.Method != http.MethodPost {
			respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract orderID from path - assumes path format like "/deduct-fee/ORD-12345"
		orderID := r.URL.Path[len("/deduct-fee/"):]
		if orderID == "" {
			respondWithError(w, "Order ID is required", http.StatusBadRequest)
			return
		}

		// Parse request body
		var req FeeDeductionRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Unmarshal JSON
		if err := json.Unmarshal(body, &req); err != nil {
			respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		// Validate request fields
		if req.AccountID == "" {
			respondWithError(w, "Account ID is required", http.StatusBadRequest)
			return
		}

		if req.Amount <= 0 {
			respondWithError(w, "Amount must be greater than zero", http.StatusBadRequest)
			return
		}

		// Deduct the fee
		newBalance, err := store.DeductFee(req.AccountID, orderID, req.Amount)
		if err != nil {
			respondWithError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return the new balance
		response := FeeDeductionResponse{
			NewBalance: newBalance,
			Success:    true,
			Message:    "Fee deducted successfully",
		}

		respondWithJSON(w, response, http.StatusOK)
	}
}

// Helper function to respond with an error
func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	response := FeeDeductionResponse{
		Success: false,
		Message: message,
	}
	respondWithJSON(w, response, statusCode)
}

// Helper function to respond with JSON
func respondWithJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.WriteHeader(statusCode)
	responseJSON, err := json.Marshal(data)
	if err != nil {
		// If we can't marshal the response, fall back to a simple error
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"success":false,"message":"Internal server error"}`) 
		return
	}
	w.Write(responseJSON)
}
