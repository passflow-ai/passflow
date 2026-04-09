package googledrive

import (
	"context"
)

// ListFilesTool lists files in a Google Drive folder.
type ListFilesTool struct {
	apiURL       string
	serviceToken string
	workspaceID  string
	agentID      string
}

// NewListFilesTool creates a new ListFilesTool instance.
func NewListFilesTool(apiURL, serviceToken, workspaceID, agentID string) *ListFilesTool {
	return &ListFilesTool{
		apiURL:       apiURL,
		serviceToken: serviceToken,
		workspaceID:  workspaceID,
		agentID:      agentID,
	}
}

// Name returns the tool name used by the LLM.
func (t *ListFilesTool) Name() string {
	return "gdrive_list_files"
}

// Description returns a human-readable description of the tool.
func (t *ListFilesTool) Description() string {
	return "List files in a Google Drive folder. Returns file names, IDs, and types. " +
		"Use this to explore the contents of a folder or find files to read."
}

// Parameters returns the JSON Schema for the tool's parameters.
func (t *ListFilesTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"folder_id": map[string]interface{}{
				"type":        "string",
				"description": "The ID of the folder to list. Leave empty or use 'root' for the root folder.",
			},
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Optional filter query to narrow down results (e.g., file name pattern).",
			},
		},
	}
}

// Execute runs the tool with the given arguments.
func (t *ListFilesTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	// Build request payload
	payload := map[string]interface{}{
		"workspace_id": t.workspaceID,
		"agent_id":     t.agentID,
	}

	// Add optional parameters
	if folderID, ok := args["folder_id"].(string); ok && folderID != "" {
		payload["folder_id"] = folderID
	}
	if query, ok := args["query"].(string); ok && query != "" {
		payload["query"] = query
	}

	// Call the internal API
	result, err := apiRequest(ctx, t.apiURL, t.serviceToken, "/internal/google-drive/list", payload)
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
