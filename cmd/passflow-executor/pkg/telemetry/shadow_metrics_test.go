package telemetry

import (
	"context"
	"testing"
)

func TestInitShadowMetrics(t *testing.T) {
	m, err := InitShadowMetrics()
	if err != nil {
		t.Fatalf("InitShadowMetrics failed: %v", err)
	}
	if m == nil {
		t.Fatal("InitShadowMetrics returned nil")
	}

	if m.ShadowRunsTotal == nil {
		t.Error("ShadowRunsTotal not initialized")
	}
	if m.ShadowRunsSuccess == nil {
		t.Error("ShadowRunsSuccess not initialized")
	}
	if m.ShadowToolMocksHit == nil {
		t.Error("ShadowToolMocksHit not initialized")
	}
	if m.ShadowToolMocksMiss == nil {
		t.Error("ShadowToolMocksMiss not initialized")
	}
	if m.ShadowDuration == nil {
		t.Error("ShadowDuration not initialized")
	}
	if m.ShadowTokensTotal == nil {
		t.Error("ShadowTokensTotal not initialized")
	}
}

func TestShadowMetrics_Record(t *testing.T) {
	m, err := InitShadowMetrics()
	if err != nil {
		t.Fatalf("InitShadowMetrics failed: %v", err)
	}
	ctx := context.Background()

	// Verify counters can be incremented without panic.
	m.ShadowRunsTotal.Add(ctx, 1)
	m.ShadowRunsSuccess.Add(ctx, 1)
	m.ShadowToolMocksHit.Add(ctx, 5)
	m.ShadowToolMocksMiss.Add(ctx, 2)
	m.ShadowTokensTotal.Add(ctx, 1500)
	m.ShadowDuration.Record(ctx, 2.5)
}
