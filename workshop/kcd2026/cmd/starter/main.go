package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.temporal.io/sdk/client"

	"kcd2026/pkg/workflows"
)

const (
	DefaultTemporalHost = "localhost:7233"
	DefaultNamespace    = "default"
	DefaultTaskQueue    = "kcd2026-task-queue"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Parse subcommand
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	temporalHost := getEnv("TEMPORAL_HOST", DefaultTemporalHost)
	namespace := getEnv("TEMPORAL_NAMESPACE", DefaultNamespace)
	taskQueue := getEnv("TASK_QUEUE", DefaultTaskQueue)

	// Create Temporal client
	c, err := client.Dial(client.Options{
		HostPort:  temporalHost,
		Namespace: namespace,
	})
	if err != nil {
		logger.Error("failed to create Temporal client", "error", err)
		os.Exit(1)
	}
	defer c.Close()

	ctx := context.Background()

	switch subcommand {
	case "pgroll":
		runPgrollMigration(ctx, c, taskQueue, logger)
	case "schema":
		runEnsureSchema(ctx, c, taskQueue, logger)
	case "replica":
		runProvisionReplica(ctx, c, taskQueue, logger)
	case "scale":
		runScaleReplicas(ctx, c, taskQueue, logger)
	case "failover":
		runFailover(ctx, c, taskQueue, logger)
	case "maintenance":
		runMaintenance(ctx, c, taskQueue, logger)
	case "signal":
		runSignal(ctx, c, logger)
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: starter <command> [flags]

Commands:
  pgroll        Start a pgroll migration workflow
  schema        Start an EnsureSchema workflow (sqldef)
  replica       Provision a read replica (pgstream)
  scale         Scale read replicas
  failover      Failover to a replica
  maintenance   Start a maintenance mode workflow
  signal        Send a signal to a running workflow

Environment:
  TEMPORAL_HOST      Temporal host (default: localhost:7233)
  TEMPORAL_NAMESPACE Temporal namespace (default: default)
  TASK_QUEUE         Task queue (default: kcd2026-task-queue)`)
}

func runPgrollMigration(ctx context.Context, c client.Client, taskQueue string, logger *slog.Logger) {
	fs := flag.NewFlagSet("pgroll", flag.ExitOnError)
	connStr := fs.String("conn", "", "PostgreSQL connection string")
	migrationFile := fs.String("file", "migrations/001_create_users.json", "Migration file path")
	fs.Parse(os.Args[2:])

	if *connStr == "" {
		logger.Error("--conn is required")
		os.Exit(1)
	}

	logger.Info("starting pgroll migration workflow", "file", *migrationFile)

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        fmt.Sprintf("pgroll-migration-%d", time.Now().UnixMilli()),
		TaskQueue: taskQueue,
	}, workflows.MigrateWithPgrollWorkflow, workflows.SchemaChangeRequest{
		ConnString:    *connStr,
		MigrationFile: *migrationFile,
	})
	if err != nil {
		logger.Error("failed to start workflow", "error", err)
		os.Exit(1)
	}

	logger.Info("workflow started", "workflow_id", run.GetID(), "run_id", run.GetRunID())

	var result workflows.MigrationResult
	if err := run.Get(ctx, &result); err != nil {
		logger.Error("workflow failed", "error", err)
		os.Exit(1)
	}

	logger.Info("workflow completed", "success", result.Success, "message", result.Message)
}

func runEnsureSchema(ctx context.Context, c client.Client, taskQueue string, logger *slog.Logger) {
	fs := flag.NewFlagSet("schema", flag.ExitOnError)
	connStr := fs.String("conn", "", "PostgreSQL connection string")
	schemaFile := fs.String("file", "migrations/desired_schema.sql", "Schema file path")
	fs.Parse(os.Args[2:])

	if *connStr == "" {
		logger.Error("--conn is required")
		os.Exit(1)
	}

	logger.Info("starting ensure schema workflow", "file", *schemaFile)

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        fmt.Sprintf("ensure-schema-%d", time.Now().UnixMilli()),
		TaskQueue: taskQueue,
	}, workflows.EnsureSchemaWorkflow, workflows.DesiredSchemaInput{
		ConnString: *connStr,
		SchemaFile: *schemaFile,
	})
	if err != nil {
		logger.Error("failed to start workflow", "error", err)
		os.Exit(1)
	}

	logger.Info("workflow started", "workflow_id", run.GetID(), "run_id", run.GetRunID())

	var result workflows.SchemaResult
	if err := run.Get(ctx, &result); err != nil {
		logger.Error("workflow failed", "error", err)
		os.Exit(1)
	}

	logger.Info("workflow completed", "success", result.Success, "has_changes", result.HasChanges, "message", result.Message)
}

func runProvisionReplica(ctx context.Context, c client.Client, taskQueue string, logger *slog.Logger) {
	fs := flag.NewFlagSet("replica", flag.ExitOnError)
	replicaName := fs.String("name", "replica-1", "Replica instance name")
	primaryConn := fs.String("primary", "", "Primary connection string")
	fs.Parse(os.Args[2:])

	if *primaryConn == "" {
		logger.Error("--primary is required")
		os.Exit(1)
	}

	logger.Info("starting provision replica workflow", "name", *replicaName)

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        fmt.Sprintf("provision-replica-%s-%d", *replicaName, time.Now().UnixMilli()),
		TaskQueue: taskQueue,
	}, workflows.ProvisionReplicaWorkflow, workflows.ReplicaRequest{
		ReplicaName:       *replicaName,
		PrimaryConnString: *primaryConn,
	})
	if err != nil {
		logger.Error("failed to start workflow", "error", err)
		os.Exit(1)
	}

	logger.Info("workflow started", "workflow_id", run.GetID(), "run_id", run.GetRunID())

	var result workflows.ReplicaResult
	if err := run.Get(ctx, &result); err != nil {
		logger.Error("workflow failed", "error", err)
		os.Exit(1)
	}

	logger.Info("workflow completed",
		"success", result.Success,
		"replica_name", result.ReplicaName,
		"replica_port", result.ReplicaPort,
		"conn_string", result.ReplicaConnString,
	)
}

func runScaleReplicas(ctx context.Context, c client.Client, taskQueue string, logger *slog.Logger) {
	fs := flag.NewFlagSet("scale", flag.ExitOnError)
	count := fs.Int("count", 2, "Number of replicas to provision")
	primaryConn := fs.String("primary", "", "Primary connection string")
	prefix := fs.String("prefix", "replica", "Replica name prefix")
	fs.Parse(os.Args[2:])

	if *primaryConn == "" {
		logger.Error("--primary is required")
		os.Exit(1)
	}

	logger.Info("starting scale replicas workflow", "count", *count)

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        fmt.Sprintf("scale-replicas-%d", time.Now().UnixMilli()),
		TaskQueue: taskQueue,
	}, workflows.ScaleReadReplicasWorkflow, workflows.ScaleRequest{
		PrimaryConnString: *primaryConn,
		NewReplicaCount:   *count,
		ReplicaNamePrefix: *prefix,
	})
	if err != nil {
		logger.Error("failed to start workflow", "error", err)
		os.Exit(1)
	}

	logger.Info("workflow started", "workflow_id", run.GetID(), "run_id", run.GetRunID())

	var result workflows.ScaleResult
	if err := run.Get(ctx, &result); err != nil {
		logger.Error("workflow failed", "error", err)
		os.Exit(1)
	}

	logger.Info("workflow completed", "success", result.Success, "message", result.Message)
}

func runFailover(ctx context.Context, c client.Client, taskQueue string, logger *slog.Logger) {
	fs := flag.NewFlagSet("failover", flag.ExitOnError)
	replicaName := fs.String("replica", "replica-1", "Replica to promote")
	replicaConn := fs.String("replica-conn", "", "Replica connection string")
	appEndpoint := fs.String("app", "http://localhost:8080", "App endpoint")
	fs.Parse(os.Args[2:])

	if *replicaConn == "" {
		logger.Error("--replica-conn is required")
		os.Exit(1)
	}

	logger.Info("starting failover workflow", "replica", *replicaName)

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        fmt.Sprintf("failover-%d", time.Now().UnixMilli()),
		TaskQueue: taskQueue,
	}, workflows.FailoverWorkflow, workflows.FailoverRequest{
		AppEndpoint:       *appEndpoint,
		ReplicaName:       *replicaName,
		ReplicaConnString: *replicaConn,
	})
	if err != nil {
		logger.Error("failed to start workflow", "error", err)
		os.Exit(1)
	}

	logger.Info("workflow started", "workflow_id", run.GetID(), "run_id", run.GetRunID())

	var result workflows.FailoverResult
	if err := run.Get(ctx, &result); err != nil {
		logger.Error("workflow failed", "error", err)
		os.Exit(1)
	}

	logger.Info("workflow completed", "success", result.Success, "message", result.Message)
}

func runMaintenance(ctx context.Context, c client.Client, taskQueue string, logger *slog.Logger) {
	fs := flag.NewFlagSet("maintenance", flag.ExitOnError)
	appEndpoint := fs.String("app", "http://localhost:8080", "App endpoint")
	fs.Parse(os.Args[2:])

	logger.Info("starting maintenance mode workflow")

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        fmt.Sprintf("maintenance-%d", time.Now().UnixMilli()),
		TaskQueue: taskQueue,
	}, workflows.MaintenanceModeWorkflow, workflows.MaintenanceRequest{
		AppEndpoint: *appEndpoint,
	})
	if err != nil {
		logger.Error("failed to start workflow", "error", err)
		os.Exit(1)
	}

	logger.Info("maintenance workflow started - send signals to control it",
		"workflow_id", run.GetID(),
		"run_id", run.GetRunID(),
	)
	logger.Info("use: starter signal --workflow-id " + run.GetID() + " --signal enter_maintenance")
	logger.Info("use: starter signal --workflow-id " + run.GetID() + " --signal resume")
}

func runSignal(ctx context.Context, c client.Client, logger *slog.Logger) {
	fs := flag.NewFlagSet("signal", flag.ExitOnError)
	workflowID := fs.String("workflow-id", "", "Workflow ID to signal")
	signalName := fs.String("signal", "", "Signal name (enter_maintenance, resume)")
	fs.Parse(os.Args[2:])

	if *workflowID == "" || *signalName == "" {
		logger.Error("--workflow-id and --signal are required")
		os.Exit(1)
	}

	logger.Info("sending signal", "workflow_id", *workflowID, "signal", *signalName)

	err := c.SignalWorkflow(ctx, *workflowID, "", *signalName, nil)
	if err != nil {
		logger.Error("failed to send signal", "error", err)
		os.Exit(1)
	}

	logger.Info("signal sent successfully")
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
