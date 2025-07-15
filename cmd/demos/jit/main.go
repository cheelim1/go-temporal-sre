package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"app/internal/atlas"
	"app/internal/features/jit"
	"app/internal/jitaccess"
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
	fmt.Println("Welcome to JIT Access Demo using Centralized Worker!")

	// Create structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create Temporal logger adapter
	temporalLogger := &logAdapter{logger: logger}

	// Load configuration with JIT feature enabled
	cfg := config.LoadConfig()
	cfg.EnabledFeatures = []string{"jit"} // Only enable JIT for this demo
	cfg.HTTPPort = 8080                   // Use port 8080 for HTTP server

	logger.Info("Starting JIT Access demo",
		"temporalHost", cfg.TemporalHost,
		"temporalNamespace", cfg.TemporalNamespace)

	// Create centralized worker
	centralizedWorker, err := worker.NewCentralizedWorker(cfg, temporalLogger)
	if err != nil {
		logger.Error("Failed to create centralized worker", "error", err)
		os.Exit(1)
	}

	// Register JIT feature
	jitFeature := jit.NewFeature()
	if err := centralizedWorker.RegisterFeature(jitFeature); err != nil {
		logger.Error("Failed to register JIT feature", "error", err)
		os.Exit(1)
	}

	// Start the worker
	if err := centralizedWorker.Start(); err != nil {
		logger.Error("Failed to start centralized worker", "error", err)
		os.Exit(1)
	}

	// Setup HTTP server for demo UI
	mux := http.NewServeMux()

	// Default handler shows JIT info
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<html>
			<head><title>JIT Access Demo</title></head>
			<body>
				<h1>JIT Access Demo</h1>
				<p>This is a demo of the JIT (Just-In-Time) access feature using the centralized worker.</p>
				<p>Worker Status: Running</p>
				<p>Feature: jit</p>
				<p>Task Queue: jit_access_task_queue</p>
				<p>Check the Temporal Web UI at <a href="http://localhost:8080">http://localhost:8080</a></p>
				<h2>Available Endpoints:</h2>
				<ul>
					<li><a href="/api/user-role?username=demo-user">Get User Role</a></li>
					<li><a href="/api/built-in-roles">Get Built-in Roles</a></li>
					<li><a href="/api/database-users">Get Database Users</a></li>
					<li>POST /api/jit-request - Submit JIT request</li>
				</ul>
			</body>
			</html>
		`)
	})

	// API endpoints from original JIT demo
	mux.HandleFunc("/api/user-role", func(w http.ResponseWriter, r *http.Request) {
		handleGetUserRole(w, r, logger)
	})
	mux.HandleFunc("/api/built-in-roles", func(w http.ResponseWriter, r *http.Request) {
		handleGetBuiltInRoles(w, r)
	})
	mux.HandleFunc("/api/jit-request", func(w http.ResponseWriter, r *http.Request) {
		handleJITRequest(w, r, centralizedWorker.GetClient(), logger)
	})
	mux.HandleFunc("/api/database-users", func(w http.ResponseWriter, r *http.Request) {
		handleGetDatabaseUsers(w, r, logger)
	})

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: mux,
	}

	// Start HTTP server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", "port", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	logger.Info("JIT Access demo started successfully")
	logger.Info("HTTP server", "url", fmt.Sprintf("http://localhost:%d", cfg.HTTPPort))
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

	logger.Info("JIT Access demo shut down")
}

// HTTP handlers from original JIT demo

func handleGetUserRole(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username parameter is required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	role, err := atlas.GetUserRole(ctx, username)
	if err != nil {
		logger.Error("failed to get user role", "username", username, "error", err)
		http.Error(w, fmt.Sprintf("failed to get user role: %v", err), http.StatusInternalServerError)
		return
	}
	resp := map[string]string{
		"username":     username,
		"current_role": role,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleGetBuiltInRoles(w http.ResponseWriter, r *http.Request) {
	// Hardcoded list of MongoDB Atlas built-in roles.
	roles := []string{
		"atlasAdmin",
		"readWriteAnyDatabase",
		"readAnyDatabase",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}

func handleGetDatabaseUsers(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	ctx := r.Context()
	users, err := atlas.GetDatabaseUsers(ctx)
	if err != nil {
		logger.Error("failed to get database users", "error", err)
		http.Error(w, fmt.Sprintf("failed to get database users: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// JITRequest represents the JSON payload for a JIT access request.
type JITRequest struct {
	Username string `json:"username"`
	Reason   string `json:"reason"`
	NewRole  string `json:"new_role"`
	Duration string `json:"duration"`
}

func handleJITRequest(w http.ResponseWriter, r *http.Request, temporalClient client.Client, logger *slog.Logger) {
	var req JITRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	// Validate required fields.
	if req.Username == "" || req.NewRole == "" || req.Duration == "" {
		http.Error(w, "username, new_role, and duration are required", http.StatusBadRequest)
		return
	}
	// Validate duration format.
	d, err := time.ParseDuration(req.Duration)
	if err != nil {
		http.Error(w, "invalid duration format", http.StatusBadRequest)
		return
	}
	// Check that new_role is different from current role.
	ctx := r.Context()
	currentRole, err := atlas.GetUserRole(ctx, req.Username)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get current role: %v", err), http.StatusInternalServerError)
		return
	}
	if currentRole == req.NewRole {
		http.Error(w, "new_role cannot be the same as current role", http.StatusBadRequest)
		return
	}
	// Build the workflow request.
	workflowRequest := jitaccess.JITAccessRequest{
		Username: req.Username,
		Reason:   req.Reason,
		NewRole:  req.NewRole,
		Duration: d,
	}
	workflowID := "jit_access_" + req.Username + "_" + fmt.Sprintf("%d", time.Now().Unix())
	options := client.StartWorkflowOptions{
		ID:                                       workflowID,
		TaskQueue:                                "jit_access_task_queue",
		WorkflowExecutionErrorWhenAlreadyStarted: true,
	}
	we, err := temporalClient.ExecuteWorkflow(context.Background(), options, jitaccess.JITAccessWorkflow, workflowRequest)
	if err != nil {
		logger.Error("failed to start workflow", "error", err)
		http.Error(w, fmt.Sprintf("failed to start workflow: %v", err), http.StatusInternalServerError)
		return
	}
	resp := map[string]string{
		"status":     "accepted",
		"workflowID": we.GetID(),
		"runID":      we.GetRunID(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
