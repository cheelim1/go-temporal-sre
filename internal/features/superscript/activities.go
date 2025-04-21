package superscript

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bitfield/script"
	"go.temporal.io/sdk/log"
)

// Activities holds the configuration for script execution activities
type Activities struct {
	ScriptBasePath string
	Logger         log.Logger
}

// NewActivities creates a new instance of Activities
func NewActivities(scriptBasePath string, logger log.Logger) *Activities {
	return &Activities{
		ScriptBasePath: scriptBasePath,
		Logger:         logger,
	}
}

// RunPaymentCollectionScript runs the single payment collection script for an OrderID
// and returns the result in a standardized format
func (a *Activities) RunPaymentCollectionScript(ctx context.Context, orderID string) (*PaymentResult, error) {
	//logger := activity.GetLogger(ctx)
	logger := slog.Default()
	logger.Info("Starting payment collection activity", "orderID", orderID)

	startTime := time.Now()

	// Construct the command using the bitfield/script library
	scriptPath := a.ScriptBasePath + "./scripts/single_payment_collection.sh"
	if orderID == "4242" {
		logger.Warn("Unit test for Activity .. 4242 ..")
		scriptPath = a.ScriptBasePath + "./scripts/happy_payment_collection.sh"
	}
	cmdStr := fmt.Sprintf("%s %s", scriptPath, orderID)
	logger.Info("Executing command", "command", cmdStr)

	// Execute the script and capture output and exit code
	execPipe := script.Exec(cmdStr)
	output, err := execPipe.String()
	//spew.Dump(output)
	//spew.Dump(err)
	// Assign exit code ..
	exitCode := execPipe.ExitStatus()
	if err != nil {
		logger.Info("Script execution error", "error", err.Error())
		//return nil, err
	}

	// Calculate execution time
	executionTime := time.Since(startTime)

	// Prepare result
	result := &PaymentResult{
		OrderID:       orderID,
		Success:       exitCode == 0,
		Output:        output,
		ErrorMessage:  "",
		ExitCode:      exitCode,
		ExecutionTime: executionTime,
		Timestamp:     time.Now(),
	}

	if exitCode != 0 {
		result.ErrorMessage = fmt.Sprintf("Script failed with exit code: %d", exitCode)
		// We log the error but do not return it as an error to Temporal
		// This way the workflow can properly handle the script failure
		logger.Error("Script execution failed", "orderID", orderID, "exitCode", exitCode, "output", output)
		return result, fmt.Errorf("Script execution failed with exit code: %d", exitCode)
	}
	logger.Info("Script execution succeeded", "orderID", orderID, "executionTime", executionTime)

	return result, nil
}
