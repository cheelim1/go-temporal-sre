package jitaccess_test

import (
	"testing"
	"time"

	"github.com/cheelim1/go-temporal-sre/demo/jit/demo-be/internal/jitaccess"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestJITAccessWorkflow_Success(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	// Stub GetUserRoleActivity: For any context and any string parameter, return "originalRole".
	env.OnActivity(jitaccess.GetUserRoleActivity, mock.Anything, mock.AnythingOfType("string")).Return("originalRole", nil)
	// Stub SetUserRoleActivity: For any context and any two string parameters, return nil.
	env.OnActivity(jitaccess.SetUserRoleActivity, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)

	// Create a workflow request.
	req := jitaccess.JITAccessRequest{
		Username: "testuser",
		Reason:   "testing",
		NewRole:  "elevatedRole",
		Duration: 1 * time.Second,
	}

	env.ExecuteWorkflow(jitaccess.JITAccessWorkflow, req)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
