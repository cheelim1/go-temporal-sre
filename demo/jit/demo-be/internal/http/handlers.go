package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"log/slog"

	"github.com/cheelim1/go-temporal-sre/demo/jit/demo-be/internal/atlas"
	"github.com/cheelim1/go-temporal-sre/demo/jit/demo-be/internal/jitaccess"
	"go.temporal.io/sdk/client"
)

// Handler holds the Temporal client so that HTTP endpoints can trigger workflows.
type Handler struct {
	TemporalClient client.Client
}

// NewHandler creates a new HTTP handler.
func NewHandler(temporalClient client.Client) *Handler {
	return &Handler{
		TemporalClient: temporalClient,
	}
}

// GetUserRole handles GET /api/user-role?username=<db_user>
func (h *Handler) GetUserRole(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username parameter is required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	role, err := atlas.GetUserRole(ctx, username)
	if err != nil {
		slog.Error("failed to get user role", "username", username, "error", err)
		http.Error(w, fmt.Sprintf("failed to get user role: %v", err), http.StatusInternalServerError)
		return
	}
	resp := map[string]string{
		"username":     username,
		"current_role": role,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetBuiltInRoles handles GET /api/built-in-roles
func (h *Handler) GetBuiltInRoles(w http.ResponseWriter, r *http.Request) {
	// Hardcoded list of MongoDB Atlas built-in roles.
	roles := []string{
		"atlasAdmin",
		"readWriteAnyDatabase",
		"readAnyDatabase",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}

// JITRequest represents the JSON payload for a JIT access request.
type JITRequest struct {
	Username string `json:"username"`
	Reason   string `json:"reason"`
	NewRole  string `json:"new_role"`
	Duration string `json:"duration"`
}

// PostJITRequest handles POST /api/jit-request.
func (h *Handler) PostJITRequest(w http.ResponseWriter, r *http.Request) {
	var req JITRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	// Validate required fields.
	if req.Username == "" || req.NewRole == "" || req.Duration == "" {
		http.Error(w, "username, new_role, and duration are required", http.StatusBadRequest)
		return
	}
	// Validate duration format.
	d, err := time.ParseDuration(req.Duration)
	if err != nil {
		http.Error(w, "invalid duration format", http.StatusBadRequest)
		return
	}
	// Check that new_role is different from current role.
	ctx := r.Context()
	currentRole, err := atlas.GetUserRole(ctx, req.Username)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get current role: %v", err), http.StatusInternalServerError)
		return
	}
	if currentRole == req.NewRole {
		http.Error(w, "new_role cannot be the same as current role", http.StatusBadRequest)
		return
	}
	// Build the workflow request.
	workflowRequest := jitaccess.JITAccessRequest{
		Username: req.Username,
		Reason:   req.Reason,
		NewRole:  req.NewRole,
		Duration: d,
	}
	workflowID := "jit_access_" + req.Username + "_" + fmt.Sprintf("%d", time.Now().Unix())
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "jit_access_task_queue",
		WorkflowExecutionErrorWhenAlreadyStarted: true,
	}
	we, err := h.TemporalClient.ExecuteWorkflow(context.Background(), options, jitaccess.JITAccessWorkflow, workflowRequest)
	if err != nil {
		slog.Error("failed to start workflow", "error", err)
		http.Error(w, fmt.Sprintf("failed to start workflow: %v", err), http.StatusInternalServerError)
		return
	}
	resp := map[string]string{
		"status":     "accepted",
		"workflowID": we.GetID(),
		"runID":      we.GetRunID(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
