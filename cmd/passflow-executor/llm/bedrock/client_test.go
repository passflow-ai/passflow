package bedrock

import (
	"encoding/json"
	"testing"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
)

func TestClient_ImplementsInterface(t *testing.T) {
	var _ llm.Client = (*Client)(nil)
}

func TestConvertMessagesForClaude(t *testing.T) {
	tests := []struct {
		name     string
		messages []llm.Message
		wantLen  int
	}{
		{
			name: "simple user message",
			messages: []llm.Message{
				{Role: "user", Content: "Hello"},
			},
			wantLen: 1,
		},
		{
			name: "user and assistant conversation",
			messages: []llm.Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there"},
			},
			wantLen: 2,
		},
		{
			name: "system message is excluded from messages array",
			messages: []llm.Message{
				{Role: "system", Content: "You are helpful"},
				{Role: "user", Content: "Hello"},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertMessagesForClaude(tt.messages)
			if len(got) != tt.wantLen {
				t.Errorf("convertMessagesForClaude() returned %d messages, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestExtractSystemPrompt(t *testing.T) {
	tests := []struct {
		name     string
		messages []llm.Message
		want     string
	}{
		{
			name: "extracts system message",
			messages: []llm.Message{
				{Role: "system", Content: "You are a helpful assistant"},
				{Role: "user", Content: "Hello"},
			},
			want: "You are a helpful assistant",
		},
		{
			name: "returns empty when no system message",
			messages: []llm.Message{
				{Role: "user", Content: "Hello"},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSystemPrompt(tt.messages)
			if got != tt.want {
				t.Errorf("extractSystemPrompt() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConvertToolsForClaude(t *testing.T) {
	tools := []llm.ToolDefinition{
		{
			Name:        "get_weather",
			Description: "Get weather for a location",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"location": map[string]interface{}{
						"type":        "string",
						"description": "The city and state",
					},
				},
				"required": []interface{}{"location"},
			},
		},
	}

	got := convertToolsForClaude(tools)
	if len(got) != 1 {
		t.Fatalf("convertToolsForClaude() returned %d tools, want 1", len(got))
	}

	if got[0].Name != "get_weather" {
		t.Errorf("tool name = %q, want %q", got[0].Name, "get_weather")
	}

	if got[0].Description != "Get weather for a location" {
		t.Errorf("tool description = %q, want %q", got[0].Description, "Get weather for a location")
	}

	if got[0].InputSchema == nil {
		t.Error("tool input schema is nil")
	}
}

func TestConvertToolCalls(t *testing.T) {
	toolUse := claudeToolUse{
		Type:  "tool_use",
		ID:    "toolu_123",
		Name:  "get_weather",
		Input: json.RawMessage(`{"location":"San Francisco, CA"}`),
	}

	got := convertToolCall(toolUse)

	if got.ID != "toolu_123" {
		t.Errorf("tool call ID = %q, want %q", got.ID, "toolu_123")
	}

	if got.Name != "get_weather" {
		t.Errorf("tool call name = %q, want %q", got.Name, "get_weather")
	}

	if got.Arguments != `{"location":"San Francisco, CA"}` {
		t.Errorf("tool call arguments = %q, want %q", got.Arguments, `{"location":"San Francisco, CA"}`)
	}
}

func TestParseClaudeResponse(t *testing.T) {
	tests := []struct {
		name         string
		response     claudeResponse
		wantContent  string
		wantToolCall bool
	}{
		{
			name: "text response",
			response: claudeResponse{
				Content: []claudeContent{
					{Type: "text", Text: "Hello!"},
				},
				Usage: claudeUsage{
					InputTokens:  10,
					OutputTokens: 5,
				},
			},
			wantContent:  "Hello!",
			wantToolCall: false,
		},
		{
			name: "tool use response",
			response: claudeResponse{
				Content: []claudeContent{
					{
						Type: "tool_use",
						ToolUse: &claudeToolUse{
							Type:  "tool_use",
							ID:    "toolu_abc",
							Name:  "search",
							Input: json.RawMessage(`{"query":"test"}`),
						},
					},
				},
			},
			wantContent:  "",
			wantToolCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseClaudeResponse(&tt.response)

			if got.Content != tt.wantContent {
				t.Errorf("content = %q, want %q", got.Content, tt.wantContent)
			}

			if tt.wantToolCall && len(got.ToolCalls) == 0 {
				t.Error("expected tool calls but got none")
			}
		})
	}
}

func TestConvertToolResultMessage(t *testing.T) {
	msg := llm.Message{
		Role:       "tool",
		Content:    "The weather is sunny",
		ToolCallID: "toolu_123",
	}

	got := convertToolResultForClaude(msg)

	if got.Role != "user" {
		t.Errorf("role = %q, want %q", got.Role, "user")
	}

	if len(got.Content) != 1 {
		t.Fatalf("content length = %d, want 1", len(got.Content))
	}

	if got.Content[0].Type != "tool_result" {
		t.Errorf("content type = %q, want %q", got.Content[0].Type, "tool_result")
	}

	if got.Content[0].ToolUseID != "toolu_123" {
		t.Errorf("tool_use_id = %q, want %q", got.Content[0].ToolUseID, "toolu_123")
	}
}

func TestIsClaudeModel(t *testing.T) {
	tests := []struct {
		modelID string
		want    bool
	}{
		{"anthropic.claude-3-5-sonnet-20241022-v2:0", true},
		{"anthropic.claude-3-haiku-20240307-v1:0", true},
		{"anthropic.claude-v2", true},
		{"amazon.titan-text-express-v1", false},
		{"meta.llama3-8b-instruct-v1:0", false},
	}

	for _, tt := range tests {
		t.Run(tt.modelID, func(t *testing.T) {
			got := isClaudeModel(tt.modelID)
			if got != tt.want {
				t.Errorf("isClaudeModel(%q) = %v, want %v", tt.modelID, got, tt.want)
			}
		})
	}
}
