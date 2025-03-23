package batch

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// Account represents a user account with a balance
type Account struct {
	ID      string
	Balance float64
}

// AccountStore is a simple in-memory store for accounts
type AccountStore struct {
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
	account, exists := s.accounts[id]
	if !exists {
		return nil, fmt.Errorf("account %s not found", id)
	}
	return account, nil
}

// CreateAccount creates a new account with the given ID and initial balance
func (s *AccountStore) CreateAccount(id string, initialBalance float64) *Account {
	account := &Account{
		ID:      id,
		Balance: initialBalance,
	}
	s.accounts[id] = account
	return account
}

// DeductFee deducts a fee from an account after a random delay
// This function is NOT idempotent, causing double-deduction issues
func DeductFee(store *AccountStore, accountID, orderID string, amount float64) (float64, error) {
	// Introduce a random delay between 200ms and 2s
	delayMs := 200 + rand.Intn(1800)
	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	// Get the account
	account, err := store.GetAccount(accountID)
	if err != nil {
		return 0, err
	}

	// Deduct the fee
	if account.Balance < amount {
		return account.Balance, fmt.Errorf("insufficient funds: balance %.2f, fee %.2f", account.Balance, amount)
	}

	account.Balance -= amount
	return account.Balance, nil
}

// HTTPHandler returns a handler for the DeductFee endpoint
func DeductFeeHTTPHandler(store *AccountStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract orderID from path - assumes path format like "/deduct-fee/ORD-12345"
		orderID := r.URL.Path[len("/deduct-fee/"):]
		if orderID == "" {
			http.Error(w, "Order ID is required", http.StatusBadRequest)
			return
		}

		// For simplicity, we'll use a fixed account ID and amount
		accountID := "ACCT-12345"
		amount := 10.0 // USD 10

		// Deduct the fee
		newBalance, err := DeductFee(store, accountID, orderID, amount)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return the new balance
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%.2f", newBalance)
	}
}
