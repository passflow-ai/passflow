package telemetry

import (
	"context"
	"testing"
)

func TestInitProvider_Disabled(t *testing.T) {
	cfg := &Config{
		ServiceName:    "test",
		TracesEnabled:  false,
		MetricsEnabled: false,
	}

	shutdown, err := InitProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("InitProvider failed: %v", err)
	}
	if shutdown == nil {
		t.Fatal("shutdown function should not be nil")
	}

	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
}

func TestGetTracer(t *testing.T) {
	tracer := GetTracer()
	if tracer == nil {
		t.Fatal("GetTracer returned nil")
	}
}

func TestGetMeter(t *testing.T) {
	meter := GetMeter()
	if meter == nil {
		t.Fatal("GetMeter returned nil")
	}
}
