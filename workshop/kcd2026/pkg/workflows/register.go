package workflows

import (
	"go.temporal.io/sdk/worker"
)

// RegisterWorkflows registers all workflows with the worker
func RegisterWorkflows(w worker.Worker) {
	// Demo 1a: pgroll migration
	w.RegisterWorkflow(MigrateWithPgrollWorkflow)

	// Demo 1b: sqldef schema
	w.RegisterWorkflow(EnsureSchemaWorkflow)

	// Demo 2: provision replica
	w.RegisterWorkflow(ProvisionReplicaWorkflow)

	// Demo 3: scale replicas
	w.RegisterWorkflow(ScaleReadReplicasWorkflow)

	// Demo 4: failover
	w.RegisterWorkflow(FailoverWorkflow)

	// Demo 5: maintenance mode
	w.RegisterWorkflow(MaintenanceModeWorkflow)
}

// RegisterActivities registers all activities with the worker
func RegisterActivities(w worker.Worker) {
	activities := &Activities{}

	// pgroll activities
	w.RegisterActivity(activities.PgrollInitActivity)
	w.RegisterActivity(activities.PgrollStartActivity)
	w.RegisterActivity(activities.PgrollVerifyActivity)
	w.RegisterActivity(activities.PgrollCompleteActivity)

	// sqldef activities
	w.RegisterActivity(activities.PsqldefDryRunActivity)
	w.RegisterActivity(activities.PsqldefApplyActivity)

	// pg0 activities
	w.RegisterActivity(activities.SpinUpReplicaActivity)
	w.RegisterActivity(activities.StopInstanceActivity)

	// pgstream activities
	w.RegisterActivity(activities.StartPgstreamProducerActivity)
	w.RegisterActivity(activities.StartPgstreamConsumerActivity)
	w.RegisterActivity(activities.WaitForSnapshotActivity)

	// app config activities
	w.RegisterActivity(activities.UpdateReadPoolActivity)
	w.RegisterActivity(activities.PauseWritesActivity)
	w.RegisterActivity(activities.ResumeWritesActivity)
	w.RegisterActivity(activities.PromoteReplicaActivity)
}
