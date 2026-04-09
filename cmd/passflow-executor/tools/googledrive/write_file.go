package googledrive

import (
	"context"
	"fmt"
)

// WriteFileTool writes content to a Google Drive file.
type WriteFileTool struct {
	apiURL       string
	serviceToken string
	workspaceID  string
	agentID      string
}

// NewWriteFileTool creates a new WriteFileTool instance.
func NewWriteFileTool(apiURL, serviceToken, workspaceID, agentID string) *WriteFileTool {
	return &WriteFileTool{
		apiURL:       apiURL,
		serviceToken: serviceToken,
		workspaceID:  workspaceID,
		agentID:      agentID,
	}
}

// Name returns the tool name used by the LLM.
func (t *WriteFileTool) Name() string {
	return "gdrive_write_file"
}

// Description returns a human-readable description of the tool.
func (t *WriteFileTool) Description() string {
	return "Write or update content in a Google Drive file. " +
		"Use this to modify existing files. The agent must have write permission for the file."
}

// Parameters returns the JSON Schema for the tool's parameters.
func (t *WriteFileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_id": map[string]interface{}{
				"type":        "string",
				"description": "The ID of the file to write to. Get this from gdrive_list_files or gdrive_search.",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The new content to write to the file.",
			},
		},
		"required": []string{"file_id", "content"},
	}
}

// Execute runs the tool with the given arguments.
func (t *WriteFileTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	// Validate required parameters
	fileID, ok := args["file_id"].(string)
	if !ok || fileID == "" {
		return "", fmt.Errorf("file_id is required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("content is required")
	}

	// Build request payload
	payload := map[string]interface{}{
		"workspace_id": t.workspaceID,
		"agent_id":     t.agentID,
		"file_id":      fileID,
		"content":      content,
	}

	// Call the internal API
	result, err := apiRequest(ctx, t.apiURL, t.serviceToken, "/internal/google-drive/write", payload)
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
