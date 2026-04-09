package secrets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ControlPlaneResponse is the response from the control-plane.
type ControlPlaneResponse struct {
	SelectedProvider string          `json:"selected_provider"`
	SelectedModel    string          `json:"selected_model"`
	CredentialRef    string          `json:"credential_ref"`
	Credentials      *CredentialData `json:"credentials,omitempty"`
	FallbackChain    []FallbackEntry `json:"fallback_chain"`
	Constraints      struct {
		MaxCostPerRequest float64 `json:"max_cost_per_request"`
		RegulatoryScope   string  `json:"regulatory_scope"`
		AllowFallback     bool    `json:"allow_fallback"`
	} `json:"constraints"`
}

type CredentialData struct {
	AuthType string `json:"auth_type"`
	APIKey   string `json:"api_key,omitempty"`
	Token    string `json:"token,omitempty"`
	Status   string `json:"status"`
}

type FallbackEntry struct {
	Provider      string `json:"provider"`
	Model         string `json:"model"`
	CredentialRef string `json:"credential_ref"`
	Priority      int    `json:"priority"`
}

// ControlPlaneSource resolves routes and credentials from the control-plane.
type ControlPlaneSource struct {
	apiURL       string
	serviceToken string
	httpClient   *http.Client
}

// NewControlPlaneSource creates a new ControlPlaneSource.
func NewControlPlaneSource(apiURL, serviceToken string) *ControlPlaneSource {
	return &ControlPlaneSource{
		apiURL:       apiURL,
		serviceToken: serviceToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ResolveRequest is the request body for route resolution.
type ResolveRequest struct {
	TenantID       string `json:"tenant_id"`
	WorkspaceID    string `json:"workspace_id"`
	UseCase        string `json:"use_case"`
	RequestedModel string `json:"requested_model,omitempty"`
}

// ResolveRoute calls the control-plane to resolve routing.
func (c *ControlPlaneSource) ResolveRoute(ctx context.Context, tenantID, workspaceID, useCase, requestedModel string) (*ControlPlaneResponse, error) {
	reqBody := ResolveRequest{
		TenantID:       tenantID,
		WorkspaceID:    workspaceID,
		UseCase:        useCase,
		RequestedModel: requestedModel,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.apiURL + "/api/v1/internal/routing/resolve"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", c.serviceToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call control-plane: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("control-plane returned status %d", resp.StatusCode)
	}

	var result ControlPlaneResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetCredentials implements WorkspaceCredentialSource interface.
// It calls the control-plane to resolve credentials for a workspace and provider.
func (c *ControlPlaneSource) GetCredentials(ctx context.Context, workspaceID, provider string) (map[string]string, error) {
	resp, err := c.ResolveRoute(ctx, "", workspaceID, "llm", "")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve credentials from control-plane: %w", err)
	}

	if resp.Credentials == nil {
		return nil, fmt.Errorf("no credentials returned for workspace %s provider %s", workspaceID, provider)
	}

	creds := map[string]string{}
	if resp.Credentials.APIKey != "" {
		creds["apiKey"] = resp.Credentials.APIKey
	}
	if resp.Credentials.Token != "" {
		creds["apiKey"] = resp.Credentials.Token
	}

	if len(creds) == 0 {
		return nil, fmt.Errorf("empty credentials for workspace %s provider %s", workspaceID, provider)
	}

	return creds, nil
}

// ExtractCredentials extracts credentials from a control-plane response (helper method).
func (c *ControlPlaneSource) ExtractCredentials(response *ControlPlaneResponse) (*Credentials, error) {
	if response.Credentials == nil {
		return nil, fmt.Errorf("no credentials in response")
	}

	return &Credentials{
		APIKey: response.Credentials.APIKey,
	}, nil
}
