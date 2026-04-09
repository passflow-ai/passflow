package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExecutorExecute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tools/call" {
			var req CallRequest
			json.NewDecoder(r.Body).Decode(&req)

			resp := CallResponse{
				Content: []Content{{Type: "text", Text: "Executed: " + req.Name}},
				IsError: false,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	executor := NewExecutor(client)

	result, err := executor.Execute(context.Background(), "slack.send_message", map[string]interface{}{
		"channel": "general",
		"text":    "Hello",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != "Executed: slack.send_message" {
		t.Errorf("expected 'Executed: slack.send_message', got %s", result)
	}
}

func TestExecutorExecuteError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := CallResponse{
			Content: []Content{{Type: "text", Text: "Tool error: channel not found"}},
			IsError: true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	executor := NewExecutor(client)

	_, err := executor.Execute(context.Background(), "slack.send_message", nil)
	if err == nil {
		t.Error("expected error for IsError=true response")
	}
	if err.Error() != "tool error: Tool error: channel not found" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExecutorListTools(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tools/list" {
			resp := ListToolsResponse{
				Tools: []Tool{
					{Name: "slack.send_message", Description: "Send a message"},
					{Name: "github.create_issue", Description: "Create an issue"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	executor := NewExecutor(client)

	tools, err := executor.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}
}

func TestExecutorCanHandle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tools/list" {
			resp := ListToolsResponse{
				Tools: []Tool{
					{Name: "slack.send_message", Description: "Send a message"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	executor := NewExecutor(client)

	// First call loads tools
	canHandle, err := executor.CanHandle(context.Background(), "slack.send_message")
	if err != nil {
		t.Fatalf("CanHandle failed: %v", err)
	}
	if !canHandle {
		t.Error("expected CanHandle=true for slack.send_message")
	}

	// Unknown tool
	canHandle, err = executor.CanHandle(context.Background(), "unknown.tool")
	if err != nil {
		t.Fatalf("CanHandle failed: %v", err)
	}
	if canHandle {
		t.Error("expected CanHandle=false for unknown.tool")
	}
}
