package googledrive

import (
	"context"
	"fmt"
)

// ReadFileTool reads the content of a Google Drive file.
type ReadFileTool struct {
	apiURL       string
	serviceToken string
	workspaceID  string
	agentID      string
}

// NewReadFileTool creates a new ReadFileTool instance.
func NewReadFileTool(apiURL, serviceToken, workspaceID, agentID string) *ReadFileTool {
	return &ReadFileTool{
		apiURL:       apiURL,
		serviceToken: serviceToken,
		workspaceID:  workspaceID,
		agentID:      agentID,
	}
}

// Name returns the tool name used by the LLM.
func (t *ReadFileTool) Name() string {
	return "gdrive_read_file"
}

// Description returns a human-readable description of the tool.
func (t *ReadFileTool) Description() string {
	return "Read the content of a Google Drive file. Supports various file formats " +
		"including PDFs, documents, spreadsheets, and text files. " +
		"The content is automatically parsed and returned as text."
}

// Parameters returns the JSON Schema for the tool's parameters.
func (t *ReadFileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_id": map[string]interface{}{
				"type":        "string",
				"description": "The ID of the file to read. Get this from gdrive_list_files or gdrive_search.",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "Optional output format: 'text' (default), 'json', or 'markdown'.",
				"enum":        []string{"text", "json", "markdown"},
			},
		},
		"required": []string{"file_id"},
	}
}

// Execute runs the tool with the given arguments.
func (t *ReadFileTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	// Validate required parameters
	fileID, ok := args["file_id"].(string)
	if !ok || fileID == "" {
		return "", fmt.Errorf("file_id is required")
	}

	// Build request payload
	payload := map[string]interface{}{
		"workspace_id": t.workspaceID,
		"agent_id":     t.agentID,
		"file_id":      fileID,
	}

	// Add optional format parameter
	if format, ok := args["format"].(string); ok && format != "" {
		payload["format"] = format
	}

	// Call the internal API
	result, err := apiRequest(ctx, t.apiURL, t.serviceToken, "/internal/google-drive/read", payload)
	if err != nil {
		return "", err
	}

	// Extract and format the response
	data, err := extractData(result)
	if err != nil {
		return "", err
	}

	return formatResult(data)
}
