package registry

import (
	"errors"
	"testing"
	"time"

	"github.com/jaak-ai/passflow-mcp-gateway/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker_ExecuteSuccess(t *testing.T) {
	cb := NewCircuitBreaker("test", DefaultCircuitBreakerConfig())

	result, err := cb.Execute(func() (interface{}, error) {
		return "success", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_ExecuteFailure(t *testing.T) {
	cb := NewCircuitBreaker("test", DefaultCircuitBreakerConfig())

	result, err := cb.Execute(func() (interface{}, error) {
		return nil, errors.New("failure")
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "failure", err.Error())
}

func TestCircuitBreaker_Name(t *testing.T) {
	cb := NewCircuitBreaker("test-server", DefaultCircuitBreakerConfig())
	assert.Equal(t, "test-server", cb.Name())
}

func TestState_String(t *testing.T) {
	assert.Equal(t, "closed", StateClosed.String())
	assert.Equal(t, "open", StateOpen.String())
	assert.Equal(t, "half-open", StateHalfOpen.String())

	// Test unknown state
	var unknownState State = 999
	assert.Equal(t, "unknown", unknownState.String())
}

func TestRegistry_RegisterWithCircuitBreaker_MultipleServers(t *testing.T) {
	r := New()

	servers := []mcp.ServerInfo{
		{Name: "slack", Endpoint: "http://localhost:8080", Tools: []string{"send"}, Healthy: true},
		{Name: "github", Endpoint: "http://localhost:8081", Tools: []string{"create_pr"}, Healthy: true},
		{Name: "kubernetes", Endpoint: "http://localhost:8082", Tools: []string{"get_pods"}, Healthy: true},
	}

	for _, server := range servers {
		r.RegisterWithCircuitBreaker(server)
	}

	// Verify all circuit breakers exist
	for _, server := range servers {
		cb, exists := r.GetCircuitBreaker(server.Name)
		assert.True(t, exists)
		assert.NotNil(t, cb)
		assert.Equal(t, StateClosed, cb.State())
	}
}

func TestRegistry_RegisterWithCircuitBreaker_Idempotent(t *testing.T) {
	r := New()

	server := mcp.ServerInfo{
		Name:     "slack",
		Endpoint: "http://localhost:8080",
		Tools:    []string{"send_message"},
		Healthy:  true,
	}

	// Register twice
	r.RegisterWithCircuitBreaker(server)
	r.RegisterWithCircuitBreaker(server)

	// Should still have only one circuit breaker
	cb, exists := r.GetCircuitBreaker("slack")
	assert.True(t, exists)
	assert.NotNil(t, cb)
}

func TestRegistry_GetCircuitBreaker_NotFound(t *testing.T) {
	r := New()

	cb, exists := r.GetCircuitBreaker("non-existent")
	assert.False(t, exists)
	assert.Nil(t, cb)
}

func TestRegistry_GetCircuitBreaker_NilMap(t *testing.T) {
	r := &Registry{
		servers:         make(map[string]mcp.ServerInfo),
		circuitBreakers: nil, // Explicitly nil
	}

	cb, exists := r.GetCircuitBreaker("any-server")
	assert.False(t, exists)
	assert.Nil(t, cb)
}

func TestCircuitBreakerConfig_DefaultReadyToTrip(t *testing.T) {
	config := DefaultCircuitBreakerConfig()

	// Test various failure counts
	tests := []struct {
		failures      uint32
		shouldTrip    bool
		description   string
	}{
		{0, false, "no failures"},
		{1, false, "one failure"},
		{4, false, "four failures (threshold - 1)"},
		{5, true, "five failures (at threshold)"},
		{10, true, "ten failures (above threshold)"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			counts := Counts{TotalFailures: tt.failures}
			result := config.ReadyToTrip(counts)
			assert.Equal(t, tt.shouldTrip, result, "failures=%d", tt.failures)
		})
	}
}

func TestCircuitBreaker_CustomReadyToTrip(t *testing.T) {
	// Custom function that trips after 2 consecutive failures
	config := CircuitBreakerConfig{
		MaxRequests: 1,
		Interval:    100 * time.Millisecond,
		Timeout:     50 * time.Millisecond,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 2
		},
	}

	cb := NewCircuitBreaker("test", config)

	// One failure - should not trip
	_, _ = cb.Execute(func() (interface{}, error) { return nil, errors.New("fail 1") })
	assert.Equal(t, StateClosed, cb.State())

	// Second consecutive failure - should trip
	_, _ = cb.Execute(func() (interface{}, error) { return nil, errors.New("fail 2") })
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreaker_StateChangeLogging(t *testing.T) {
	type stateChange struct {
		name string
		from State
		to   State
	}

	changes := []stateChange{}

	config := CircuitBreakerConfig{
		MaxRequests: 1,
		Interval:    100 * time.Millisecond,
		Timeout:     50 * time.Millisecond,
		ReadyToTrip: func(counts Counts) bool { return counts.ConsecutiveFailures >= 2 },
		OnStateChange: func(name string, from, to State) {
			changes = append(changes, stateChange{name, from, to})
		},
	}

	cb := NewCircuitBreaker("test-server", config)

	// Trip the circuit
	_, _ = cb.Execute(func() (interface{}, error) { return nil, errors.New("fail 1") })
	_, _ = cb.Execute(func() (interface{}, error) { return nil, errors.New("fail 2") })

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Recover
	_, _ = cb.Execute(func() (interface{}, error) { return "success", nil })

	// Verify all state changes were logged
	assert.True(t, len(changes) >= 3, "Expected at least 3 state changes, got %d", len(changes))

	// Find the transitions
	foundOpenTransition := false
	foundHalfOpenTransition := false
	foundClosedTransition := false

	for _, change := range changes {
		assert.Equal(t, "test-server", change.name)
		if change.to == StateOpen {
			foundOpenTransition = true
		}
		if change.to == StateHalfOpen {
			foundHalfOpenTransition = true
		}
		if change.from == StateHalfOpen && change.to == StateClosed {
			foundClosedTransition = true
		}
	}

	assert.True(t, foundOpenTransition, "Should have CLOSED -> OPEN transition")
	assert.True(t, foundHalfOpenTransition, "Should have OPEN -> HALF_OPEN transition")
	assert.True(t, foundClosedTransition, "Should have HALF_OPEN -> CLOSED transition")
}

func TestRegistry_StateChangeMarksUnhealthyWhenOpen(t *testing.T) {
	r := New()

	server := mcp.ServerInfo{
		Name:     "test",
		Endpoint: "http://localhost:8080",
		Tools:    []string{"tool"},
		Healthy:  true,
	}

	r.RegisterWithCircuitBreaker(server)

	// Verify server starts healthy
	initial, _ := r.Get("test")
	assert.True(t, initial.Healthy)

	// Trip the circuit
	cb, _ := r.GetCircuitBreaker("test")
	for i := 0; i < 5; i++ {
		_, _ = cb.Execute(func() (interface{}, error) { return nil, errors.New("fail") })
	}

	// Circuit should be open
	assert.Equal(t, StateOpen, cb.State())

	// Server should be marked unhealthy when circuit is open
	updated, _ := r.Get("test")
	assert.False(t, updated.Healthy, "Server should be unhealthy when circuit is open")
}
