package acf

import (
	"encoding/json"
	"fmt"
)

// ToolCall represents a single tool invocation from the LLM.
type ToolCall struct {
	Tool   string                 `json:"tool"`
	Params map[string]interface{} `json:"params"`
	CallID string                 `json:"call_id"`
}

// ToolResult holds the output of a tool execution.
type ToolResult struct {
	CallID string `json:"call_id"`
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

// Policy defines workspace-level tool restrictions applied to a SystemToolProvider.
type Policy struct {
	AllowedTools   []string `json:"allowed_tools,omitempty"`
	DeniedTools    []string `json:"denied_tools,omitempty"`
	MaxBashTimeout int      `json:"max_bash_timeout,omitempty"` // seconds
	AllowNetwork   bool     `json:"allow_network"`
	AllowWrite     bool     `json:"allow_write"`
	PathWhitelist  []string `json:"path_whitelist,omitempty"`
}

// ParseToolCall converts a provider-specific tool call into a universal ToolCall.
func ParseToolCall(provider, callID string, raw json.RawMessage) (*ToolCall, error) {
	switch provider {
	case "anthropic":
		return parseAnthropicToolCall(callID, raw)
	case "openai", "azure":
		return parseOpenAIToolCall(callID, raw)
	case "gemini":
		return parseGeminiToolCall(callID, raw)
	default:
		// Fallback: try OpenAI format
		return parseOpenAIToolCall(callID, raw)
	}
}

// FormatToolResult converts a ToolResult into the format expected by the provider.
func FormatToolResult(provider string, result *ToolResult) (json.RawMessage, error) {
	content := result.Output
	if result.Error != "" {
		content = fmt.Sprintf("Error: %s", result.Error)
	}

	switch provider {
	case "anthropic":
		return json.Marshal(map[string]interface{}{
			"type":        "tool_result",
			"tool_use_id": result.CallID,
			"content":     content,
		})
	case "openai", "azure":
		return json.Marshal(map[string]interface{}{
			"role":         "tool",
			"tool_call_id": result.CallID,
			"content":      content,
		})
	case "gemini":
		return json.Marshal(map[string]interface{}{
			"name": result.CallID,
			"response": map[string]interface{}{
				"content": content,
			},
		})
	default:
		return json.Marshal(map[string]interface{}{
			"tool_call_id": result.CallID,
			"content":      content,
		})
	}
}

// --- Provider-specific parsers ---

func parseAnthropicToolCall(callID string, raw json.RawMessage) (*ToolCall, error) {
	var data struct {
		Name  string                 `json:"name"`
		Input map[string]interface{} `json:"input"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("parse anthropic tool call: %w", err)
	}
	return &ToolCall{Tool: data.Name, Params: data.Input, CallID: callID}, nil
}

func parseOpenAIToolCall(callID string, raw json.RawMessage) (*ToolCall, error) {
	var data struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("parse openai tool call: %w", err)
	}
	params := make(map[string]interface{})
	if data.Arguments != "" {
		if err := json.Unmarshal([]byte(data.Arguments), &params); err != nil {
			return nil, fmt.Errorf("parse openai arguments: %w", err)
		}
	}
	return &ToolCall{Tool: data.Name, Params: params, CallID: callID}, nil
}

func parseGeminiToolCall(callID string, raw json.RawMessage) (*ToolCall, error) {
	var data struct {
		Name string                 `json:"name"`
		Args map[string]interface{} `json:"args"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("parse gemini tool call: %w", err)
	}
	return &ToolCall{Tool: data.Name, Params: data.Args, CallID: callID}, nil
}
