package main

import (
	"app/internal/iwf-superscript"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/indeedeng/iwf-golang-sdk/gen/iwfidl"
	"github.com/indeedeng/iwf-golang-sdk/iwf"
)

const (
	defaultApiPort    = "8081"
	defaultWorkerPort = "8802"             // Default iWF worker port
	workflowType      = "YourWorkflowType" // Replace with actual type if needed by client
)

//// simpleLogger adapts to iwf.Logger
//type simpleLogger struct{}
//
//// Debugf implements iwf.Logger
//func (l *simpleLogger) Debug(msg string, keyvals ...interface{}) {
//	slog.Debug(msg, keyvals...)
//}
//
//// Infof implements iwf.Logger
//func (l *simpleLogger) Info(msg string, keyvals ...interface{}) {
//	slog.Info(msg, keyvals...)
//}
//
//// Warnf implements iwf.Logger
//func (l *simpleLogger) Warn(msg string, keyvals ...interface{}) {
//	slog.Warn(msg, keyvals...)
//}
//
//// Errorf implements iwf.Logger
//func (l *simpleLogger) Error(msg string, keyvals ...interface{}) {
//	slog.Error(msg, keyvals...)
//}

func main() {
	err := run()
	if err != nil {
		slog.Error("Application run failed", "error", err)
		os.Exit(1)
	}
	slog.Info("Application finished gracefully")
}

func run() error {
	logger := slog.Logger{}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil))) // Configure slog

	// --- Worker Setup ---
	iwfServerUrl := "http://localhost:8801" // Default iWF server URL
	workerPort := "8802"
	workerUrl := "http://localhost:" + workerPort
	logger.Info("Setting up iWF worker", "iwfServerUrl", iwfServerUrl, "workerUrl", workerUrl)

	// Define worker options
	workerOptions := iwf.WorkerOptions{
		ObjectEncoder: iwf.GetDefaultObjectEncoder(),
	}

	// Call RegisterWorkflows to get the configured worker service and registry
	// Pass the logger instance directly
	workerService, registry := iwfsuperscript.RegisterWorkflows(workerOptions, "", logger)

	// Create iWF Client using the registry from the worker service
	clientOptions := iwf.ClientOptions{
		ServerUrl:     iwfServerUrl,
		WorkerUrl:     workerUrl,
		ObjectEncoder: iwf.GetDefaultObjectEncoder(),
	}
	iwfClient := iwf.NewClient(registry, &clientOptions) // Use registry from worker setup

	// --- API Server Setup ---
	apiPort := "8081" // Port for the API endpoints
	apiAddr := ":" + apiPort
	logger.Info("Setting up API server", "port", apiPort)
	apiServerWrapper := NewServer(iwfClient, logger, apiAddr) // Pass iwfClient

	// --- Worker Server Setup ---
	workerAddr := ":" + workerPort
	logger.Info("Setting up Worker HTTP server", "port", workerPort)

	workerServerMux := http.NewServeMux()

	// Define handler functions for iWF APIs
	handleStateWaitUntil := func(w http.ResponseWriter, r *http.Request) {
		var req iwfidl.WorkflowStateWaitUntilRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "error decoding request: %v", err)
			return
		}

		resp, err := workerService.HandleWorkflowStateWaitUntil(r.Context(), req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error handling state wait until: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Error encoding response for state wait until", "error", err)
		}
	}

	handleStateExecute := func(w http.ResponseWriter, r *http.Request) {
		var req iwfidl.WorkflowStateExecuteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "error decoding request: %v", err)
			return
		}

		resp, err := workerService.HandleWorkflowStateExecute(r.Context(), req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error handling state execute: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Error encoding response for state execute", "error", err)
		}
	}

	handleWorkerRpc := func(w http.ResponseWriter, r *http.Request) {
		var req iwfidl.WorkflowWorkerRpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "error decoding request: %v", err)
			return
		}

		resp, err := workerService.HandleWorkflowWorkerRPC(r.Context(), req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error handling worker rpc: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Error encoding response for worker rpc", "error", err)
		}
	}

	// Register handlers
	workerServerMux.HandleFunc(iwf.WorkflowStateWaitUntilApi, handleStateWaitUntil)
	workerServerMux.HandleFunc(iwf.WorkflowStateExecuteApi, handleStateExecute)
	workerServerMux.HandleFunc(iwf.WorkflowWorkerRPCAPI, handleWorkerRpc)

	workerServer := &http.Server{
		Addr:    workerAddr,
		Handler: workerServerMux, // Use the mux with registered handlers
	}

	// --- Server Lifecycle Management ---
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errChan := make(chan error, 2) // Buffer for errors from both servers

	// Start API server in a goroutine
	go func() {
		logger.Info("Starting API server", "addr", apiServerWrapper.Addr)
		if err := apiServerWrapper.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("API server error", "error", err)
			errChan <- fmt.Errorf("API server error: %w", err)
		} else {
			logger.Info("API server stopped cleanly")
			errChan <- nil // Signal clean exit
		}
		logger.Info("API server stopped listening", "addr", apiServerWrapper.Addr) // Use Addr field
	}()

	// Start Worker server in a goroutine
	go func() {
		logger.Info("Starting Worker HTTP server", "addr", workerServer.Addr)
		if err := workerServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Worker server error", "error", err)
			errChan <- fmt.Errorf("worker server error: %w", err)
		} else {
			logger.Info("Worker server stopped cleanly")
			errChan <- nil // Signal clean exit
		}
		logger.Info("Worker server stopped listening", "addr", workerServer.Addr)
	}()

	// Wait for interrupt signal or server error
	select {
	case <-ctx.Done():
		logger.Info("Shutdown signal received")
		stop()
	case err := <-errChan:
		if err != nil {
			logger.Error("Server failed", "error", err)
		} else {
			logger.Info("One of the servers stopped cleanly, initiating shutdown")
		}
		stop() // Ensure context is cancelled even if one server stops cleanly
	}

	// --- Graceful Shutdown ---
	logger.Info("Shutting down servers...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		logger.Info("Shutting down API server")
		if err := apiServerWrapper.Shutdown(shutdownCtx); err != nil {
			logger.Error("API server shutdown error", "error", err)
		} else {
			logger.Info("API server shutdown complete")
		}
	}()

	go func() {
		defer wg.Done()
		logger.Info("Shutting down Worker HTTP server")
		if err := workerServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("Worker server shutdown error", "error", err)
		} else {
			logger.Info("Worker server shutdown complete")
		}
	}()

	wg.Wait()
	logger.Info("All servers shutdown complete")

	// Optionally stop the worker service itself if needed (check SDK docs)
	// workerService.Stop() ?

	logger.Info("Application exiting")
	return nil
}
