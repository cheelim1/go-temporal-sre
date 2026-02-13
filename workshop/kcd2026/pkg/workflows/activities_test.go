package workflows

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

// testActivityContext creates a context with activity info for testing
func testActivityContext(t *testing.T) context.Context {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	// Use RegisterActivity to set up the environment
	a := &Activities{}
	env.RegisterActivity(a.PgrollInitActivity)
	return context.Background()
}

// newTestActivitiesWithExec creates an Activities struct with a mock exec function
func newTestActivitiesWithExec(output string, err error) *Activities {
	return &Activities{
		ExecFn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return []byte(output), err
		},
	}
}

func TestPgrollInitActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	a := newTestActivitiesWithExec("pgroll init complete", nil)
	env.RegisterActivity(a.PgrollInitActivity)

	result, err := env.ExecuteActivity(a.PgrollInitActivity, PgrollInitInput{
		ConnString: "postgresql://postgres:postgres@localhost:5432/test",
	})
	require.NoError(t, err)

	var output PgrollInitOutput
	require.NoError(t, result.Get(&output))
	assert.True(t, output.Success)
	assert.Equal(t, "pgroll init complete", output.Message)
}

func TestPgrollInitActivity_Error(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	a := newTestActivitiesWithExec("error output", fmt.Errorf("exec failed"))
	env.RegisterActivity(a.PgrollInitActivity)

	_, err := env.ExecuteActivity(a.PgrollInitActivity, PgrollInitInput{
		ConnString: "postgresql://postgres:postgres@localhost:5432/test",
	})
	require.Error(t, err)
}

func TestPgrollStartActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	a := newTestActivitiesWithExec("migration started", nil)
	env.RegisterActivity(a.PgrollStartActivity)

	result, err := env.ExecuteActivity(a.PgrollStartActivity, PgrollStartInput{
		ConnString:    "postgresql://postgres:postgres@localhost:5432/test",
		MigrationFile: "001_create_users.json",
	})
	require.NoError(t, err)

	var output PgrollStartOutput
	require.NoError(t, result.Get(&output))
	assert.True(t, output.Success)
}

func TestPgrollVerifyActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	a := &Activities{}
	env.RegisterActivity(a.PgrollVerifyActivity)

	result, err := env.ExecuteActivity(a.PgrollVerifyActivity, PgrollVerifyInput{
		ConnString: "postgresql://postgres:postgres@localhost:5432/test",
	})
	require.NoError(t, err)

	var output PgrollVerifyOutput
	require.NoError(t, result.Get(&output))
	assert.True(t, output.Success)
}

func TestPgrollCompleteActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	a := newTestActivitiesWithExec("migration completed", nil)
	env.RegisterActivity(a.PgrollCompleteActivity)

	result, err := env.ExecuteActivity(a.PgrollCompleteActivity, PgrollCompleteInput{
		ConnString: "postgresql://postgres:postgres@localhost:5432/test",
	})
	require.NoError(t, err)

	var output PgrollCompleteOutput
	require.NoError(t, result.Get(&output))
	assert.True(t, output.Success)
}

func TestPsqldefDryRunActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	a := newTestActivitiesWithExec("CREATE TABLE users...", nil)
	env.RegisterActivity(a.PsqldefDryRunActivity)

	result, err := env.ExecuteActivity(a.PsqldefDryRunActivity, PsqldefDryRunInput{
		ConnString: "postgresql://postgres:postgres@localhost:5432/test",
		SchemaFile: "desired_schema.sql",
	})
	require.NoError(t, err)

	var output PsqldefDryRunOutput
	require.NoError(t, result.Get(&output))
	assert.True(t, output.HasChanges)
	assert.Contains(t, output.Changes, "CREATE TABLE")
}

func TestPsqldefApplyActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	a := newTestActivitiesWithExec("schema applied", nil)
	env.RegisterActivity(a.PsqldefApplyActivity)

	result, err := env.ExecuteActivity(a.PsqldefApplyActivity, PsqldefApplyInput{
		ConnString: "postgresql://postgres:postgres@localhost:5432/test",
		SchemaFile: "desired_schema.sql",
	})
	require.NoError(t, err)

	var output PsqldefApplyOutput
	require.NoError(t, result.Get(&output))
	assert.True(t, output.Success)
}

func TestStartPgstreamProducerActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	a := &Activities{}
	env.RegisterActivity(a.StartPgstreamProducerActivity)

	result, err := env.ExecuteActivity(a.StartPgstreamProducerActivity, StartPgstreamProducerInput{
		PrimaryConnString: "postgresql://postgres:postgres@localhost:5432/test",
		ReplicationSlot:   "test_slot",
	})
	require.NoError(t, err)

	var output StartPgstreamProducerOutput
	require.NoError(t, result.Get(&output))
	assert.True(t, output.Success)
}

func TestStartPgstreamConsumerActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	a := &Activities{}
	env.RegisterActivity(a.StartPgstreamConsumerActivity)

	result, err := env.ExecuteActivity(a.StartPgstreamConsumerActivity, StartPgstreamConsumerInput{
		ReplicaConnString: "postgresql://postgres:postgres@localhost:5433/test",
		ReplicationSlot:   "test_slot",
	})
	require.NoError(t, err)

	var output StartPgstreamConsumerOutput
	require.NoError(t, result.Get(&output))
	assert.True(t, output.Success)
}

func TestUpdateReadPoolActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	a := &Activities{}
	env.RegisterActivity(a.UpdateReadPoolActivity)

	result, err := env.ExecuteActivity(a.UpdateReadPoolActivity, UpdateReadPoolInput{
		ReplicaConnStrings: []string{"conn1", "conn2"},
	})
	require.NoError(t, err)

	var output UpdateReadPoolOutput
	require.NoError(t, result.Get(&output))
	assert.True(t, output.Success)
}

func TestPauseWritesActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	a := &Activities{}
	env.RegisterActivity(a.PauseWritesActivity)

	result, err := env.ExecuteActivity(a.PauseWritesActivity, PauseWritesInput{
		AppEndpoint: "http://localhost:8080",
	})
	require.NoError(t, err)

	var output PauseWritesOutput
	require.NoError(t, result.Get(&output))
	assert.True(t, output.Success)
}

func TestResumeWritesActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	a := &Activities{}
	env.RegisterActivity(a.ResumeWritesActivity)

	result, err := env.ExecuteActivity(a.ResumeWritesActivity, ResumeWritesInput{
		AppEndpoint:       "http://localhost:8080",
		NewPrimaryConnStr: "conn-string",
	})
	require.NoError(t, err)

	var output ResumeWritesOutput
	require.NoError(t, result.Get(&output))
	assert.True(t, output.Success)
}

func TestPromoteReplicaActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	a := &Activities{}
	env.RegisterActivity(a.PromoteReplicaActivity)

	result, err := env.ExecuteActivity(a.PromoteReplicaActivity, PromoteReplicaInput{
		ReplicaName: "replica-1",
	})
	require.NoError(t, err)

	var output PromoteReplicaOutput
	require.NoError(t, result.Get(&output))
	assert.True(t, output.Success)
}

func TestMaskPassword(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "postgresql://user:pass@localhost:5432/db",
			expected: "postgresql://***:***@localhost:5432/db",
		},
		{
			input:    "no-password-string",
			expected: "no-password-string",
		},
	}

	for _, tt := range tests {
		result := maskPassword(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}
