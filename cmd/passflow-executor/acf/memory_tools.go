package acf

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

// MemoryToolProvider implements agent memory system tools that interact with passflow-api.
// These tools allow agents to store and retrieve context between executions.
type MemoryToolProvider struct {
	agentID    string
	apiBaseURL string
	apiToken   string
	httpClient *http.Client
}

// NewMemoryToolProvider creates a new memory tool provider for an agent.
func NewMemoryToolProvider(agentID string) *MemoryToolProvider {
	apiBaseURL := os.Getenv("CHRONOS_API_URL")
	if apiBaseURL == "" {
		apiBaseURL = "http://passflow-api"
	}

	apiToken := os.Getenv("SERVICE_TOKEN")

	return &MemoryToolProvider{
		agentID:    agentID,
		apiBaseURL: apiBaseURL,
		apiToken:   apiToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Execute runs a memory tool call and returns a ToolResult.
func (m *MemoryToolProvider) Execute(ctx context.Context, call *ToolCall) (*ToolResult, error) {
	var result *ToolResult
	switch call.Tool {
	case "memory_store":
		result = m.execStore(ctx, call)
	case "memory_retrieve":
		result = m.execRetrieve(ctx, call)
	case "memory_list":
		result = m.execList(ctx, call)
	case "memory_delete":
		result = m.execDelete(ctx, call)
	default:
		result = &ToolResult{Error: fmt.Sprintf("unknown memory tool: %q", call.Tool)}
	}

	result.CallID = call.CallID
	return result, nil
}

// execStore stores a key-value memory entry.
// Required params: key (string), value (string)
// Optional params: metadata (object), ttl (number, seconds)
func (m *MemoryToolProvider) execStore(ctx context.Context, call *ToolCall) *ToolResult {
	key, ok := call.Params["key"].(string)
	if !ok || key == "" {
		return &ToolResult{Error: "memory_store requires 'key' parameter"}
	}

	value, ok := call.Params["value"].(string)
	if !ok || value == "" {
		return &ToolResult{Error: "memory_store requires 'value' parameter"}
	}

	payload := map[string]interface{}{
		"key":   key,
		"value": value,
	}

	if metadata, ok := call.Params["metadata"].(map[string]interface{}); ok {
		payload["metadata"] = metadata
	}

	if ttl, ok := call.Params["ttl"].(float64); ok && ttl > 0 {
		payload["ttl"] = int(ttl)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("failed to marshal request: %v", err)}
	}

	url := fmt.Sprintf("%s/api/v1/agents/%s/memory", m.apiBaseURL, m.agentID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("failed to create request: %v", err)}
	}

	req.Header.Set("Content-Type", "application/json")
	if m.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+m.apiToken)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("failed to store memory: %v", err)}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return &ToolResult{Error: fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody))}
	}

	return &ToolResult{Output: fmt.Sprintf("Memory stored successfully for key '%s'", key)}
}

// execRetrieve retrieves a memory entry by key.
// Required params: key (string)
func (m *MemoryToolProvider) execRetrieve(ctx context.Context, call *ToolCall) *ToolResult {
	key, ok := call.Params["key"].(string)
	if !ok || key == "" {
		return &ToolResult{Error: "memory_retrieve requires 'key' parameter"}
	}

	url := fmt.Sprintf("%s/api/v1/agents/%s/memory/%s", m.apiBaseURL, m.agentID, key)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("failed to create request: %v", err)}
	}

	if m.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+m.apiToken)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("failed to retrieve memory: %v", err)}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 404 {
		return &ToolResult{Output: fmt.Sprintf("No memory found for key '%s'", key)}
	}

	if resp.StatusCode >= 400 {
		return &ToolResult{Error: fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody))}
	}

	var memory map[string]interface{}
	if err := json.Unmarshal(respBody, &memory); err != nil {
		return &ToolResult{Error: fmt.Sprintf("failed to parse response: %v", err)}
	}

	// Return the value field from the response
	if value, ok := memory["value"].(string); ok {
		output := fmt.Sprintf("Memory for key '%s': %s", key, value)
		if metadata, ok := memory["metadata"].(map[string]interface{}); ok && len(metadata) > 0 {
			metadataJSON, _ := json.Marshal(metadata)
			output += fmt.Sprintf("\nMetadata: %s", string(metadataJSON))
		}
		return &ToolResult{Output: output}
	}

	return &ToolResult{Error: "invalid response format"}
}

// execList lists all memory entries for the agent.
func (m *MemoryToolProvider) execList(ctx context.Context, call *ToolCall) *ToolResult {
	url := fmt.Sprintf("%s/api/v1/agents/%s/memory", m.apiBaseURL, m.agentID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("failed to create request: %v", err)}
	}

	if m.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+m.apiToken)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("failed to list memories: %v", err)}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return &ToolResult{Error: fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody))}
	}

	var response struct {
		Memories []map[string]interface{} `json:"memories"`
		Count    int                      `json:"count"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return &ToolResult{Error: fmt.Sprintf("failed to parse response: %v", err)}
	}

	if response.Count == 0 {
		return &ToolResult{Output: "No memories found"}
	}

	output := fmt.Sprintf("Found %d memory entries:\n", response.Count)
	for _, mem := range response.Memories {
		key, _ := mem["key"].(string)
		value, _ := mem["value"].(string)
		output += fmt.Sprintf("  - %s: %s\n", key, value)
	}

	return &ToolResult{Output: output}
}

// execDelete deletes a memory entry by key.
// Required params: key (string)
func (m *MemoryToolProvider) execDelete(ctx context.Context, call *ToolCall) *ToolResult {
	key, ok := call.Params["key"].(string)
	if !ok || key == "" {
		return &ToolResult{Error: "memory_delete requires 'key' parameter"}
	}

	url := fmt.Sprintf("%s/api/v1/agents/%s/memory/%s", m.apiBaseURL, m.agentID, key)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("failed to create request: %v", err)}
	}

	if m.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+m.apiToken)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("failed to delete memory: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return &ToolResult{Output: fmt.Sprintf("No memory found for key '%s'", key)}
	}

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return &ToolResult{Error: fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody))}
	}

	return &ToolResult{Output: fmt.Sprintf("Memory deleted successfully for key '%s'", key)}
}

// GetToolDefinitions returns Claude-compatible tool definitions for memory tools.
func GetMemoryToolDefinitions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "memory_store",
			"description": "Store a key-value memory entry for the agent. Use this to remember context, learnings, or state between executions.",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]interface{}{
						"type":        "string",
						"description": "Unique key for the memory entry (max 256 characters)",
					},
					"value": map[string]interface{}{
						"type":        "string",
						"description": "Value to store",
					},
					"metadata": map[string]interface{}{
						"type":        "object",
						"description": "Optional metadata for the memory entry",
					},
					"ttl": map[string]interface{}{
						"type":        "number",
						"description": "Optional time-to-live in seconds. Memory will auto-expire after this duration.",
					},
				},
				"required": []string{"key", "value"},
			},
		},
		{
			"name":        "memory_retrieve",
			"description": "Retrieve a memory entry by key.",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]interface{}{
						"type":        "string",
						"description": "Key of the memory entry to retrieve",
					},
				},
				"required": []string{"key"},
			},
		},
		{
			"name":        "memory_list",
			"description": "List all memory entries for this agent.",
			"input_schema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			"name":        "memory_delete",
			"description": "Delete a memory entry by key.",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]interface{}{
						"type":        "string",
						"description": "Key of the memory entry to delete",
					},
				},
				"required": []string{"key"},
			},
		},
	}
}
