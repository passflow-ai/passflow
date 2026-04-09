package azure

import (
	"encoding/json"
	"testing"

	"github.com/jaak-ai/passflow-agent-executor/llm"
)

func TestClient_ImplementsInterface(t *testing.T) {
	var _ llm.Client = (*Client)(nil)
}

func TestConvertMessages(t *testing.T) {
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
			name: "system message is included",
			messages: []llm.Message{
				{Role: "system", Content: "You are helpful"},
				{Role: "user", Content: "Hello"},
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertMessages(tt.messages)
			if len(got) != tt.wantLen {
				t.Errorf("convertMessages() returned %d messages, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestConvertMessages_Roles(t *testing.T) {
	messages := []llm.Message{
		{Role: "system", Content: "System prompt"},
		{Role: "user", Content: "User message"},
		{Role: "assistant", Content: "Assistant response"},
	}

	got := convertMessages(messages)

	if len(got) != 3 {
		t.Fatalf("convertMessages() returned %d messages, want 3", len(got))
	}

	expectedRoles := []string{"system", "user", "assistant"}
	for i, expected := range expectedRoles {
		if string(got[i].Role) != expected {
			t.Errorf("message %d role = %q, want %q", i, got[i].Role, expected)
		}
	}
}

func TestConvertMessages_WithToolCalls(t *testing.T) {
	messages := []llm.Message{
		{
			Role:    "assistant",
			Content: "",
			ToolCalls: []llm.ToolCall{
				{
					ID:        "call_123",
					Name:      "get_weather",
					Arguments: `{"location":"SF"}`,
				},
			},
		},
	}

	got := convertMessages(messages)

	if len(got) != 1 {
		t.Fatalf("convertMessages() returned %d messages, want 1", len(got))
	}

	if len(got[0].ToolCalls) != 1 {
		t.Fatalf("message has %d tool calls, want 1", len(got[0].ToolCalls))
	}

	tc := got[0].ToolCalls[0]
	if tc.ID != "call_123" {
		t.Errorf("tool call ID = %q, want %q", tc.ID, "call_123")
	}
	if tc.Function.Name != "get_weather" {
		t.Errorf("tool call name = %q, want %q", tc.Function.Name, "get_weather")
	}
	if tc.Function.Arguments != `{"location":"SF"}` {
		t.Errorf("tool call arguments = %q, want %q", tc.Function.Arguments, `{"location":"SF"}`)
	}
}

func TestConvertMessages_WithToolResult(t *testing.T) {
	messages := []llm.Message{
		{
			Role:       "tool",
			Content:    "The weather is sunny",
			ToolCallID: "call_123",
		},
	}

	got := convertMessages(messages)

	if len(got) != 1 {
		t.Fatalf("convertMessages() returned %d messages, want 1", len(got))
	}

	if string(got[0].Role) != "tool" {
		t.Errorf("message role = %q, want %q", got[0].Role, "tool")
	}
	if got[0].ToolCallID != "call_123" {
		t.Errorf("tool call ID = %q, want %q", got[0].ToolCallID, "call_123")
	}
	if got[0].Content != "The weather is sunny" {
		t.Errorf("content = %q, want %q", got[0].Content, "The weather is sunny")
	}
}

func TestConvertTools(t *testing.T) {
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

	got := convertTools(tools)

	if len(got) != 1 {
		t.Fatalf("convertTools() returned %d tools, want 1", len(got))
	}

	if got[0].Function.Name != "get_weather" {
		t.Errorf("tool name = %q, want %q", got[0].Function.Name, "get_weather")
	}

	if got[0].Function.Description != "Get weather for a location" {
		t.Errorf("tool description = %q, want %q", got[0].Function.Description, "Get weather for a location")
	}
}

func TestConvertTools_EmptyParameters(t *testing.T) {
	tools := []llm.ToolDefinition{
		{
			Name:        "list_users",
			Description: "List all users",
			Parameters:  nil,
		},
	}

	got := convertTools(tools)

	if len(got) != 1 {
		t.Fatalf("convertTools() returned %d tools, want 1", len(got))
	}

	// Should have empty object schema when nil
	params := got[0].Function.Parameters
	if params == nil {
		t.Error("parameters should not be nil, expected empty object schema")
	}
}

func TestParseResponse_TextContent(t *testing.T) {
	content := "Hello, how can I help you?"
	choice := azureChatChoice{
		Message: azureChatMessage{
			Content: &content,
		},
	}
	usage := azureUsage{
		PromptTokens:     10,
		CompletionTokens: 5,
		TotalTokens:      15,
	}

	got := parseResponse(choice, usage)

	if got.Content != content {
		t.Errorf("content = %q, want %q", got.Content, content)
	}
	if got.Usage.PromptTokens != 10 {
		t.Errorf("prompt tokens = %d, want %d", got.Usage.PromptTokens, 10)
	}
	if got.Usage.CompletionTokens != 5 {
		t.Errorf("completion tokens = %d, want %d", got.Usage.CompletionTokens, 5)
	}
	if got.Usage.TotalTokens != 15 {
		t.Errorf("total tokens = %d, want %d", got.Usage.TotalTokens, 15)
	}
}

func TestParseResponse_WithToolCalls(t *testing.T) {
	choice := azureChatChoice{
		Message: azureChatMessage{
			ToolCalls: []azureToolCall{
				{
					ID:   "call_abc",
					Type: "function",
					Function: azureFunctionCall{
						Name:      "get_weather",
						Arguments: `{"location":"NYC"}`,
					},
				},
			},
		},
	}
	usage := azureUsage{}

	got := parseResponse(choice, usage)

	if len(got.ToolCalls) != 1 {
		t.Fatalf("tool calls length = %d, want 1", len(got.ToolCalls))
	}

	tc := got.ToolCalls[0]
	if tc.ID != "call_abc" {
		t.Errorf("tool call ID = %q, want %q", tc.ID, "call_abc")
	}
	if tc.Name != "get_weather" {
		t.Errorf("tool call name = %q, want %q", tc.Name, "get_weather")
	}
	if tc.Arguments != `{"location":"NYC"}` {
		t.Errorf("tool call arguments = %q, want %q", tc.Arguments, `{"location":"NYC"}`)
	}
}

func TestParseResponse_NilContent(t *testing.T) {
	choice := azureChatChoice{
		Message: azureChatMessage{
			Content: nil,
		},
	}
	usage := azureUsage{}

	got := parseResponse(choice, usage)

	if got.Content != "" {
		t.Errorf("content = %q, want empty string", got.Content)
	}
}

func TestSerializeParameters(t *testing.T) {
	params := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type": "string",
			},
		},
	}

	got := serializeParameters(params)

	// Should be valid JSON
	var parsed map[string]interface{}
	err := json.Unmarshal(got, &parsed)
	if err != nil {
		t.Errorf("serializeParameters() returned invalid JSON: %v", err)
	}

	if parsed["type"] != "object" {
		t.Errorf("parsed type = %v, want object", parsed["type"])
	}
}

func TestSerializeParameters_Nil(t *testing.T) {
	got := serializeParameters(nil)

	var parsed map[string]interface{}
	err := json.Unmarshal(got, &parsed)
	if err != nil {
		t.Errorf("serializeParameters(nil) returned invalid JSON: %v", err)
	}

	if parsed["type"] != "object" {
		t.Errorf("parsed type = %v, want object", parsed["type"])
	}
}

func TestNew_ValidatesEndpoint(t *testing.T) {
	_, err := New("", "api-key", "deployment")
	if err == nil {
		t.Error("New() with empty endpoint should return error")
	}
}

func TestNew_ValidatesAPIKey(t *testing.T) {
	_, err := New("https://example.openai.azure.com", "", "deployment")
	if err == nil {
		t.Error("New() with empty API key should return error")
	}
}

func TestNew_ValidatesDeployment(t *testing.T) {
	_, err := New("https://example.openai.azure.com", "api-key", "")
	if err == nil {
		t.Error("New() with empty deployment should return error")
	}
}
