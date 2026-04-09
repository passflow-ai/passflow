package telemetry

import (
	"os"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	cfg := LoadConfig()

	if cfg.ServiceName == "" {
		t.Error("ServiceName should have a default value")
	}
	if cfg.OTLPEndpoint == "" {
		t.Error("OTLPEndpoint should have a default value")
	}
}

func TestLoadConfig_FromEnv(t *testing.T) {
	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "collector:4317")
	os.Setenv("OTEL_TRACES_ENABLED", "false")
	defer func() {
		os.Unsetenv("OTEL_SERVICE_NAME")
		os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
		os.Unsetenv("OTEL_TRACES_ENABLED")
	}()

	cfg := LoadConfig()

	if cfg.ServiceName != "test-service" {
		t.Errorf("ServiceName = %s, want test-service", cfg.ServiceName)
	}
	if cfg.OTLPEndpoint != "collector:4317" {
		t.Errorf("OTLPEndpoint = %s, want collector:4317", cfg.OTLPEndpoint)
	}
	if cfg.TracesEnabled {
		t.Error("TracesEnabled should be false")
	}
}
