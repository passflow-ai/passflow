package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jaak-ai/passflow-mcp-gateway/mcp"
)

func TestProxyForward(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/call" {
			t.Errorf("expected path /call, got %s", r.URL.Path)
		}

		var req mcp.CallRequest
		json.NewDecoder(r.Body).Decode(&req)

		toolName := req.Name
		if idx := strings.Index(toolName, "."); idx != -1 {
			toolName = toolName[idx+1:]
		}

		if toolName != "send_message" {
			t.Errorf("expected tool 'send_message', got %q", toolName)
		}

		resp := mcp.CallResponse{
			Content: []mcp.Content{{Type: "text", Text: `{"ok":true}`}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	p := New(nil)

	server := mcp.ServerInfo{
		Name:     "slack",
		Endpoint: mockServer.URL,
	}

	req := mcp.CallRequest{
		Name:      "slack.send_message",
		Arguments: map[string]interface{}{"channel": "#test", "text": "hello"},
	}

	resp, err := p.Forward(server, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Content) != 1 {
		t.Errorf("expected 1 content block, got %d", len(resp.Content))
	}
	if resp.Content[0].Text != `{"ok":true}` {
		t.Errorf("unexpected response: %s", resp.Content[0].Text)
	}
}
