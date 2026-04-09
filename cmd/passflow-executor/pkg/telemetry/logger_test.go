package telemetry

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger("test-service")
	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}
}

func TestLoggerWithTraceContext_NoSpan(t *testing.T) {
	logger := NewLogger("test-service")
	ctx := context.Background()

	ctxLogger := LoggerWithTraceContext(ctx, logger)
	if ctxLogger == nil {
		t.Fatal("LoggerWithTraceContext returned nil")
	}
}

func TestLoggerWithTraceContext_WithSpan(t *testing.T) {
	logger := NewLogger("test-service")

	traceID, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := trace.SpanIDFromHex("0102030405060708")
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)

	ctxLogger := LoggerWithTraceContext(ctx, logger)
	if ctxLogger == nil {
		t.Fatal("LoggerWithTraceContext returned nil")
	}
}
