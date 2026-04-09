package bedrock

import "encoding/json"

// claudeRequest is the request format for Claude models on Bedrock.
type claudeRequest struct {
	AnthropicVersion string          `json:"anthropic_version"`
	MaxTokens        int             `json:"max_tokens"`
	System           string          `json:"system,omitempty"`
	Messages         []claudeMessage `json:"messages"`
	Tools            []claudeTool    `json:"tools,omitempty"`
}

// claudeMessage represents a message in the Claude conversation.
type claudeMessage struct {
	Role    string          `json:"role"`
	Content []claudeContent `json:"content"`
}

// claudeContent represents content within a message.
type claudeContent struct {
	Type string `json:"type"`

	// For text content
	Text string `json:"text,omitempty"`

	// For tool_use content
	ToolUse *claudeToolUse `json:"tool_use,omitempty"`

	// For tool_result content
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
}

// claudeToolUse represents a tool invocation by the model.
type claudeToolUse struct {
	Type  string          `json:"type"`
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// claudeTool represents a tool definition for Claude.
type claudeTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// claudeResponse is the response format from Claude models on Bedrock.
type claudeResponse struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Role         string          `json:"role"`
	Content      []claudeContent `json:"content"`
	Model        string          `json:"model"`
	StopReason   string          `json:"stop_reason"`
	StopSequence string          `json:"stop_sequence,omitempty"`
	Usage        claudeUsage     `json:"usage"`
}

// claudeUsage tracks token consumption for Claude.
type claudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
