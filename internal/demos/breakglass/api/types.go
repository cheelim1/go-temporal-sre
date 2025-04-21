package api

import "time"

// BreakglassRequest represents the request body for breakglass actions
type BreakglassRequest struct {
	ServiceID   string            `json:"service_id"`
	Action      string            `json:"action"` // e.g., "restart", "scale", "rollback"
	Parameters  map[string]string `json:"parameters"`
	RequestedBy string            `json:"requested_by"`
}

// BreakglassResponse represents the response from breakglass actions
type BreakglassResponse struct {
	WorkflowID  string    `json:"workflow_id"`
	ServiceID   string    `json:"service_id"`
	Action      string    `json:"action"`
	Status      string    `json:"status"`
	Message     string    `json:"message,omitempty"`
	RequestedAt time.Time `json:"requested_at"`
}

// WorkflowStatus represents the status of a workflow
type WorkflowStatus struct {
	WorkflowID  string    `json:"workflow_id"`
	ServiceID   string    `json:"service_id"`
	Action      string    `json:"action"`
	Status      string    `json:"status"`
	Success     bool      `json:"success"`
	Message     string    `json:"message,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}
