package redis

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/passflow-ai/passflow/pkg/common/eventbus"
	"github.com/passflow-ai/passflow/pkg/common/eventbus/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// newTestProcessor builds a StreamProcessor wired to a fresh bus with an
// observed zap logger so tests can inspect log entries.
func newTestProcessor(t *testing.T, mock *MockRedisClient) (*StreamProcessor, *RedisEventBus, *observer.ObservedLogs) {
	t.Helper()

	core, logs := observer.New(zap.ErrorLevel)
	logger := zap.New(core)

	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)
	proc := NewStreamProcessor(mock, config, bus, logger)
	return proc, bus, logs
}

// --- LOW 4: processRetries must NOT hold RLock while launching goroutines ---

// TestProcessRetries_NoDeadlockUnderRLock verifies that processRetries releases
// the read lock before dispatching goroutines.  If the lock were still held,
// a handler that tries to Ack/Nack (which acquires a write lock) would deadlock.
func TestProcessRetries_NoDeadlockUnderRLock(t *testing.T) {
	mock := NewMockRedisClient()
	proc, bus, _ := newTestProcessor(t, mock)

	ctx := context.Background()

	// Publish an event and nack it so RetryCount > 0.
	event := types.NewEvent(types.EventAgentFailed, "agent", "workspace")
	require.NoError(t, bus.Publish(ctx, *event))
	require.NoError(t, bus.NackEvent(ctx, event.ID, "initial failure"))

	// Register a handler that calls AckEvent (write lock) during processing.
	// If processRetries holds RLock when the goroutine runs, this deadlocks.
	var ackCalled bool
	var mu sync.Mutex
	_, err := bus.Subscribe(ctx, types.EventAgentFailed, func(ctx context.Context, e types.Event) error {
		_ = bus.AckEvent(ctx, e.ID) // requires write lock — must not deadlock
		mu.Lock()
		ackCalled = true
		mu.Unlock()
		return nil
	})
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		proc.processRetries(ctx)
		close(done)
	}()

	select {
	case <-done:
		// processRetries returned; give goroutines time to finish.
		time.Sleep(100 * time.Millisecond)
	case <-time.After(2 * time.Second):
		t.Fatal("processRetries deadlocked while holding RLock")
	}

	mu.Lock()
	got := ackCalled
	mu.Unlock()
	assert.True(t, got, "handler should have been called for the retryable event")
}

// TestProcessRetries_SkipsNonRetryable verifies that events that cannot be
// retried (MaxRetries exhausted) are not dispatched.
func TestProcessRetries_SkipsNonRetryable(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig() // MaxRetries == 3
	core, _ := observer.New(zap.ErrorLevel)
	proc := NewStreamProcessor(mock, config, NewRedisEventBusWithMock(mock, config), zap.New(core))
	bus := proc.bus

	ctx := context.Background()

	event := types.NewEvent(types.EventAgentFailed, "agent", "workspace")
	require.NoError(t, bus.Publish(ctx, *event))

	// Exhaust retries.
	for i := 0; i <= config.MaxRetries; i++ {
		require.NoError(t, bus.NackEvent(ctx, event.ID, "failure"))
	}

	var handlerCalled bool
	_, _ = bus.Subscribe(ctx, types.EventAgentFailed, func(ctx context.Context, e types.Event) error {
		handlerCalled = true
		return nil
	})

	proc.processRetries(ctx)
	time.Sleep(50 * time.Millisecond)

	assert.False(t, handlerCalled, "exhausted event must not be retried")
}

// --- LOW 2: processRetries must log errors from processEvent ---

// TestProcessRetries_LogsProcessEventError verifies that when processEvent
// returns an error, the error is logged with the event ID rather than silenced.
func TestProcessRetries_LogsProcessEventError(t *testing.T) {
	mock := NewMockRedisClient()
	proc, bus, logs := newTestProcessor(t, mock)

	ctx := context.Background()

	// Publish and nack so the event is queued for retry.
	event := types.NewEvent(types.EventAgentFailed, "agent", "workspace")
	require.NoError(t, bus.Publish(ctx, *event))
	require.NoError(t, bus.NackEvent(ctx, event.ID, "first failure"))

	handlerErr := errors.New("handler exploded")
	_, err := bus.Subscribe(ctx, types.EventAgentFailed, func(ctx context.Context, e types.Event) error {
		return handlerErr
	})
	require.NoError(t, err)

	proc.processRetries(ctx)
	// Give the goroutine time to run and log.
	time.Sleep(150 * time.Millisecond)

	// At least one log entry should mention the event ID.
	require.GreaterOrEqual(t, logs.Len(), 1, "expected at least one error log entry")
	found := false
	for _, entry := range logs.All() {
		for _, field := range entry.Context {
			if field.Key == "event_id" && field.String == event.ID {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "error log entry must include event_id field with the failing event's ID")
}

// TestNewStreamProcessor_WithLogger verifies that NewStreamProcessor accepts
// a logger and does not panic.
func TestNewStreamProcessor_WithLogger(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)
	logger := zap.NewNop()

	proc := NewStreamProcessor(mock, config, bus, logger)
	require.NotNil(t, proc)
}
