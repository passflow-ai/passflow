package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

func InjectTraceContext(ctx context.Context, payload map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return
	}

	payload["trace_id"] = span.SpanContext().TraceID().String()
	payload["span_id"] = span.SpanContext().SpanID().String()
}

func ExtractTraceContext(ctx context.Context, payload map[string]interface{}) context.Context {
	traceIDStr, ok1 := payload["trace_id"].(string)
	spanIDStr, ok2 := payload["span_id"].(string)
	if !ok1 || !ok2 {
		return ctx
	}

	traceID, err := trace.TraceIDFromHex(traceIDStr)
	if err != nil {
		return ctx
	}

	spanID, err := trace.SpanIDFromHex(spanIDStr)
	if err != nil {
		return ctx
	}

	remoteCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})

	return trace.ContextWithRemoteSpanContext(ctx, remoteCtx)
}

func ExtractTraceContextFromSpec(ctx context.Context, traceID, spanID string) context.Context {
	if traceID == "" || spanID == "" {
		return ctx
	}

	payload := map[string]interface{}{
		"trace_id": traceID,
		"span_id":  spanID,
	}
	return ExtractTraceContext(ctx, payload)
}
