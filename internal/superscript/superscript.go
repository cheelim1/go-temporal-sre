package superscript

import (
	"time"
)

const (
	// SuperscriptTaskQueue TaskQueue names
	SuperscriptTaskQueue = "superscript-task-queue"

	// SinglePaymentWorkflowType Workflow types
	SinglePaymentWorkflowType = "single-payment-workflow"
	// OrchestratorWorkflowType Workflow types
	OrchestratorWorkflowType = "orchestrator-workflow"

	// SampleOrderID Sample OrderID for demonstration
	SampleOrderID = "ORD-DEMO-123"

	// SinglePaymentScriptPath Script paths
	SinglePaymentScriptPath = "./internal/superscript/scripts/single_payment_collection.sh"
	// TraditionalBatchScriptPath script paths
	TraditionalBatchScriptPath = "./internal/superscript/scripts/traditional_payment_collection.sh"
)

// PaymentResult contains information about a payment collection attempt
type PaymentResult struct {
	OrderID       string        `json:"order_id"`
	Success       bool          `json:"success"`
	Output        string        `json:"output"`
	ErrorMessage  string        `json:"error_message,omitempty"`
	ExitCode      int           `json:"exit_code"`
	ExecutionTime time.Duration `json:"execution_time"`
	Timestamp     time.Time     `json:"timestamp"`
}

// BatchResult contains information about a batch of payment collections
type BatchResult struct {
	OrderIDs     []string        `json:"order_ids"`
	Results      []PaymentResult `json:"results"`
	TotalCount   int             `json:"total_count"`
	SuccessCount int             `json:"success_count"`
	FailCount    int             `json:"fail_count"`
	StartTime    time.Time       `json:"start_time"`
	EndTime      time.Time       `json:"end_time"`
}

// GetSuccessRate calculates the success rate as a percentage
func (br *BatchResult) GetSuccessRate() int {
	if br.TotalCount == 0 {
		return 0
	}
	return (br.SuccessCount * 100) / br.TotalCount
}
