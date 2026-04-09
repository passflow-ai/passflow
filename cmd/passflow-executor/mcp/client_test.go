package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientListTools(t *testing.T) {
	expected := ListToolsResponse{
		Tools: []Tool{
			{Name: "slack.send_message", Description: "Send a message"},
			{Name: "slack.list_channels", Description: "List channels"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tools/list" {
			t.Errorf("expected path /tools/list, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	tools, err := client.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}
	if tools[0].Name != "slack.send_message" {
		t.Errorf("expected slack.send_message, got %s", tools[0].Name)
	}
}

func TestClientCallTool(t *testing.T) {
	expected := CallResponse{
		Content: []Content{{Type: "text", Text: "Message sent"}},
		IsError: false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tools/call" {
			t.Errorf("expected path /tools/call, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req CallRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}
		if req.Name != "slack.send_message" {
			t.Errorf("expected tool slack.send_message, got %s", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	resp, err := client.CallTool(context.Background(), "slack.send_message", map[string]interface{}{
		"channel": "general",
		"text":    "Hello",
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	if resp.IsError {
		t.Error("expected IsError=false")
	}
	if len(resp.Content) != 1 {
		t.Fatalf("expected 1 content, got %d", len(resp.Content))
	}
	if resp.Content[0].Text != "Message sent" {
		t.Errorf("expected 'Message sent', got %s", resp.Content[0].Text)
	}
}

func TestClientCallToolError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.CallTool(context.Background(), "slack.send_message", nil)
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestClientSendsBearerToken(t *testing.T) {
	const token = "test-gateway-token"
	t.Setenv("MCP_GATEWAY_TOKEN", token)

	expected := ListToolsResponse{
		Tools: []Tool{{Name: "slack.send_message", Description: "Send a message"}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer "+token {
			t.Errorf("expected Authorization: Bearer %s, got %q", token, authHeader)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}
}

// TestListTools_RespectsContextCancellation verifies that a ListTools call that
// is already cancelled before it fires does not reach the server.
func TestListTools_RespectsContextCancellation(t *testing.T) {
	reached := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ListToolsResponse{})
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	client := NewClient(server.URL)
	_, err := client.ListTools(ctx)
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
	if reached {
		t.Error("server should not have been reached when context is already cancelled")
	}
}

// TestCallTool_RespectsContextCancellation verifies that a CallTool call that
// is already cancelled before it fires does not reach the server.
func TestCallTool_RespectsContextCancellation(t *testing.T) {
	reached := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CallResponse{})
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	client := NewClient(server.URL)
	_, err := client.CallTool(ctx, "slack.send_message", nil)
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
	if reached {
		t.Error("server should not have been reached when context is already cancelled")
	}
}

// TestClientNoTokenWhenEnvUnset verifies that no Authorization header is sent
// when MCP_GATEWAY_TOKEN is empty.
func TestClientNoTokenWhenEnvUnset(t *testing.T) {
	t.Setenv("MCP_GATEWAY_TOKEN", "")

	expected := ListToolsResponse{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "" {
			t.Errorf("expected no Authorization header when token is empty, got %q", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}
}
