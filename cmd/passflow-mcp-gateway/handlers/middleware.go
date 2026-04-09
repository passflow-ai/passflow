package handlers

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Version of the middleware package for tracking deployments.
const middlewareVersion = "1.2.0" // Renamed to Passflow API key support

// AuthUserResponse represents the response from passflow-api /auth/me endpoint.
type AuthUserResponse struct {
	ID           string   `json:"id"`
	Email        string   `json:"email"`
	Name         string   `json:"name"`
	Role         string   `json:"role"`
	WorkspaceIDs []string `json:"workspaceIds"`
}

// BearerAuth returns an HTTP middleware that validates authentication on every
// incoming request. It supports two authentication methods:
//
//  1. Passflow API keys (pf_* prefix): Validated via passflow-api /auth/me endpoint
//  2. Static gateway token (MCP_GATEWAY_TOKEN): For service-to-service communication
//
// The /health endpoint is exempt so liveness/readiness probes continue to work
// without credentials.
func BearerAuth(next http.Handler) http.Handler {
	staticToken := os.Getenv("MCP_GATEWAY_TOKEN")
	passflowAPIURL := os.Getenv("PASSFLOW_API_URL")

	// HTTP client with reasonable timeout for API key validation
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow unauthenticated access to the health probe endpoint.
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		token := extractToken(r)
		if token == "" {
			w.Header().Set("WWW-Authenticate", `Bearer realm="mcp-gateway"`)
			http.Error(w, `{"error":"missing Authorization header or X-API-Key"}`, http.StatusUnauthorized)
			return
		}

		// Route to appropriate auth method based on token prefix
		if strings.HasPrefix(token, "pf_") {
			// Validate via passflow-api
			if passflowAPIURL == "" {
				http.Error(w, `{"error":"server misconfigured: PASSFLOW_API_URL not set for API key validation"}`, http.StatusInternalServerError)
				return
			}
			if err := validatePassflowAPIKey(httpClient, passflowAPIURL, token); err != nil {
				w.Header().Set("WWW-Authenticate", `Bearer realm="mcp-gateway"`)
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusUnauthorized)
				return
			}
		} else {
			// Validate against static gateway token
			if staticToken == "" {
				http.Error(w, `{"error":"server misconfigured: MCP_GATEWAY_TOKEN not set"}`, http.StatusInternalServerError)
				return
			}
			if subtle.ConstantTimeCompare([]byte(token), []byte(staticToken)) != 1 {
				w.Header().Set("WWW-Authenticate", `Bearer realm="mcp-gateway"`)
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// extractToken extracts the authentication token from the request.
// It checks X-API-Key header first, then Authorization: Bearer header.
func extractToken(r *http.Request) string {
	// First check X-API-Key header
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}

	// Then check Authorization: Bearer header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}

	return ""
}

// validatePassflowAPIKey validates a pf_* API key by calling passflow-api.
func validatePassflowAPIKey(client *http.Client, apiURL, token string) error {
	url := strings.TrimSuffix(apiURL, "/") + "/api/v1/auth/me"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("X-API-Key", token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate API key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid or expired API key")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API key validation failed with status %d", resp.StatusCode)
	}

	// Optionally decode the response to verify it's valid JSON
	var authUser AuthUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&authUser); err != nil {
		return fmt.Errorf("invalid response from auth service: %w", err)
	}

	return nil
}
