package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/passflow-ai/passflow/pkg/common/eventbus/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLogger implements a mock logger for testing.
type MockLogger struct {
	infoCalls  []string
	debugCalls []string
	warnCalls  []string
	errorCalls []string
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		infoCalls:  []string{},
		debugCalls: []string{},
		warnCalls:  []string{},
		errorCalls: []string{},
	}
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.infoCalls = append(m.infoCalls, msg)
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.debugCalls = append(m.debugCalls, msg)
}

func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.warnCalls = append(m.warnCalls, msg)
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.errorCalls = append(m.errorCalls, msg)
}

func TestNewEventLogger(t *testing.T) {
	mockLog := NewMockLogger()
	logger := NewEventLogger(mockLog)

	assert.NotNil(t, logger)
	assert.Equal(t, mockLog, logger.log)
}

func TestLogPublish(t *testing.T) {
	mockLog := NewMockLogger()
	logger := NewEventLogger(mockLog)

	event := types.Event{
		ID:          "event-1",
		Type:        types.EventAgentStarted,
		SourceAgent: "agent-1",
		TargetAgent: "agent-2",
		WorkspaceID: "workspace-1",
		Priority:    types.PriorityNormal,
	}

	logger.LogPublish(event)

	assert.Len(t, mockLog.infoCalls, 1)
}

func TestLogReceive(t *testing.T) {
	mockLog := NewMockLogger()
	logger := NewEventLogger(mockLog)

	event := types.Event{
		ID:          "event-1",
		Type:        types.EventAgentStarted,
		SourceAgent: "agent-1",
	}

	logger.LogReceive(event)

	assert.Len(t, mockLog.debugCalls, 1)
}

func TestLogProcessed(t *testing.T) {
	mockLog := NewMockLogger()
	logger := NewEventLogger(mockLog)

	event := types.Event{
		ID:   "event-1",
		Type: types.EventAgentStarted,
	}

	logger.LogProcessed(event, 100*time.Millisecond)

	assert.Len(t, mockLog.infoCalls, 1)
}

func TestLogFailed(t *testing.T) {
	mockLog := NewMockLogger()
	logger := NewEventLogger(mockLog)

	event := types.Event{
		ID:         "event-1",
		Type:       types.EventAgentStarted,
		RetryCount: 1,
	}

	logger.LogFailed(event, errors.New("test error"), true)

	assert.Len(t, mockLog.errorCalls, 1)
}

func TestLogRetry(t *testing.T) {
	mockLog := NewMockLogger()
	logger := NewEventLogger(mockLog)

	event := types.Event{
		ID:         "event-1",
		Type:       types.EventAgentStarted,
		RetryCount: 1,
		MaxRetries: 3,
	}

	logger.LogRetry(event)

	assert.Len(t, mockLog.warnCalls, 1)
}

func TestWrapHandler(t *testing.T) {
	mockLog := NewMockLogger()
	logger := NewEventLogger(mockLog)

	event := types.Event{
		ID:          "event-1",
		Type:        types.EventAgentStarted,
		SourceAgent: "agent-1",
	}

	handler := func(ctx context.Context, e types.Event) error {
		return nil
	}

	wrappedHandler := logger.WrapHandler(handler)

	err := wrappedHandler(context.Background(), event)

	assert.NoError(t, err)
	assert.Len(t, mockLog.debugCalls, 1)
	assert.Len(t, mockLog.infoCalls, 1)
}

func TestWrapHandlerWithError(t *testing.T) {
	mockLog := NewMockLogger()
	logger := NewEventLogger(mockLog)

	event := types.Event{
		ID:          "event-1",
		Type:        types.EventAgentStarted,
		SourceAgent: "agent-1",
	}

	testErr := errors.New("handler error")
	handler := func(ctx context.Context, e types.Event) error {
		return testErr
	}

	wrappedHandler := logger.WrapHandler(handler)

	err := wrappedHandler(context.Background(), event)

	assert.Equal(t, testErr, err)
	assert.Len(t, mockLog.debugCalls, 1)
	assert.Len(t, mockLog.errorCalls, 1)
}

func TestLoggingMiddleware(t *testing.T) {
	mockLog := NewMockLogger()
	middleware := LoggingMiddleware(mockLog)

	assert.NotNil(t, middleware)
}

func TestMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()

	assert.NotNil(t, collector)

	collector.RecordPublished()
	collector.RecordPublished()

	metrics := collector.GetMetrics()

	assert.Equal(t, int64(2), metrics["published"])
	assert.Equal(t, int64(0), metrics["processed"])
	assert.Equal(t, int64(0), metrics["failed"])
}

func TestMetricsCollectorProcessed(t *testing.T) {
	collector := NewMetricsCollector()

	collector.RecordPublished()
	collector.RecordProcessed(100)
	collector.RecordProcessed(200)

	metrics := collector.GetMetrics()

	assert.Equal(t, int64(1), metrics["published"])
	assert.Equal(t, int64(2), metrics["processed"])
	assert.Equal(t, int64(150), metrics["avg_duration_ms"])
	assert.Equal(t, int64(300), metrics["total_duration_ms"])
}

func TestMetricsCollectorFailed(t *testing.T) {
	collector := NewMetricsCollector()

	collector.RecordPublished()
	collector.RecordFailed()
	collector.RecordFailed()

	metrics := collector.GetMetrics()

	assert.Equal(t, int64(1), metrics["published"])
	assert.Equal(t, int64(2), metrics["failed"])
}

func TestWrapHandlerWithMetrics(t *testing.T) {
	collector := NewMetricsCollector()

	event := types.Event{
		ID:   "event-1",
		Type: types.EventAgentStarted,
	}

	handler := func(ctx context.Context, e types.Event) error {
		return nil
	}

	wrappedHandler := collector.WrapHandler(handler)

	err := wrappedHandler(context.Background(), event)

	require.NoError(t, err)

	metrics := collector.GetMetrics()
	assert.Equal(t, int64(1), metrics["processed"])
}
