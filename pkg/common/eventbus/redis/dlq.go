package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redisv8 "github.com/go-redis/redis/v8"
	"github.com/passflow-ai/passflow/pkg/common/eventbus"
	"github.com/passflow-ai/passflow/pkg/common/eventbus/types"
)

const (
	dlqStreamName   = "passflow:dlq"
	dlqEventPrefix  = "passflow:dlq:event:"
	dlqMetadataKey  = "passflow:dlq:metadata"
	maxDLQEventSize = 1000 // Maximum number of events to keep in DLQ
)

// DLQEvent represents an event in the Dead Letter Queue with metadata.
type DLQEvent struct {
	Event      types.Event `json:"event"`
	Errors     []string    `json:"errors"`
	RetryCount int         `json:"retry_count"`
	MaxRetries int         `json:"max_retries"`
	SentToDLQ  time.Time   `json:"sent_to_dlq"`
}

// DLQ manages the Dead Letter Queue for failed events.
type DLQ struct {
	client RedisClient
	config eventbus.EventBusConfig
}

// NewDLQ creates a new Dead Letter Queue manager.
func NewDLQ(client RedisClient, config eventbus.EventBusConfig) *DLQ {
	return &DLQ{
		client: client,
		config: config,
	}
}

// SendToDLQ sends a failed event to the Dead Letter Queue with metadata.
func (d *DLQ) SendToDLQ(ctx context.Context, event types.Event, errors []string) error {
	dlqEvent := DLQEvent{
		Event:      event,
		Errors:     errors,
		RetryCount: event.RetryCount,
		MaxRetries: event.MaxRetries,
		SentToDLQ:  time.Now().UTC(),
	}

	eventJSON, err := json.Marshal(dlqEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal DLQ event: %w", err)
	}

	// Add to DLQ stream
	streamID, err := d.client.XAdd(ctx, &redisv8.XAddArgs{
		Stream: dlqStreamName,
		MaxLen: maxDLQEventSize,
		Approx: true,
		Values: map[string]interface{}{
			"event_id": event.ID,
			"type":     string(event.Type),
			"data":     string(eventJSON),
		},
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to add event to DLQ stream: %w", err)
	}

	// Store metadata with stream ID for retrieval
	metadataKey := dlqEventPrefix + event.ID
	if err := d.client.HSet(ctx, metadataKey,
		"stream_id", streamID,
		"event_type", string(event.Type),
		"workspace_id", event.WorkspaceID,
		"sent_at", dlqEvent.SentToDLQ.Format(time.RFC3339),
		"data", string(eventJSON),
	).Err(); err != nil {
		return fmt.Errorf("failed to store DLQ metadata: %w", err)
	}

	return nil
}

// ListDLQEvents retrieves events from the Dead Letter Queue.
func (d *DLQ) ListDLQEvents(ctx context.Context, limit, offset int) ([]DLQEvent, error) {
	// Read from DLQ stream in reverse order (newest first)
	messages, err := d.client.XRevRange(ctx, dlqStreamName, "+", "-").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read DLQ stream: %w", err)
	}

	var events []DLQEvent
	for i, msg := range messages {
		if i < offset {
			continue
		}
		if len(events) >= limit {
			break
		}

		eventData, ok := msg.Values["data"].(string)
		if !ok {
			continue
		}

		var dlqEvent DLQEvent
		if err := json.Unmarshal([]byte(eventData), &dlqEvent); err != nil {
			continue
		}

		events = append(events, dlqEvent)
	}

	return events, nil
}

// RetryEvent removes an event from DLQ and re-publishes it to the main stream.
func (d *DLQ) RetryEvent(ctx context.Context, eventID string) error {
	// Get event metadata
	metadataKey := dlqEventPrefix + eventID
	metadata, err := d.client.HGetAll(ctx, metadataKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get DLQ event metadata: %w", err)
	}

	if len(metadata) == 0 {
		return fmt.Errorf("event not found in DLQ: %s", eventID)
	}

	eventData, ok := metadata["data"]
	if !ok {
		return fmt.Errorf("event data not found for: %s", eventID)
	}

	var dlqEvent DLQEvent
	if err := json.Unmarshal([]byte(eventData), &dlqEvent); err != nil {
		return fmt.Errorf("failed to unmarshal DLQ event: %w", err)
	}

	// Reset retry count for fresh attempt
	dlqEvent.Event.RetryCount = 0
	dlqEvent.Event.Processed = false
	dlqEvent.Event.ProcessedAt = nil

	// Re-publish to main stream
	eventJSON, err := json.Marshal(dlqEvent.Event)
	if err != nil {
		return fmt.Errorf("failed to marshal event for retry: %w", err)
	}

	if err := d.client.XAdd(ctx, &redisv8.XAddArgs{
		Stream: d.config.StreamName,
		Values: map[string]interface{}{
			"event_id": dlqEvent.Event.ID,
			"type":     string(dlqEvent.Event.Type),
			"data":     string(eventJSON),
		},
	}).Err(); err != nil {
		return fmt.Errorf("failed to re-publish event: %w", err)
	}

	// Remove from DLQ
	streamID, ok := metadata["stream_id"]
	if ok {
		_ = d.client.XDel(ctx, dlqStreamName, streamID).Err()
	}

	// Remove metadata
	if err := d.client.Del(ctx, metadataKey).Err(); err != nil {
		return fmt.Errorf("failed to remove DLQ metadata: %w", err)
	}

	return nil
}

// DiscardEvent permanently removes an event from the DLQ.
func (d *DLQ) DiscardEvent(ctx context.Context, eventID string) error {
	metadataKey := dlqEventPrefix + eventID

	// Get stream ID
	metadata, err := d.client.HGetAll(ctx, metadataKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get DLQ event metadata: %w", err)
	}

	// Remove from stream
	if streamID, ok := metadata["stream_id"]; ok {
		if err := d.client.XDel(ctx, dlqStreamName, streamID).Err(); err != nil {
			return fmt.Errorf("failed to remove from DLQ stream: %w", err)
		}
	}

	// Remove metadata
	if err := d.client.Del(ctx, metadataKey).Err(); err != nil {
		return fmt.Errorf("failed to remove DLQ metadata: %w", err)
	}

	return nil
}

// GetDLQStats returns statistics about the Dead Letter Queue.
func (d *DLQ) GetDLQStats(ctx context.Context) (DLQStats, error) {
	// Get stream length
	info, err := d.client.XInfoStream(ctx, dlqStreamName).Result()
	if err != nil {
		return DLQStats{}, fmt.Errorf("failed to get DLQ stream info: %w", err)
	}

	return DLQStats{
		TotalEvents: info.Length,
		OldestEvent: info.FirstEntry.ID,
		NewestEvent: info.LastEntry.ID,
	}, nil
}

// DLQStats contains statistics about the Dead Letter Queue.
type DLQStats struct {
	TotalEvents int64
	OldestEvent string
	NewestEvent string
}
