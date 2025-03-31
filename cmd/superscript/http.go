package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"app/internal/superscript"

	"github.com/bitfield/script"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
)

// Server represents an HTTP server for the Superscript API
type Server struct {
	Client client.Client
	Logger log.Logger
	Addr   string
	Server *http.Server
}

// NewServer creates a new HTTP server
func NewServer(c client.Client, logger log.Logger, addr string) *Server {
	return &Server{
		Client: c,
		Logger: logger,
		Addr:   addr,
	}
}

// Start begins the HTTP server
func (s *Server) Start() error {
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

// Stop gracefully shuts down the HTTP server
func (s *Server) Stop(ctx context.Context) error {
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
	workflowID := fmt.Sprintf("%s-%s", superscript.SinglePaymentWorkflowType, request.OrderID)

	// Start the workflow with idempotency guaranteed by Temporal
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: superscript.SuperscriptTaskQueue,
		// Reject duplicate ensures idempotency
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	s.Logger.Info("Starting single payment workflow", "orderID", request.OrderID, "workflowID", workflowID)

	workflowRun, err := s.Client.ExecuteWorkflow(r.Context(), workflowOptions, superscript.SinglePaymentCollectionWorkflow, superscript.SinglePaymentWorkflowParams{
		OrderID: request.OrderID,
	})

	if err != nil {
		// DEBUG
		//spew.Dump(err)
		if temporal.IsWorkflowExecutionAlreadyStartedError(err) {
			// This is expected when calling the same workflow ID multiple times
			// We can get the existing run and return its info
			s.Logger.Info("Workflow already started - retrieving existing run", "workflowID", workflowID)

			// Return response indicating that this was a duplicate request that was handled idempotently
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message":      "Workflow already started and handled idempotently",
				"workflow_id":  workflowID,
				"is_duplicate": true,
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
	if err := workflowRun.Get(r.Context(), &result); err != nil {
		s.Logger.Error("Workflow execution failed", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Return the workflow result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"workflow_id": workflowRun.GetID(),
		"run_id":      workflowRun.GetRunID(),
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
		// Full test below
		request.OrderIDs = []string{"7307", "5493", "7387", "2614", "5999", "3078", "8577", "5479", "6606", "8448"}
		// Below shortened
		//request.OrderIDs = []string{"7307", "5493"}
	}

	// Create a workflow ID based on the current date
	currentDate := time.Now().Format("2006-01-02")
	workflowID := fmt.Sprintf("%s-%s", superscript.OrchestratorWorkflowType, currentDate)

	// Start the orchestrator workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: superscript.SuperscriptTaskQueue,
		// We use AllowDuplicate here since we might want to run multiple batches per day
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}

	s.Logger.Info("Starting orchestrator workflow", "orderCount", len(request.OrderIDs), "workflowID", workflowID)

	workflowRun, err := s.Client.ExecuteWorkflow(r.Context(),
		workflowOptions,
		superscript.OrchestratorWorkflow,
		superscript.OrchestratorWorkflowParams{
			OrderIDs: request.OrderIDs,
			RunDate:  time.Now(),
		})

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
		"workflow_id": workflowRun.GetID(),
		"run_id":      workflowRun.GetRunID(),
		"status":      "running",
	})
}

// handleRunTraditional executes the traditional non-idempotent script directly
// This is used for demonstration purposes to show the non-idempotent behavior
func (s *Server) handleRunTraditional(w http.ResponseWriter, r *http.Request) {
	// Execute the traditional script directly using the script library
	s.Logger.Info("Executing traditional script directly")

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Running traditional script directly (non-idempotent)...\n"))
	w.Write([]byte("Check server logs for script output\n"))

	// Execute the script asynchronously to not block the HTTP response
	go func() {
		cmd := fmt.Sprintf("%s", superscript.TraditionalBatchScriptPath)
		s.Logger.Info("Running command", "cmd", cmd)

		// Use the script package to execute the command and get the output
		output, err := script.Exec(cmd).String()
		if err != nil {
			s.Logger.Error("Script execution failed", "error", err, "output", output)
		} else {
			s.Logger.Info("Script execution completed", "output", output)
		}
	}()
}
