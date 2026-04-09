package input

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func newTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return mr, rdb
}

func TestWebhookEventLogger_Log_StoresEvent(t *testing.T) {
	mr, rdb := newTestRedis(t)
	logger := NewWebhookEventLogger(rdb)
	ctx := context.Background()

	event := WebhookEvent{
		RuleID:         "rule-123",
		Timestamp:      time.Now(),
		StatusCode:     200,
		PayloadPreview: `{"action":"test"}`,
		ErrorMessage:   "",
	}

	err := logger.Log(ctx, event)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	key := "webhook:events:rule-123"
	list, err := mr.List(key)
	if err != nil {
		t.Fatalf("failed to get list: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 event in list, got %d", len(list))
	}
}

func TestWebhookEventLogger_Log_TruncatesTo50Events(t *testing.T) {
	_, rdb := newTestRedis(t)
	logger := NewWebhookEventLogger(rdb)
	ctx := context.Background()

	// Log 55 events
	for i := 0; i < 55; i++ {
		event := WebhookEvent{
			RuleID:         "rule-456",
			Timestamp:      time.Now(),
			StatusCode:     200,
			PayloadPreview: "payload",
		}
		if err := logger.Log(ctx, event); err != nil {
			t.Fatalf("Log failed on iteration %d: %v", i, err)
		}
	}

	// Verify only 50 are kept
	events, err := logger.GetRecent(ctx, "rule-456", 100)
	if err != nil {
		t.Fatalf("GetRecent failed: %v", err)
	}
	if len(events) != 50 {
		t.Errorf("expected 50 events after truncation, got %d", len(events))
	}
}

func TestWebhookEventLogger_Log_SetsExpiry(t *testing.T) {
	mr, rdb := newTestRedis(t)
	logger := NewWebhookEventLogger(rdb)
	ctx := context.Background()

	event := WebhookEvent{
		RuleID:     "rule-789",
		Timestamp:  time.Now(),
		StatusCode: 200,
	}

	if err := logger.Log(ctx, event); err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	key := "webhook:events:rule-789"
	ttl := mr.TTL(key)
	// TTL should be approximately 24 hours (86400 seconds)
	if ttl < 23*time.Hour || ttl > 25*time.Hour {
		t.Errorf("expected TTL around 24h, got %v", ttl)
	}
}

func TestWebhookEventLogger_GetRecent_ReturnsEventsInOrder(t *testing.T) {
	_, rdb := newTestRedis(t)
	logger := NewWebhookEventLogger(rdb)
	ctx := context.Background()

	// Log events in order
	for i := 1; i <= 5; i++ {
		event := WebhookEvent{
			RuleID:         "rule-order",
			Timestamp:      time.Now().Add(time.Duration(i) * time.Second),
			StatusCode:     200 + i,
			PayloadPreview: "test",
		}
		if err := logger.Log(ctx, event); err != nil {
			t.Fatalf("Log failed: %v", err)
		}
	}

	events, err := logger.GetRecent(ctx, "rule-order", 10)
	if err != nil {
		t.Fatalf("GetRecent failed: %v", err)
	}

	if len(events) != 5 {
		t.Fatalf("expected 5 events, got %d", len(events))
	}

	// Most recent should be first (LPUSH means newest at head)
	if events[0].StatusCode != 205 {
		t.Errorf("expected most recent event (status 205) first, got %d", events[0].StatusCode)
	}
	if events[4].StatusCode != 201 {
		t.Errorf("expected oldest event (status 201) last, got %d", events[4].StatusCode)
	}
}

func TestWebhookEventLogger_GetRecent_RespectsLimit(t *testing.T) {
	_, rdb := newTestRedis(t)
	logger := NewWebhookEventLogger(rdb)
	ctx := context.Background()

	// Log 10 events
	for i := 0; i < 10; i++ {
		event := WebhookEvent{
			RuleID:     "rule-limit",
			Timestamp:  time.Now(),
			StatusCode: 200,
		}
		if err := logger.Log(ctx, event); err != nil {
			t.Fatalf("Log failed: %v", err)
		}
	}

	events, err := logger.GetRecent(ctx, "rule-limit", 3)
	if err != nil {
		t.Fatalf("GetRecent failed: %v", err)
	}

	if len(events) != 3 {
		t.Errorf("expected 3 events with limit, got %d", len(events))
	}
}

func TestWebhookEventLogger_GetRecent_EmptyList(t *testing.T) {
	_, rdb := newTestRedis(t)
	logger := NewWebhookEventLogger(rdb)
	ctx := context.Background()

	events, err := logger.GetRecent(ctx, "nonexistent-rule", 10)
	if err != nil {
		t.Fatalf("GetRecent failed: %v", err)
	}

	if len(events) != 0 {
		t.Errorf("expected empty list for nonexistent rule, got %d events", len(events))
	}
}

func TestWebhookEventLogger_Log_WithError(t *testing.T) {
	_, rdb := newTestRedis(t)
	logger := NewWebhookEventLogger(rdb)
	ctx := context.Background()

	event := WebhookEvent{
		RuleID:         "rule-error",
		Timestamp:      time.Now(),
		StatusCode:     500,
		PayloadPreview: "{}",
		ErrorMessage:   "internal server error",
	}

	if err := logger.Log(ctx, event); err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	events, err := logger.GetRecent(ctx, "rule-error", 1)
	if err != nil {
		t.Fatalf("GetRecent failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].ErrorMessage != "internal server error" {
		t.Errorf("expected error message 'internal server error', got %q", events[0].ErrorMessage)
	}
}

func TestTruncatePayload_ShortPayload(t *testing.T) {
	input := `{"short":"value"}`
	result := truncatePayload(input, 100)
	if result != input {
		t.Errorf("expected unchanged payload, got %q", result)
	}
}

func TestTruncatePayload_LongPayload(t *testing.T) {
	input := "a very long payload that exceeds the maximum length"
	result := truncatePayload(input, 20)
	// First 20 chars: "a very long payload " (includes trailing space)
	expected := "a very long payload ..."
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTruncatePayload_ExactLength(t *testing.T) {
	input := "exactly20characters!"
	result := truncatePayload(input, 20)
	if result != input {
		t.Errorf("expected unchanged payload at exact length, got %q", result)
	}
}

func TestTruncatePayload_EmptyString(t *testing.T) {
	result := truncatePayload("", 100)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}
