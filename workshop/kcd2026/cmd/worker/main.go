package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"kcd2026/pkg/workflows"
)

const (
	DefaultTemporalHost      = "localhost:7233"
	DefaultTemporalNamespace = "default"
	DefaultTaskQueue         = "kcd2026-task-queue"
)

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("starting kcd2026 temporal worker")

	// Get configuration from environment
	temporalHost := getEnv("TEMPORAL_HOST", DefaultTemporalHost)
	temporalNamespace := getEnv("TEMPORAL_NAMESPACE", DefaultTemporalNamespace)
	taskQueue := getEnv("TASK_QUEUE", DefaultTaskQueue)
	buildID := getEnv("BUILD_ID", "dev-build")

	logger.Info("worker configuration",
		"temporal_host", temporalHost,
		"namespace", temporalNamespace,
		"task_queue", taskQueue,
		"build_id", buildID,
	)

	// Create Temporal client
	c, err := client.Dial(client.Options{
		HostPort:  temporalHost,
		Namespace: temporalNamespace,
		Logger:    NewTemporalLogger(logger),
	})
	if err != nil {
		logger.Error("failed to create temporal client", "error", err)
		os.Exit(1)
	}
	defer c.Close()

	// Create worker with versioning support
	w := worker.New(c, taskQueue, worker.Options{
		BuildID:                                buildID,
		UseBuildIDForVersioning:                true,
		MaxConcurrentActivityExecutionSize:     10,
		MaxConcurrentWorkflowTaskExecutionSize: 10,
	})

	// Register workflows
	workflows.RegisterWorkflows(w)

	// Register activities
	workflows.RegisterActivities(w)

	logger.Info("registered workflows and activities")

	// Start worker
	err = w.Start()
	if err != nil {
		logger.Error("failed to start worker", "error", err)
		os.Exit(1)
	}

	logger.Info("worker started successfully")

	// Wait for interrupt signal for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("shutting down worker")
	w.Stop()
	logger.Info("worker stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TemporalLogger adapts slog.Logger to Temporal's logger interface
type TemporalLogger struct {
	logger *slog.Logger
}

func NewTemporalLogger(logger *slog.Logger) *TemporalLogger {
	return &TemporalLogger{logger: logger}
}

func (l *TemporalLogger) Debug(msg string, keyvals ...any) {
	l.logger.Debug(msg, keyvals...)
}

func (l *TemporalLogger) Info(msg string, keyvals ...any) {
	l.logger.Info(msg, keyvals...)
}

func (l *TemporalLogger) Warn(msg string, keyvals ...any) {
	l.logger.Warn(msg, keyvals...)
}

func (l *TemporalLogger) Error(msg string, keyvals ...any) {
	l.logger.Error(msg, keyvals...)
}
