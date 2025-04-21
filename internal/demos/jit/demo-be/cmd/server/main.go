package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/cheelim1/go-temporal-sre/demo/jit/demo-be/internal/atlas"
	"github.com/cheelim1/go-temporal-sre/demo/jit/demo-be/internal/config"
	httphandlers "github.com/cheelim1/go-temporal-sre/demo/jit/demo-be/internal/http"
	"go.temporal.io/sdk/client"
)

func main() {
	slog.Info("Starting JIT Access Demo HTTP server")
	cfg := config.LoadConfig()

	// Initialize Atlas client.
	if err := atlas.InitAtlasClient(); err != nil {
		slog.Error("failed to initialize Atlas client", "error", err)
		os.Exit(1)
	}

	// Create Temporal client.
	temporalClient, err := client.Dial(client.Options{
		HostPort:  cfg.TemporalHost,
		Namespace: cfg.TemporalNamespace,
	})
	if err != nil {
		slog.Error("failed to create Temporal client", "error", err)
		os.Exit(1)
	}
	defer temporalClient.Close()

	// Setup HTTP handlers.
	handler := httphandlers.NewHandler(temporalClient)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/user-role", handler.GetUserRole)
	mux.HandleFunc("/api/built-in-roles", handler.GetBuiltInRoles)
	mux.HandleFunc("/api/jit-request", handler.PostJITRequest)
	mux.HandleFunc("/api/database-users", handler.GetDatabaseUsers)

	addr := ":" + cfg.Port
	slog.Info("HTTP server listening", "address", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("HTTP server failed", "error", err)
		os.Exit(1)
	}
}
