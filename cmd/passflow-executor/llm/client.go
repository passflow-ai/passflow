package llm

import (
	"context"
)

// Message is a chat message.
type Message struct {
	Role       string     `json:"role"`                   // system | user | assistant | tool
	Content    string     `json:"content"`
	ToolCallID string     `json:"tool_call_id,omitempty"` // for role=tool
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`   // for role=assistant
}

// ToolCall represents a tool invocation requested by the LLM.
type ToolCall struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Arguments is a JSON string (as returned by the LLM).
	Arguments string `json:"arguments"`
}

// ToolDefinition describes a callable tool for the LLM.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
}

// Response is the LLM's reply.
type Response struct {
	Content   string     // text content (may be empty when ToolCalls is set)
	ToolCalls []ToolCall // non-empty when the LLM wants to call a tool
	Usage     Usage
}

// Usage tracks token consumption.
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Client is the unified LLM interface.
type Client interface {
	// Complete sends messages and tool definitions; returns the model's response.
	Complete(ctx context.Context, messages []Message, tools []ToolDefinition) (*Response, error)
}
