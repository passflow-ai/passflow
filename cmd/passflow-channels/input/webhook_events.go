package input

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// webhookEventsKeyPrefix is the Redis key prefix for webhook event logs.
	webhookEventsKeyPrefix = "webhook:events:"
	// webhookEventsMaxItems is the maximum number of events to keep per rule.
	webhookEventsMaxItems = 50
	// webhookEventsTTL is the time-to-live for webhook event logs.
	webhookEventsTTL = 24 * time.Hour
)

// WebhookEvent represents a logged webhook event.
type WebhookEvent struct {
	RuleID         string    `json:"rule_id"`
	Timestamp      time.Time `json:"timestamp"`
	StatusCode     int       `json:"status_code"`
	PayloadPreview string    `json:"payload_preview"`
	ErrorMessage   string    `json:"error_message,omitempty"`
}

// WebhookEventLogger logs webhook events to Redis.
type WebhookEventLogger struct {
	rdb *redis.Client
}

// NewWebhookEventLogger creates a new WebhookEventLogger.
func NewWebhookEventLogger(rdb *redis.Client) *WebhookEventLogger {
	return &WebhookEventLogger{rdb: rdb}
}

// Log stores a webhook event in Redis.
// Events are stored using LPUSH (newest first), trimmed to webhookEventsMaxItems,
// and expire after webhookEventsTTL.
func (l *WebhookEventLogger) Log(ctx context.Context, event WebhookEvent) error {
	key := webhookEventsKeyPrefix + event.RuleID

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook event: %w", err)
	}

	pipe := l.rdb.Pipeline()
	pipe.LPush(ctx, key, string(data))
	pipe.LTrim(ctx, key, 0, webhookEventsMaxItems-1)
	pipe.Expire(ctx, key, webhookEventsTTL)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to log webhook event: %w", err)
	}

	return nil
}

// GetRecent retrieves the most recent webhook events for a rule.
// Events are returned in reverse chronological order (newest first).
func (l *WebhookEventLogger) GetRecent(ctx context.Context, ruleID string, limit int64) ([]WebhookEvent, error) {
	key := webhookEventsKeyPrefix + ruleID

	// LRANGE 0 to limit-1 gets the first `limit` items
	data, err := l.rdb.LRange(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook events: %w", err)
	}

	events := make([]WebhookEvent, 0, len(data))
	for _, item := range data {
		var event WebhookEvent
		if err := json.Unmarshal([]byte(item), &event); err != nil {
			continue // Skip malformed entries
		}
		events = append(events, event)
	}

	return events, nil
}

// truncatePayload truncates a payload string to the specified maximum length.
// If truncated, "..." is appended to indicate the content was cut.
func truncatePayload(payload string, maxLen int) string {
	if len(payload) <= maxLen {
		return payload
	}
	return payload[:maxLen] + "..."
}
