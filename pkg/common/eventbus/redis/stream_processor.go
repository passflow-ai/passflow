package redis

import (
	"context"
	"encoding/json"
	"time"

	redisv8 "github.com/go-redis/redis/v8"
	"github.com/passflow-ai/passflow/pkg/common/eventbus"
	"github.com/passflow-ai/passflow/pkg/common/eventbus/types"
	"go.uber.org/zap"
)

// StreamProcessor handles Redis stream operations for event processing.
type StreamProcessor struct {
	client RedisClient
	config eventbus.EventBusConfig
	bus    *RedisEventBus
	logger *zap.Logger
}

// NewStreamProcessor creates a new StreamProcessor.
// logger is used to record retry errors; pass zap.NewNop() to discard logs.
func NewStreamProcessor(client RedisClient, config eventbus.EventBusConfig, bus *RedisEventBus, logger *zap.Logger) *StreamProcessor {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &StreamProcessor{
		client: client,
		config: config,
		bus:    bus,
		logger: logger,
	}
}

// EnsureConsumerGroup creates the consumer group if it doesn't exist.
func (p *StreamProcessor) EnsureConsumerGroup(ctx context.Context) error {
	err := p.client.XGroupCreateMkStream(
		ctx,
		p.config.StreamName,
		p.config.ConsumerGroup,
		"0",
	).Err()

	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return err
	}

	return nil
}

// ProcessLoop continuously reads and processes events from the stream.
func (p *StreamProcessor) ProcessLoop(ctx context.Context, stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		default:
			p.readAndProcessEvents(ctx)
		}
	}
}

// RetryLoop periodically retries failed events.
func (p *StreamProcessor) RetryLoop(ctx context.Context, stopCh <-chan struct{}) {
	ticker := time.NewTicker(time.Duration(p.config.RetryDelaySeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			p.processRetries(ctx)
		}
	}
}

func (p *StreamProcessor) readAndProcessEvents(ctx context.Context) {
	streams := p.client.XReadGroup(ctx, &redisv8.XReadGroupArgs{
		Group:    p.config.ConsumerGroup,
		Consumer: p.config.ConsumerName,
		Streams:  []string{p.config.StreamName, ">"},
		Count:    int64(p.config.BatchSize),
		Block:    time.Duration(p.config.BlockTimeoutMillis) * time.Millisecond,
	})

	result, err := streams.Result()
	if err != nil {
		return
	}

	for _, stream := range result {
		for _, message := range stream.Messages {
			p.processMessage(ctx, message)
		}
	}
}

func (p *StreamProcessor) processMessage(ctx context.Context, message redisv8.XMessage) {
	eventData, ok := message.Values["data"].(string)
	if !ok {
		return
	}

	var event types.Event
	if err := json.Unmarshal([]byte(eventData), &event); err != nil {
		return
	}

	if err := p.bus.processEvent(ctx, event); err != nil {
		_ = p.bus.NackEvent(ctx, event.ID, err.Error())
		return
	}

	_ = p.client.XAck(ctx, p.config.StreamName, p.config.ConsumerGroup, message.ID).Err()
	_ = p.bus.AckEvent(ctx, event.ID)
}

func (p *StreamProcessor) processRetries(ctx context.Context) {
	// LOW 4: Collect retryable events under the lock, then release it before
	// launching goroutines.  Holding RLock while spawning goroutines that
	// themselves need a write lock (e.g. AckEvent, NackEvent) causes a
	// deadlock because sync.RWMutex does not allow concurrent writers while
	// a reader holds the lock.
	p.bus.mu.RLock()
	var retryable []types.Event
	for _, event := range p.bus.events {
		if !event.Processed && event.CanRetry() && event.RetryCount > 0 {
			retryable = append(retryable, *event)
		}
	}
	p.bus.mu.RUnlock()

	// LOW 2: Log errors from processEvent so they are not silently discarded.
	for _, e := range retryable {
		go func(e types.Event) {
			if err := p.bus.processEvent(ctx, e); err != nil {
				p.logger.Error("retry processing failed",
					zap.String("event_id", e.ID),
					zap.String("event_type", string(e.Type)),
					zap.Error(err),
				)
			}
		}(e)
	}
}
