// Package googledrive provides Google Drive tools for the agent executor.
// These tools enable agents to list, read, write, and search files in
// Google Drive workspaces where they have been granted access.
package googledrive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// httpClient is the shared HTTP client for all Google Drive tool requests.
var httpClient = &http.Client{Timeout: 30 * time.Second}

// apiRequest sends a POST request to the passflow-api internal endpoint.
// It handles authentication via X-Service-Token and JSON encoding/decoding.
func apiRequest(ctx context.Context, apiURL, serviceToken, path string, payload interface{}) (map[string]interface{}, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", serviceToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // 1MB limit
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// extractData extracts the "data" field from an API response.
// passflow-api wraps responses in {"data": ...} structure.
func extractData(result map[string]interface{}) (map[string]interface{}, error) {
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		// Some responses may not have the wrapper
		return result, nil
	}
	return data, nil
}

// formatResult converts the API response data to a JSON string for the LLM.
func formatResult(data map[string]interface{}) (string, error) {
	out, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to format result: %w", err)
	}
	return string(out), nil
}
