package atlas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/mongodb-forks/digest"
	"github.com/mongodb/atlas-sdk-go/admin"
)

var (
	atlasClient *admin.APIClient
	projectID   string
)

func InitAtlasClient() error {
	publicKey := os.Getenv("ATLAS_PUBLIC_KEY")
	privateKey := os.Getenv("ATLAS_PRIVATE_KEY")
	projectID = os.Getenv("ATLAS_PROJECT_ID")
	if publicKey == "" || privateKey == "" || projectID == "" {
		return fmt.Errorf("mongo atlas credentials or project id not set in environment variables")
	}

	cfg := admin.NewConfiguration()
	if cfg.DefaultHeader == nil {
		cfg.DefaultHeader = make(map[string]string)
	}
	// Set the Accept header required for Atlas API v2.
	cfg.DefaultHeader["Accept"] = "application/vnd.atlas.2023-02-01+json"
	cfg.DefaultHeader["Content-Type"] = "application/json"

	cfg.Servers = []admin.ServerConfiguration{
		{
			URL: "https://cloud.mongodb.com",
		},
	}

	// Create an HTTP client that uses Digest Authentication.
	httpClient := &http.Client{
		Transport: digest.NewTransport(publicKey, privateKey),
	}
	cfg.HTTPClient = httpClient

	// Create the Atlas API client.
	atlasClient = admin.NewAPIClient(cfg)
	return nil
}

// GetUserRole fetches the current role for a given database user in the "admin" database.
func GetUserRole(ctx context.Context, username string) (string, error) {
	resp, _, err := atlasClient.DatabaseUsersApi.
		GetDatabaseUser(ctx, projectID, "admin", username).
		Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get user from Atlas: %w", err)
	}
	if resp.Roles == nil || len(*resp.Roles) == 0 {
		return "", fmt.Errorf("no roles found for user %s", username)
	}
	return (*resp.Roles)[0].RoleName, nil
}

// SetUserRole updates the role for a given database user in the "admin" database.
func SetUserRole(ctx context.Context, username, role string) error {
	// Build the payload.
	payload := map[string]any{
		"roles": []map[string]string{
			{
				"databaseName": "admin",
				"roleName":     role,
			},
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Get the base URL from the configuration.
	baseURL := atlasClient.GetConfig().Servers[0].URL
	url := fmt.Sprintf("%s/api/atlas/v2/groups/%s/databaseUsers/admin/%s", baseURL, projectID, username)

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	// Set the headers from the configuration.
	for k, v := range atlasClient.GetConfig().DefaultHeader {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))

	// Send the request.
	resp, err := atlasClient.GetConfig().HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update user role in Atlas (status %d): %s", resp.StatusCode, string(body))
	}
	return nil
}
