package registry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/passflow-ai/passflow/cmd/passflow-mcp-gateway/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	cb := NewCircuitBreaker("test-server", CircuitBreakerConfig{
		MaxRequests:   1,
		Interval:      100 * time.Millisecond,
		Timeout:       50 * time.Millisecond,
		ReadyToTrip:   func(counts Counts) bool { return counts.ConsecutiveFailures >= 2 },
		OnStateChange: nil,
	})

	// Initial state should be CLOSED
	assert.Equal(t, StateClosed, cb.State())

	// Execute successful request
	result, err := cb.Execute(func() (interface{}, error) {
		return "success", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, StateClosed, cb.State())

	// Execute failing requests
	_, err = cb.Execute(func() (interface{}, error) {
		return nil, errors.New("failure 1")
	})
	assert.Error(t, err)
	assert.Equal(t, StateClosed, cb.State()) // Still closed after 1 failure

	_, err = cb.Execute(func() (interface{}, error) {
		return nil, errors.New("failure 2")
	})
	assert.Error(t, err)
	assert.Equal(t, StateOpen, cb.State()) // Now open after 2 consecutive failures

	// Requests should fail immediately while open
	_, err = cb.Execute(func() (interface{}, error) {
		return "should not execute", nil
	})
	assert.Error(t, err)
	assert.Equal(t, ErrCircuitBreakerOpen, err)

	// Wait for timeout to transition to HALF_OPEN
	time.Sleep(60 * time.Millisecond)

	// Next request should transition to HALF_OPEN and execute
	result, err = cb.Execute(func() (interface{}, error) {
		return "half-open success", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "half-open success", result)
	assert.Equal(t, StateClosed, cb.State()) // Back to closed after success
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	cb := NewCircuitBreaker("test-server", CircuitBreakerConfig{
		MaxRequests:   1,
		Interval:      100 * time.Millisecond,
		Timeout:       50 * time.Millisecond,
		ReadyToTrip:   func(counts Counts) bool { return counts.ConsecutiveFailures >= 2 },
		OnStateChange: nil,
	})

	// Trip the circuit
	_, _ = cb.Execute(func() (interface{}, error) { return nil, errors.New("failure 1") })
	_, _ = cb.Execute(func() (interface{}, error) { return nil, errors.New("failure 2") })
	assert.Equal(t, StateOpen, cb.State())

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Fail during HALF_OPEN should return to OPEN
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, errors.New("half-open failure")
	})
	assert.Error(t, err)
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreaker_MaxRequestsInHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker("test-server", CircuitBreakerConfig{
		MaxRequests:   2, // Allow 2 requests in half-open
		Interval:      100 * time.Millisecond,
		Timeout:       50 * time.Millisecond,
		ReadyToTrip:   func(counts Counts) bool { return counts.ConsecutiveFailures >= 1 },
		OnStateChange: nil,
	})

	// Trip the circuit
	_, _ = cb.Execute(func() (interface{}, error) { return nil, errors.New("failure") })
	assert.Equal(t, StateOpen, cb.State())

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// First request in HALF_OPEN
	result1, err1 := cb.Execute(func() (interface{}, error) {
		return "request 1", nil
	})
	require.NoError(t, err1)
	assert.Equal(t, "request 1", result1)

	// Second request in HALF_OPEN
	result2, err2 := cb.Execute(func() (interface{}, error) {
		return "request 2", nil
	})
	require.NoError(t, err2)
	assert.Equal(t, "request 2", result2)

	// After MaxRequests successful calls, should be CLOSED
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_IntervalReset(t *testing.T) {
	cb := NewCircuitBreaker("test-server", CircuitBreakerConfig{
		MaxRequests: 1,
		Interval:    100 * time.Millisecond,
		Timeout:     50 * time.Millisecond,
		ReadyToTrip: func(counts Counts) bool { return counts.TotalFailures >= 2 },
	})

	// Single failure
	_, _ = cb.Execute(func() (interface{}, error) { return nil, errors.New("failure 1") })
	assert.Equal(t, StateClosed, cb.State())

	// Wait for interval to reset counters
	time.Sleep(110 * time.Millisecond)

	// Success after interval - counters should be reset
	result, err := cb.Execute(func() (interface{}, error) {
		return "success", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, StateClosed, cb.State())

	// Failure count should have been reset, so one failure won't trip
	_, _ = cb.Execute(func() (interface{}, error) { return nil, errors.New("failure 2") })
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_StateChangeCallback(t *testing.T) {
	stateChanges := []State{}

	cb := NewCircuitBreaker("test-server", CircuitBreakerConfig{
		MaxRequests: 1,
		Interval:    100 * time.Millisecond,
		Timeout:     50 * time.Millisecond,
		ReadyToTrip: func(counts Counts) bool { return counts.ConsecutiveFailures >= 2 },
		OnStateChange: func(name string, from, to State) {
			stateChanges = append(stateChanges, to)
		},
	})

	// Trip circuit
	_, _ = cb.Execute(func() (interface{}, error) { return nil, errors.New("failure 1") })
	_, _ = cb.Execute(func() (interface{}, error) { return nil, errors.New("failure 2") })

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Success in HALF_OPEN
	_, _ = cb.Execute(func() (interface{}, error) { return "success", nil })

	assert.Contains(t, stateChanges, StateOpen)
	assert.Contains(t, stateChanges, StateHalfOpen)
	assert.Contains(t, stateChanges, StateClosed)
}

func TestDefaultCircuitBreakerConfig(t *testing.T) {
	config := DefaultCircuitBreakerConfig()

	assert.Equal(t, uint32(1), config.MaxRequests)
	assert.Equal(t, 60*time.Second, config.Interval)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.NotNil(t, config.ReadyToTrip)

	// Test default ReadyToTrip function
	trippable := config.ReadyToTrip(Counts{TotalFailures: 5})
	assert.True(t, trippable)

	notTrippable := config.ReadyToTrip(Counts{TotalFailures: 4})
	assert.False(t, notTrippable)
}

func TestRegistryWithCircuitBreaker_HealthyServer(t *testing.T) {
	r := New()

	server := mcp.ServerInfo{
		Name:     "slack",
		Endpoint: "http://localhost:8080",
		Tools:    []string{"send_message"},
		Healthy:  true,
	}

	r.RegisterWithCircuitBreaker(server)

	// Get the server
	retrieved, ok := r.Get("slack")
	assert.True(t, ok)
	assert.Equal(t, server.Name, retrieved.Name)

	// Verify circuit breaker exists
	cb, exists := r.GetCircuitBreaker("slack")
	assert.True(t, exists)
	assert.NotNil(t, cb)
	assert.Equal(t, StateClosed, cb.State())
}

func TestRegistryWithCircuitBreaker_ExecuteWithCircuitBreaker(t *testing.T) {
	r := New()

	server := mcp.ServerInfo{
		Name:     "slack",
		Endpoint: "http://localhost:8080",
		Tools:    []string{"send_message"},
		Healthy:  true,
	}

	r.RegisterWithCircuitBreaker(server)

	ctx := context.Background()

	// Execute successful request
	result, err := r.ExecuteWithCircuitBreaker(ctx, "slack", func(ctx context.Context) (interface{}, error) {
		return "success", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "success", result)

	// Verify server is still healthy
	retrieved, _ := r.Get("slack")
	assert.True(t, retrieved.Healthy)
}

func TestRegistryWithCircuitBreaker_CircuitTrips(t *testing.T) {
	r := New()

	server := mcp.ServerInfo{
		Name:     "slack",
		Endpoint: "http://localhost:8080",
		Tools:    []string{"send_message"},
		Healthy:  true,
	}

	r.RegisterWithCircuitBreaker(server)

	ctx := context.Background()

	// Execute failing requests to trip circuit
	for i := 0; i < 5; i++ {
		_, _ = r.ExecuteWithCircuitBreaker(ctx, "slack", func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("server error")
		})
	}

	// Circuit should be open
	cb, _ := r.GetCircuitBreaker("slack")
	assert.Equal(t, StateOpen, cb.State())

	// Server should be marked unhealthy
	retrieved, _ := r.Get("slack")
	assert.False(t, retrieved.Healthy)

	// Next request should fail immediately
	_, err := r.ExecuteWithCircuitBreaker(ctx, "slack", func(ctx context.Context) (interface{}, error) {
		return "should not execute", nil
	})
	assert.Error(t, err)
	assert.Equal(t, ErrCircuitBreakerOpen, err)
}

func TestRegistryWithCircuitBreaker_ServerNotFound(t *testing.T) {
	r := New()
	ctx := context.Background()

	_, err := r.ExecuteWithCircuitBreaker(ctx, "non-existent", func(ctx context.Context) (interface{}, error) {
		return nil, nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server not found")
}
