package acf

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryToolProvider_Store(t *testing.T) {
	t.Run("stores memory successfully", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/agents/agent-123/memory", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"mem-1","key":"test-key","value":"test-value"}`))
		}))
		defer server.Close()

		provider := &MemoryToolProvider{
			agentID:    "agent-123",
			apiBaseURL: server.URL,
			httpClient: server.Client(),
		}

		call := &ToolCall{
			CallID: "call-1",
			Tool:   "memory_store",
			Params: map[string]interface{}{
				"key":   "test-key",
				"value": "test-value",
			},
		}

		result, err := provider.Execute(context.Background(), call)
		require.NoError(t, err)
		assert.Empty(t, result.Error)
		assert.Contains(t, result.Output, "stored successfully")
	})

	t.Run("requires key parameter", func(t *testing.T) {
		provider := NewMemoryToolProvider("agent-123")
		call := &ToolCall{
			Tool:   "memory_store",
			Params: map[string]interface{}{"value": "test"},
		}

		result, err := provider.Execute(context.Background(), call)
		require.NoError(t, err)
		assert.Contains(t, result.Error, "requires 'key' parameter")
	})

	t.Run("requires value parameter", func(t *testing.T) {
		provider := NewMemoryToolProvider("agent-123")
		call := &ToolCall{
			Tool:   "memory_store",
			Params: map[string]interface{}{"key": "test"},
		}

		result, err := provider.Execute(context.Background(), call)
		require.NoError(t, err)
		assert.Contains(t, result.Error, "requires 'value' parameter")
	})
}

func TestMemoryToolProvider_Retrieve(t *testing.T) {
	t.Run("retrieves memory successfully", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/agents/agent-123/memory/test-key", r.URL.Path)
			assert.Equal(t, "GET", r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"mem-1","key":"test-key","value":"test-value","metadata":{"source":"test"}}`))
		}))
		defer server.Close()

		provider := &MemoryToolProvider{
			agentID:    "agent-123",
			apiBaseURL: server.URL,
			httpClient: server.Client(),
		}

		call := &ToolCall{
			CallID: "call-1",
			Tool:   "memory_retrieve",
			Params: map[string]interface{}{"key": "test-key"},
		}

		result, err := provider.Execute(context.Background(), call)
		require.NoError(t, err)
		assert.Empty(t, result.Error)
		assert.Contains(t, result.Output, "test-value")
		assert.Contains(t, result.Output, "Metadata")
	})

	t.Run("handles not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"not_found"}`))
		}))
		defer server.Close()

		provider := &MemoryToolProvider{
			agentID:    "agent-123",
			apiBaseURL: server.URL,
			httpClient: server.Client(),
		}

		call := &ToolCall{
			Tool:   "memory_retrieve",
			Params: map[string]interface{}{"key": "missing-key"},
		}

		result, err := provider.Execute(context.Background(), call)
		require.NoError(t, err)
		assert.Empty(t, result.Error)
		assert.Contains(t, result.Output, "No memory found")
	})
}

func TestMemoryToolProvider_List(t *testing.T) {
	t.Run("lists memories successfully", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/agents/agent-123/memory", r.URL.Path)
			assert.Equal(t, "GET", r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"memories":[{"key":"key1","value":"value1"},{"key":"key2","value":"value2"}],"count":2}`))
		}))
		defer server.Close()

		provider := &MemoryToolProvider{
			agentID:    "agent-123",
			apiBaseURL: server.URL,
			httpClient: server.Client(),
		}

		call := &ToolCall{
			CallID: "call-1",
			Tool:   "memory_list",
			Params: map[string]interface{}{},
		}

		result, err := provider.Execute(context.Background(), call)
		require.NoError(t, err)
		assert.Empty(t, result.Error)
		assert.Contains(t, result.Output, "Found 2 memory entries")
		assert.Contains(t, result.Output, "key1: value1")
		assert.Contains(t, result.Output, "key2: value2")
	})

	t.Run("handles empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"memories":[],"count":0}`))
		}))
		defer server.Close()

		provider := &MemoryToolProvider{
			agentID:    "agent-123",
			apiBaseURL: server.URL,
			httpClient: server.Client(),
		}

		call := &ToolCall{
			Tool:   "memory_list",
			Params: map[string]interface{}{},
		}

		result, err := provider.Execute(context.Background(), call)
		require.NoError(t, err)
		assert.Empty(t, result.Error)
		assert.Contains(t, result.Output, "No memories found")
	})
}

func TestMemoryToolProvider_Delete(t *testing.T) {
	t.Run("deletes memory successfully", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/agents/agent-123/memory/test-key", r.URL.Path)
			assert.Equal(t, "DELETE", r.Method)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		provider := &MemoryToolProvider{
			agentID:    "agent-123",
			apiBaseURL: server.URL,
			httpClient: server.Client(),
		}

		call := &ToolCall{
			CallID: "call-1",
			Tool:   "memory_delete",
			Params: map[string]interface{}{"key": "test-key"},
		}

		result, err := provider.Execute(context.Background(), call)
		require.NoError(t, err)
		assert.Empty(t, result.Error)
		assert.Contains(t, result.Output, "deleted successfully")
	})

	t.Run("handles not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		provider := &MemoryToolProvider{
			agentID:    "agent-123",
			apiBaseURL: server.URL,
			httpClient: server.Client(),
		}

		call := &ToolCall{
			Tool:   "memory_delete",
			Params: map[string]interface{}{"key": "missing-key"},
		}

		result, err := provider.Execute(context.Background(), call)
		require.NoError(t, err)
		assert.Empty(t, result.Error)
		assert.Contains(t, result.Output, "No memory found")
	})
}

func TestGetMemoryToolDefinitions(t *testing.T) {
	defs := GetMemoryToolDefinitions()

	assert.Len(t, defs, 4)

	tools := make(map[string]bool)
	for _, def := range defs {
		name, ok := def["name"].(string)
		require.True(t, ok)
		tools[name] = true

		// Verify each tool has required fields
		assert.NotEmpty(t, def["description"])
		assert.NotNil(t, def["input_schema"])
	}

	assert.True(t, tools["memory_store"])
	assert.True(t, tools["memory_retrieve"])
	assert.True(t, tools["memory_list"])
	assert.True(t, tools["memory_delete"])
}
