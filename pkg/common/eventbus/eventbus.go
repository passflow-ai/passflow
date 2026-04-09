package eventbus

import (
	"context"
	"errors"

	"github.com/jaak-ai/passflow-common/eventbus/types"
)

// Common errors
var (
	ErrEventBusClosed     = errors.New("event bus is closed")
	ErrHandlerNotFound    = errors.New("handler not found for event type")
	ErrEventNotFound      = errors.New("event not found")
	ErrSubscriptionExists = errors.New("subscription already exists for event type")
	ErrInvalidEvent       = errors.New("invalid event")
	ErrPublishFailed      = errors.New("failed to publish event")
	ErrSubscribeFailed    = errors.New("failed to subscribe to event")
	ErrBackpressure       = errors.New("event bus is under backpressure, too many pending events")
)

// EventHandler is a function that processes an event.
type EventHandler func(ctx context.Context, event types.Event) error

// Subscription represents an active subscription to events.
type Subscription struct {
	ID        string
	EventType types.EventType
	Handler   EventHandler
	Active    bool
}

// EventBus defines the interface for event-based communication between agents.
type EventBus interface {
	// Publish sends an event to the event bus.
	Publish(ctx context.Context, event types.Event) error

	// PublishAsync sends an event asynchronously without waiting for processing.
	PublishAsync(ctx context.Context, event types.Event) error

	// Subscribe registers a handler for a specific event type.
	// Deprecated: Use SubscribeToStream for topic-based routing.
	Subscribe(ctx context.Context, eventType types.EventType, handler EventHandler) (*Subscription, error)

	// SubscribeToStream registers a handler for events from a specific stream.
	// This enables topic-based routing where consumers only receive events from their domain.
	SubscribeToStream(ctx context.Context, streamName string, handler EventHandler) (*Subscription, error)

	// SubscribeAll registers a handler for all events.
	SubscribeAll(ctx context.Context, handler EventHandler) (*Subscription, error)

	// SubscribeWithFilter registers a handler with a custom filter.
	SubscribeWithFilter(ctx context.Context, filter EventFilter, handler EventHandler) (*Subscription, error)

	// Unsubscribe removes a subscription by event type.
	Unsubscribe(eventType types.EventType) error

	// UnsubscribeByID removes a subscription by ID.
	UnsubscribeByID(subscriptionID string) error

	// GetPendingEvents retrieves unprocessed events for a specific agent.
	GetPendingEvents(ctx context.Context, agentID string, limit int) ([]types.Event, error)

	// AckEvent marks an event as processed.
	AckEvent(ctx context.Context, eventID string) error

	// NackEvent marks an event as failed and schedules retry if possible.
	NackEvent(ctx context.Context, eventID string, reason string) error

	// GetEventByID retrieves an event by its ID.
	GetEventByID(ctx context.Context, eventID string) (*types.Event, error)

	// GetEventHistory retrieves processed events for a workspace.
	GetEventHistory(ctx context.Context, workspaceID string, limit int, offset int) ([]types.Event, error)

	// Start begins processing events.
	Start(ctx context.Context) error

	// Stop gracefully stops the event bus.
	Stop(ctx context.Context) error

	// IsRunning returns true if the event bus is running.
	IsRunning() bool

	// Stats returns statistics about the event bus.
	Stats() EventBusStats
}

// EventFilter defines a filter for event subscriptions.
type EventFilter struct {
	EventTypes  []types.EventType
	SourceAgent string
	TargetAgent string
	WorkspaceID string
	MinPriority types.Priority
}

// Matches checks if an event matches the filter criteria.
func (f *EventFilter) Matches(event types.Event) bool {
	if len(f.EventTypes) > 0 {
		found := false
		for _, et := range f.EventTypes {
			if et == event.Type {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if f.SourceAgent != "" && f.SourceAgent != event.SourceAgent {
		return false
	}

	if f.TargetAgent != "" && f.TargetAgent != event.TargetAgent {
		return false
	}

	if f.WorkspaceID != "" && f.WorkspaceID != event.WorkspaceID {
		return false
	}

	if event.Priority < f.MinPriority {
		return false
	}

	return true
}

// EventBusStats contains statistics about the event bus.
type EventBusStats struct {
	TotalPublished    int64
	TotalProcessed    int64
	TotalFailed       int64
	TotalPending      int64
	ActiveSubscribers int
	IsRunning         bool
}

// EventBusConfig holds configuration for the event bus.
type EventBusConfig struct {
	RedisURL           string
	StreamName         string
	ConsumerGroup      string
	ConsumerName       string
	MaxRetries         int
	RetryDelaySeconds  int
	BatchSize          int
	BlockTimeoutMillis int
	MaxPendingEvents   int // Maximum pending events before backpressure kicks in
}

// DefaultConfig returns the default configuration for the event bus.
func DefaultConfig() EventBusConfig {
	return EventBusConfig{
		StreamName:         "passflow:events",
		ConsumerGroup:      "passflow-agents",
		ConsumerName:       "agent-consumer",
		MaxRetries:         3,
		RetryDelaySeconds:  5,
		BatchSize:          10,
		BlockTimeoutMillis: 5000,
		MaxPendingEvents:   10000, // Default: 10,000 pending events
	}
}
