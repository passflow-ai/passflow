package nats

import (
	"context"
	"sync"

	"github.com/passflow-ai/passflow/pkg/common/eventbus"
	"github.com/passflow-ai/passflow/pkg/common/eventbus/types"
	"github.com/nats-io/nats.go"
	
)

// NATSEventBus implements EventBus using NATS JetStream.
//
// This is a skeleton implementation for evaluation purposes (ADR-002).
// It defines the structure and interface but does not include the full
// implementation of all methods. This allows for design validation and
// benchmark planning before committing to a full migration from Redis Streams.
type NATSEventBus struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	logger interface{}
	config NATSConfig

	mu            sync.RWMutex
	subscriptions map[string]*eventbus.Subscription
	running       bool
	stopCh        chan struct{}
	stopOnce      sync.Once

	stats eventbus.EventBusStats
}

// NATSConfig holds configuration for NATS JetStream EventBus.
type NATSConfig struct {
	// NATS server URLs (e.g., "nats://localhost:4222")
	URL string

	// Stream name prefix (e.g., "passflow")
	StreamPrefix string

	// Consumer group name
	ConsumerGroup string

	// Consumer name (unique per instance)
	ConsumerName string

	// Maximum retries for failed events
	MaxRetries int

	// Subject mapping strategy
	// - "flat": single subject "passflow.events"
	// - "domain": subject per domain "passflow.agents.*", "passflow.tasks.*"
	SubjectStrategy string

	// JetStream configuration
	// Replicas: number of stream replicas (for HA)
	Replicas int

	// Storage: "file" or "memory"
	Storage string

	// MaxAge: maximum age of messages in the stream
	MaxAge int64 // seconds

	// MaxBytes: maximum size of the stream
	MaxBytes int64
}

// DefaultNATSConfig returns the default configuration for NATS EventBus.
func DefaultNATSConfig() NATSConfig {
	return NATSConfig{
		URL:             "nats://localhost:4222",
		StreamPrefix:    "passflow",
		ConsumerGroup:   "passflow-agents",
		ConsumerName:    "agent-consumer",
		MaxRetries:      3,
		SubjectStrategy: "domain",
		Replicas:        3,
		Storage:         "file",
		MaxAge:          86400, // 24 hours
		MaxBytes:        10 * 1024 * 1024 * 1024, // 10GB
	}
}

// NewNATSEventBus creates a new NATSEventBus instance.
//
// Implementation steps (to be completed):
// 1. Connect to NATS server
// 2. Create JetStream context
// 3. Create or update stream with subjects
// 4. Initialize subscriptions map
//
// Example:
//   conn, _ := nats.Connect(config.URL)
//   js, _ := conn.JetStream()
//   return &NATSEventBus{conn: conn, js: js, config: config}
func NewNATSEventBus(config NATSConfig, logger interface{}) (*NATSEventBus, error) {
	// TODO: Implement connection logic
	return &NATSEventBus{
		config:        config,
		logger:        logger,
		subscriptions: make(map[string]*eventbus.Subscription),
		stopCh:        make(chan struct{}),
	}, nil
}

// Publish sends an event to the event bus.
//
// Subject routing strategy:
// - Flat: All events go to "passflow.events"
// - Domain: Events routed to "passflow.{domain}.{event_type}"
//   Examples:
//     - EventAgentStarted → "passflow.agents.started"
//     - EventTaskCreated → "passflow.tasks.created"
//     - EventTriggerAgent → "passflow.triggers.agent"
//
// Implementation steps:
// 1. Validate event
// 2. Determine subject based on strategy
// 3. Serialize event to JSON
// 4. Publish to JetStream with message ID for deduplication
// 5. Update stats
//
// Example:
//   subject := n.getSubject(event)
//   data, _ := json.Marshal(event)
//   _, err := n.js.Publish(subject, data, nats.MsgId(event.ID))
func (n *NATSEventBus) Publish(ctx context.Context, event types.Event) error {
	// TODO: Implement publish logic
	return eventbus.ErrPublishFailed
}

// PublishAsync sends an event asynchronously.
//
// Similar to Publish but uses PublishAsync for better throughput.
// The callback can be used to track publish confirmation.
//
// Implementation:
//   n.js.PublishAsync(subject, data, func(ack *nats.PubAck) {
//       if ack.Error != nil {
//           n.logger.Error("publish failed", zap.Error(ack.Error))
//       }
//   })
func (n *NATSEventBus) PublishAsync(ctx context.Context, event types.Event) error {
	// TODO: Implement async publish logic
	return eventbus.ErrPublishFailed
}

// Subscribe registers a handler for a specific event type.
//
// Creates a durable consumer for the event type's subject.
// Uses consumer groups for load balancing across instances.
//
// Implementation steps:
// 1. Determine subject from event type
// 2. Create durable consumer with consumer group
// 3. Subscribe to consumer with handler
// 4. Store subscription
//
// Example:
//   subject := n.getSubjectForEventType(eventType)
//   sub, _ := n.js.QueueSubscribe(subject, n.config.ConsumerGroup, handler)
func (n *NATSEventBus) Subscribe(
	ctx context.Context,
	eventType types.EventType,
	handler eventbus.EventHandler,
) (*eventbus.Subscription, error) {
	// TODO: Implement subscribe logic
	return nil, eventbus.ErrSubscribeFailed
}

// SubscribeAll registers a handler for all events.
//
// Subscribes to wildcard subject: "passflow.*.*" or "passflow.>"
//
// Implementation:
//   n.js.QueueSubscribe("passflow.>", n.config.ConsumerGroup, handler)
func (n *NATSEventBus) SubscribeAll(
	ctx context.Context,
	handler eventbus.EventHandler,
) (*eventbus.Subscription, error) {
	// TODO: Implement subscribe all logic
	return nil, eventbus.ErrSubscribeFailed
}

// SubscribeWithFilter registers a handler with a custom filter.
//
// Uses subject wildcards to filter at the broker level.
// Example filters:
// - All agent events: "passflow.agents.*"
// - All events for workspace: requires custom filtering in handler
//
// Implementation:
//   subject := n.buildSubjectFromFilter(filter)
//   n.js.QueueSubscribe(subject, n.config.ConsumerGroup, filteredHandler)
func (n *NATSEventBus) SubscribeWithFilter(
	ctx context.Context,
	filter eventbus.EventFilter,
	handler eventbus.EventHandler,
) (*eventbus.Subscription, error) {
	// TODO: Implement filtered subscribe logic
	return nil, eventbus.ErrSubscribeFailed
}

// Unsubscribe removes a subscription by event type.
func (n *NATSEventBus) Unsubscribe(eventType types.EventType) error {
	// TODO: Implement unsubscribe logic
	return eventbus.ErrHandlerNotFound
}

// UnsubscribeByID removes a subscription by ID.
func (n *NATSEventBus) UnsubscribeByID(subscriptionID string) error {
	// TODO: Implement unsubscribe by ID logic
	return eventbus.ErrHandlerNotFound
}

// GetPendingEvents retrieves unprocessed events for a specific agent.
//
// Uses consumer info to get pending count and fetch messages.
//
// Implementation:
//   info, _ := sub.ConsumerInfo()
//   pending := info.NumPending
//   msgs, _ := sub.Fetch(limit)
func (n *NATSEventBus) GetPendingEvents(
	ctx context.Context,
	agentID string,
	limit int,
) ([]types.Event, error) {
	// TODO: Implement get pending events logic
	return nil, nil
}

// AckEvent marks an event as processed.
//
// Acknowledges the message in NATS.
//
// Implementation:
//   msg.Ack()
func (n *NATSEventBus) AckEvent(ctx context.Context, eventID string) error {
	// TODO: Implement ack logic
	return nil
}

// NackEvent marks an event as failed and schedules retry if possible.
//
// Uses Nak() to requeue the message for retry.
// If max retries exceeded, send to dead letter queue.
//
// Implementation:
//   if retryCount < maxRetries {
//       msg.Nak()
//   } else {
//       n.publishToDLQ(event)
//       msg.Ack()
//   }
func (n *NATSEventBus) NackEvent(ctx context.Context, eventID string, reason string) error {
	// TODO: Implement nack logic
	return nil
}

// GetEventByID retrieves an event by its ID.
//
// NATS does not have native get-by-ID. Options:
// 1. Use sequence number if tracked
// 2. Query stream by message ID (requires additional indexing)
// 3. Return error (events are ephemeral in NATS)
func (n *NATSEventBus) GetEventByID(ctx context.Context, eventID string) (*types.Event, error) {
	// TODO: Implement get event by ID logic
	return nil, eventbus.ErrEventNotFound
}

// GetEventHistory retrieves processed events for a workspace.
//
// Similar to GetEventByID, NATS is optimized for streaming, not queries.
// For event history, recommend using MongoDB to store processed events.
func (n *NATSEventBus) GetEventHistory(
	ctx context.Context,
	workspaceID string,
	limit int,
	offset int,
) ([]types.Event, error) {
	// TODO: Implement get event history logic
	return nil, nil
}

// Start begins processing events.
//
// Starts all active subscriptions and begins consuming.
func (n *NATSEventBus) Start(ctx context.Context) error {
	n.mu.Lock()
	n.running = true
	n.stats.IsRunning = true
	n.mu.Unlock()

	// TODO: Start consumer loops

	<-n.stopCh
	return nil
}

// Stop gracefully stops the event bus.
func (n *NATSEventBus) Stop(ctx context.Context) error {
	n.mu.Lock()
	n.running = false
	n.stats.IsRunning = false
	n.mu.Unlock()

	n.stopOnce.Do(func() {
		close(n.stopCh)
	})

	// Drain subscriptions
	if n.conn != nil {
		n.conn.Drain()
	}

	return nil
}

// IsRunning returns true if the event bus is running.
func (n *NATSEventBus) IsRunning() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.running
}

// Stats returns statistics about the event bus.
func (n *NATSEventBus) Stats() eventbus.EventBusStats {
	n.mu.RLock()
	defer n.mu.RUnlock()

	stats := n.stats
	stats.ActiveSubscribers = len(n.subscriptions)

	// TODO: Query JetStream for pending count
	// jsInfo, _ := n.js.StreamInfo(n.config.StreamPrefix)
	// stats.TotalPending = jsInfo.State.Msgs

	return stats
}

// getSubject determines the NATS subject for an event based on strategy.
func (n *NATSEventBus) getSubject(event types.Event) string {
	switch n.config.SubjectStrategy {
	case "domain":
		return n.getSubjectDomain(event.Type)
	default:
		return n.config.StreamPrefix + ".events"
	}
}

// getSubjectDomain maps event types to domain-based subjects.
//
// Mapping:
// - agent.* → passflow.agents.*
// - task.* → passflow.tasks.*
// - incident.* → passflow.incidents.*
// - lead.* → passflow.leads.*
// - pr.* → passflow.prs.*
// - trigger.* → passflow.triggers.*
func (n *NATSEventBus) getSubjectDomain(eventType types.EventType) string {
	// Parse event type: "{domain}.{action}"
	// Examples:
	// - "agent.started" → "passflow.agents.started"
	// - "task.created" → "passflow.tasks.created"

	// TODO: Implement parsing logic
	return n.config.StreamPrefix + ".events"
}

// ensureStream creates or updates the JetStream stream.
//
// Stream configuration:
// - Name: "CHRONOS"
// - Subjects: ["passflow.>"] (all subjects under passflow)
// - Replicas: config.Replicas
// - Storage: config.Storage
// - MaxAge: config.MaxAge
// - MaxBytes: config.MaxBytes
//
// Implementation:
//   streamConfig := &nats.StreamConfig{
//       Name:     "CHRONOS",
//       Subjects: []string{"passflow.>"},
//       Replicas: n.config.Replicas,
//       Storage:  nats.FileStorage,
//       MaxAge:   time.Duration(n.config.MaxAge) * time.Second,
//       MaxBytes: n.config.MaxBytes,
//   }
//   n.js.AddStream(streamConfig)
func (n *NATSEventBus) ensureStream() error {
	// TODO: Implement stream creation logic
	return nil
}
