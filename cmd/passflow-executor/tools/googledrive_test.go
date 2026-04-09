package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestRunGoogleDriveAction_ListFiles tests the list_files action.
func TestRunGoogleDriveAction_ListFiles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/google-drive/list" {
			t.Errorf("Expected path /internal/google-drive/list, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.Header.Get("X-Service-Token") != "test-token" {
			t.Errorf("Expected X-Service-Token header")
		}

		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req["workspace_id"] != "ws-123" {
			t.Errorf("Expected workspace_id=ws-123, got %v", req["workspace_id"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"files": []map[string]interface{}{
					{"id": "file-1", "name": "document.pdf"},
				},
				"folder_id": "root",
			},
		})
	}))
	defer srv.Close()

	creds := map[string]string{
		"api_url":       srv.URL,
		"service_token": "test-token",
		"workspace_id":  "ws-123",
		"agent_id":      "agent-1",
	}

	result, err := runGoogleDriveAction(context.Background(), "list_files", creds, map[string]interface{}{})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !strings.Contains(result, "document.pdf") {
		t.Errorf("Expected result to contain document.pdf, got %s", result)
	}
}

// TestRunGoogleDriveAction_ListFilesWithFolder tests list_files with folder_id.
func TestRunGoogleDriveAction_ListFilesWithFolder(t *testing.T) {
	var capturedReq map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedReq)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"files":     []interface{}{},
				"folder_id": "folder-abc",
			},
		})
	}))
	defer srv.Close()

	creds := map[string]string{
		"api_url":       srv.URL,
		"service_token": "token",
		"workspace_id":  "ws-1",
	}

	_, err := runGoogleDriveAction(context.Background(), "list_files", creds, map[string]interface{}{
		"folder_id": "folder-abc",
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if capturedReq["folder_id"] != "folder-abc" {
		t.Errorf("Expected folder_id=folder-abc, got %v", capturedReq["folder_id"])
	}
}

// TestRunGoogleDriveAction_ReadFile tests the read_file action.
func TestRunGoogleDriveAction_ReadFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/google-drive/read" {
			t.Errorf("Expected path /internal/google-drive/read, got %s", r.URL.Path)
		}

		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req["file_id"] != "file-123" {
			t.Errorf("Expected file_id=file-123, got %v", req["file_id"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"file_id":   "file-123",
				"name":      "report.pdf",
				"mime_type": "application/pdf",
				"content":   "This is the document content.",
				"format":    "text",
			},
		})
	}))
	defer srv.Close()

	creds := map[string]string{
		"api_url":       srv.URL,
		"service_token": "token",
		"workspace_id":  "ws-1",
	}

	result, err := runGoogleDriveAction(context.Background(), "read_file", creds, map[string]interface{}{
		"file_id": "file-123",
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !strings.Contains(result, "document content") {
		t.Errorf("Expected result to contain content, got %s", result)
	}
}

// TestRunGoogleDriveAction_ReadFile_MissingFileID tests error for missing file_id.
func TestRunGoogleDriveAction_ReadFile_MissingFileID(t *testing.T) {
	creds := map[string]string{
		"api_url":       "http://api",
		"service_token": "token",
		"workspace_id":  "ws-1",
	}

	_, err := runGoogleDriveAction(context.Background(), "read_file", creds, map[string]interface{}{})

	if err == nil {
		t.Error("Expected error for missing file_id")
	}
	if !strings.Contains(err.Error(), "file_id") {
		t.Errorf("Error should mention file_id, got %v", err)
	}
}

// TestRunGoogleDriveAction_WriteFile tests the write_file action.
func TestRunGoogleDriveAction_WriteFile(t *testing.T) {
	var capturedReq map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/google-drive/write" {
			t.Errorf("Expected path /internal/google-drive/write, got %s", r.URL.Path)
		}

		_ = json.NewDecoder(r.Body).Decode(&capturedReq)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"file_id":   "file-123",
				"name":      "report.txt",
				"mime_type": "text/plain",
				"size":      1234,
			},
		})
	}))
	defer srv.Close()

	creds := map[string]string{
		"api_url":       srv.URL,
		"service_token": "token",
		"workspace_id":  "ws-1",
	}

	result, err := runGoogleDriveAction(context.Background(), "write_file", creds, map[string]interface{}{
		"file_id": "file-123",
		"content": "Updated content",
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if capturedReq["content"] != "Updated content" {
		t.Errorf("Expected content to be sent, got %v", capturedReq["content"])
	}

	if !strings.Contains(result, "file-123") {
		t.Errorf("Expected result to contain file_id, got %s", result)
	}
}

// TestRunGoogleDriveAction_WriteFile_MissingContent tests error for missing content.
func TestRunGoogleDriveAction_WriteFile_MissingContent(t *testing.T) {
	creds := map[string]string{
		"api_url":       "http://api",
		"service_token": "token",
		"workspace_id":  "ws-1",
	}

	_, err := runGoogleDriveAction(context.Background(), "write_file", creds, map[string]interface{}{
		"file_id": "file-123",
	})

	if err == nil {
		t.Error("Expected error for missing content")
	}
	if !strings.Contains(err.Error(), "content") {
		t.Errorf("Error should mention content, got %v", err)
	}
}

// TestRunGoogleDriveAction_SearchFiles tests the search_files action.
func TestRunGoogleDriveAction_SearchFiles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/google-drive/search" {
			t.Errorf("Expected path /internal/google-drive/search, got %s", r.URL.Path)
		}

		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req["query"] != "quarterly report" {
			t.Errorf("Expected query='quarterly report', got %v", req["query"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"files": []map[string]interface{}{
					{"id": "file-1", "name": "Q1 Report.pdf"},
				},
				"query": "quarterly report",
			},
		})
	}))
	defer srv.Close()

	creds := map[string]string{
		"api_url":       srv.URL,
		"service_token": "token",
		"workspace_id":  "ws-1",
	}

	result, err := runGoogleDriveAction(context.Background(), "search_files", creds, map[string]interface{}{
		"query": "quarterly report",
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !strings.Contains(result, "Q1 Report") {
		t.Errorf("Expected result to contain Q1 Report, got %s", result)
	}
}

// TestRunGoogleDriveAction_SearchFiles_WithMimeTypes tests search with MIME type filter.
func TestRunGoogleDriveAction_SearchFiles_WithMimeTypes(t *testing.T) {
	var capturedReq map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedReq)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"files": []interface{}{},
				"query": "test",
			},
		})
	}))
	defer srv.Close()

	creds := map[string]string{
		"api_url":       srv.URL,
		"service_token": "token",
		"workspace_id":  "ws-1",
	}

	_, err := runGoogleDriveAction(context.Background(), "search_files", creds, map[string]interface{}{
		"query":      "test",
		"mime_types": []interface{}{"application/pdf", "text/plain"},
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	mimeTypes, ok := capturedReq["mime_types"].([]interface{})
	if !ok || len(mimeTypes) != 2 {
		t.Errorf("Expected mime_types with 2 items, got %v", capturedReq["mime_types"])
	}
}

// TestRunGoogleDriveAction_SearchFiles_MissingQuery tests error for missing query.
func TestRunGoogleDriveAction_SearchFiles_MissingQuery(t *testing.T) {
	creds := map[string]string{
		"api_url":       "http://api",
		"service_token": "token",
		"workspace_id":  "ws-1",
	}

	_, err := runGoogleDriveAction(context.Background(), "search_files", creds, map[string]interface{}{})

	if err == nil {
		t.Error("Expected error for missing query")
	}
	if !strings.Contains(err.Error(), "query") {
		t.Errorf("Error should mention query, got %v", err)
	}
}

// TestRunGoogleDriveAction_MissingAPIURL tests error when API URL is not provided.
func TestRunGoogleDriveAction_MissingAPIURL(t *testing.T) {
	// Ensure env var is not set
	os.Unsetenv("CHRONOS_API_URL")

	creds := map[string]string{
		"service_token": "token",
		"workspace_id":  "ws-1",
	}

	_, err := runGoogleDriveAction(context.Background(), "list_files", creds, map[string]interface{}{})

	if err == nil {
		t.Error("Expected error for missing api_url")
	}
	if !strings.Contains(err.Error(), "api_url") {
		t.Errorf("Error should mention api_url, got %v", err)
	}
}

// TestRunGoogleDriveAction_MissingServiceToken tests error when service token is not provided.
func TestRunGoogleDriveAction_MissingServiceToken(t *testing.T) {
	// Ensure env var is not set
	os.Unsetenv("CHRONOS_SERVICE_TOKEN")

	creds := map[string]string{
		"api_url":      "http://api",
		"workspace_id": "ws-1",
	}

	_, err := runGoogleDriveAction(context.Background(), "list_files", creds, map[string]interface{}{})

	if err == nil {
		t.Error("Expected error for missing service_token")
	}
	if !strings.Contains(err.Error(), "service_token") {
		t.Errorf("Error should mention service_token, got %v", err)
	}
}

// TestRunGoogleDriveAction_MissingWorkspaceID tests error when workspace_id is not provided.
func TestRunGoogleDriveAction_MissingWorkspaceID(t *testing.T) {
	creds := map[string]string{
		"api_url":       "http://api",
		"service_token": "token",
	}

	_, err := runGoogleDriveAction(context.Background(), "list_files", creds, map[string]interface{}{})

	if err == nil {
		t.Error("Expected error for missing workspace_id")
	}
	if !strings.Contains(err.Error(), "workspace_id") {
		t.Errorf("Error should mention workspace_id, got %v", err)
	}
}

// TestRunGoogleDriveAction_UnsupportedAction tests error for unsupported action.
func TestRunGoogleDriveAction_UnsupportedAction(t *testing.T) {
	creds := map[string]string{
		"api_url":       "http://api",
		"service_token": "token",
		"workspace_id":  "ws-1",
	}

	_, err := runGoogleDriveAction(context.Background(), "unsupported_action", creds, map[string]interface{}{})

	if err == nil {
		t.Error("Expected error for unsupported action")
	}
	if !strings.Contains(err.Error(), "unsupported action") {
		t.Errorf("Error should mention unsupported action, got %v", err)
	}
}

// TestRunGoogleDriveAction_APIError tests handling of API errors.
func TestRunGoogleDriveAction_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "internal server error",
		})
	}))
	defer srv.Close()

	creds := map[string]string{
		"api_url":       srv.URL,
		"service_token": "token",
		"workspace_id":  "ws-1",
	}

	_, err := runGoogleDriveAction(context.Background(), "list_files", creds, map[string]interface{}{})

	if err == nil {
		t.Error("Expected error for API failure")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Error should mention status code, got %v", err)
	}
}

// TestRunGoogleDriveAction_FromEnvironment tests reading config from environment.
func TestRunGoogleDriveAction_FromEnvironment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Service-Token") != "env-token" {
			t.Errorf("Expected token from env, got %s", r.Header.Get("X-Service-Token"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"files": []interface{}{}},
		})
	}))
	defer srv.Close()

	// Set environment variables
	os.Setenv("CHRONOS_API_URL", srv.URL)
	os.Setenv("CHRONOS_SERVICE_TOKEN", "env-token")
	defer os.Unsetenv("CHRONOS_API_URL")
	defer os.Unsetenv("CHRONOS_SERVICE_TOKEN")

	creds := map[string]string{
		"workspace_id": "ws-1",
	}

	_, err := runGoogleDriveAction(context.Background(), "list_files", creds, map[string]interface{}{})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}
