package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/passflow-ai/passflow/cmd/passflow-mcp-gateway/mcp"
	"github.com/passflow-ai/passflow/cmd/passflow-mcp-gateway/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForwardWithCircuitBreaker_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := mcp.CallResponse{
			Content: []mcp.Content{{Type: "text", Text: "success"}},
			IsError: false,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Setup registry with circuit breaker
	reg := registry.New()
	serverInfo := mcp.ServerInfo{
		Name:     "test-server",
		Endpoint: server.URL,
		Tools:    []string{"test_tool"},
		Healthy:  true,
	}
	reg.RegisterWithCircuitBreaker(serverInfo)

	// Create proxy with registry
	p := NewWithRegistry(nil, reg)

	// Execute request
	req := mcp.CallRequest{
		Name:      "test-server.test_tool",
		Arguments: map[string]interface{}{"arg1": "value1"},
	}

	resp, err := p.ForwardWithCircuitBreaker(serverInfo, req)
	require.NoError(t, err)
	assert.False(t, resp.IsError)
	assert.Len(t, resp.Content, 1)
	assert.Equal(t, "success", resp.Content[0].Text)

	// Verify circuit breaker state is still closed
	cb, _ := reg.GetCircuitBreaker("test-server")
	assert.Equal(t, registry.StateClosed, cb.State())
}

func TestForwardWithCircuitBreaker_ServerError(t *testing.T) {
	failCount := 0

	// Create test server that fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Setup registry with circuit breaker
	reg := registry.New()
	serverInfo := mcp.ServerInfo{
		Name:     "test-server",
		Endpoint: server.URL,
		Tools:    []string{"test_tool"},
		Healthy:  true,
	}
	reg.RegisterWithCircuitBreaker(serverInfo)

	// Create proxy with registry
	p := NewWithRegistry(nil, reg)

	req := mcp.CallRequest{
		Name:      "test-server.test_tool",
		Arguments: map[string]interface{}{},
	}

	// Execute multiple failing requests to trip circuit
	for i := 0; i < 5; i++ {
		_, _ = p.ForwardWithCircuitBreaker(serverInfo, req)
	}

	// Verify circuit breaker is open
	cb, _ := reg.GetCircuitBreaker("test-server")
	assert.Equal(t, registry.StateOpen, cb.State())

	// Next request should fail immediately without calling server
	currentFailCount := failCount
	_, err := p.ForwardWithCircuitBreaker(serverInfo, req)
	assert.Error(t, err)
	assert.Equal(t, registry.ErrCircuitBreakerOpen, err)

	// Verify server wasn't called (failCount didn't increase)
	assert.Equal(t, currentFailCount, failCount)
}

func TestForwardWithCircuitBreaker_FallbackToNonCircuitBreaker(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := mcp.CallResponse{
			Content: []mcp.Content{{Type: "text", Text: "success"}},
			IsError: false,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Setup registry WITHOUT circuit breaker (using regular Register)
	reg := registry.New()
	serverInfo := mcp.ServerInfo{
		Name:     "test-server",
		Endpoint: server.URL,
		Tools:    []string{"test_tool"},
		Healthy:  true,
	}
	reg.Register(serverInfo)

	// Create proxy with registry
	p := NewWithRegistry(nil, reg)

	req := mcp.CallRequest{
		Name:      "test-server.test_tool",
		Arguments: map[string]interface{}{},
	}

	// Should fallback to regular Forward
	resp, err := p.ForwardWithCircuitBreaker(serverInfo, req)
	require.NoError(t, err)
	assert.False(t, resp.IsError)
}

func TestProxy_WithRegistry(t *testing.T) {
	reg := registry.New()
	p := NewWithRegistry(nil, reg)

	assert.NotNil(t, p)
	assert.Equal(t, reg, p.registry)
}
