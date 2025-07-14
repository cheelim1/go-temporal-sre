package iwfsuperscript

import (
	"context"
	"testing"
	"time"

	"app/internal/superscript"

	"github.com/indeedeng/iwf-golang-sdk/iwf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// testLogger implements the iwf.Logger interface for testing
type testLogger struct {
	t *testing.T
}

func (l *testLogger) Debug(msg string, keyvals ...interface{}) {
	l.t.Logf("[DEBUG] %s %v", msg, keyvals)
}

func (l *testLogger) Info(msg string, keyvals ...interface{}) {
	l.t.Logf("[INFO] %s %v", msg, keyvals)
}

func (l *testLogger) Warn(msg string, keyvals ...interface{}) {
	l.t.Logf("[WARN] %s %v", msg, keyvals)
}

func (l *testLogger) Error(msg string, keyvals ...interface{}) {
	l.t.Logf("[ERROR] %s %v", msg, keyvals)
}

// IWFWorkflowTestSuite is a test suite for iWF workflows
type IWFWorkflowTestSuite struct {
	suite.Suite
	//TestServer     iwf.TestWorkflowEnvironment
	TestClient iwf.Client
	//TestWorker     iwf.Worker
	//WorkerConfig   iwf.WorkerConfig
	MockActivities *MockActivities
	OrigActivities *superscript.Activities
}

// MockActivities is a mock implementation of the Activities struct
type MockActivities struct {
	t                          *testing.T
	ExpectedOrderID            string
	ExpectedResult             *superscript.PaymentResult
	ExpectedError              error
	RunPaymentCollectionCalled bool
}

// RunPaymentCollectionScript is a mock implementation of the RunPaymentCollectionScript method
func (m *MockActivities) RunPaymentCollectionScript(ctx context.Context, orderID string) (*superscript.PaymentResult, error) {
	m.RunPaymentCollectionCalled = true
	m.t.Logf("Mock RunPaymentCollectionScript called with orderID: %s", orderID)

	assert.Equal(m.t, m.ExpectedOrderID, orderID, "OrderID should match expected value")

	return m.ExpectedResult, m.ExpectedError
}

// SetupSuite sets up the test suite
func (s *IWFWorkflowTestSuite) SetupSuite() {
	// Create a test logger
	logger := &testLogger{t: s.T()}

	//// Create worker config
	//s.WorkerConfig = iwf.WorkerConfig{
	//	IwfServiceAddress: "localhost:7171", // Not used in tests
	//	IwfWorkerAddress:  "localhost:7172", // Not used in tests
	//	Namespace:         "default",
	//	TaskQueue:         TaskQueueName,
	//	Logger:            logger,
	//}
	//
	// Create mock activities
	s.MockActivities = &MockActivities{
		t: s.T(),
	}
	//
	//// Create original activities for comparison
	//s.OrigActivities = superscript.NewActivities("./", logger)
	//
	//// Create test environment
	//s.TestServer = iwf.NewTestWorkflowEnvironment()
	//
	//// Create test client
	//s.TestClient = s.TestServer.GetIwfClient()
	//
	//// Register workflows with mock activities
	//singlePaymentWorkflow := &SinglePaymentWorkflow{
	//	Activities: s.MockActivities,
	//}
	//orchestratorWorkflow := NewOrchestratorWorkflow(s.MockActivities) // Pass activities
	//
	//// Register workflows with test environment
	//s.TestServer.RegisterWorkflow(singlePaymentWorkflow)
	//s.TestServer.RegisterWorkflow(orchestratorWorkflow)
	logger.Warn("TODO!!")
}

// TearDownSuite tears down the test suite
func (s *IWFWorkflowTestSuite) TearDownSuite() {
	// Nothing to do here
}

// TestSinglePaymentWorkflow tests the SinglePaymentWorkflow
func (s *IWFWorkflowTestSuite) TestSinglePaymentWorkflow() {
	// Set up mock activities
	s.MockActivities.ExpectedOrderID = "1234"
	s.MockActivities.ExpectedResult = &superscript.PaymentResult{
		OrderID:       "1234",
		Success:       true,
		Output:        "Payment processed successfully",
		ExitCode:      0,
		ExecutionTime: 100 * time.Millisecond,
		Timestamp:     time.Now(),
	}
	s.MockActivities.ExpectedError = nil
	s.MockActivities.RunPaymentCollectionCalled = false

	// What is the correct idiomaic way to write tests in iWF? Unknonw ..
	//// Start workflow
	//workflowID := "test-single-payment-1234"
	//runID, err := s.TestClient.StartWorkflow(context.Background(), SinglePaymentWorkflowType,
	//	iwf.WorkflowOptions{
	//		WorkflowID: workflowID,
	//		TaskQueue:  TaskQueueName,
	//	},
	//	superscript.SinglePaymentWorkflowParams{
	//		OrderID: "1234",
	//	})
	//
	//require.NoError(s.T(), err)
	//require.NotEmpty(s.T(), runID)
	//
	//// Execute workflow
	//s.TestServer.ExecuteWorkflow()
	//
	//// Verify mock was called
	//assert.True(s.T(), s.MockActivities.RunPaymentCollectionCalled, "RunPaymentCollectionScript should be called")
	//
	// Get workflow result
	var result superscript.PaymentResult
	//err = s.TestClient.GetWorkflowResult(context.Background(), workflowID, runID.RunID, &result)
	//require.NoError(s.T(), err)

	// Verify result
	assert.Equal(s.T(), "1234", result.OrderID)
	assert.True(s.T(), result.Success)
	assert.Equal(s.T(), "Payment processed successfully", result.Output)
	assert.Equal(s.T(), 0, result.ExitCode)
}

// TestSinglePaymentWorkflowFailure tests the SinglePaymentWorkflow with a failure
func (s *IWFWorkflowTestSuite) TestSinglePaymentWorkflowFailure() {
	// Set up mock activities
	s.MockActivities.ExpectedOrderID = "5678"
	s.MockActivities.ExpectedResult = &superscript.PaymentResult{
		OrderID:       "5678",
		Success:       false,
		Output:        "Payment processing failed",
		ErrorMessage:  "Invalid account",
		ExitCode:      1,
		ExecutionTime: 100 * time.Millisecond,
		Timestamp:     time.Now(),
	}
	s.MockActivities.ExpectedError = nil
	s.MockActivities.RunPaymentCollectionCalled = false

	// Follow the idiomatic iWF way to do tests
	//// Start workflow
	//workflowID := "test-single-payment-5678"
	//runID, err := s.TestClient.StartWorkflow(context.Background(), workflowID, SinglePaymentWorkflowType,
	//	iwf.WorkflowOptions{
	//		WorkflowID: workflowID,
	//		TaskQueue:  TaskQueueName,
	//	},
	//	superscript.SinglePaymentWorkflowParams{
	//		OrderID: "5678",
	//	})
	//
	//require.NoError(s.T(), err)
	//require.NotEmpty(s.T(), runID)
	//
	//// Execute workflow
	//s.TestServer.ExecuteWorkflow()
	//
	//// Verify mock was called
	//assert.True(s.T(), s.MockActivities.RunPaymentCollectionCalled, "RunPaymentCollectionScript should be called")
	//
	// Get workflow result
	var result superscript.PaymentResult
	//err = s.TestClient.GetWorkflowResult(context.Background(), workflowID, runID.RunID, &result)
	//require.NoError(s.T(), err)

	// Verify result
	assert.Equal(s.T(), "5678", result.OrderID)
	assert.False(s.T(), result.Success)
	assert.Equal(s.T(), "Payment processing failed", result.Output)
	assert.Equal(s.T(), 1, result.ExitCode)
}

// TestOrchestratorWorkflow tests the OrchestratorWorkflow
func (s *IWFWorkflowTestSuite) TestOrchestratorWorkflow() {
	// All broken .. will fix single run first ..
	//	// Mock the child workflow execution
	//	s.TestServer.OnWorkflow(SinglePaymentWorkflowType, iwf.AnyWorkflowOptions(), func(ctx iwf.WorkflowContext, input iwf.Object) (interface{}, error) {
	//		var params superscript.SinglePaymentWorkflowParams
	//		err := input.Get(&params)
	//		require.NoError(s.T(), err)
	//
	//		// Return a success result for even order IDs, failure for odd
	//		orderID := params.OrderID
	//		isEven := (orderID[len(orderID)-1]-'0')%2 == 0
	//
	//		return &superscript.PaymentResult{
	//			OrderID:       orderID,
	//			Success:       isEven,
	//			Output:        "Test output for " + orderID,
	//			Error:         "Test error for " + orderID,
	//			ExecutionTime: 100 * time.Millisecond,
	//			Timestamp:     time.Now(),
	//		}, nil
	//	})
	//
	//	// Start workflow
	//	workflowID := "test-orchestrator"
	//	runID, err := s.TestClient.StartWorkflow(context.Background(), OrchestratorWorkflowType,
	//		iwf.WorkflowOptions{
	//			WorkflowID: workflowID,
	//			TaskQueue:  TaskQueueName,
	//		},
	//		superscript.OrchestratorWorkflowParams{
	//			OrderIDs: []string{"1001", "1002", "1003", "1004"},
	//			RunDate:  time.Now(),
	//		})
	//
	//	require.NoError(s.T(), err)
	//	require.NotEmpty(s.T(), runID)
	//
	//	// Execute workflow
	//	s.TestServer.ExecuteWorkflow()
	//
	// Get workflow result
	var result superscript.BatchResult
	//	err = s.TestClient.GetWorkflowResult(context.Background(), workflowID, runID.RunID, &result)
	//	require.NoError(s.T(), err)
	//
	// Verify result
	assert.Equal(s.T(), 4, result.TotalCount)
	assert.Equal(s.T(), 2, result.SuccessCount) // Even order IDs succeed
	assert.Equal(s.T(), 2, result.FailCount)    // Odd order IDs fail
	assert.Len(s.T(), result.Results, 4)

	// Check individual results
	for i, res := range result.Results {
		orderID := result.OrderIDs[i]
		assert.Equal(s.T(), orderID, res.OrderID)

		isEven := (orderID[len(orderID)-1]-'0')%2 == 0
		assert.Equal(s.T(), isEven, res.Success)
	}
}

// TestIdempotency tests the idempotency behavior of the SinglePaymentWorkflow
func (s *IWFWorkflowTestSuite) TestIdempotency() {
	// Set up mock activities
	s.MockActivities.ExpectedOrderID = "9999"
	s.MockActivities.ExpectedResult = &superscript.PaymentResult{
		OrderID:       "9999",
		Success:       true,
		Output:        "Payment processed successfully",
		ExitCode:      0,
		ExecutionTime: 100 * time.Millisecond,
		Timestamp:     time.Now(),
	}
	s.MockActivities.ExpectedError = nil
	s.MockActivities.RunPaymentCollectionCalled = false

	// What is the idiomatic way to wirte tests in iWF?
	//// Start workflow first time
	//workflowID := "test-idempotency-9999"
	//runID1, err := s.TestClient.StartWorkflow(context.Background(), workflowID, SinglePaymentWorkflowType,
	//	iwf.WorkflowOptions{
	//		WorkflowID:            workflowID,
	//		TaskQueue:             TaskQueueName,
	//		WorkflowIdReusePolicy: iwf.WorkflowIdReusePolicyRejectDuplicate,
	//	},
	//	superscript.SinglePaymentWorkflowParams{
	//		OrderID: "9999",
	//	})
	//
	//require.NoError(s.T(), err)
	//require.NotEmpty(s.T(), runID1)
	//
	//// Execute workflow
	//s.TestServer.ExecuteWorkflow()
	//
	//// Verify mock was called
	//assert.True(s.T(), s.MockActivities.RunPaymentCollectionCalled, "RunPaymentCollectionScript should be called")
	//
	//// Reset mock
	//s.MockActivities.RunPaymentCollectionCalled = false
	//
	//// Try to start the same workflow again (should be rejected due to idempotency)
	//_, err = s.TestClient.StartWorkflow(context.Background(), workflowID, SinglePaymentWorkflowType,
	//	iwf.WorkflowOptions{
	//		WorkflowID:            workflowID,
	//		TaskQueue:             TaskQueueName,
	//		WorkflowIdReusePolicy: iwf.WorkflowIdReusePolicyRejectDuplicate,
	//	},
	//	superscript.SinglePaymentWorkflowParams{
	//		OrderID: "9999",
	//	})
	//
	// Verify error is WorkflowAlreadyStartedError
	//require.Error(s.T(), err)
	//assert.True(s.T(), iwf.IsWorkflowAlreadyStartedError(err), "Expected WorkflowAlreadyStartedError")

	// Verify mock was NOT called again
	assert.False(s.T(), s.MockActivities.RunPaymentCollectionCalled, "RunPaymentCollectionScript should not be called again")
	s.T().Fatal("TODO: Implement ,..")
}

// TestIWFWorkflows runs the test suite
func TestIWFWorkflows(t *testing.T) {
	t.Skip("IWF tests are not implemented yet - skipping to keep build clean")
	suite.Run(t, new(IWFWorkflowTestSuite))
}
