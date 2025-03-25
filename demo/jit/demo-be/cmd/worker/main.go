package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cheelim1/go-temporal-sre/demo/jit/demo-be/internal/atlas"
	"github.com/cheelim1/go-temporal-sre/demo/jit/demo-be/internal/config"
	"github.com/cheelim1/go-temporal-sre/demo/jit/demo-be/internal/jitaccess"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	slog.Info("Starting Temporal worker")
	cfg := config.LoadConfig()

	// Create Temporal client.
	c, err := client.Dial(client.Options{
		HostPort:  cfg.TemporalHost,
		Namespace: cfg.TemporalNamespace,
	})
	if err != nil {
		slog.Error("Unable to create Temporal client", "error", err)
		os.Exit(1)
	}
	defer c.Close()

	// IMPORTANT: Initialize the Atlas client so that activities can call it.
	if err := atlas.InitAtlasClient(); err != nil {
		slog.Error("Failed to initialize Atlas client", "error", err)
		os.Exit(1)
	}

	// Create worker for the jit_access_task_queue.
	w := worker.New(c, "jit_access_task_queue", worker.Options{})

	// Register workflow and activities.
	w.RegisterWorkflow(jitaccess.JITAccessWorkflow)
	w.RegisterActivity(jitaccess.GetUserRoleActivity)
	w.RegisterActivity(jitaccess.SetUserRoleActivity)

	// Run the worker in a separate goroutine.
	go func() {
		slog.Info("Temporal worker running")
		if err := w.Run(worker.InterruptCh()); err != nil {
			slog.Error("Worker failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for termination signal.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	slog.Info("Worker shutting down")
	w.Stop()
}
