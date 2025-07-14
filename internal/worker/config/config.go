package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// WorkerConfig holds configuration for the centralized Temporal worker
type WorkerConfig struct {
	// Temporal connection settings
	TemporalHost      string
	TemporalNamespace string

	// Worker settings
	MaxConcurrentActivities int
	MaxConcurrentWorkflows  int

	// Feature enablement
	EnabledFeatures []string

	// Logging
	LogLevel string

	// HTTP server settings (for demos)
	HTTPPort int
	HTTPHost string

	// Feature-specific settings
	SuperscriptBasePath  string
	JITTaskQueue         string
	BatchProcessingQueue string
	KilcronTaskQueue     string

	// Atlas/MongoDB settings (for JIT feature)
	AtlasPublicKey  string
	AtlasPrivateKey string
	AtlasProjectID  string
}

// LoadConfig loads configuration from environment variables with sensible defaults
func LoadConfig() *WorkerConfig {
	config := &WorkerConfig{
		// Default Temporal settings
		TemporalHost:      getEnv("TEMPORAL_HOST", "localhost:7233"),
		TemporalNamespace: getEnv("TEMPORAL_NAMESPACE", "default"),

		// Default worker settings
		MaxConcurrentActivities: getEnvInt("MAX_CONCURRENT_ACTIVITIES", 10),
		MaxConcurrentWorkflows:  getEnvInt("MAX_CONCURRENT_WORKFLOWS", 10),

		// Default enabled features (all enabled by default)
		EnabledFeatures: getEnvSlice("ENABLED_FEATURES", []string{
			"kilcron", "superscript", "jit", "batch", "data-enrichment",
		}),

		// Default logging
		LogLevel: getEnv("LOG_LEVEL", "INFO"),

		// Default HTTP settings
		HTTPPort: getEnvInt("HTTP_PORT", 8080),
		HTTPHost: getEnv("HTTP_HOST", "localhost"),

		// Feature-specific defaults
		SuperscriptBasePath:  getEnv("SUPERSCRIPT_BASE_PATH", "./internal/features/superscript/scripts/"),
		JITTaskQueue:         getEnv("JIT_TASK_QUEUE", "jit_access_task_queue"),
		BatchProcessingQueue: getEnv("BATCH_PROCESSING_QUEUE", "batch_processing_task_queue"),
		KilcronTaskQueue:     getEnv("KILCRON_TASK_QUEUE", "kilcron_task_queue"),

		// Atlas/MongoDB settings (for JIT feature)
		AtlasPublicKey:  getEnv("ATLAS_PUBLIC_KEY", ""),
		AtlasPrivateKey: getEnv("ATLAS_PRIVATE_KEY", ""),
		AtlasProjectID:  getEnv("ATLAS_PROJECT_ID", ""),
	}

	return config
}

// IsFeatureEnabled checks if a feature is enabled
func (c *WorkerConfig) IsFeatureEnabled(feature string) bool {
	for _, enabled := range c.EnabledFeatures {
		if enabled == feature {
			return true
		}
	}
	return false
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
