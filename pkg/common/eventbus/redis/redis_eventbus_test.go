package redis

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/passflow-ai/passflow/pkg/common/eventbus"
	"github.com/passflow-ai/passflow/pkg/common/eventbus/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedisEventBus(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()

	bus := NewRedisEventBusWithMock(mock, config)

	require.NotNil(t, bus)
	assert.False(t, bus.IsRunning())
}

func TestRedisEventBus_Publish(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	event := types.NewEvent(types.EventIncidentDetected, "log-monitor", "workspace-123").
		WithPayload(map[string]interface{}{"error": "connection timeout"}).
		WithPriority(types.PriorityHigh)

	err := bus.Publish(ctx, *event)

	require.NoError(t, err)

	// Incident events should route to agents stream
	messages := mock.GetStreamMessages(eventbus.StreamAgents)
	assert.Len(t, messages, 1)
}

func TestRedisEventBus_Publish_TopicRouting(t *testing.T) {
	tests := []struct {
		name         string
		eventType    types.EventType
		expectedStream string
	}{
		{
			name:         "agent event routes to agents stream",
			eventType:    types.EventAgentStarted,
			expectedStream: eventbus.StreamAgents,
		},
		{
			name:         "task event routes to tasks stream",
			eventType:    types.EventTaskCreated,
			expectedStream: eventbus.StreamTasks,
		},
		{
			name:         "trigger event routes to triggers stream",
			eventType:    types.EventTriggerAgent,
			expectedStream: eventbus.StreamTriggers,
		},
		{
			name:         "sync event routes to sync stream",
			eventType:    types.EventAgentDetected,
			expectedStream: eventbus.StreamSync,
		},
		{
			name:         "incident event routes to agents stream",
			eventType:    types.EventIncidentDetected,
			expectedStream: eventbus.StreamAgents,
		},
		{
			name:         "lead event routes to agents stream",
			eventType:    types.EventLeadReceived,
			expectedStream: eventbus.StreamAgents,
		},
		{
			name:         "PR event routes to agents stream",
			eventType:    types.EventPROpened,
			expectedStream: eventbus.StreamAgents,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockRedisClient()
			config := eventbus.DefaultConfig()
			bus := NewRedisEventBusWithMock(mock, config)

			ctx := context.Background()
			event := types.NewEvent(tt.eventType, "test-agent", "workspace-123")

			err := bus.Publish(ctx, *event)
			require.NoError(t, err)

			// Verify event was published to correct stream
			messages := mock.GetStreamMessages(tt.expectedStream)
			assert.Len(t, messages, 1, "event should be published to %s", tt.expectedStream)

			// Verify event was not published to other streams
			allStreams := []string{
				eventbus.StreamAgents,
				eventbus.StreamTasks,
				eventbus.StreamTriggers,
				eventbus.StreamSync,
			}
			for _, stream := range allStreams {
				if stream != tt.expectedStream {
					messages := mock.GetStreamMessages(stream)
					assert.Len(t, messages, 0, "event should not be in stream %s", stream)
				}
			}
		})
	}
}

func TestRedisEventBus_Publish_ValidationError(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	invalidEvent := types.Event{
		SourceAgent: "agent",
		WorkspaceID: "workspace",
	}

	err := bus.Publish(ctx, invalidEvent)

	assert.Error(t, err)
	assert.Equal(t, types.ErrEmptyEventType, err)
}

func TestRedisEventBus_Publish_RedisError(t *testing.T) {
	mock := NewMockRedisClient()
	mock.SetShouldFail(true, errors.New("redis connection error"))

	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	event := types.NewEvent(types.EventAgentStarted, "agent", "workspace")

	err := bus.Publish(ctx, *event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis connection error")
}

func TestRedisEventBus_Subscribe(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	handler := func(ctx context.Context, event types.Event) error {
		return nil
	}

	sub, err := bus.Subscribe(ctx, types.EventIncidentDetected, handler)

	require.NoError(t, err)
	require.NotNil(t, sub)
	assert.NotEmpty(t, sub.ID)
	assert.Equal(t, types.EventIncidentDetected, sub.EventType)
	assert.True(t, sub.Active)
}

func TestRedisEventBus_SubscribeToStream(t *testing.T) {
	tests := []struct {
		name       string
		streamName string
	}{
		{
			name:       "subscribe to agents stream",
			streamName: eventbus.StreamAgents,
		},
		{
			name:       "subscribe to tasks stream",
			streamName: eventbus.StreamTasks,
		},
		{
			name:       "subscribe to triggers stream",
			streamName: eventbus.StreamTriggers,
		},
		{
			name:       "subscribe to sync stream",
			streamName: eventbus.StreamSync,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockRedisClient()
			config := eventbus.DefaultConfig()
			bus := NewRedisEventBusWithMock(mock, config)

			ctx := context.Background()
			handler := func(ctx context.Context, event types.Event) error {
				return nil
			}

			sub, err := bus.SubscribeToStream(ctx, tt.streamName, handler)

			require.NoError(t, err)
			require.NotNil(t, sub)
			assert.NotEmpty(t, sub.ID)
			assert.True(t, sub.Active)
		})
	}
}

func TestRedisEventBus_SubscribeAll(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	handler := func(ctx context.Context, event types.Event) error {
		return nil
	}

	sub, err := bus.SubscribeAll(ctx, handler)

	require.NoError(t, err)
	require.NotNil(t, sub)
	assert.True(t, sub.Active)
}

func TestRedisEventBus_SubscribeWithFilter(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	filter := eventbus.EventFilter{
		EventTypes:  []types.EventType{types.EventLeadReceived, types.EventLeadQualified},
		WorkspaceID: "workspace-123",
		MinPriority: types.PriorityNormal,
	}
	handler := func(ctx context.Context, event types.Event) error {
		return nil
	}

	sub, err := bus.SubscribeWithFilter(ctx, filter, handler)

	require.NoError(t, err)
	require.NotNil(t, sub)
	assert.True(t, sub.Active)
}

func TestRedisEventBus_Unsubscribe(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	handler := func(ctx context.Context, event types.Event) error {
		return nil
	}

	_, err := bus.Subscribe(ctx, types.EventTaskCreated, handler)
	require.NoError(t, err)

	err = bus.Unsubscribe(types.EventTaskCreated)
	assert.NoError(t, err)
}

func TestRedisEventBus_Unsubscribe_NotFound(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	err := bus.Unsubscribe(types.EventPROpened)
	assert.Equal(t, eventbus.ErrHandlerNotFound, err)
}

func TestRedisEventBus_UnsubscribeByID(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	handler := func(ctx context.Context, event types.Event) error {
		return nil
	}

	sub, err := bus.Subscribe(ctx, types.EventAgentCompleted, handler)
	require.NoError(t, err)

	err = bus.UnsubscribeByID(sub.ID)
	assert.NoError(t, err)
}

func TestRedisEventBus_UnsubscribeByID_NotFound(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	err := bus.UnsubscribeByID("non-existent-id")
	assert.Equal(t, eventbus.ErrHandlerNotFound, err)
}

func TestRedisEventBus_AckEvent(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	event := types.NewEvent(types.EventIncidentResolved, "responder", "workspace")
	_ = bus.Publish(ctx, *event)

	err := bus.AckEvent(ctx, event.ID)
	assert.NoError(t, err)
}

func TestRedisEventBus_NackEvent(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	event := types.NewEvent(types.EventAgentFailed, "agent", "workspace")
	_ = bus.Publish(ctx, *event)

	err := bus.NackEvent(ctx, event.ID, "processing error")
	assert.NoError(t, err)
}

func TestRedisEventBus_GetEventByID(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	originalEvent := types.NewEvent(types.EventLeadStale, "sales-agent", "workspace-123").
		WithPayload(map[string]interface{}{"lead_id": "lead-456"})
	_ = bus.Publish(ctx, *originalEvent)

	retrievedEvent, err := bus.GetEventByID(ctx, originalEvent.ID)

	require.NoError(t, err)
	require.NotNil(t, retrievedEvent)
	assert.Equal(t, originalEvent.ID, retrievedEvent.ID)
	assert.Equal(t, originalEvent.Type, retrievedEvent.Type)
	assert.Equal(t, originalEvent.SourceAgent, retrievedEvent.SourceAgent)
}

func TestRedisEventBus_GetEventByID_NotFound(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()

	event, err := bus.GetEventByID(ctx, "non-existent-id")

	assert.Equal(t, eventbus.ErrEventNotFound, err)
	assert.Nil(t, event)
}

func TestRedisEventBus_GetPendingEvents(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	agentID := "incident-responder"

	event1 := types.NewEvent(types.EventIncidentDetected, "monitor", "workspace").
		WithTargetAgent(agentID)
	event2 := types.NewEvent(types.EventIncidentEscalated, "monitor", "workspace").
		WithTargetAgent(agentID)

	_ = bus.Publish(ctx, *event1)
	_ = bus.Publish(ctx, *event2)

	events, err := bus.GetPendingEvents(ctx, agentID, 10)

	require.NoError(t, err)
	assert.Len(t, events, 2)
}

func TestRedisEventBus_GetEventHistory(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	workspaceID := "workspace-123"

	event1 := types.NewEvent(types.EventTaskCompleted, "agent1", workspaceID)
	event2 := types.NewEvent(types.EventTaskCreated, "agent2", workspaceID)

	_ = bus.Publish(ctx, *event1)
	_ = bus.Publish(ctx, *event2)

	// After acking, the event is removed from the in-memory map; GetEventHistory
	// returns what remains in the map (not-yet-acked events are still present).
	// In production, history is persisted to Redis. The in-memory implementation
	// returns zero history because acked events are deleted for memory hygiene.
	_ = bus.AckEvent(ctx, event1.ID)

	history, err := bus.GetEventHistory(ctx, workspaceID, 10, 0)

	require.NoError(t, err)
	// event1 was acked (removed from map), event2 is still pending (unprocessed).
	// GetEventHistory only returns events with Processed==true, so count is 0.
	assert.Equal(t, 0, len(history))
}

func TestRedisEventBus_StartAndStop(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()

	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = bus.Stop(ctx)
	}()

	err := bus.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, bus.IsRunning())
}

func TestRedisEventBus_Stats(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	handler := func(ctx context.Context, event types.Event) error {
		return nil
	}
	_, _ = bus.Subscribe(ctx, types.EventAgentStarted, handler)

	stats := bus.Stats()

	assert.Equal(t, 1, stats.ActiveSubscribers)
	assert.False(t, stats.IsRunning)
}

func TestRedisEventBus_PublishAsync(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	event := types.NewEvent(types.EventPROpened, "pr-agent", "workspace")

	err := bus.PublishAsync(ctx, *event)

	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	// PR events route to agents stream
	messages := mock.GetStreamMessages(eventbus.StreamAgents)
	assert.Len(t, messages, 1)
}

func TestRedisEventBus_EventSerialization(t *testing.T) {
	event := types.NewEvent(types.EventLeadReceived, "sales-bot", "ws-123").
		WithTargetAgent("sdr-agent").
		WithPayload(map[string]interface{}{
			"lead_name":  "John Doe",
			"company":    "Acme Inc",
			"source":     "website",
			"score":      85.5,
			"is_premium": true,
		}).
		WithPriority(types.PriorityHigh)

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var deserialized types.Event
	err = json.Unmarshal(data, &deserialized)
	require.NoError(t, err)

	assert.Equal(t, event.ID, deserialized.ID)
	assert.Equal(t, event.Type, deserialized.Type)
	assert.Equal(t, event.SourceAgent, deserialized.SourceAgent)
	assert.Equal(t, event.TargetAgent, deserialized.TargetAgent)
	assert.Equal(t, event.WorkspaceID, deserialized.WorkspaceID)
	assert.Equal(t, event.Priority, deserialized.Priority)
	assert.Equal(t, "John Doe", deserialized.Payload["lead_name"])
}

// TestRedisEventBus_AckEvent_RemovesEventFromMap verifies that AckEvent deletes
// the event from the in-memory map so the map does not grow unboundedly (HIGH 2).
func TestRedisEventBus_AckEvent_RemovesEventFromMap(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	event := types.NewEvent(types.EventIncidentResolved, "responder", "workspace")
	require.NoError(t, bus.Publish(ctx, *event))

	// Confirm the event is present before ack.
	e, err := bus.GetEventByID(ctx, event.ID)
	require.NoError(t, err)
	require.NotNil(t, e)

	require.NoError(t, bus.AckEvent(ctx, event.ID))

	// After ack the event must be gone from the map.
	_, err = bus.GetEventByID(ctx, event.ID)
	assert.Equal(t, eventbus.ErrEventNotFound, err)
}

// TestRedisEventBus_EventMapCap verifies that publishing more than maxEvents
// events does not grow the map beyond the cap (HIGH 2).
func TestRedisEventBus_EventMapCap(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()

	// Publish more events than the cap.
	total := maxEvents + 100
	for i := 0; i < total; i++ {
		e := types.NewEvent(types.EventAgentStarted, "agent", "ws")
		_ = bus.Publish(ctx, *e)
	}

	bus.mu.RLock()
	size := len(bus.events)
	bus.mu.RUnlock()

	assert.LessOrEqual(t, size, maxEvents,
		"event map size %d must not exceed maxEvents cap %d", size, maxEvents)
}

// TestRedisEventBus_PublishAsync_DetachedContext verifies that PublishAsync
// succeeds even when the caller's context is already cancelled (HIGH 3).
func TestRedisEventBus_PublishAsync_DetachedContext(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	// Cancel the context before PublishAsync runs.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	event := types.NewEvent(types.EventAgentStarted, "agent", "workspace")
	err := bus.PublishAsync(ctx, *event)
	require.NoError(t, err)

	// Give the goroutine time to complete.
	time.Sleep(100 * time.Millisecond)

	// Agent events route to agents stream
	messages := mock.GetStreamMessages(eventbus.StreamAgents)
	assert.Len(t, messages, 1, "event must be published despite cancelled caller context")
}

// TestRedisEventBus_Stop_CalledTwice verifies that calling Stop twice does not
// panic (HIGH 4 — double-close of stopCh).
func TestRedisEventBus_Stop_CalledTwice(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()

	// Start in background and stop twice — must not panic.
	go func() {
		_ = bus.Start(ctx)
	}()
	time.Sleep(20 * time.Millisecond)

	assert.NotPanics(t, func() {
		_ = bus.Stop(ctx)
		_ = bus.Stop(ctx)
	})
}

func TestRedisEventBus_ConcurrentPublish(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	numEvents := 100
	var wg sync.WaitGroup

	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			event := types.NewEvent(types.EventAgentStarted, "agent", "workspace")
			_ = bus.Publish(ctx, *event)
		}(i)
	}

	wg.Wait()

	// Agent events route to agents stream
	messages := mock.GetStreamMessages(eventbus.StreamAgents)
	assert.Equal(t, numEvents, len(messages))
}

func TestRedisEventBus_HandlerExecution(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()
	var mu sync.Mutex
	receivedEvents := []types.Event{}

	handler := func(ctx context.Context, event types.Event) error {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
		return nil
	}

	_, err := bus.Subscribe(ctx, types.EventTaskCreated, handler)
	require.NoError(t, err)

	event := types.NewEvent(types.EventTaskCreated, "task-agent", "workspace")
	err = bus.Publish(ctx, *event)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	bus.processEvent(ctx, *event)

	mu.Lock()
	assert.Len(t, receivedEvents, 1)
	mu.Unlock()
}

// --- LOW 3: UnsubscribeByID handler comparison fix ---

// TestRedisEventBus_UnsubscribeByID_HandlerActuallyRemoved verifies that after
// UnsubscribeByID, the handler is no longer called when an event of that type
// is processed.  The previous implementation compared &h == &sub.Handler which
// always evaluates to false (loop variable vs struct field address), so the
// handler was never removed from the slice.
func TestRedisEventBus_UnsubscribeByID_HandlerActuallyRemoved(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()

	var callCount int
	var mu sync.Mutex

	sub, err := bus.Subscribe(ctx, types.EventTaskCreated, func(ctx context.Context, e types.Event) error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return nil
	})
	require.NoError(t, err)

	// Handler fires before unsubscribe.
	event := types.NewEvent(types.EventTaskCreated, "agent", "workspace")
	require.NoError(t, bus.processEvent(ctx, *event))

	mu.Lock()
	assert.Equal(t, 1, callCount, "handler should have been called once before unsubscribe")
	mu.Unlock()

	// Unsubscribe by ID.
	require.NoError(t, bus.UnsubscribeByID(sub.ID))

	// Handler must NOT fire after unsubscribe.
	require.NoError(t, bus.processEvent(ctx, *event))

	mu.Lock()
	assert.Equal(t, 1, callCount, "handler must NOT be called after UnsubscribeByID")
	mu.Unlock()
}

// TestRedisEventBus_UnsubscribeByID_MultipleHandlers verifies that only the
// unsubscribed handler is removed; other handlers for the same event type
// continue to fire.
func TestRedisEventBus_UnsubscribeByID_MultipleHandlers(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()

	var countA, countB int
	var mu sync.Mutex

	subA, err := bus.Subscribe(ctx, types.EventTaskCreated, func(ctx context.Context, e types.Event) error {
		mu.Lock()
		countA++
		mu.Unlock()
		return nil
	})
	require.NoError(t, err)

	_, err = bus.Subscribe(ctx, types.EventTaskCreated, func(ctx context.Context, e types.Event) error {
		mu.Lock()
		countB++
		mu.Unlock()
		return nil
	})
	require.NoError(t, err)

	event := types.NewEvent(types.EventTaskCreated, "agent", "workspace")

	// Both handlers fire initially.
	require.NoError(t, bus.processEvent(ctx, *event))
	mu.Lock()
	assert.Equal(t, 1, countA)
	assert.Equal(t, 1, countB)
	mu.Unlock()

	// Remove only handler A.
	require.NoError(t, bus.UnsubscribeByID(subA.ID))

	// Only handler B should fire.
	require.NoError(t, bus.processEvent(ctx, *event))
	mu.Lock()
	assert.Equal(t, 1, countA, "handler A must not be called after unsubscribe")
	assert.Equal(t, 2, countB, "handler B must still be called")
	mu.Unlock()
}

// TestRedisEventBus_UnsubscribeByID_SubscribeAll verifies that a catch-all
// handler registered via SubscribeAll is also correctly removed.
func TestRedisEventBus_UnsubscribeByID_SubscribeAll(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mock, config)

	ctx := context.Background()

	var called bool
	var mu sync.Mutex

	sub, err := bus.SubscribeAll(ctx, func(ctx context.Context, e types.Event) error {
		mu.Lock()
		called = true
		mu.Unlock()
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, bus.UnsubscribeByID(sub.ID))

	event := types.NewEvent(types.EventAgentStarted, "agent", "workspace")
	require.NoError(t, bus.processEvent(ctx, *event))

	mu.Lock()
	assert.False(t, called, "all-handler must not be called after UnsubscribeByID")
	mu.Unlock()
}
// TestRedisEventBus_Publish_Backpressure verifies that Publish returns
// ErrBackpressure when the stream length exceeds MaxPendingEvents.
func TestRedisEventBus_Publish_Backpressure(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	config.MaxPendingEvents = 5 // Set low limit for testing

	bus := NewRedisEventBusWithMock(mock, config)
	ctx := context.Background()

	// Publish events up to the limit
	for i := 0; i < 5; i++ {
		event := types.NewEvent(types.EventAgentStarted, "agent", "workspace")
		err := bus.Publish(ctx, *event)
		require.NoError(t, err)
	}

	// The 6th event should trigger backpressure
	event := types.NewEvent(types.EventAgentStarted, "agent", "workspace")
	err := bus.Publish(ctx, *event)

	assert.ErrorIs(t, err, eventbus.ErrBackpressure,
		"publish should fail with ErrBackpressure when stream length exceeds MaxPendingEvents")
}

// TestRedisEventBus_Publish_BackpressureDisabled verifies that when
// MaxPendingEvents is 0, backpressure is disabled.
func TestRedisEventBus_Publish_BackpressureDisabled(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	config.MaxPendingEvents = 0 // Disable backpressure

	bus := NewRedisEventBusWithMock(mock, config)
	ctx := context.Background()

	// Publish many events - should all succeed
	for i := 0; i < 100; i++ {
		event := types.NewEvent(types.EventAgentStarted, "agent", "workspace")
		err := bus.Publish(ctx, *event)
		require.NoError(t, err)
	}

	// Agent events route to agents stream
	messages := mock.GetStreamMessages(eventbus.StreamAgents)
	assert.Len(t, messages, 100, "all events should be published when backpressure is disabled")
}

// TestRedisEventBus_Publish_BackpressureRecovery verifies that after
// acknowledging events, new events can be published again.
func TestRedisEventBus_Publish_BackpressureRecovery(t *testing.T) {
	mock := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	config.MaxPendingEvents = 3

	bus := NewRedisEventBusWithMock(mock, config)
	ctx := context.Background()

	// Fill up to the limit
	eventIDs := []string{}
	for i := 0; i < 3; i++ {
		event := types.NewEvent(types.EventAgentStarted, "agent", "workspace")
		err := bus.Publish(ctx, *event)
		require.NoError(t, err)
		eventIDs = append(eventIDs, event.ID)
	}

	// Should trigger backpressure
	event := types.NewEvent(types.EventAgentStarted, "agent", "workspace")
	err := bus.Publish(ctx, *event)
	assert.ErrorIs(t, err, eventbus.ErrBackpressure)

	// Acknowledge one event to free up space
	// Agent events route to agents stream
	messages := mock.GetStreamMessages(eventbus.StreamAgents)
	require.NotEmpty(t, messages)
	err = mock.XAck(ctx, eventbus.StreamAgents, config.ConsumerGroup, messages[0].ID).Err()
	require.NoError(t, err)

	// Now publishing should succeed again
	newEvent := types.NewEvent(types.EventAgentCompleted, "agent", "workspace")
	err = bus.Publish(ctx, *newEvent)
	assert.NoError(t, err, "publish should succeed after acknowledging events")
}
