package googledrive

import (
	"context"
	"fmt"
)

// SearchFilesTool searches for files across Google Drive.
type SearchFilesTool struct {
	apiURL       string
	serviceToken string
	workspaceID  string
	agentID      string
}

// NewSearchFilesTool creates a new SearchFilesTool instance.
func NewSearchFilesTool(apiURL, serviceToken, workspaceID, agentID string) *SearchFilesTool {
	return &SearchFilesTool{
		apiURL:       apiURL,
		serviceToken: serviceToken,
		workspaceID:  workspaceID,
		agentID:      agentID,
	}
}

// Name returns the tool name used by the LLM.
func (t *SearchFilesTool) Name() string {
	return "gdrive_search"
}

// Description returns a human-readable description of the tool.
func (t *SearchFilesTool) Description() string {
	return "Search for files across Google Drive using full-text search. " +
		"Returns matching files with their IDs, names, and types. " +
		"Optionally filter by MIME types to narrow down results."
}

// Parameters returns the JSON Schema for the tool's parameters.
func (t *SearchFilesTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query. Searches file names and content.",
			},
			"mime_types": map[string]interface{}{
				"type":        "array",
				"description": "Optional array of MIME types to filter results (e.g., 'application/pdf', 'text/plain').",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"query"},
	}
}

// Execute runs the tool with the given arguments.
func (t *SearchFilesTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	// Validate required parameters
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("query is required")
	}

	// Build request payload
	payload := map[string]interface{}{
		"workspace_id": t.workspaceID,
		"agent_id":     t.agentID,
		"query":        query,
	}

	// Add optional mime_types parameter
	if mimeTypes, ok := args["mime_types"].([]interface{}); ok && len(mimeTypes) > 0 {
		// Convert []interface{} to []string
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

	// Call the internal API
	result, err := apiRequest(ctx, t.apiURL, t.serviceToken, "/internal/google-drive/search", payload)
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
