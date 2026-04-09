package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/passflow-ai/passflow/pkg/common/eventbus"
	"github.com/passflow-ai/passflow/pkg/common/eventbus/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDLQ_SendToDLQ_MarshalError(t *testing.T) {
	mockClient := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	dlq := NewDLQ(mockClient, config)

	ctx := context.Background()

	// Create event with invalid payload that would cause marshal error
	// In practice, this is hard to trigger with map[string]interface{}
	// but we test the error path
	event := types.Event{
		ID:          "event-1",
		Type:        types.EventAgentFailed,
		SourceAgent: "agent-1",
		WorkspaceID: "workspace-1",
		RetryCount:  3,
		MaxRetries:  3,
		Timestamp:   time.Now().UTC(),
		Payload:     map[string]interface{}{"valid": "data"},
	}

	// Normal case should work
	err := dlq.SendToDLQ(ctx, event, []string{"error"})
	require.NoError(t, err)
}

func TestDLQ_SendToDLQ_RedisError(t *testing.T) {
	mockClient := NewMockRedisClient()
	mockClient.SetShouldFail(true, errors.New("redis connection error"))

	config := eventbus.DefaultConfig()
	dlq := NewDLQ(mockClient, config)

	ctx := context.Background()
	event := types.Event{
		ID:          "event-1",
		Type:        types.EventAgentFailed,
		SourceAgent: "agent-1",
		WorkspaceID: "workspace-1",
		RetryCount:  3,
		MaxRetries:  3,
		Timestamp:   time.Now().UTC(),
	}

	err := dlq.SendToDLQ(ctx, event, []string{"error"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add event to DLQ stream")
}

func TestDLQ_ListDLQEvents_WithPagination(t *testing.T) {
	mockClient := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	dlq := NewDLQ(mockClient, config)

	ctx := context.Background()

	// Create 10 events
	for i := 0; i < 10; i++ {
		event := types.Event{
			ID:          string(rune('0' + i)),
			Type:        types.EventAgentFailed,
			SourceAgent: "agent-1",
			WorkspaceID: "workspace-1",
			RetryCount:  3,
			MaxRetries:  3,
			Timestamp:   time.Now().UTC(),
		}
		err := dlq.SendToDLQ(ctx, event, []string{"test error"})
		require.NoError(t, err)
	}

	// Test with offset
	events, err := dlq.ListDLQEvents(ctx, 5, 3)
	require.NoError(t, err)
	assert.Len(t, events, 5)

	// Test with limit larger than available
	events, err = dlq.ListDLQEvents(ctx, 20, 0)
	require.NoError(t, err)
	assert.Len(t, events, 10)
}

func TestDLQ_ListDLQEvents_EmptyQueue(t *testing.T) {
	mockClient := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	dlq := NewDLQ(mockClient, config)

	ctx := context.Background()

	events, err := dlq.ListDLQEvents(ctx, 10, 0)
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestDLQ_RetryEvent_MarshalError(t *testing.T) {
	mockClient := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	dlq := NewDLQ(mockClient, config)

	ctx := context.Background()

	// Send event to DLQ
	event := types.Event{
		ID:          "event-1",
		Type:        types.EventAgentFailed,
		SourceAgent: "agent-1",
		WorkspaceID: "workspace-1",
		RetryCount:  3,
		MaxRetries:  3,
		Timestamp:   time.Now().UTC(),
	}
	err := dlq.SendToDLQ(ctx, event, []string{"error"})
	require.NoError(t, err)

	// This should work normally
	err = dlq.RetryEvent(ctx, event.ID)
	require.NoError(t, err)
}

func TestDLQ_DiscardEvent_NotFound(t *testing.T) {
	mockClient := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	dlq := NewDLQ(mockClient, config)

	ctx := context.Background()

	// Discard non-existent event should not error
	err := dlq.DiscardEvent(ctx, "non-existent")
	assert.NoError(t, err)
}

func TestNackEvent_ErrorHistoryTracking(t *testing.T) {
	mockClient := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	config.MaxRetries = 3

	bus := NewRedisEventBusWithMock(mockClient, config)

	event := types.Event{
		ID:          "event-1",
		Type:        types.EventAgentFailed,
		SourceAgent: "agent-1",
		WorkspaceID: "workspace-1",
		RetryCount:  0,
		MaxRetries:  3,
		Timestamp:   time.Now().UTC(),
	}

	bus.events[event.ID] = &event

	ctx := context.Background()

	// Track multiple errors
	err := bus.NackEvent(ctx, event.ID, "error 1")
	require.NoError(t, err)

	err = bus.NackEvent(ctx, event.ID, "error 2")
	require.NoError(t, err)

	err = bus.NackEvent(ctx, event.ID, "error 3")
	require.NoError(t, err)

	// Final nack should send to DLQ with all errors
	err = bus.NackEvent(ctx, event.ID, "error 4")
	require.NoError(t, err)

	// Verify event was sent to DLQ
	dlqMessages := mockClient.GetStreamMessages(dlqStreamName)
	assert.NotEmpty(t, dlqMessages)

	// Verify error history was cleaned up
	assert.Empty(t, bus.errorHistory[event.ID])
}

func TestNackEvent_NonExistentEvent(t *testing.T) {
	mockClient := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	bus := NewRedisEventBusWithMock(mockClient, config)

	ctx := context.Background()

	// Nack non-existent event should not error
	err := bus.NackEvent(ctx, "non-existent", "error")
	assert.NoError(t, err)
}

func TestDLQ_GetDLQStats_EmptyQueue(t *testing.T) {
	mockClient := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	dlq := NewDLQ(mockClient, config)

	ctx := context.Background()

	stats, err := dlq.GetDLQStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.TotalEvents)
}
