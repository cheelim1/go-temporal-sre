package main

import (
	"log/slog"
	"os"

	"app/internal/features/jit"
	"app/internal/features/kilcron"
	"app/internal/features/superscript"
	"app/internal/worker"
	"app/internal/worker/config"
)

// logAdapter adapts slog.Logger to Temporal's log.Logger interface
type logAdapter struct {
	logger *slog.Logger
}

func (l *logAdapter) Debug(msg string, keyvals ...interface{}) {
	l.logger.Debug(msg, keyvals...)
}

func (l *logAdapter) Info(msg string, keyvals ...interface{}) {
	l.logger.Info(msg, keyvals...)
}

func (l *logAdapter) Warn(msg string, keyvals ...interface{}) {
	l.logger.Warn(msg, keyvals...)
}

func (l *logAdapter) Error(msg string, keyvals ...interface{}) {
	l.logger.Error(msg, keyvals...)
}

func main() {
	// Create structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create Temporal logger adapter
	temporalLogger := &logAdapter{logger: logger}

	logger.Info("Starting centralized Temporal worker")

	// Load configuration
	cfg := config.LoadConfig()
	logger.Info("Loaded configuration",
		"temporalHost", cfg.TemporalHost,
		"temporalNamespace", cfg.TemporalNamespace,
		"enabledFeatures", cfg.EnabledFeatures)

	// Create centralized worker
	centralizedWorker, err := worker.NewCentralizedWorker(cfg, temporalLogger)
	if err != nil {
		logger.Error("Failed to create centralized worker", "error", err)
		os.Exit(1)
	}

	// Register features
	if cfg.IsFeatureEnabled("kilcron") {
		kilcronFeature := kilcron.NewFeature()
		if err := centralizedWorker.RegisterFeature(kilcronFeature); err != nil {
			logger.Error("Failed to register kilcron feature", "error", err)
			os.Exit(1)
		}
	}

	if cfg.IsFeatureEnabled("superscript") {
		superscriptFeature := superscript.NewFeature(temporalLogger)
		if err := centralizedWorker.RegisterFeature(superscriptFeature); err != nil {
			logger.Error("Failed to register superscript feature", "error", err)
			os.Exit(1)
		}
	}

	if cfg.IsFeatureEnabled("jit") {
		jitFeature := jit.NewFeature()
		if err := centralizedWorker.RegisterFeature(jitFeature); err != nil {
			logger.Error("Failed to register JIT feature", "error", err)
			os.Exit(1)
		}
	}

	// Start the worker
	if err := centralizedWorker.Start(); err != nil {
		logger.Error("Failed to start centralized worker", "error", err)
		os.Exit(1)
	}

	logger.Info("Centralized worker started successfully")

	// Wait for shutdown
	centralizedWorker.WaitForShutdown()

	logger.Info("Centralized worker shut down")
}
