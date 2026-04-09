package googledrive

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestListFilesTool_Name verifies the tool name matches the expected value.
func TestListFilesTool_Name(t *testing.T) {
	tool := NewListFilesTool("http://api", "token", "ws-1", "agent-1")
	if got := tool.Name(); got != "gdrive_list_files" {
		t.Errorf("Name() = %q, want %q", got, "gdrive_list_files")
	}
}

// TestListFilesTool_Description verifies the tool has a non-empty description.
func TestListFilesTool_Description(t *testing.T) {
	tool := NewListFilesTool("http://api", "token", "ws-1", "agent-1")
	if tool.Description() == "" {
		t.Error("Description() should not be empty")
	}
}

// TestListFilesTool_Parameters verifies the JSON schema for parameters.
func TestListFilesTool_Parameters(t *testing.T) {
	tool := NewListFilesTool("http://api", "token", "ws-1", "agent-1")
	params := tool.Parameters()

	if params["type"] != "object" {
		t.Errorf("Parameters type = %v, want object", params["type"])
	}

	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Parameters properties should be a map")
	}

	// folder_id should be optional
	if _, ok := props["folder_id"]; !ok {
		t.Error("Parameters should have folder_id property")
	}

	// query should be optional
	if _, ok := props["query"]; !ok {
		t.Error("Parameters should have query property")
	}
}

// TestListFilesTool_Execute_Success tests successful file listing.
func TestListFilesTool_Execute_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/internal/google-drive/list" {
			t.Errorf("Expected /internal/google-drive/list, got %s", r.URL.Path)
		}
		if r.Header.Get("X-Service-Token") != "test-token" {
			t.Errorf("Expected X-Service-Token header")
		}

		// Parse request body
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if req["workspace_id"] != "ws-123" {
			t.Errorf("Expected workspace_id=ws-123, got %v", req["workspace_id"])
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"files": []map[string]interface{}{
					{"id": "file-1", "name": "document.pdf", "mimeType": "application/pdf"},
					{"id": "file-2", "name": "sheet.xlsx", "mimeType": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
				},
				"folder_id": "root",
			},
		})
	}))
	defer srv.Close()

	tool := NewListFilesTool(srv.URL, "test-token", "ws-123", "agent-1")
	result, err := tool.Execute(context.Background(), map[string]interface{}{})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(result, "document.pdf") {
		t.Errorf("Result should contain document.pdf, got: %s", result)
	}
	if !strings.Contains(result, "sheet.xlsx") {
		t.Errorf("Result should contain sheet.xlsx, got: %s", result)
	}
}

// TestListFilesTool_Execute_WithFolderID tests listing files in a specific folder.
func TestListFilesTool_Execute_WithFolderID(t *testing.T) {
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

	tool := NewListFilesTool(srv.URL, "token", "ws-1", "agent-1")
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"folder_id": "folder-abc",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if capturedReq["folder_id"] != "folder-abc" {
		t.Errorf("Expected folder_id=folder-abc, got %v", capturedReq["folder_id"])
	}
}

// TestListFilesTool_Execute_APIError tests handling of API errors.
func TestListFilesTool_Execute_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "internal server error",
		})
	}))
	defer srv.Close()

	tool := NewListFilesTool(srv.URL, "token", "ws-1", "agent-1")
	_, err := tool.Execute(context.Background(), map[string]interface{}{})

	if err == nil {
		t.Error("Expected error for API failure")
	}
}

// TestReadFileTool_Name verifies the tool name matches the expected value.
func TestReadFileTool_Name(t *testing.T) {
	tool := NewReadFileTool("http://api", "token", "ws-1", "agent-1")
	if got := tool.Name(); got != "gdrive_read_file" {
		t.Errorf("Name() = %q, want %q", got, "gdrive_read_file")
	}
}

// TestReadFileTool_Parameters verifies the JSON schema for parameters.
func TestReadFileTool_Parameters(t *testing.T) {
	tool := NewReadFileTool("http://api", "token", "ws-1", "agent-1")
	params := tool.Parameters()

	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Parameters properties should be a map")
	}

	// file_id should be present
	if _, ok := props["file_id"]; !ok {
		t.Error("Parameters should have file_id property")
	}

	// format should be optional
	if _, ok := props["format"]; !ok {
		t.Error("Parameters should have format property")
	}

	// Check required field
	required, ok := params["required"].([]string)
	if !ok {
		t.Fatal("Parameters should have required array")
	}

	hasFileID := false
	for _, r := range required {
		if r == "file_id" {
			hasFileID = true
		}
	}
	if !hasFileID {
		t.Error("file_id should be required")
	}
}

// TestReadFileTool_Execute_Success tests successful file reading.
func TestReadFileTool_Execute_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/google-drive/read" {
			t.Errorf("Expected /internal/google-drive/read, got %s", r.URL.Path)
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
				"content":   "This is the parsed content of the document.",
				"format":    "text",
			},
		})
	}))
	defer srv.Close()

	tool := NewReadFileTool(srv.URL, "token", "ws-1", "agent-1")
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"file_id": "file-123",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(result, "parsed content") {
		t.Errorf("Result should contain file content, got: %s", result)
	}
}

// TestReadFileTool_Execute_MissingFileID tests error handling for missing file_id.
func TestReadFileTool_Execute_MissingFileID(t *testing.T) {
	tool := NewReadFileTool("http://api", "token", "ws-1", "agent-1")
	_, err := tool.Execute(context.Background(), map[string]interface{}{})

	if err == nil {
		t.Error("Expected error for missing file_id")
	}
	if !strings.Contains(err.Error(), "file_id") {
		t.Errorf("Error should mention file_id, got: %v", err)
	}
}

// TestWriteFileTool_Name verifies the tool name matches the expected value.
func TestWriteFileTool_Name(t *testing.T) {
	tool := NewWriteFileTool("http://api", "token", "ws-1", "agent-1")
	if got := tool.Name(); got != "gdrive_write_file" {
		t.Errorf("Name() = %q, want %q", got, "gdrive_write_file")
	}
}

// TestWriteFileTool_Parameters verifies the JSON schema for parameters.
func TestWriteFileTool_Parameters(t *testing.T) {
	tool := NewWriteFileTool("http://api", "token", "ws-1", "agent-1")
	params := tool.Parameters()

	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Parameters properties should be a map")
	}

	// file_id and content should be present
	if _, ok := props["file_id"]; !ok {
		t.Error("Parameters should have file_id property")
	}
	if _, ok := props["content"]; !ok {
		t.Error("Parameters should have content property")
	}

	// Check required fields
	required, ok := params["required"].([]string)
	if !ok {
		t.Fatal("Parameters should have required array")
	}

	hasFileID := false
	hasContent := false
	for _, r := range required {
		if r == "file_id" {
			hasFileID = true
		}
		if r == "content" {
			hasContent = true
		}
	}
	if !hasFileID || !hasContent {
		t.Error("file_id and content should be required")
	}
}

// TestWriteFileTool_Execute_Success tests successful file writing.
func TestWriteFileTool_Execute_Success(t *testing.T) {
	var capturedReq map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/google-drive/write" {
			t.Errorf("Expected /internal/google-drive/write, got %s", r.URL.Path)
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

	tool := NewWriteFileTool(srv.URL, "token", "ws-1", "agent-1")
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"file_id": "file-123",
		"content": "Updated document content",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if capturedReq["content"] != "Updated document content" {
		t.Errorf("Expected content to be sent, got: %v", capturedReq["content"])
	}

	if !strings.Contains(result, "file-123") {
		t.Errorf("Result should contain file_id, got: %s", result)
	}
}

// TestWriteFileTool_Execute_MissingContent tests error handling for missing content.
func TestWriteFileTool_Execute_MissingContent(t *testing.T) {
	tool := NewWriteFileTool("http://api", "token", "ws-1", "agent-1")
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"file_id": "file-123",
	})

	if err == nil {
		t.Error("Expected error for missing content")
	}
	if !strings.Contains(err.Error(), "content") {
		t.Errorf("Error should mention content, got: %v", err)
	}
}

// TestSearchFilesTool_Name verifies the tool name matches the expected value.
func TestSearchFilesTool_Name(t *testing.T) {
	tool := NewSearchFilesTool("http://api", "token", "ws-1", "agent-1")
	if got := tool.Name(); got != "gdrive_search" {
		t.Errorf("Name() = %q, want %q", got, "gdrive_search")
	}
}

// TestSearchFilesTool_Parameters verifies the JSON schema for parameters.
func TestSearchFilesTool_Parameters(t *testing.T) {
	tool := NewSearchFilesTool("http://api", "token", "ws-1", "agent-1")
	params := tool.Parameters()

	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Parameters properties should be a map")
	}

	// query should be required
	if _, ok := props["query"]; !ok {
		t.Error("Parameters should have query property")
	}

	// mime_types should be optional
	if _, ok := props["mime_types"]; !ok {
		t.Error("Parameters should have mime_types property")
	}

	required, ok := params["required"].([]string)
	if !ok {
		t.Fatal("Parameters should have required array")
	}

	hasQuery := false
	for _, r := range required {
		if r == "query" {
			hasQuery = true
		}
	}
	if !hasQuery {
		t.Error("query should be required")
	}
}

// TestSearchFilesTool_Execute_Success tests successful file search.
func TestSearchFilesTool_Execute_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/google-drive/search" {
			t.Errorf("Expected /internal/google-drive/search, got %s", r.URL.Path)
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
					{"id": "file-2", "name": "Q2 Report.pdf"},
				},
				"query": "quarterly report",
			},
		})
	}))
	defer srv.Close()

	tool := NewSearchFilesTool(srv.URL, "token", "ws-1", "agent-1")
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"query": "quarterly report",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(result, "Q1 Report") {
		t.Errorf("Result should contain Q1 Report, got: %s", result)
	}
}

// TestSearchFilesTool_Execute_WithMimeTypes tests search with MIME type filter.
func TestSearchFilesTool_Execute_WithMimeTypes(t *testing.T) {
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

	tool := NewSearchFilesTool(srv.URL, "token", "ws-1", "agent-1")
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"query":      "test",
		"mime_types": []interface{}{"application/pdf", "text/plain"},
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	mimeTypes, ok := capturedReq["mime_types"].([]interface{})
	if !ok || len(mimeTypes) != 2 {
		t.Errorf("Expected mime_types with 2 items, got: %v", capturedReq["mime_types"])
	}
}

// TestSearchFilesTool_Execute_MissingQuery tests error handling for missing query.
func TestSearchFilesTool_Execute_MissingQuery(t *testing.T) {
	tool := NewSearchFilesTool("http://api", "token", "ws-1", "agent-1")
	_, err := tool.Execute(context.Background(), map[string]interface{}{})

	if err == nil {
		t.Error("Expected error for missing query")
	}
	if !strings.Contains(err.Error(), "query") {
		t.Errorf("Error should mention query, got: %v", err)
	}
}
