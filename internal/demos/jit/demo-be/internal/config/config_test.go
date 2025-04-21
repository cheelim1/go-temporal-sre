package config_test

import (
	"os"
	"testing"

	"github.com/cheelim1/go-temporal-sre/demo/jit/demo-be/internal/config"
)

func TestLoadConfig(t *testing.T) {
	// Set environment variables for testing.
	os.Setenv("ATLAS_PUBLIC_KEY", "test_public")
	os.Setenv("ATLAS_PRIVATE_KEY", "test_private")
	os.Setenv("ATLAS_PROJECT_ID", "test_project")
	os.Setenv("TEMPORAL_HOST", "localhost:7233")
	os.Setenv("TEMPORAL_NAMESPACE", "default")
	os.Setenv("PORT", "8080")

	cfg := config.LoadConfig()

	if cfg.AtlasPublicKey != "test_public" {
		t.Errorf("expected test_public, got %s", cfg.AtlasPublicKey)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected 8080, got %s", cfg.Port)
	}
}
