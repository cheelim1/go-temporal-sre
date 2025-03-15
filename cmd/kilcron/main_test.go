package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestServerResponseWithJitterAndHeaderCount(t *testing.T) {
	// Initialize a slice of words to be used as responses
	words := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon"}

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Define a handler function that changes behavior based on the count received from the header
	h1 := func(w http.ResponseWriter, r *http.Request) {
		// Extract count from the HTTP header
		countHeader := r.Header.Get("X-Request-Count")
		count, err := strconv.Atoi(countHeader)
		if err != nil {
			// Handle the error if the count is not a valid integer
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Invalid request count")
			return
		}

		// Introduce a random delay of 1-5 seconds for the first five responses
		if count <= 5 {
			delay := time.Duration(rand.Intn(5)+1) * time.Second
			time.Sleep(delay)
			responseWord := words[count-1]
			fmt.Println("RESPONSE: ID:", count, "AFTER: ", delay, " OUTPUT:", responseWord)
			fmt.Fprint(w, responseWord)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable) // 503 Service Unavailable
			fmt.Println("RESPONSE: ERROR!!!")
			fmt.Fprint(w, "Server error")
		}
	}

	// Attach handler function to the ServeMux
	mux.HandleFunc("/", h1)

	// Create a test server using the ServeMux
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Create a custom HTTP client with a longer timeout
	client := &http.Client{
		Timeout: 10 * time.Second, // Adjust the timeout to be longer than the maximum expected delay
	}

	// Use a wait group to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Test the response for the first five calls and then the sixth call
	for i := 1; i <= 6; i++ {
		wg.Add(1)
		go func(callNumber int) {
			defer wg.Done()

			req, err := http.NewRequest("GET", ts.URL, nil)
			if err != nil {
				t.Fatalf("Error creating request: %v", err)
			}
			// Set the request count in the header
			req.Header.Set("X-Request-Count", strconv.Itoa(callNumber))

			startTime := time.Now()
			res, err := client.Do(req)
			if err != nil {
				t.Fatalf("Unexpected error on call %d: %v", callNumber, err)
			}
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("Unexpected error reading response body on call %d: %v", callNumber, err)
			}

			// Check the elapsed time for the first five calls
			elapsed := time.Since(startTime)
			if callNumber <= 5 {
				if elapsed < time.Second || elapsed > 6*time.Second {
					t.Errorf("Handler did not delay for the expected time on call %d: delayed for %v", callNumber, elapsed)
				}
				expected := words[callNumber-1]
				if string(body) != expected {
					t.Errorf("Handler returned unexpected body on call %d: got %v want %v",
						callNumber, string(body), expected)
				}
			} else {
				// Expect a server error on the sixth call
				if res.StatusCode != http.StatusServiceUnavailable {
					t.Errorf("Expected StatusCode %d on call %d, got %d", http.StatusServiceUnavailable, callNumber, res.StatusCode)
				}
				expected := "Server error"
				if string(body) != expected {
					t.Errorf("Handler returned unexpected body on call %d: got %v want %v",
						callNumber, string(body), expected)
				}
			}
		}(i)
	}

	wg.Wait() // Wait for all goroutines to complete
}
