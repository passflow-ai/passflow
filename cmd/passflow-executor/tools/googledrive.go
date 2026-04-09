package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// googleDriveHTTP is the HTTP client for Google Drive internal API requests.
var googleDriveHTTP = &http.Client{Timeout: 30 * time.Second}

// runGoogleDriveAction dispatches Google Drive tool calls to the passflow-api
// internal endpoints. Unlike other integrations that call external APIs directly,
// Google Drive tools route through passflow-api which handles:
//   - OAuth2 token management and refresh
//   - Permission checking for agent file access
//   - Document parsing and content extraction
//
// Actions:
//   - list_files: List files in a folder
//   - read_file: Read and parse file content
//   - write_file: Update file content
//   - search_files: Full-text search across Drive
func runGoogleDriveAction(ctx context.Context, action string, creds map[string]string, args map[string]interface{}) (string, error) {
	// Get API URL and service token from credentials or environment
	apiURL := creds["api_url"]
	if apiURL == "" {
		apiURL = os.Getenv("CHRONOS_API_URL")
	}
	if apiURL == "" {
		return "", fmt.Errorf("google_drive: missing api_url")
	}

	serviceToken := creds["service_token"]
	if serviceToken == "" {
		serviceToken = os.Getenv("CHRONOS_SERVICE_TOKEN")
	}
	if serviceToken == "" {
		return "", fmt.Errorf("google_drive: missing service_token")
	}

	workspaceID := creds["workspace_id"]
	if workspaceID == "" {
		if ws, ok := args["workspace_id"].(string); ok {
			workspaceID = ws
		}
	}
	if workspaceID == "" {
		return "", fmt.Errorf("google_drive: missing workspace_id")
	}

	agentID := creds["agent_id"]
	if agentID == "" {
		if aid, ok := args["agent_id"].(string); ok {
			agentID = aid
		}
	}

	switch action {
	case "list_files":
		return googleDriveListFiles(ctx, apiURL, serviceToken, workspaceID, agentID, args)
	case "read_file":
		return googleDriveReadFile(ctx, apiURL, serviceToken, workspaceID, agentID, args)
	case "write_file":
		return googleDriveWriteFile(ctx, apiURL, serviceToken, workspaceID, agentID, args)
	case "search_files":
		return googleDriveSearchFiles(ctx, apiURL, serviceToken, workspaceID, agentID, args)
	default:
		return "", fmt.Errorf("google_drive: unsupported action %q", action)
	}
}

// googleDriveListFiles lists files in a Google Drive folder.
func googleDriveListFiles(ctx context.Context, apiURL, serviceToken, workspaceID, agentID string, args map[string]interface{}) (string, error) {
	payload := map[string]interface{}{
		"workspace_id": workspaceID,
		"agent_id":     agentID,
	}

	if folderID, ok := args["folder_id"].(string); ok && folderID != "" {
		payload["folder_id"] = folderID
	}
	if query, ok := args["query"].(string); ok && query != "" {
		payload["query"] = query
	}

	return googleDriveAPIRequest(ctx, apiURL, serviceToken, "/internal/google-drive/list", payload)
}

// googleDriveReadFile reads the content of a Google Drive file.
func googleDriveReadFile(ctx context.Context, apiURL, serviceToken, workspaceID, agentID string, args map[string]interface{}) (string, error) {
	fileID, ok := args["file_id"].(string)
	if !ok || fileID == "" {
		return "", fmt.Errorf("google_drive.read_file: file_id is required")
	}

	payload := map[string]interface{}{
		"workspace_id": workspaceID,
		"agent_id":     agentID,
		"file_id":      fileID,
	}

	if format, ok := args["format"].(string); ok && format != "" {
		payload["format"] = format
	}

	return googleDriveAPIRequest(ctx, apiURL, serviceToken, "/internal/google-drive/read", payload)
}

// googleDriveWriteFile updates the content of a Google Drive file.
func googleDriveWriteFile(ctx context.Context, apiURL, serviceToken, workspaceID, agentID string, args map[string]interface{}) (string, error) {
	fileID, ok := args["file_id"].(string)
	if !ok || fileID == "" {
		return "", fmt.Errorf("google_drive.write_file: file_id is required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("google_drive.write_file: content is required")
	}

	payload := map[string]interface{}{
		"workspace_id": workspaceID,
		"agent_id":     agentID,
		"file_id":      fileID,
		"content":      content,
	}

	return googleDriveAPIRequest(ctx, apiURL, serviceToken, "/internal/google-drive/write", payload)
}

// googleDriveSearchFiles performs a full-text search across Google Drive.
func googleDriveSearchFiles(ctx context.Context, apiURL, serviceToken, workspaceID, agentID string, args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("google_drive.search_files: query is required")
	}

	payload := map[string]interface{}{
		"workspace_id": workspaceID,
		"agent_id":     agentID,
		"query":        query,
	}

	// Handle mime_types array
	if mimeTypes, ok := args["mime_types"].([]interface{}); ok && len(mimeTypes) > 0 {
		stringMimeTypes := make([]string, 0, len(mimeTypes))
		for _, mt := range mimeTypes {
			if s, ok := mt.(string); ok {
				stringMimeTypes = append(stringMimeTypes, s)
			}
		}
		if len(stringMimeTypes) > 0 {
			payload["mime_types"] = stringMimeTypes
		}
	}

	return googleDriveAPIRequest(ctx, apiURL, serviceToken, "/internal/google-drive/search", payload)
}

// googleDriveAPIRequest sends a POST request to the passflow-api internal endpoint.
func googleDriveAPIRequest(ctx context.Context, apiURL, serviceToken, path string, payload interface{}) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("google_drive: failed to encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL+path, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("google_drive: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", serviceToken)

	resp, err := googleDriveHTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("google_drive: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // 1MB limit
	if err != nil {
		return "", fmt.Errorf("google_drive: failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("google_drive: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Extract the "data" field from the response if present
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("google_drive: failed to decode response: %w", err)
	}

	// passflow-api wraps responses in {"data": ...} structure
	if data, ok := result["data"]; ok {
		out, err := json.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("google_drive: failed to format result: %w", err)
		}
		return string(out), nil
	}

	return string(respBody), nil
}
