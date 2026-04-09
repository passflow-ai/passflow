package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jaak-ai/passflow-mcp-gateway/mcp"
	"github.com/jaak-ai/passflow-mcp-gateway/registry"
)

// Proxy forwards requests to MCP servers.
type Proxy struct {
	client   *http.Client
	registry *registry.Registry
}

// New creates a new Proxy.
func New(client *http.Client) *Proxy {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &Proxy{client: client}
}

// NewWithRegistry creates a new Proxy with a registry for circuit breaker support.
func NewWithRegistry(client *http.Client, reg *registry.Registry) *Proxy {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &Proxy{
		client:   client,
		registry: reg,
	}
}

// Forward sends a call request to an MCP server.
func (p *Proxy) Forward(server mcp.ServerInfo, req mcp.CallRequest) (mcp.CallResponse, error) {
	localReq := mcp.CallRequest{
		Name:      stripPrefix(req.Name),
		Arguments: req.Arguments,
	}

	body, err := json.Marshal(localReq)
	if err != nil {
		return mcp.CallResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := strings.TrimSuffix(server.Endpoint, "/") + "/call"
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return mcp.CallResponse{}, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		return mcp.CallResponse{}, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	var resp mcp.CallResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return mcp.CallResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return resp, nil
}

// ForwardWithCircuitBreaker sends a call request through circuit breaker if available.
// Falls back to regular Forward if circuit breaker is not configured.
func (p *Proxy) ForwardWithCircuitBreaker(server mcp.ServerInfo, req mcp.CallRequest) (mcp.CallResponse, error) {
	// If no registry or no circuit breaker, use regular Forward
	if p.registry == nil {
		return p.Forward(server, req)
	}

	cb, exists := p.registry.GetCircuitBreaker(server.Name)
	if !exists {
		return p.Forward(server, req)
	}

	// Execute through circuit breaker
	result, err := cb.Execute(func() (interface{}, error) {
		return p.Forward(server, req)
	})

	if err == registry.ErrCircuitBreakerOpen {
		return mcp.CallResponse{}, err
	}

	if err != nil {
		return mcp.CallResponse{}, err
	}

	return result.(mcp.CallResponse), nil
}

func stripPrefix(toolName string) string {
	if idx := strings.Index(toolName, "."); idx != -1 {
		return toolName[idx+1:]
	}
	return toolName
}
