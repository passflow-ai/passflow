package telemetry

import (
	"os"
	"strconv"
)

type Config struct {
	ServiceName    string
	Environment    string
	OTLPEndpoint   string
	TracesEnabled  bool
	MetricsEnabled bool
	LogsEnabled    bool
	SampleRate     float64
}

func LoadConfig() *Config {
	return &Config{
		ServiceName:    getEnv("OTEL_SERVICE_NAME", "passflow-agent-executor"),
		Environment:    getEnv("OTEL_ENVIRONMENT", "development"),
		OTLPEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		TracesEnabled:  getEnvBool("OTEL_TRACES_ENABLED", true),
		MetricsEnabled: getEnvBool("OTEL_METRICS_ENABLED", true),
		LogsEnabled:    getEnvBool("OTEL_LOGS_ENABLED", true),
		SampleRate:     getEnvFloat("OTEL_TRACES_SAMPLER_ARG", 1.0),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		b, err := strconv.ParseBool(val)
		if err == nil {
			return b
		}
	}
	return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		f, err := strconv.ParseFloat(val, 64)
		if err == nil {
			return f
		}
	}
	return defaultVal
}
