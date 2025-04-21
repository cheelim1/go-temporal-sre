package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	data_enrichment "app/internal/features/data-enrichment"
	"app/internal/worker"
	"app/internal/worker/config"
)

func main() {
	fmt.Println("Welcome to Data Enrichment Demo ...")
	Run()
}

// debugAccessHandler handles requests to the debug endpoint
func debugAccessHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Data Enrichment Debug Page\n")
	fmt.Fprintf(w, "=========================\n")
	fmt.Fprintf(w, "Status: Running\n")
	fmt.Fprintf(w, "Task Queue: %s\n", data_enrichment.TQ)
}

func Run() {
	// Create worker configuration
	cfg := config.DefaultConfig().
		WithTaskQueue(data_enrichment.TQ).
		WithActivityTimeout(5 * time.Second)

	// Create worker
	w, err := worker.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}

	// Register workflows and activities
	data_enrichment.RegisterWorkflows(w)

	// Start worker in a goroutine
	go func() {
		if err := w.Start(); err != nil {
			log.Printf("Worker error: %v", err)
		}
	}()

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Define a handler function
	defaultHandler := func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/demo/debug/", http.StatusFound)
		return
	}

	// Attach handler function to the ServeMux
	mux.HandleFunc("/", defaultHandler)
	mux.HandleFunc("/demo/debug/", debugAccessHandler)

	// HTTP Server Setup
	server := &http.Server{
		Addr:    ":8888",
		Handler: mux,
	}

	// Start HTTP server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil {
			fmt.Println("Server error:", err)
		}
	}()

	// Prepare for handling signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	interruptSignal := <-interrupt
	fmt.Printf("Received %s, shutting down.\n", interruptSignal)

	// Shutdown the server gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("Error shutting down:", err)
	} else {
		fmt.Println("Server shutdown gracefully.")
	}

	// Stop the worker
	w.Stop()
}
