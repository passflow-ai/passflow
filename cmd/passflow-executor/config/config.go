package config

import "os"

// Config holds all configuration the executor pod needs.
// Values come from environment variables injected by Kubernetes.
type Config struct {
	// Redis Stream
	RedisURL   string // REDIS_URL
	StreamKey  string // CHRONOS_STREAM_KEY  (default: passflow.tasks.queue)
	ConsumerGroup string // CHRONOS_CONSUMER_GROUP (default: passflow-agent-executor)
	ConsumerID    string // CHRONOS_CONSUMER_ID    (default: hostname)

	// passflow-api base URL + service token
	APIURL       string // CHRONOS_API_URL
	ServiceToken string // CHRONOS_SERVICE_TOKEN

	// LLM provider base URLs (override for custom/self-hosted)
	OpenAIBaseURL    string // OPENAI_BASE_URL    (default: https://api.openai.com/v1)
	OpenAIAPIKey     string // OPENAI_API_KEY
	AnthropicAPIKey  string // ANTHROPIC_API_KEY
	OllamaBaseURL    string // OLLAMA_BASE_URL    (default: http://localhost:11434)
	LLMGatewayURL    string // LLM_GATEWAY_URL    (default: http://passflow-llm-gateway)

	// Worker pool
	Workers int // CHRONOS_WORKERS (default: 4)

	// LLM Fallback configuration
	EnableLLMFallback bool // CHRONOS_LLM_FALLBACK (default: false)
	LLMMaxRetries     int  // CHRONOS_LLM_MAX_RETRIES (default: 3)
}

func Load() *Config {
	hostname, _ := os.Hostname()
	return &Config{
		RedisURL:         getEnv("REDIS_URL", "localhost:6379"),
		StreamKey:        getEnv("CHRONOS_STREAM_KEY", "passflow.tasks.queue"),
		ConsumerGroup:    getEnv("CHRONOS_CONSUMER_GROUP", "passflow-agent-executor"),
		ConsumerID:       getEnv("CHRONOS_CONSUMER_ID", hostname),
		APIURL:           getEnv("CHRONOS_API_URL", "http://passflow-api:8080"),
		ServiceToken:     getEnv("CHRONOS_SERVICE_TOKEN", ""),
		OpenAIBaseURL:    getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIAPIKey:     getEnv("OPENAI_API_KEY", ""),
		AnthropicAPIKey:  getEnv("ANTHROPIC_API_KEY", ""),
		OllamaBaseURL:    getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
		LLMGatewayURL:    getEnv("LLM_GATEWAY_URL", "http://passflow-llm-gateway"),
		Workers:           getEnvInt("CHRONOS_WORKERS", 4),
		EnableLLMFallback: getEnvBool("CHRONOS_LLM_FALLBACK", false),
		LLMMaxRetries:     getEnvInt("CHRONOS_LLM_MAX_RETRIES", 3),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n := 0
	for _, c := range v {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func getEnvBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v == "true" || v == "1" || v == "yes"
}
