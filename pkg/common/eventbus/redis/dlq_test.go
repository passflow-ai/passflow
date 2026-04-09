package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/passflow-ai/passflow/pkg/common/eventbus"
	"github.com/passflow-ai/passflow/pkg/common/eventbus/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDLQ_SendToDLQ(t *testing.T) {
	tests := []struct {
		name   string
		event  types.Event
		errors []string
	}{
		{
			name: "should send event to DLQ with metadata",
			event: types.Event{
				ID:          "event-1",
				Type:        types.EventAgentFailed,
				SourceAgent: "agent-1",
				WorkspaceID: "workspace-1",
				RetryCount:  3,
				MaxRetries:  3,
				Timestamp:   time.Now().UTC(),
			},
			errors: []string{"error 1", "error 2", "error 3"},
		},
		{
			name: "should include retry count metadata",
			event: types.Event{
				ID:          "event-2",
				Type:        types.EventTaskBlocked,
				SourceAgent: "agent-2",
				WorkspaceID: "workspace-2",
				RetryCount:  5,
				MaxRetries:  5,
				Timestamp:   time.Now().UTC(),
			},
			errors: []string{"connection timeout"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			config := eventbus.DefaultConfig()
			dlq := NewDLQ(mockClient, config)

			ctx := context.Background()
			err := dlq.SendToDLQ(ctx, tt.event, tt.errors)

			require.NoError(t, err)

			// Verify event was added to DLQ stream
			messages := mockClient.GetStreamMessages(dlqStreamName)
			assert.NotEmpty(t, messages)

			// Verify metadata was stored
			metadataKey := dlqEventPrefix + tt.event.ID
			metadata, exists := mockClient.GetHashValue(metadataKey, "event_type")
			assert.True(t, exists)
			assert.Equal(t, string(tt.event.Type), metadata)
		})
	}
}

func TestDLQ_ListDLQEvents(t *testing.T) {
	mockClient := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	dlq := NewDLQ(mockClient, config)

	// Create test events
	ctx := context.Background()
	for i := 1; i <= 3; i++ {
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

	// Test listing
	events, err := dlq.ListDLQEvents(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, events, 3)
}

func TestDLQ_RetryEvent(t *testing.T) {
	mockClient := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	dlq := NewDLQ(mockClient, config)

	// Send event to DLQ
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
	err := dlq.SendToDLQ(ctx, event, []string{"error 1", "error 2"})
	require.NoError(t, err)

	// Retry the event
	err = dlq.RetryEvent(ctx, event.ID)
	require.NoError(t, err)

	// Verify event was re-published to main stream
	messages := mockClient.GetStreamMessages(config.StreamName)
	assert.NotEmpty(t, messages)

	// Verify the event in main stream has reset retry count
	var republishedEvent types.Event
	eventData := messages[0].Values["data"].(string)
	err = json.Unmarshal([]byte(eventData), &republishedEvent)
	require.NoError(t, err)
	assert.Equal(t, 0, republishedEvent.RetryCount)
}

func TestDLQ_RetryEvent_NotFound(t *testing.T) {
	mockClient := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	dlq := NewDLQ(mockClient, config)

	ctx := context.Background()
	err := dlq.RetryEvent(ctx, "non-existent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event not found in DLQ")
}

func TestDLQ_DiscardEvent(t *testing.T) {
	mockClient := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	dlq := NewDLQ(mockClient, config)

	// Send event to DLQ
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
	err := dlq.SendToDLQ(ctx, event, []string{"error 1"})
	require.NoError(t, err)

	// Discard the event
	err = dlq.DiscardEvent(ctx, event.ID)
	require.NoError(t, err)

	// Verify metadata was removed
	metadataKey := dlqEventPrefix + event.ID
	_, exists := mockClient.GetHashValue(metadataKey, "event_type")
	assert.False(t, exists)
}

func TestNackEvent_SendsToDLQAfterMaxRetries(t *testing.T) {
	mockClient := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	config.MaxRetries = 3

	bus := NewRedisEventBusWithMock(mockClient, config)

	event := types.Event{
		ID:          "event-1",
		Type:        types.EventAgentFailed,
		SourceAgent: "agent-1",
		WorkspaceID: "workspace-1",
		RetryCount:  2, // One retry left
		MaxRetries:  3,
		Timestamp:   time.Now().UTC(),
	}

	bus.events[event.ID] = &event

	ctx := context.Background()

	// First nack - should increment to 3 retries (still can retry)
	err := bus.NackEvent(ctx, event.ID, "processing failed")
	require.NoError(t, err)

	// Check if event still exists (should exist if can retry)
	if ev, exists := bus.events[event.ID]; exists {
		assert.Equal(t, 3, ev.RetryCount)
		assert.False(t, ev.CanRetry()) // At max retries, but hasn't been nacked again yet
	}

	// Second nack - should send to DLQ (exceeded max retries)
	err = bus.NackEvent(ctx, event.ID, "processing failed again")
	require.NoError(t, err)

	// Verify event was sent to DLQ
	dlqMessages := mockClient.GetStreamMessages(dlqStreamName)
	assert.NotEmpty(t, dlqMessages)

	// Verify event was removed from main events map
	_, exists := bus.events[event.ID]
	assert.False(t, exists)
}

func TestDLQ_GetDLQStats(t *testing.T) {
	mockClient := NewMockRedisClient()
	config := eventbus.DefaultConfig()
	dlq := NewDLQ(mockClient, config)

	// Send events to DLQ
	ctx := context.Background()
	for i := 1; i <= 5; i++ {
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

	// Get stats
	stats, err := dlq.GetDLQStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(5), stats.TotalEvents)
	assert.NotEmpty(t, stats.OldestEvent)
	assert.NotEmpty(t, stats.NewestEvent)
}
