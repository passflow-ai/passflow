package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultHTTPTimeout is the default timeout for HTTP requests.
const DefaultHTTPTimeout = 30 * time.Second

// APIClient handles communication with passflow-api.
type APIClient struct {
	baseURL      string
	serviceToken string
	httpClient   *http.Client
}

// NewAPIClient creates a new API client for passflow-api.
func NewAPIClient(baseURL, serviceToken string) *APIClient {
	return &APIClient{
		baseURL:      baseURL,
		serviceToken: serviceToken,
		httpClient: &http.Client{
			Timeout: DefaultHTTPTimeout,
		},
	}
}

// NewAPIClientWithHTTPClient creates a new API client with a custom HTTP client.
func NewAPIClientWithHTTPClient(baseURL, serviceToken string, httpClient *http.Client) *APIClient {
	return &APIClient{
		baseURL:      baseURL,
		serviceToken: serviceToken,
		httpClient:   httpClient,
	}
}

// GetVaultConfig retrieves a vault configuration from passflow-api.
func (c *APIClient) GetVaultConfig(ctx context.Context, configID string) (*VaultConfig, error) {
	url := fmt.Sprintf("%s/internal/v1/vault/config/%s", c.baseURL, configID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var config VaultConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &config, nil
}

// UpdateSyncStatus reports the sync status to passflow-api.
func (c *APIClient) UpdateSyncStatus(ctx context.Context, configID string, status SyncStatus) error {
	url := fmt.Sprintf("%s/internal/v1/vault/config/%s/sync-status", c.baseURL, configID)

	body, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// setHeaders sets common headers for API requests.
func (c *APIClient) setHeaders(req *http.Request) {
	req.Header.Set("X-Service-Token", c.serviceToken)
	req.Header.Set("Accept", "application/json")
}
