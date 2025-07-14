package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	superscriptFeature "app/internal/features/superscript"
	"app/internal/superscript"
	"app/internal/worker"
	"app/internal/worker/config"

	"github.com/bitfield/script"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

// logAdapter adapts slog.Logger to Temporal's log.Logger interface
type logAdapter struct {
	logger *slog.Logger
}

func (l *logAdapter) Debug(msg string, keyvals ...interface{}) {
	l.logger.Debug(msg, keyvals...)
}

func (l *logAdapter) Info(msg string, keyvals ...interface{}) {
	l.logger.Info(msg, keyvals...)
}

func (l *logAdapter) Warn(msg string, keyvals ...interface{}) {
	l.logger.Warn(msg, keyvals...)
}

func (l *logAdapter) Error(msg string, keyvals ...interface{}) {
	l.logger.Error(msg, keyvals...)
}

func main() {
	fmt.Println("Welcome to SuperScript Demo using Centralized Worker!")

	// Create structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create Temporal logger adapter
	temporalLogger := &logAdapter{logger: logger}

	// Load configuration with superscript feature enabled
	cfg := config.LoadConfig()
	cfg.EnabledFeatures = []string{"superscript"} // Only enable superscript for this demo
	cfg.HTTPPort = 8080                           // Use port 8080 for HTTP server

	logger.Info("Starting SuperScript demo",
		"temporalHost", cfg.TemporalHost,
		"temporalNamespace", cfg.TemporalNamespace)

	// Create centralized worker
	centralizedWorker, err := worker.NewCentralizedWorker(cfg, temporalLogger)
	if err != nil {
		logger.Error("Failed to create centralized worker", "error", err)
		os.Exit(1)
	}

	// Register superscript feature
	superscriptFeature := superscriptFeature.NewFeature(temporalLogger)
	if err := centralizedWorker.RegisterFeature(superscriptFeature); err != nil {
		logger.Error("Failed to register superscript feature", "error", err)
		os.Exit(1)
	}

	// Start the worker
	if err := centralizedWorker.Start(); err != nil {
		logger.Error("Failed to start centralized worker", "error", err)
		os.Exit(1)
	}

	// Setup HTTP server for demo UI
	mux := http.NewServeMux()

	// Default handler shows SuperScript info
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<html>
			<head><title>SuperScript Demo</title></head>
			<body>
				<h1>SuperScript Demo</h1>
				<p>This is a demo of the SuperScript feature using the centralized worker.</p>
				<p>Worker Status: Running</p>
				<p>Feature: superscript</p>
				<p>Task Queue: superscript_task_queue</p>
				<p>Check the Temporal Web UI at <a href="http://localhost:8080">http://localhost:8080</a></p>
				<h2>Available Endpoints:</h2>
				<ul>
					<li><a href="/health">Health Check</a></li>
					<li><a href="/status">Worker Status</a></li>
				</ul>
			</body>
			</html>
		`)
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "healthy", "service": "superscript-demo", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	// Status endpoint
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status := centralizedWorker.GetStatus()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"worker_status": %t,
			"enabled_features": %v,
			"task_queues": %d,
			"registered_workflows": %d,
			"registered_activities": %d
		}`,
			status["isRunning"],
			status["enabledFeatures"],
			status["taskQueues"],
			status["registeredWorkflows"],
			status["registeredActivities"])
	})

	// API endpoints from original superscript
	mux.HandleFunc("/run/single", func(w http.ResponseWriter, r *http.Request) {
		handleRunSingle(w, r, centralizedWorker.GetClient(), temporalLogger)
	})
	mux.HandleFunc("/run/batch", func(w http.ResponseWriter, r *http.Request) {
		handleRunBatch(w, r, centralizedWorker.GetClient(), temporalLogger)
	})
	mux.HandleFunc("/run/traditional", func(w http.ResponseWriter, r *http.Request) {
		handleRunTraditional(w, r, temporalLogger)
	})

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: mux,
	}

	// Start HTTP server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", "port", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	logger.Info("SuperScript demo started successfully")
	logger.Info("HTTP server", "url", fmt.Sprintf("http://localhost:%d", cfg.HTTPPort))
	logger.Info("Press Ctrl+C to exit...")

	// Wait for interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt

	logger.Info("Shutting down gracefully...")

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Error shutting down HTTP server", "error", err)
	}

	// Stop centralized worker
	centralizedWorker.Stop()

	logger.Info("SuperScript demo shut down")
}

// handleRunSingle starts a single payment collection workflow
func handleRunSingle(w http.ResponseWriter, r *http.Request, c client.Client, logger *logAdapter) {
	var request struct {
		OrderID string `json:"order_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}

	if request.OrderID == "" {
		request.OrderID = superscript.SampleOrderID
	}

	workflowID := fmt.Sprintf("%s-%s", superscript.SinglePaymentWorkflowType, request.OrderID)
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             superscript.SuperscriptTaskQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	workflowRun, err := c.ExecuteWorkflow(r.Context(), workflowOptions, superscript.SinglePaymentCollectionWorkflow, superscript.SinglePaymentWorkflowParams{
		OrderID: request.OrderID,
	})

	if err != nil {
		if temporal.IsWorkflowExecutionAlreadyStartedError(err) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message":      "Workflow already started and handled idempotently",
				"workflow_id":  workflowID,
				"is_duplicate": true,
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"workflow_id": workflowRun.GetID(),
		"run_id":      workflowRun.GetRunID(),
		"status":      "started",
	})
}

// handleRunBatch starts the orchestrator workflow
func handleRunBatch(w http.ResponseWriter, r *http.Request, c client.Client, logger *logAdapter) {
	var request struct {
		OrderIDs []string `json:"order_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}

	if len(request.OrderIDs) == 0 {
		request.OrderIDs = []string{"7307", "5493", "7387", "2614", "5999"}
	}

	workflowID := fmt.Sprintf("%s-%s", superscript.OrchestratorWorkflowType, time.Now().Format("2006-01-02"))
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             superscript.SuperscriptTaskQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}

	workflowRun, err := c.ExecuteWorkflow(r.Context(), workflowOptions, superscript.OrchestratorWorkflow, superscript.OrchestratorWorkflowParams{
		OrderIDs: request.OrderIDs,
		RunDate:  time.Now(),
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Orchestrator workflow started successfully",
		"workflow_id": workflowRun.GetID(),
		"run_id":      workflowRun.GetRunID(),
		"status":      "running",
	})
}

// handleRunTraditional executes the traditional script directly
func handleRunTraditional(w http.ResponseWriter, r *http.Request, logger *logAdapter) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Running traditional script directly (non-idempotent)...\n"))
	w.Write([]byte("Check server logs for script output\n"))

	go func() {
		cmd := fmt.Sprintf("%s", superscript.TraditionalBatchScriptPath)
		output, err := script.Exec(cmd).String()
		if err != nil {
			logger.Error("Script execution failed", "error", err, "output", output)
		} else {
			logger.Info("Script execution completed", "output", output)
		}
	}()
}
