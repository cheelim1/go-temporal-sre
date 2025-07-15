package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	kilcronFeature "app/internal/features/kilcron"
	"app/internal/kilcron"
	"app/internal/worker"
	"app/internal/worker/config"

	"go.temporal.io/sdk/client"
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
	fmt.Println("Welcome to kilcron Demo using Centralized Worker...")

	// Create structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create Temporal logger adapter
	temporalLogger := &logAdapter{logger: logger}

	// Load configuration with kilcron feature enabled
	cfg := config.LoadConfig()
	cfg.EnabledFeatures = []string{"kilcron"} // Only enable kilcron for this demo

	logger.Info("Starting kilcron demo",
		"temporalHost", cfg.TemporalHost,
		"temporalNamespace", cfg.TemporalNamespace)

	// Create centralized worker
	centralizedWorker, err := worker.NewCentralizedWorker(cfg, temporalLogger)
	if err != nil {
		logger.Error("Failed to create centralized worker", "error", err)
		os.Exit(1)
	}

	// Register kilcron feature
	kilcronFeature := kilcronFeature.NewFeature()
	if err := centralizedWorker.RegisterFeature(kilcronFeature); err != nil {
		logger.Error("Failed to register kilcron feature", "error", err)
		os.Exit(1)
	}

	// Start the worker
	if err := centralizedWorker.Start(); err != nil {
		logger.Error("Failed to start centralized worker", "error", err)
		os.Exit(1)
	}

	// Setup HTTP server for demo UI
	mux := http.NewServeMux()

	// Default handler redirects to debug page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/demo/debug/", http.StatusFound)
	})

	// Debug handler (enhanced version from original debug.go)
	mux.HandleFunc("/demo/debug/", func(w http.ResponseWriter, r *http.Request) {
		debugAccessHandler(w, r, centralizedWorker.GetClient(), cfg.KilcronTaskQueue)
	})

	// Create HTTP server
	server := &http.Server{
		Addr:    ":8888",
		Handler: mux,
	}

	// Start HTTP server in a goroutine
	go func() {
		logger.Info("Starting HTTP server on :8888")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	logger.Info("Kilcron demo started successfully")
	logger.Info("HTTP server: http://localhost:8888")
	logger.Info("Press Ctrl+C to exit...")

	// Wait for interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt

	logger.Info("Shutting down gracefully...")

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Error shutting down HTTP server", "error", err)
	}

	// Stop centralized worker
	centralizedWorker.Stop()

	logger.Info("Kilcron demo shut down")
}

// debugAccessHandler handles the debug endpoint with workflow triggering
// This is moved from the original cmd/kilcron/debug.go
func debugAccessHandler(w http.ResponseWriter, r *http.Request, c client.Client, taskQueue string) {
	const orgID = "GopherPayNET"
	payID := "GoBux"

	// Check if action is happening ... after done redirect back ..
	q := r.URL.Query()
	if q.Has("action") {
		switch q.Get("action") {
		case "flaky":
			payID += "Flaky"
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	wfr, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:        orgID,
		TaskQueue: taskQueue,
	}, kilcron.PaymentWorkflow, payID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to start workflow: %v", err)
		return
	}

	render := `
<html>
<head><title>Kilcron Demo</title></head>
<body>
<h1>Kilcron Demo</h1>
<p>This is a demo of the kilcron feature using the centralized worker.</p>
<p>Worker Status: Running</p>
<p>Feature: kilcron</p>
<p>Task Queue: %s</p>
<p>Check the Temporal Web UI at <a href="http://localhost:8080">http://localhost:8080</a></p>
<hr>
<h2>Workflow Started</h2>
<p>Workflow ID: %s</p>
<p>Run ID: %s</p>
<div>
<p><a href="/demo/debug/">Run Normal</a></p>
<p><a href="/demo/debug/?action=flaky">Run Flaky</a></p>
</div>
</body>
</html>
`
	fmt.Fprintf(w, render, taskQueue, wfr.GetID(), wfr.GetRunID())
}
