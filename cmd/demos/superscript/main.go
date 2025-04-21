package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"app/internal/worker"
	"app/internal/worker/config"

	"go.temporal.io/sdk/client"
)

// logAdapter is a simple adapter that implements the Temporal logger interface
type logAdapter struct {
	logger *log.Logger
}

// Debug logs a debug message
func (l *logAdapter) Debug(msg string, keyvals ...interface{}) {
	l.logger.Printf("[DEBUG] %s %v", msg, keyvals)
}

// Info logs an info message
func (l *logAdapter) Info(msg string, keyvals ...interface{}) {
	l.logger.Printf("[INFO] %s %v", msg, keyvals)
}

// Warn logs a warning message
func (l *logAdapter) Warn(msg string, keyvals ...interface{}) {
	l.logger.Printf("[WARN] %s %v", msg, keyvals)
}

// Error logs an error message
func (l *logAdapter) Error(msg string, keyvals ...interface{}) {
	l.logger.Printf("[ERROR] %s %v", msg, keyvals)
}

func main() {
	fmt.Println("Welcome to superscript!")
	Run()
}

func Run() {
	// Create a basic logger - we'll use a simple adapter for standard logger
	logger := &logAdapter{log.New(os.Stdout, "[SuperScript] ", log.LstdFlags)}
	logger.Info("Starting SuperScript application")

	// Create Temporal client
	clientOptions := client.Options{
		HostPort:  "localhost:7233",
		Namespace: "default",
		Logger:    logger,
	}

	tempClient, err := client.Dial(clientOptions)
	if err != nil {
		logger.Error("Unable to create Temporal client", "error", err)
		os.Exit(1)
	}
	defer tempClient.Close()

	// Start worker
	w, err := worker.New(config.DefaultConfig().
		WithTaskQueue("superscript-task-queue").
		WithNamespace("default"))
	if err != nil {
		logger.Error("Unable to start worker", "error", err)
		os.Exit(1)
	}
	defer w.Stop()

	if err := w.Start(); err != nil {
		logger.Error("Unable to start worker", "error", err)
		os.Exit(1)
	}

	// Start HTTP server
	server := NewServer(tempClient, logger, "localhost:8080")

	// Start server in a goroutine so we can handle graceful shutdown
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	fmt.Println("SuperScript system is ready!")
	fmt.Println("HTTP Server running at http://localhost:8080")
	fmt.Println("Press Ctrl+C to exit...")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Shutdown gracefully
	fmt.Println("\nShutting down gracefully...")

	// Stop HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Stop(ctx); err != nil {
		logger.Error("Error stopping HTTP server", "error", err)
	}

	fmt.Println("Shutdown complete")
	os.Exit(0)
}
