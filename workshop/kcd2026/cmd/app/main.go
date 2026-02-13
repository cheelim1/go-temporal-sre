package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"kcd2026/pkg/workload"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("starting kcd2026 workload app")

	primaryConn := getEnv("PRIMARY_CONN", "")
	listenAddr := getEnv("LISTEN_ADDR", ":8080")

	cfg := workload.AppConfig{
		PrimaryConnStr: primaryConn,
		ListenAddr:     listenAddr,
		Logger:         logger,
	}

	// Parse replica connections (comma-separated)
	if replicas := os.Getenv("REPLICA_CONNS"); replicas != "" {
		for _, r := range splitNonEmpty(replicas, ',') {
			cfg.ReplicaConnStrs = append(cfg.ReplicaConnStrs, r)
		}
	}

	app := workload.New(cfg)
	ctx := context.Background()

	if err := app.Start(ctx, cfg); err != nil {
		logger.Error("failed to start app", "error", err)
		os.Exit(1)
	}

	logger.Info("workload app started", "addr", listenAddr)

	// Wait for shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("shutting down workload app")
	if err := app.Stop(ctx); err != nil {
		logger.Error("shutdown error", "error", err)
	}
	logger.Info("workload app stopped")
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func splitNonEmpty(s string, sep byte) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			if part := s[start:i]; part != "" {
				result = append(result, part)
			}
			start = i + 1
		}
	}
	if part := s[start:]; part != "" {
		result = append(result, part)
	}
	return result
}
