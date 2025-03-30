package superscript

import (
	"context"
	"fmt"
	"time"

	"github.com/bitfield/script"
	"go.temporal.io/sdk/activity"
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
	logger := activity.GetLogger(ctx)
	logger.Info("Starting payment collection activity", "orderID", orderID)
	
	startTime := time.Now()
	
	// Construct the command using the bitfield/script library
	cmdStr := fmt.Sprintf("%s %s", SinglePaymentScriptPath, orderID)
	logger.Info("Executing command", "command", cmdStr)
	
	// Execute the script and capture output and exit code
	execPipe := script.Exec(cmdStr)
	output, err := execPipe.String()
	exitCode := 0
	if err != nil {
		// The error from String() will be of the form "exit status X"
		// So we need to parse that to get the actual exit code
		logger.Info("Script execution error", "error", err.Error())
		
		// For simplicity, we'll just set a non-zero exit code
		// In a real implementation, we could parse the "exit status X" string
		exitCode = 1
		
		// If the error message contains an actual exit code, we could extract it
		// But for now, we'll just use a generic error code
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
	} else {
		logger.Info("Script execution succeeded", "orderID", orderID, "executionTime", executionTime)
	}
	
	return result, nil
}
