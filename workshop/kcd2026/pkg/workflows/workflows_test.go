package workflows

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

// newTestActivities creates an Activities struct for test registration.
// Tests always mock the activities, so the struct itself is never called.
func newTestActivities() *Activities {
	return &Activities{}
}

func TestMigrateWithPgrollWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := newTestActivities()
	env.RegisterActivity(a.PgrollInitActivity)
	env.RegisterActivity(a.PgrollStartActivity)
	env.RegisterActivity(a.PgrollVerifyActivity)
	env.RegisterActivity(a.PgrollCompleteActivity)

	// All mocks registered upfront BEFORE ExecuteWorkflow
	env.OnActivity(a.PgrollInitActivity, mock.Anything, mock.Anything).Return(&PgrollInitOutput{
		Success: true,
		Message: "pgroll initialized",
	}, nil)
	env.OnActivity(a.PgrollStartActivity, mock.Anything, mock.Anything).Return(&PgrollStartOutput{
		Success: true,
		Message: "migration started",
	}, nil)
	env.OnActivity(a.PgrollVerifyActivity, mock.Anything, mock.Anything).Return(&PgrollVerifyOutput{
		Success: true,
		Message: "migration verified",
	}, nil)
	env.OnActivity(a.PgrollCompleteActivity, mock.Anything, mock.Anything).Return(&PgrollCompleteOutput{
		Success: true,
		Message: "migration completed",
	}, nil)

	env.ExecuteWorkflow(MigrateWithPgrollWorkflow, SchemaChangeRequest{
		ConnString:    "postgresql://postgres:postgres@localhost:5432/test",
		MigrationFile: "migrations/001_create_users.json",
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result MigrationResult
	require.NoError(t, env.GetWorkflowResult(&result))
	require.True(t, result.Success)
	require.True(t, result.OldSchemaValid)
	require.True(t, result.NewSchemaValid)
}

func TestEnsureSchemaWorkflow_NoChanges(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := newTestActivities()
	env.RegisterActivity(a.PsqldefDryRunActivity)

	env.OnActivity(a.PsqldefDryRunActivity, mock.Anything, mock.Anything).Return(&PsqldefDryRunOutput{
		HasChanges: false,
		Changes:    "",
	}, nil)

	env.ExecuteWorkflow(EnsureSchemaWorkflow, DesiredSchemaInput{
		ConnString: "postgresql://postgres:postgres@localhost:5432/test",
		SchemaFile: "migrations/desired_schema.sql",
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result SchemaResult
	require.NoError(t, env.GetWorkflowResult(&result))
	require.True(t, result.Success)
	require.False(t, result.HasChanges)
	require.Contains(t, result.Message, "idempotent")
}

func TestEnsureSchemaWorkflow_WithChanges(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := newTestActivities()
	env.RegisterActivity(a.PsqldefDryRunActivity)
	env.RegisterActivity(a.PsqldefApplyActivity)

	env.OnActivity(a.PsqldefDryRunActivity, mock.Anything, mock.Anything).Return(&PsqldefDryRunOutput{
		HasChanges: true,
		Changes:    "CREATE TABLE users...",
	}, nil)
	env.OnActivity(a.PsqldefApplyActivity, mock.Anything, mock.Anything).Return(&PsqldefApplyOutput{
		Success: true,
		Message: "schema applied",
	}, nil)

	env.ExecuteWorkflow(EnsureSchemaWorkflow, DesiredSchemaInput{
		ConnString: "postgresql://postgres:postgres@localhost:5432/test",
		SchemaFile: "migrations/desired_schema.sql",
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result SchemaResult
	require.NoError(t, env.GetWorkflowResult(&result))
	require.True(t, result.Success)
	require.True(t, result.HasChanges)
}

func TestProvisionReplicaWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := newTestActivities()
	env.RegisterActivity(a.SpinUpReplicaActivity)
	env.RegisterActivity(a.StartPgstreamProducerActivity)
	env.RegisterActivity(a.StartPgstreamConsumerActivity)
	env.RegisterActivity(a.WaitForSnapshotActivity)

	// ALL mocks upfront -- no delayed callback for mocks
	env.OnActivity(a.SpinUpReplicaActivity, mock.Anything, mock.Anything).Return(&SpinUpReplicaOutput{
		Name:       "replica-1",
		Port:       5433,
		ConnString: "postgresql://postgres:postgres@localhost:5433/postgres",
	}, nil)
	env.OnActivity(a.StartPgstreamProducerActivity, mock.Anything, mock.Anything).Return(&StartPgstreamProducerOutput{
		Success: true,
		Message: "producer started",
	}, nil)
	env.OnActivity(a.StartPgstreamConsumerActivity, mock.Anything, mock.Anything).Return(&StartPgstreamConsumerOutput{
		Success: true,
		Message: "consumer started",
	}, nil)
	env.OnActivity(a.WaitForSnapshotActivity, mock.Anything, mock.Anything).Return(&WaitForSnapshotOutput{
		Success: true,
		Message: "snapshot completed",
	}, nil)

	env.ExecuteWorkflow(ProvisionReplicaWorkflow, ReplicaRequest{
		ReplicaName:         "replica-1",
		PrimaryConnString:   "postgresql://postgres:postgres@localhost:5432/postgres",
		SnapshotTimeoutSecs: 300,
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result ReplicaResult
	require.NoError(t, env.GetWorkflowResult(&result))
	require.True(t, result.Success)
	require.Equal(t, "replica-1", result.ReplicaName)
	require.Equal(t, 5433, result.ReplicaPort)
}

func TestScaleReadReplicasWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := newTestActivities()
	env.RegisterActivity(a.UpdateReadPoolActivity)

	// Mock child workflow
	env.OnWorkflow(ProvisionReplicaWorkflow, mock.Anything, mock.Anything).Return(&ReplicaResult{
		Success:           true,
		ReplicaName:       "replica-1",
		ReplicaPort:       5433,
		ReplicaConnString: "postgresql://postgres:postgres@localhost:5433/postgres",
	}, nil)

	env.OnActivity(a.UpdateReadPoolActivity, mock.Anything, mock.Anything).Return(&UpdateReadPoolOutput{
		Success: true,
	}, nil)

	env.ExecuteWorkflow(ScaleReadReplicasWorkflow, ScaleRequest{
		PrimaryConnString: "postgresql://postgres:postgres@localhost:5432/postgres",
		NewReplicaCount:   2,
		ReplicaNamePrefix: "replica",
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result ScaleResult
	require.NoError(t, env.GetWorkflowResult(&result))
	require.True(t, result.Success)
	require.Len(t, result.NewReplicaConnStrings, 2)
}

func TestFailoverWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := newTestActivities()
	env.RegisterActivity(a.PauseWritesActivity)
	env.RegisterActivity(a.PromoteReplicaActivity)
	env.RegisterActivity(a.ResumeWritesActivity)

	env.OnActivity(a.PauseWritesActivity, mock.Anything, mock.Anything).Return(&PauseWritesOutput{
		Success: true,
	}, nil)
	env.OnActivity(a.PromoteReplicaActivity, mock.Anything, mock.Anything).Return(&PromoteReplicaOutput{
		Success: true,
		Message: "replica promoted",
	}, nil)
	env.OnActivity(a.ResumeWritesActivity, mock.Anything, mock.Anything).Return(&ResumeWritesOutput{
		Success: true,
	}, nil)

	env.ExecuteWorkflow(FailoverWorkflow, FailoverRequest{
		AppEndpoint:       "http://localhost:8080",
		ReplicaName:       "replica-1",
		ReplicaConnString: "postgresql://postgres:postgres@localhost:5433/postgres",
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result FailoverResult
	require.NoError(t, env.GetWorkflowResult(&result))
	require.True(t, result.Success)
	require.Equal(t, "replica-1", result.NewPrimaryName)
}

// TestMaintenanceModeWorkflow tests the signal-driven maintenance workflow.
// RegisterDelayedCallback is used ONLY to inject signals at specific workflow times.
// The test env auto-advances time to fire these callbacks instantly.
func TestMaintenanceModeWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Inject signals at specific workflow-time offsets
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(SignalEnterMaintenance, nil)
	}, 1*time.Second)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(SignalResume, nil)
	}, 3*time.Second)

	env.ExecuteWorkflow(MaintenanceModeWorkflow, MaintenanceRequest{
		AppEndpoint: "http://localhost:8080",
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result MaintenanceResult
	require.NoError(t, env.GetWorkflowResult(&result))
	require.True(t, result.Success)
	require.Equal(t, 2*time.Second, result.MaintenanceTime)
}
