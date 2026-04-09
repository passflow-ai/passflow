package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/passflow-ai/passflow/cmd/passflow-mcp-gateway/mcp"
	"github.com/passflow-ai/passflow/cmd/passflow-mcp-gateway/registry"
)

// TestListServers_OmitsEndpoint verifies that GET /servers does not expose
// the internal Endpoint URL of registered servers (LOW 7).
func TestListServers_OmitsEndpoint(t *testing.T) {
	reg := registry.New()
	reg.Register(mcp.ServerInfo{
		Name:     "slack",
		Endpoint: "http://mcp-slack.internal:8080", // must NOT appear in response
		Tools:    []string{"send_message", "list_channels"},
		Healthy:  true,
	})

	h := New(reg, nil)

	req := httptest.NewRequest(http.MethodGet, "/servers", nil)
	rec := httptest.NewRecorder()
	h.ListServers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// The raw response body must not contain the internal endpoint.
	body := rec.Body.String()
	if strings.Contains(body, "mcp-slack.internal") {
		t.Errorf("response body exposes internal endpoint URL; got: %s", body)
	}

	// Decode into the public response type.
	var servers []PublicServerInfo
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&servers); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}

	s := servers[0]
	if s.Name != "slack" {
		t.Errorf("Name = %q, want %q", s.Name, "slack")
	}
	if s.ToolCount != 2 {
		t.Errorf("ToolCount = %d, want 2", s.ToolCount)
	}
}

// TestListServers_MultipleServers_AllOmitEndpoints verifies that endpoint
// suppression works for multiple servers simultaneously.
func TestListServers_MultipleServers_AllOmitEndpoints(t *testing.T) {
	reg := registry.New()
	reg.Register(mcp.ServerInfo{
		Name:     "slack",
		Endpoint: "http://mcp-slack.internal:8080",
		Tools:    []string{"send_message"},
		Healthy:  true,
	})
	reg.Register(mcp.ServerInfo{
		Name:     "github",
		Endpoint: "http://mcp-github.internal:9090",
		Tools:    []string{"create_issue", "list_issues"},
		Healthy:  false,
	})

	h := New(reg, nil)

	req := httptest.NewRequest(http.MethodGet, "/servers", nil)
	rec := httptest.NewRecorder()
	h.ListServers(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, "mcp-slack.internal") || strings.Contains(body, "mcp-github.internal") {
		t.Errorf("response body exposes internal endpoints: %s", body)
	}

	var servers []PublicServerInfo
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&servers); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}
}

// TestListServers_PublicServerInfo_HasNoEndpointField verifies that
// PublicServerInfo struct does not have an Endpoint field or that it is not
// serialised.
func TestListServers_PublicServerInfo_NoEndpointInJSON(t *testing.T) {
	info := PublicServerInfo{
		Name:        "test",
		Description: "Test server",
		ToolCount:   3,
	}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if strings.Contains(string(data), "endpoint") {
		t.Errorf("PublicServerInfo JSON contains 'endpoint' field: %s", data)
	}
}

// TestCallTool_UpstreamError_DoesNotLeakInternalDetails verifies that when
// the upstream MCP server returns an error, the raw Go error message (which
// may contain internal hostnames or stack details) is NOT returned to the
// caller — only a generic message (LOW 6).
func TestCallTool_UpstreamError_DoesNotLeakInternalDetails(t *testing.T) {
	reg := registry.New()
	reg.Register(mcp.ServerInfo{
		Name:     "slack",
		Endpoint: "http://mcp-slack.internal:8080",
		Tools:    []string{"send_message"},
		Healthy:  true,
	})

	// Proxy that returns a raw internal error.
	internalErr := "dial tcp mcp-slack.internal:8080: connect: connection refused"
	proxy := &stubProxy{err: &internalProxyError{msg: internalErr}}

	h := New(reg, proxy)

	body, _ := json.Marshal(mcp.CallRequest{
		Name:      "slack.send_message",
		Arguments: map[string]interface{}{"channel": "#test", "text": "hi"},
	})
	req := httptest.NewRequest(http.MethodPost, "/tools/call", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.CallTool(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}

	responseBody := rec.Body.String()
	// The raw internal error with hostname must NOT be in the response.
	if strings.Contains(responseBody, "mcp-slack.internal") {
		t.Errorf("response leaks internal hostname: %s", responseBody)
	}
	if strings.Contains(responseBody, "connection refused") {
		t.Errorf("response leaks internal error detail: %s", responseBody)
	}

	// The response should contain a generic error message.
	var resp mcp.CallResponse
	if err := json.NewDecoder(strings.NewReader(responseBody)).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.IsError {
		t.Error("expected IsError = true in response")
	}
	if len(resp.Content) == 0 {
		t.Fatal("expected at least one Content block")
	}
	// Generic message — must not contain the internal detail.
	msg := resp.Content[0].Text
	if strings.Contains(msg, "mcp-slack.internal") || strings.Contains(msg, "connection refused") {
		t.Errorf("Content[0].Text leaks internal detail: %s", msg)
	}
}

// internalProxyError is a test error type that simulates a raw Go networking error.
type internalProxyError struct{ msg string }

func (e *internalProxyError) Error() string { return e.msg }

// stubProxy implements Proxy for testing.
type stubProxy struct {
	resp mcp.CallResponse
	err  error
}

func (p *stubProxy) Forward(_ mcp.ServerInfo, _ mcp.CallRequest) (mcp.CallResponse, error) {
	return p.resp, p.err
}
