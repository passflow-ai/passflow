package registry

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/passflow-ai/passflow/cmd/passflow-mcp-gateway/mcp"
	"github.com/sony/gobreaker"
)

var (
	// ErrCircuitBreakerOpen is returned when circuit is open
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
	// ErrServerNotFound is returned when server is not registered
	ErrServerNotFound = errors.New("server not found")
)

// State represents the circuit breaker state.
type State int

const (
	// StateClosed means requests are allowed and circuit is working normally
	StateClosed State = iota
	// StateOpen means requests fail immediately without calling the service
	StateOpen
	// StateHalfOpen means a limited number of requests are allowed to test recovery
	StateHalfOpen
)

// String returns the string representation of State.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Counts holds the statistics for circuit breaker.
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// CircuitBreakerConfig holds configuration for a circuit breaker.
type CircuitBreakerConfig struct {
	// MaxRequests is the maximum number of requests allowed to pass through
	// when the circuit breaker is half-open. Default: 1
	MaxRequests uint32

	// Interval is the cyclic period of the closed state for the circuit breaker
	// to clear the internal Counts. Default: 60 seconds
	Interval time.Duration

	// Timeout is the period of the open state, after which the state becomes half-open.
	// Default: 30 seconds
	Timeout time.Duration

	// ReadyToTrip is called whenever a request fails in the closed state.
	// If ReadyToTrip returns true, the circuit breaker will be placed into the open state.
	// Default: 5 consecutive failures
	ReadyToTrip func(counts Counts) bool

	// OnStateChange is called whenever the state of the circuit breaker changes.
	OnStateChange func(name string, from State, to State)
}

// DefaultCircuitBreakerConfig returns default configuration for circuit breaker.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxRequests: 1,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.TotalFailures >= 5
		},
		OnStateChange: nil,
	}
}

// CircuitBreaker wraps gobreaker.CircuitBreaker with our types.
type CircuitBreaker struct {
	name string
	cb   *gobreaker.CircuitBreaker
	mu   sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker for a server.
func NewCircuitBreaker(name string, config CircuitBreakerConfig) *CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			ourCounts := Counts{
				Requests:             counts.Requests,
				TotalSuccesses:       counts.TotalSuccesses,
				TotalFailures:        counts.TotalFailures,
				ConsecutiveSuccesses: counts.ConsecutiveSuccesses,
				ConsecutiveFailures:  counts.ConsecutiveFailures,
			}
			return config.ReadyToTrip(ourCounts)
		},
	}

	if config.OnStateChange != nil {
		settings.OnStateChange = func(name string, from gobreaker.State, to gobreaker.State) {
			config.OnStateChange(name, mapGobreakerState(from), mapGobreakerState(to))
		}
	}

	return &CircuitBreaker{
		name: name,
		cb:   gobreaker.NewCircuitBreaker(settings),
	}
}

// Execute runs the given function if the circuit breaker allows it.
func (c *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	result, err := c.cb.Execute(fn)
	if err == gobreaker.ErrOpenState {
		return nil, ErrCircuitBreakerOpen
	}
	return result, err
}

// State returns the current state of the circuit breaker.
func (c *CircuitBreaker) State() State {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return mapGobreakerState(c.cb.State())
}

// Name returns the name of the circuit breaker.
func (c *CircuitBreaker) Name() string {
	return c.name
}

func mapGobreakerState(state gobreaker.State) State {
	switch state {
	case gobreaker.StateClosed:
		return StateClosed
	case gobreaker.StateOpen:
		return StateOpen
	case gobreaker.StateHalfOpen:
		return StateHalfOpen
	default:
		return StateClosed
	}
}

// RegisterWithCircuitBreaker registers a server with circuit breaker protection.
func (r *Registry) RegisterWithCircuitBreaker(server mcp.ServerInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Register server
	r.servers[server.Name] = server

	// Create circuit breaker if not exists
	if r.circuitBreakers == nil {
		r.circuitBreakers = make(map[string]*CircuitBreaker)
	}

	if _, exists := r.circuitBreakers[server.Name]; !exists {
		config := DefaultCircuitBreakerConfig()
		config.OnStateChange = func(name string, from State, to State) {
			if to == StateOpen {
				// Mark server as unhealthy when circuit opens
				if srv, ok := r.servers[name]; ok {
					srv.Healthy = false
					r.servers[name] = srv
				}
			} else if to == StateClosed {
				// Mark server as healthy when circuit closes
				if srv, ok := r.servers[name]; ok {
					srv.Healthy = true
					r.servers[name] = srv
				}
			}
		}

		r.circuitBreakers[server.Name] = NewCircuitBreaker(server.Name, config)
	}
}

// GetCircuitBreaker returns the circuit breaker for a server.
func (r *Registry) GetCircuitBreaker(name string) (*CircuitBreaker, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.circuitBreakers == nil {
		return nil, false
	}

	cb, exists := r.circuitBreakers[name]
	return cb, exists
}

// ExecuteWithCircuitBreaker executes a function with circuit breaker protection.
func (r *Registry) ExecuteWithCircuitBreaker(
	ctx context.Context,
	serverName string,
	fn func(ctx context.Context) (interface{}, error),
) (interface{}, error) {
	cb, exists := r.GetCircuitBreaker(serverName)
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrServerNotFound, serverName)
	}

	return cb.Execute(func() (interface{}, error) {
		return fn(ctx)
	})
}
