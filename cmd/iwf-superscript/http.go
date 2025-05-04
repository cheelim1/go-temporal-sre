package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"app/internal/iwf-superscript"
	"app/internal/superscript"

	"github.com/bitfield/script"
	"github.com/indeedeng/iwf-golang-sdk/iwf"
)

// Server represents an HTTP server for the iWF SuperScript API
type Server struct {
	Client       iwf.Client
	Logger       iwf.Logger
	Addr         string
	Server       *http.Server
}

// NewServer creates a new HTTP server
func NewServer(iwfClient iwf.Client, logger iwf.Logger, addr string) *Server {
	return &Server{
		Client:       iwfClient,
		Logger:       logger,
		Addr:         addr,
	}
}

// ListenAndServe begins the HTTP server
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("POST /run/single", s.handleRunSingle)
	mux.HandleFunc("POST /run/batch", s.handleRunBatch)
	mux.HandleFunc("GET /run/traditional", s.handleRunTraditional)

	s.Server = &http.Server{
		Addr:    s.Addr,
		Handler: mux,
	}

	s.Logger.Info("Starting HTTP server", "addr", s.Addr)
	return s.Server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	s.Logger.Info("Stopping HTTP server")
	return s.Server.Shutdown(ctx)
}

// handleHealth is a simple health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleRunSingle starts a single payment collection workflow
func (s *Server) handleRunSingle(w http.ResponseWriter, r *http.Request) {
	var request struct {
		OrderID string `json:"order_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}

	if request.OrderID == "" {
		// Use sample order ID if none provided
		request.OrderID = superscript.SampleOrderID
	}

	// Create a workflow ID based on the order ID
	workflowID := fmt.Sprintf("%s-%s", iwfsuperscript.SinglePaymentWorkflowType, request.OrderID)

	// Prepare workflow options
	// WorkflowID and TaskQueue are not set here. ID is passed directly to StartWorkflow.
	// TaskQueue is configured on the worker.
	options := &iwf.WorkflowOptions{
		WorkflowIdReusePolicy: iwf.WorkflowIdReusePolicyRejectDuplicate,
	}

	s.Logger.Info("Starting single payment workflow", "orderID", request.OrderID, "workflowID", workflowID)

	// Instantiate the workflow struct
	singlePaymentWorkflow := &iwfsuperscript.SinglePaymentWorkflow{} // Assuming struct name
	workflowInput := superscript.SinglePaymentWorkflowParams{
		OrderID: request.OrderID,
	}
	startTimeout := int32(10) // 10 seconds timeout

	// Start the workflow
	// Use correct signature: StartWorkflow(ctx, workflowObject, workflowID, timeout, input, options)
	runID, err := s.Client.StartWorkflow(r.Context(), singlePaymentWorkflow, workflowID, startTimeout, workflowInput, options)

	if err != nil {
		// Updated error check function namespace
		if iwf.IsWorkflowAlreadyStartedError(err) {
			// This is expected when calling the same workflow ID multiple times
			s.Logger.Info("Workflow already started - retrieving existing run", "workflowID", workflowID)

			// Return response indicating that this was a duplicate request that was handled idempotently
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message":      "Workflow already started and handled idempotently",
				"workflow_id":  workflowID,
				// "run_id" could potentially be retrieved here if needed, but we don't have it directly
			})
			return
		}

		s.Logger.Error("Failed to start workflow", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Wait for workflow completion
	var result superscript.PaymentResult
	// Use GetSimpleWorkflowResult (assuming this is the correct method)
	if err := s.Client.GetSimpleWorkflowResult(r.Context(), workflowID, runID, &result); err != nil {
		s.Logger.Error("Workflow execution failed", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Return the workflow result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"workflow_id": workflowID,
		"run_id":      runID, // runID is returned directly by StartWorkflow
		"result":      result,
	})
}

// handleRunBatch starts the orchestrator workflow to process multiple orders
func (s *Server) handleRunBatch(w http.ResponseWriter, r *http.Request) {
	var request struct {
		OrderIDs []string `json:"order_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}

	if len(request.OrderIDs) == 0 {
		// Use default order IDs if none provided (use same IDs as in traditional script)
		request.OrderIDs = []string{"7307", "5493", "7387", "2614", "5999", "3078", "8577", "5479", "6606", "8448"}
	}

	// Create a workflow ID based on the current date
	currentDate := time.Now().Format("2006-01-02")
	workflowID := fmt.Sprintf("%s-%s", iwfsuperscript.OrchestratorWorkflowType, currentDate)

	// Prepare workflow options
	// WorkflowID and TaskQueue are not set here.
	options := &iwf.WorkflowOptions{
		WorkflowIdReusePolicy: iwf.WorkflowIdReusePolicyAllowDuplicate,
	}

	s.Logger.Info("Starting orchestrator workflow", "orderCount", len(request.OrderIDs), "workflowID", workflowID)

	// Instantiate the workflow struct
	orchestratorWorkflow := &iwfsuperscript.OrchestratorWorkflow{} // Assuming struct name
	workflowInput := superscript.OrchestratorWorkflowParams{
		OrderIDs: request.OrderIDs,
		RunDate:  time.Now(),
	}
	startTimeout := int32(10) // 10 seconds timeout

	// Start the workflow
	// Use correct signature: StartWorkflow(ctx, workflowObject, workflowID, timeout, input, options)
	runID, err := s.Client.StartWorkflow(r.Context(), orchestratorWorkflow, workflowID, startTimeout, workflowInput, options)

	if err != nil {
		s.Logger.Error("Failed to start workflow", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Return immediately with the workflow ID so user can track progress
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Orchestrator workflow started successfully",
		"workflow_id": workflowID,
		"run_id":      runID, // runID is returned directly by StartWorkflow
		"status":      "running",
	})
}

// handleRunTraditional executes the traditional non-idempotent script directly
// This is used for demonstration purposes to show the non-idempotent behavior
func (s *Server) handleRunTraditional(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("order_id")
	if orderID == "" {
		orderID = superscript.SampleOrderID // Use default if not provided
	}

	scriptPath := "./internal/superscript/payment_collection.sh"
	pipe := script.Exec(fmt.Sprintf("%s %s", scriptPath, orderID))

	output, err := pipe.String()
	exitCode := pipe.ExitStatus()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		s.Logger.Error("Traditional script execution failed", "orderID", orderID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"order_id": orderID,
			"success":  false,
			"output":   output,
			"exitCode": exitCode,
			"error":    err.Error(),
		})
		return
	}

	s.Logger.Info("Traditional script executed", "orderID", orderID, "exitCode", exitCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"order_id": orderID,
		"success":  exitCode == 0,
		"output":   output,
		"exitCode": exitCode,
	})
}

// WaitForInterrupt blocks until an interrupt signal is received
func WaitForInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
