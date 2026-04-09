package telemetry

import (
	"context"
	"testing"
)

func TestInitMetrics(t *testing.T) {
	m, err := InitMetrics()
	if err != nil {
		t.Fatalf("InitMetrics failed: %v", err)
	}
	if m == nil {
		t.Fatal("InitMetrics returned nil")
	}

	if m.ExecutionDuration == nil {
		t.Error("ExecutionDuration not initialized")
	}
	if m.LLMCallDuration == nil {
		t.Error("LLMCallDuration not initialized")
	}
	if m.ToolCallDuration == nil {
		t.Error("ToolCallDuration not initialized")
	}
	if m.ExecutionsTotal == nil {
		t.Error("ExecutionsTotal not initialized")
	}
	if m.TokensTotal == nil {
		t.Error("TokensTotal not initialized")
	}
}

func TestMetrics_Record(t *testing.T) {
	m, _ := InitMetrics()
	ctx := context.Background()

	m.ExecutionsTotal.Add(ctx, 1)
	m.TokensTotal.Add(ctx, 100)
}
