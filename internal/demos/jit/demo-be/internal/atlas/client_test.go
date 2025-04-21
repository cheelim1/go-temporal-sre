package atlas_test

import (
	"os"
	"testing"

	"github.com/cheelim1/go-temporal-sre/demo/jit/demo-be/internal/atlas"
)

func TestInitAtlasClient_MissingEnv(t *testing.T) {
	// Unset environment variables.
	os.Unsetenv("ATLAS_PUBLIC_KEY")
	os.Unsetenv("ATLAS_PRIVATE_KEY")
	os.Unsetenv("ATLAS_PROJECT_ID")

	err := atlas.InitAtlasClient()
	if err == nil {
		t.Error("expected error when env variables are missing, got nil")
	}
}
