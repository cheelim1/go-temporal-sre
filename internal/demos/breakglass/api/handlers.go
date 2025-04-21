package api

import (
	"encoding/json"
	"net/http"
	"time"

	"app/internal/demos/breakglass"

	"go.temporal.io/sdk/client"
)

// Handler handles HTTP requests for the breakglass API
type Handler struct {
	temporalClient client.Client
}

// NewHandler creates a new API handler
func NewHandler(temporalClient client.Client) *Handler {
	return &Handler{
		temporalClient: temporalClient,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/breakglass":
		if r.Method == http.MethodPost {
			h.handleBreakglass(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/api/breakglass/status":
		if r.Method == http.MethodGet {
			h.handleStatus(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.NotFound(w, r)
	}
}

// handleBreakglass handles breakglass action requests
func (h *Handler) handleBreakglass(w http.ResponseWriter, r *http.Request) {
	var req BreakglassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start workflow
	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "breakglass-task-queue",
		ID:        req.ServiceID + "-" + time.Now().Format("20060102150405"),
	}

	workflowInput := breakglass.BreakglassWorkflowInput{
		ServiceID:   req.ServiceID,
		Action:      req.Action,
		Parameters:  req.Parameters,
		RequestedBy: req.RequestedBy,
	}

	workflowRun, err := h.temporalClient.ExecuteWorkflow(r.Context(), workflowOptions, breakglass.BreakglassWorkflow, workflowInput)
	if err != nil {
		http.Error(w, "Failed to start workflow", http.StatusInternalServerError)
		return
	}

	// Return response
	resp := BreakglassResponse{
		WorkflowID:  workflowRun.GetID(),
		ServiceID:   req.ServiceID,
		Action:      req.Action,
		Status:      "STARTED",
		RequestedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleStatus handles workflow status requests
func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	workflowID := r.URL.Query().Get("workflow_id")
	if workflowID == "" {
		http.Error(w, "Workflow ID is required", http.StatusBadRequest)
		return
	}

	// Get workflow status
	workflowRun := h.temporalClient.GetWorkflow(r.Context(), workflowID, "")
	var result breakglass.BreakglassWorkflowResult
	if err := workflowRun.Get(r.Context(), &result); err != nil {
		http.Error(w, "Failed to get workflow status", http.StatusInternalServerError)
		return
	}

	// Return response
	status := WorkflowStatus{
		WorkflowID:  workflowID,
		ServiceID:   result.ServiceID,
		Action:      result.Action,
		Status:      "COMPLETED",
		Success:     result.Success,
		Message:     result.Message,
		CompletedAt: result.CompletedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
