package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/jaak-ai/passflow-common/eventbus"
	"github.com/jaak-ai/passflow-common/eventbus/types"
)

// Logger defines a minimal logging interface to avoid external dependencies.
type Logger interface {
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// SimpleLogger is a basic logger implementation using standard output.
type SimpleLogger struct{}

func (s *SimpleLogger) Info(msg string, args ...interface{}) {
	fmt.Printf("[INFO] %s %v\n", msg, args)
}

func (s *SimpleLogger) Debug(msg string, args ...interface{}) {
	fmt.Printf("[DEBUG] %s %v\n", msg, args)
}

func (s *SimpleLogger) Warn(msg string, args ...interface{}) {
	fmt.Printf("[WARN] %s %v\n", msg, args)
}

func (s *SimpleLogger) Error(msg string, args ...interface{}) {
	fmt.Printf("[ERROR] %s %v\n", msg, args)
}

// EventLogger provides logging middleware for event processing.
type EventLogger struct {
	log Logger
}

// NewEventLogger creates a new EventLogger.
func NewEventLogger(log Logger) *EventLogger {
	return &EventLogger{log: log}
}

// LogPublish logs when an event is published.
func (l *EventLogger) LogPublish(event types.Event) {
	l.log.Info(
		fmt.Sprintf("Event published: id=%s type=%s source=%s target=%s workspace=%s priority=%s",
			event.ID, event.Type, event.SourceAgent, event.TargetAgent, event.WorkspaceID, event.Priority.String()),
	)
}

// LogReceive logs when an event is received.
func (l *EventLogger) LogReceive(event types.Event) {
	l.log.Debug(
		fmt.Sprintf("Event received: id=%s type=%s source=%s",
			event.ID, event.Type, event.SourceAgent),
	)
}

// LogProcessed logs when an event is successfully processed.
func (l *EventLogger) LogProcessed(event types.Event, duration time.Duration) {
	l.log.Info(
		fmt.Sprintf("Event processed: id=%s type=%s duration=%dms",
			event.ID, event.Type, duration.Milliseconds()),
	)
}

// LogFailed logs when an event processing fails.
func (l *EventLogger) LogFailed(event types.Event, err error, willRetry bool) {
	l.log.Error(
		fmt.Sprintf("Event processing failed: id=%s type=%s error=%s retry_count=%d will_retry=%v",
			event.ID, event.Type, err.Error(), event.RetryCount, willRetry),
	)
}

// LogRetry logs when an event is being retried.
func (l *EventLogger) LogRetry(event types.Event) {
	l.log.Warn(
		fmt.Sprintf("Retrying event: id=%s type=%s retry_count=%d max_retries=%d",
			event.ID, event.Type, event.RetryCount, event.MaxRetries),
	)
}

// WrapHandler wraps an event handler with logging.
func (l *EventLogger) WrapHandler(handler eventbus.EventHandler) eventbus.EventHandler {
	return func(ctx context.Context, event types.Event) error {
		l.LogReceive(event)

		start := time.Now()
		err := handler(ctx, event)
		duration := time.Since(start)

		if err != nil {
			l.LogFailed(event, err, event.CanRetry())
			return err
		}

		l.LogProcessed(event, duration)
		return nil
	}
}

// LoggingMiddleware creates a middleware that logs all events.
func LoggingMiddleware(log Logger) func(eventbus.EventHandler) eventbus.EventHandler {
	eventLogger := NewEventLogger(log)
	return eventLogger.WrapHandler
}

// MetricsCollector collects metrics about event processing.
type MetricsCollector struct {
	publishedCount  int64
	processedCount  int64
	failedCount     int64
	totalDurationMs int64
}

// NewMetricsCollector creates a new MetricsCollector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

// RecordPublished increments the published count.
func (m *MetricsCollector) RecordPublished() {
	m.publishedCount++
}

// RecordProcessed records a successful processing.
func (m *MetricsCollector) RecordProcessed(durationMs int64) {
	m.processedCount++
	m.totalDurationMs += durationMs
}

// RecordFailed increments the failed count.
func (m *MetricsCollector) RecordFailed() {
	m.failedCount++
}

// GetMetrics returns the current metrics.
func (m *MetricsCollector) GetMetrics() map[string]int64 {
	avgDuration := int64(0)
	if m.processedCount > 0 {
		avgDuration = m.totalDurationMs / m.processedCount
	}

	return map[string]int64{
		"published":           m.publishedCount,
		"processed":           m.processedCount,
		"failed":              m.failedCount,
		"avg_duration_ms":     avgDuration,
		"total_duration_ms":   m.totalDurationMs,
	}
}

// WrapHandlerWithMetrics wraps a handler to collect metrics.
func (m *MetricsCollector) WrapHandler(handler eventbus.EventHandler) eventbus.EventHandler {
	return func(ctx context.Context, event types.Event) error {
		start := time.Now()
		err := handler(ctx, event)
		duration := time.Since(start).Milliseconds()

		if err != nil {
			m.RecordFailed()
			return err
		}

		m.RecordProcessed(duration)
		return nil
	}
}
