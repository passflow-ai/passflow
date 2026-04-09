package telemetry

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestInjectTraceContext(t *testing.T) {
	traceID, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := trace.SpanIDFromHex("0102030405060708")
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)

	payload := make(map[string]interface{})
	InjectTraceContext(ctx, payload)

	if payload["trace_id"] != "0102030405060708090a0b0c0d0e0f10" {
		t.Errorf("trace_id = %v, want 0102030405060708090a0b0c0d0e0f10", payload["trace_id"])
	}
	if payload["span_id"] != "0102030405060708" {
		t.Errorf("span_id = %v, want 0102030405060708", payload["span_id"])
	}
}

func TestInjectTraceContext_NoSpan(t *testing.T) {
	ctx := context.Background()
	payload := make(map[string]interface{})
	InjectTraceContext(ctx, payload)

	if _, ok := payload["trace_id"]; ok {
		t.Error("trace_id should not be set when no span")
	}
}

func TestExtractTraceContext(t *testing.T) {
	payload := map[string]interface{}{
		"trace_id": "0102030405060708090a0b0c0d0e0f10",
		"span_id":  "0102030405060708",
	}

	ctx := ExtractTraceContext(context.Background(), payload)
	spanCtx := trace.SpanContextFromContext(ctx)

	if spanCtx.TraceID().String() != "0102030405060708090a0b0c0d0e0f10" {
		t.Errorf("TraceID = %s, want 0102030405060708090a0b0c0d0e0f10", spanCtx.TraceID())
	}
	if spanCtx.SpanID().String() != "0102030405060708" {
		t.Errorf("SpanID = %s, want 0102030405060708", spanCtx.SpanID())
	}
	if !spanCtx.IsRemote() {
		t.Error("SpanContext should be marked as remote")
	}
}

func TestExtractTraceContext_MissingFields(t *testing.T) {
	payload := map[string]interface{}{}
	ctx := ExtractTraceContext(context.Background(), payload)
	spanCtx := trace.SpanContextFromContext(ctx)

	if spanCtx.IsValid() {
		t.Error("SpanContext should not be valid when fields are missing")
	}
}
