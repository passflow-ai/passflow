package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jaak-ai/passflow-common/eventbus"
	"github.com/jaak-ai/passflow-common/eventbus/types"
	"go.uber.org/zap"
)

const (
	eventHashPrefix     = "passflow:event:"
	pendingEventsPrefix = "passflow:pending:"
	historyPrefix       = "passflow:history:"
	retryQueueKey       = "passflow:retry"

	// maxEvents is the upper bound on in-memory event storage.
	// When the cap is reached the oldest entry is evicted before inserting
	// the new one, preventing unbounded growth (HIGH 2).
	maxEvents = 10_000
)

// RedisClient defines the interface for Redis operations.
type RedisClient interface {
	XAdd(ctx context.Context, a *redis.XAddArgs) *redis.StringCmd
	XReadGroup(ctx context.Context, a *redis.XReadGroupArgs) *redis.XStreamSliceCmd
	XAck(ctx context.Context, stream, group string, ids ...string) *redis.IntCmd
	XGroupCreateMkStream(ctx context.Context, stream, group, start string) *redis.StatusCmd
	XRevRange(ctx context.Context, stream, start, stop string) *redis.XMessageSliceCmd
	XDel(ctx context.Context, stream string, ids ...string) *redis.IntCmd
	XInfoStream(ctx context.Context, stream string) *redis.XInfoStreamCmd
	XLen(ctx context.Context, stream string) *redis.IntCmd
	HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	HGet(ctx context.Context, key, field string) *redis.StringCmd
	HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd
	ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd
	ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd
	ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Close() error
}

// filterSubscription holds a filter-based subscription.
type filterSubscription struct {
	ID      string
	Filter  eventbus.EventFilter
	Handler eventbus.EventHandler
	Active  bool
}

// RedisEventBus implements EventBus using Redis.
type RedisEventBus struct {
	client RedisClient
	config eventbus.EventBusConfig
	mu     sync.RWMutex
	// handlers maps event types to a sub-map of subscription ID → handler.
	// Using a map keyed by subscription ID instead of a slice makes
	// UnsubscribeByID O(1) and eliminates the need to compare function
	// pointers (which is unreliable in Go — LOW 3).
	handlers            map[types.EventType]map[string]eventbus.EventHandler
	subscriptions       map[string]*eventbus.Subscription
	filterSubscriptions []filterSubscription
	// allHandlers maps subscription ID → handler for catch-all subscriptions.
	allHandlers map[string]eventbus.EventHandler
	// streamHandlers maps stream name → subscription ID → handler for stream-specific subscriptions.
	streamHandlers map[string]map[string]eventbus.EventHandler
	running        bool
	stopCh         chan struct{}
	stopOnce       sync.Once // guards close(stopCh) to prevent double-close panics (HIGH 4)
	stats          eventbus.EventBusStats
	events         map[string]*types.Event
	dlq            *DLQ
	// errorHistory tracks recent errors for events to include in DLQ metadata.
	errorHistory map[string][]string
}

func newEventBus(client RedisClient, config eventbus.EventBusConfig) *RedisEventBus {
	return &RedisEventBus{
		client:              client,
		config:              config,
		handlers:            make(map[types.EventType]map[string]eventbus.EventHandler),
		subscriptions:       make(map[string]*eventbus.Subscription),
		filterSubscriptions: []filterSubscription{},
		allHandlers:         make(map[string]eventbus.EventHandler),
		streamHandlers:      make(map[string]map[string]eventbus.EventHandler),
		stopCh:              make(chan struct{}),
		events:              make(map[string]*types.Event),
		dlq:                 NewDLQ(client, config),
		errorHistory:        make(map[string][]string),
	}
}

// NewRedisEventBus creates a new RedisEventBus with a real Redis client.
func NewRedisEventBus(client *redis.Client, config eventbus.EventBusConfig) *RedisEventBus {
	return newEventBus(client, config)
}

// NewRedisEventBusWithMock creates a new RedisEventBus with a mock client for testing.
func NewRedisEventBusWithMock(client RedisClient, config eventbus.EventBusConfig) *RedisEventBus {
	return newEventBus(client, config)
}

// Publish sends an event to the event bus.
// Events are routed to domain-specific streams based on their type.
func (r *RedisEventBus) Publish(ctx context.Context, event types.Event) error {
	if err := event.Validate(); err != nil {
		return err
	}

	// Route to appropriate stream based on event type
	streamName := eventbus.GetStreamForEventType(event.Type)

	// Backpressure: Check stream length before publishing
	// Must check the actual destination stream, not the config stream
	if r.config.MaxPendingEvents > 0 {
		streamLen, err := r.client.XLen(ctx, streamName).Result()
		if err == nil && streamLen >= int64(r.config.MaxPendingEvents) {
			return eventbus.ErrBackpressure
		}
	}

	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	r.mu.Lock()
	r.storeEventCapped(event)
	r.mu.Unlock()

	err = r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: map[string]interface{}{
			"event_id": event.ID,
			"type":     string(event.Type),
			"domain":   event.Domain,
			"data":     string(eventJSON),
		},
	}).Err()

	if err != nil {
		return err
	}

	if err := r.storeEventMetadata(ctx, event); err != nil {
		return fmt.Errorf("failed to store event metadata: %w", err)
	}

	if event.TargetAgent != "" {
		if err := r.addToPendingQueue(ctx, event); err != nil {
			return fmt.Errorf("failed to add to pending queue: %w", err)
		}
	}

	r.mu.Lock()
	r.stats.TotalPublished++
	r.mu.Unlock()

	return nil
}

// PublishAsync sends an event asynchronously without waiting for processing.
// It detaches from the caller's context so that a cancelled request does not
// abort the publish goroutine (HIGH 3).
func (r *RedisEventBus) PublishAsync(ctx context.Context, event types.Event) error {
	// context.WithoutCancel returns a copy of ctx that is never cancelled,
	// preventing the goroutine from being abandoned when the HTTP handler
	// returns and cancels the original context.
	detached := context.WithoutCancel(ctx)
	go func() {
		_ = r.Publish(detached, event)
	}()
	return nil
}

// Subscribe registers a handler for a specific event type.
// Deprecated: Use SubscribeToStream for topic-based routing.
func (r *RedisEventBus) Subscribe(
	ctx context.Context,
	eventType types.EventType,
	handler eventbus.EventHandler,
) (*eventbus.Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	subID := uuid.New().String()
	subscription := &eventbus.Subscription{
		ID:        subID,
		EventType: eventType,
		Handler:   handler,
		Active:    true,
	}

	if r.handlers[eventType] == nil {
		r.handlers[eventType] = make(map[string]eventbus.EventHandler)
	}
	r.handlers[eventType][subID] = handler
	r.subscriptions[subID] = subscription

	return subscription, nil
}

// SubscribeToStream registers a handler for events from a specific stream.
// This enables topic-based routing where consumers only receive events from their domain.
func (r *RedisEventBus) SubscribeToStream(
	ctx context.Context,
	streamName string,
	handler eventbus.EventHandler,
) (*eventbus.Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	subID := uuid.New().String()
	subscription := &eventbus.Subscription{
		ID:      subID,
		Handler: handler,
		Active:  true,
	}

	if r.streamHandlers[streamName] == nil {
		r.streamHandlers[streamName] = make(map[string]eventbus.EventHandler)
	}
	r.streamHandlers[streamName][subID] = handler
	r.subscriptions[subID] = subscription

	return subscription, nil
}

// SubscribeAll registers a handler for all events.
func (r *RedisEventBus) SubscribeAll(
	ctx context.Context,
	handler eventbus.EventHandler,
) (*eventbus.Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	subID := uuid.New().String()
	subscription := &eventbus.Subscription{
		ID:      subID,
		Handler: handler,
		Active:  true,
	}

	r.allHandlers[subID] = handler
	r.subscriptions[subID] = subscription

	return subscription, nil
}

// SubscribeWithFilter registers a handler with a custom filter.
func (r *RedisEventBus) SubscribeWithFilter(
	ctx context.Context,
	filter eventbus.EventFilter,
	handler eventbus.EventHandler,
) (*eventbus.Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	subID := uuid.New().String()
	subscription := &eventbus.Subscription{
		ID:      subID,
		Handler: handler,
		Active:  true,
	}

	r.filterSubscriptions = append(r.filterSubscriptions, filterSubscription{
		ID:      subID,
		Filter:  filter,
		Handler: handler,
		Active:  true,
	})
	r.subscriptions[subID] = subscription

	return subscription, nil
}

// Unsubscribe removes a subscription by event type.
func (r *RedisEventBus) Unsubscribe(eventType types.EventType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.handlers[eventType]; !exists {
		return eventbus.ErrHandlerNotFound
	}

	delete(r.handlers, eventType)

	for id, sub := range r.subscriptions {
		if sub.EventType == eventType {
			delete(r.subscriptions, id)
		}
	}

	return nil
}

// UnsubscribeByID removes a subscription by ID.
func (r *RedisEventBus) UnsubscribeByID(subscriptionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sub, exists := r.subscriptions[subscriptionID]
	if !exists {
		return eventbus.ErrHandlerNotFound
	}

	if sub.EventType != "" {
		// LOW 3: delete by subscription ID — O(1), no function-pointer comparison.
		delete(r.handlers[sub.EventType], subscriptionID)
		if len(r.handlers[sub.EventType]) == 0 {
			delete(r.handlers, sub.EventType)
		}
	}

	// Remove from the catch-all handler map if it was registered there.
	delete(r.allHandlers, subscriptionID)

	for i, fs := range r.filterSubscriptions {
		if fs.ID == subscriptionID {
			r.filterSubscriptions = append(r.filterSubscriptions[:i], r.filterSubscriptions[i+1:]...)
			break
		}
	}

	delete(r.subscriptions, subscriptionID)

	return nil
}

// GetPendingEvents retrieves unprocessed events for a specific agent.
func (r *RedisEventBus) GetPendingEvents(
	ctx context.Context,
	agentID string,
	limit int,
) ([]types.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var pendingEvents []types.Event
	for _, event := range r.events {
		if event.TargetAgent == agentID && !event.Processed {
			pendingEvents = append(pendingEvents, *event)
			if len(pendingEvents) >= limit {
				break
			}
		}
	}

	return pendingEvents, nil
}

// AckEvent marks an event as processed and removes it from the in-memory map
// so the map does not grow unboundedly (HIGH 2).
func (r *RedisEventBus) AckEvent(ctx context.Context, eventID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if event, exists := r.events[eventID]; exists {
		event.MarkProcessed()
		r.stats.TotalProcessed++
		// Remove the event from the map after successful processing so memory
		// is reclaimed immediately rather than waiting for a periodic sweep.
		delete(r.events, eventID)
	}

	return nil
}

// NackEvent marks an event as failed and schedules retry if possible.
// If the event has exceeded MaxRetries, it is sent to the Dead Letter Queue.
func (r *RedisEventBus) NackEvent(ctx context.Context, eventID string, reason string) error {
	r.mu.Lock()

	event, exists := r.events[eventID]
	if !exists {
		r.mu.Unlock()
		return nil
	}

	// Track error history for DLQ metadata
	if r.errorHistory[eventID] == nil {
		r.errorHistory[eventID] = make([]string, 0)
	}
	r.errorHistory[eventID] = append(r.errorHistory[eventID], reason)

	event.IncrementRetry()

	// Check if event has exceeded max retries
	if !event.CanRetry() {
		r.stats.TotalFailed++

		// Get error history before unlocking
		errors := r.errorHistory[eventID]
		eventCopy := *event

		r.mu.Unlock()

		// Send to DLQ (outside lock to avoid deadlock)
		if err := r.dlq.SendToDLQ(ctx, eventCopy, errors); err != nil {
			return fmt.Errorf("failed to send event to DLQ: %w", err)
		}

		// Clean up error history
		r.mu.Lock()
		delete(r.errorHistory, eventID)
		delete(r.events, eventID)
		r.mu.Unlock()

		return nil
	}

	r.mu.Unlock()
	return nil
}

// GetEventByID retrieves an event by its ID.
func (r *RedisEventBus) GetEventByID(ctx context.Context, eventID string) (*types.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if event, exists := r.events[eventID]; exists {
		return event, nil
	}

	return nil, eventbus.ErrEventNotFound
}

// GetEventHistory retrieves processed events for a workspace.
func (r *RedisEventBus) GetEventHistory(
	ctx context.Context,
	workspaceID string,
	limit int,
	offset int,
) ([]types.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var history []types.Event
	count := 0
	skipped := 0

	for _, event := range r.events {
		if event.WorkspaceID == workspaceID && event.Processed {
			if skipped < offset {
				skipped++
				continue
			}
			history = append(history, *event)
			count++
			if count >= limit {
				break
			}
		}
	}

	return history, nil
}

// Start begins processing events.
func (r *RedisEventBus) Start(ctx context.Context) error {
	r.mu.Lock()
	r.running = true
	r.stats.IsRunning = true
	r.mu.Unlock()

	processor := NewStreamProcessor(r.client, r.config, r, zap.NewNop())
	if err := processor.EnsureConsumerGroup(ctx); err != nil {
		return err
	}

	go processor.ProcessLoop(ctx, r.stopCh)
	go processor.RetryLoop(ctx, r.stopCh)

	<-r.stopCh
	return nil
}

// Stop gracefully stops the event bus.
// It is safe to call Stop more than once; the channel is closed exactly once
// via sync.Once to prevent a double-close panic (HIGH 4).
func (r *RedisEventBus) Stop(ctx context.Context) error {
	r.mu.Lock()
	r.running = false
	r.stats.IsRunning = false
	r.mu.Unlock()

	r.stopOnce.Do(func() {
		close(r.stopCh)
	})
	return nil
}

// IsRunning returns true if the event bus is running.
func (r *RedisEventBus) IsRunning() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.running
}

// Stats returns statistics about the event bus.
func (r *RedisEventBus) Stats() eventbus.EventBusStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := r.stats
	stats.ActiveSubscribers = len(r.subscriptions)
	stats.TotalPending = int64(r.countPendingEvents())

	return stats
}

func (r *RedisEventBus) storeEventMetadata(ctx context.Context, event types.Event) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	key := eventHashPrefix + event.ID
	return r.client.HSet(ctx, key, "data", string(eventJSON)).Err()
}

func (r *RedisEventBus) addToPendingQueue(ctx context.Context, event types.Event) error {
	key := pendingEventsPrefix + event.TargetAgent
	score := float64(event.Priority)*1000000 + float64(event.Timestamp.UnixNano())/1e9
	return r.client.ZAdd(ctx, key, &redis.Z{
		Score:  score,
		Member: event.ID,
	}).Err()
}

func (r *RedisEventBus) processEvent(ctx context.Context, event types.Event) error {
	r.mu.RLock()
	// Copy the maps under the read lock to avoid holding it during handler
	// execution, which allows handlers to safely call Ack/Nack (write lock).
	typeHandlers := make(map[string]eventbus.EventHandler, len(r.handlers[event.Type]))
	for id, h := range r.handlers[event.Type] {
		typeHandlers[id] = h
	}
	allHandlers := make(map[string]eventbus.EventHandler, len(r.allHandlers))
	for id, h := range r.allHandlers {
		allHandlers[id] = h
	}
	filterSubs := r.filterSubscriptions
	r.mu.RUnlock()

	for _, handler := range typeHandlers {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}

	for _, handler := range allHandlers {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}

	for _, fs := range filterSubs {
		if fs.Active && fs.Filter.Matches(event) {
			if err := fs.Handler(ctx, event); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *RedisEventBus) countPendingEvents() int {
	count := 0
	for _, event := range r.events {
		if !event.Processed {
			count++
		}
	}
	return count
}

// storeEventCapped inserts an event into the in-memory map, evicting the oldest
// entry when the map has reached maxEvents capacity (HIGH 2).
// Must be called while holding r.mu (write lock).
func (r *RedisEventBus) storeEventCapped(event types.Event) {
	if len(r.events) >= maxEvents {
		// Evict one arbitrary entry. Map iteration order in Go is randomised, so
		// this is an O(1) eviction that avoids tracking insertion order.
		for k := range r.events {
			delete(r.events, k)
			break
		}
	}
	r.events[event.ID] = &event
}
