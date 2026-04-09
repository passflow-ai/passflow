package anthropic

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
)

func TestClient_ImplementsInterface(t *testing.T) {
	var _ llm.Client = (*Client)(nil)
}

func TestNew(t *testing.T) {
	client := New("test-api-key", "claude-3-opus-20240229")

	if client.modelID != "claude-3-opus-20240229" {
		t.Errorf("modelID = %q, want %q", client.modelID, "claude-3-opus-20240229")
	}

	if client.apiKey != "test-api-key" {
		t.Errorf("apiKey = %q, want %q", client.apiKey, "test-api-key")
	}

	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestClose_ReturnsNil(t *testing.T) {
	client := New("test-api-key", "claude-3-opus-20240229")

	err := client.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

func TestConvertMessages_UserMessages(t *testing.T) {
	messages := []llm.Message{
		{Role: "user", Content: "Hello"},
	}

	anthropicMsgs, _ := convertMessages(messages)

	if len(anthropicMsgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(anthropicMsgs))
	}

	msg := anthropicMsgs[0]
	if msg.Role != "user" {
		t.Errorf("role = %q, want %q", msg.Role, "user")
	}
	if len(msg.Content) != 1 {
		t.Fatalf("got %d content blocks, want 1", len(msg.Content))
	}
	if msg.Content[0].Type != "text" {
		t.Errorf("content type = %q, want %q", msg.Content[0].Type, "text")
	}
	if msg.Content[0].Text != "Hello" {
		t.Errorf("content text = %q, want %q", msg.Content[0].Text, "Hello")
	}
}

func TestConvertMessages_SystemMessage(t *testing.T) {
	messages := []llm.Message{
		{Role: "system", Content: "You are helpful"},
		{Role: "user", Content: "Hello"},
	}

	anthropicMsgs, system := convertMessages(messages)

	if system != "You are helpful" {
		t.Errorf("system = %q, want %q", system, "You are helpful")
	}
	if len(anthropicMsgs) != 1 {
		t.Fatalf("got %d messages, want 1 (system excluded)", len(anthropicMsgs))
	}
}

func TestConvertMessages_AssistantWithToolCalls(t *testing.T) {
	messages := []llm.Message{
		{
			Role:    "assistant",
			Content: "Let me check",
			ToolCalls: []llm.ToolCall{
				{
					ID:        "call_123",
					Name:      "get_weather",
					Arguments: `{"location":"SF"}`,
				},
			},
		},
	}

	anthropicMsgs, _ := convertMessages(messages)

	if len(anthropicMsgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(anthropicMsgs))
	}

	msg := anthropicMsgs[0]
	if msg.Role != "assistant" {
		t.Errorf("role = %q, want %q", msg.Role, "assistant")
	}
	if len(msg.Content) != 2 {
		t.Fatalf("got %d content blocks, want 2 (text + tool_use)", len(msg.Content))
	}

	// First block should be text
	if msg.Content[0].Type != "text" {
		t.Errorf("content[0] type = %q, want %q", msg.Content[0].Type, "text")
	}
	if msg.Content[0].Text != "Let me check" {
		t.Errorf("content[0] text = %q, want %q", msg.Content[0].Text, "Let me check")
	}

	// Second block should be tool_use
	if msg.Content[1].Type != "tool_use" {
		t.Errorf("content[1] type = %q, want %q", msg.Content[1].Type, "tool_use")
	}
	if msg.Content[1].ID != "call_123" {
		t.Errorf("content[1] id = %q, want %q", msg.Content[1].ID, "call_123")
	}
	if msg.Content[1].Name != "get_weather" {
		t.Errorf("content[1] name = %q, want %q", msg.Content[1].Name, "get_weather")
	}
}

func TestConvertMessages_ToolResult(t *testing.T) {
	messages := []llm.Message{
		{
			Role:       "tool",
			Content:    "Sunny, 72F",
			ToolCallID: "call_123",
		},
	}

	anthropicMsgs, _ := convertMessages(messages)

	if len(anthropicMsgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(anthropicMsgs))
	}

	msg := anthropicMsgs[0]
	if msg.Role != "user" {
		t.Errorf("role = %q, want %q (tool results go as user)", msg.Role, "user")
	}
	if len(msg.Content) != 1 {
		t.Fatalf("got %d content blocks, want 1", len(msg.Content))
	}
	if msg.Content[0].Type != "tool_result" {
		t.Errorf("content type = %q, want %q", msg.Content[0].Type, "tool_result")
	}
	if msg.Content[0].ToolUseID != "call_123" {
		t.Errorf("tool_use_id = %q, want %q", msg.Content[0].ToolUseID, "call_123")
	}
	if msg.Content[0].Content != "Sunny, 72F" {
		t.Errorf("content = %q, want %q", msg.Content[0].Content, "Sunny, 72F")
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

	if got[0].Name != "get_weather" {
		t.Errorf("tool name = %q, want %q", got[0].Name, "get_weather")
	}

	if got[0].Description != "Get weather for a location" {
		t.Errorf("tool description = %q, want %q", got[0].Description, "Get weather for a location")
	}
}

func TestParseResponse_TextContent(t *testing.T) {
	resp := &anthropicResponse{
		ID: "msg_123",
		Content: []anthropicContent{
			{Type: "text", Text: "Hello, how can I help you?"},
		},
		Usage: struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		}{
			InputTokens:  10,
			OutputTokens: 5,
		},
		StopReason: "end_turn",
	}

	result := parseResponse(resp)

	if result.Content != "Hello, how can I help you?" {
		t.Errorf("content = %q, want %q", result.Content, "Hello, how can I help you?")
	}
	if result.Usage.PromptTokens != 10 {
		t.Errorf("prompt tokens = %d, want %d", result.Usage.PromptTokens, 10)
	}
	if result.Usage.CompletionTokens != 5 {
		t.Errorf("completion tokens = %d, want %d", result.Usage.CompletionTokens, 5)
	}
	if result.Usage.TotalTokens != 15 {
		t.Errorf("total tokens = %d, want %d", result.Usage.TotalTokens, 15)
	}
}

func TestParseResponse_WithToolCalls(t *testing.T) {
	resp := &anthropicResponse{
		ID: "msg_123",
		Content: []anthropicContent{
			{
				Type:  "tool_use",
				ID:    "call_abc",
				Name:  "get_weather",
				Input: json.RawMessage(`{"location":"NYC"}`),
			},
		},
		Usage: struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		}{},
	}

	result := parseResponse(resp)

	if len(result.ToolCalls) != 1 {
		t.Fatalf("tool calls length = %d, want 1", len(result.ToolCalls))
	}

	tc := result.ToolCalls[0]
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

// TestTruncateBody verifies the helper that sanitises error-response bodies
// before they are included in log messages.
func TestTruncateBody(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int  // max length of returned string
		wantFull bool // true when input is short enough to be returned verbatim
	}{
		{
			name:     "short body returned verbatim",
			input:    "error: bad request",
			wantLen:  len("error: bad request"),
			wantFull: true,
		},
		{
			name:    "body exactly at limit returned verbatim",
			input:   strings.Repeat("x", 200),
			wantLen: 200,
		},
		{
			name:    "body longer than limit is truncated",
			input:   strings.Repeat("a", 500),
			wantLen: 200,
		},
		{
			name:    "very long body is truncated to 200",
			input:   strings.Repeat("z", 10000),
			wantLen: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateBody(tt.input)
			if len(got) > 200 {
				t.Errorf("truncateBody() length = %d, want <= 200", len(got))
			}
			if tt.wantFull && got != tt.input {
				t.Errorf("truncateBody() = %q, want %q (full input)", got, tt.input)
			}
			if !tt.wantFull && len(got) != tt.wantLen {
				t.Errorf("truncateBody() length = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

// TestComplete_ErrorBodyTruncated verifies that a large API error response body
// is truncated before appearing in the returned error message.
func TestComplete_ErrorBodyTruncated(t *testing.T) {
	largeBody := strings.Repeat("sensitive-data", 100) // > 200 chars

	c := New("bad-key", "claude-3-opus-20240229")
	c.httpClient = &http.Client{Transport: &testTransport{
		statusCode: http.StatusUnauthorized,
		body:       largeBody,
	}}

	ctx := context.Background()
	_, err := c.Complete(ctx, []llm.Message{{Role: "user", Content: "hi"}}, nil)
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
	errMsg := err.Error()
	if len(errMsg) > 400 {
		t.Errorf("error message too long (%d chars); body should be truncated. msg=%q", len(errMsg), errMsg[:200])
	}
	if strings.Contains(errMsg, largeBody) {
		t.Error("error message must not contain the full raw response body")
	}
}

// testTransport is a minimal http.RoundTripper for unit tests.
type testTransport struct {
	statusCode int
	body       string
}

func (tt *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: tt.statusCode,
		Body:       io.NopCloser(strings.NewReader(tt.body)),
		Header:     make(http.Header),
	}, nil
}

func TestParseResponse_MixedContent(t *testing.T) {
	resp := &anthropicResponse{
		ID: "msg_123",
		Content: []anthropicContent{
			{Type: "text", Text: "I'll check the weather. "},
			{
				Type:  "tool_use",
				ID:    "call_abc",
				Name:  "get_weather",
				Input: json.RawMessage(`{"location":"NYC"}`),
			},
		},
		Usage: struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		}{},
	}

	result := parseResponse(resp)

	if result.Content != "I'll check the weather. " {
		t.Errorf("content = %q, want %q", result.Content, "I'll check the weather. ")
	}
	if len(result.ToolCalls) != 1 {
		t.Fatalf("tool calls length = %d, want 1", len(result.ToolCalls))
	}
}
