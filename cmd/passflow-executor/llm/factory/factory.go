package factory

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jaak-ai/passflow-agent-executor/config"
	"github.com/jaak-ai/passflow-agent-executor/llm"
	"github.com/jaak-ai/passflow-agent-executor/llm/anthropic"
	"github.com/jaak-ai/passflow-agent-executor/llm/azure"
	"github.com/jaak-ai/passflow-agent-executor/llm/bedrock"
	"github.com/jaak-ai/passflow-agent-executor/llm/gemini"
	"github.com/jaak-ai/passflow-agent-executor/llm/openai"
	"github.com/jaak-ai/passflow-agent-executor/llm/registry"
	"github.com/jaak-ai/passflow-agent-executor/llm/secrets"
)

// Factory creates LLM clients with intelligent fallback support.
type Factory struct {
	registry *registry.Registry
	matcher  *registry.Matcher
	secrets  *secrets.Resolver
}

// FactoryConfig holds configuration for client creation.
type FactoryConfig struct {
	EnableFallback bool          // Wrap client in ResilientClient with fallbacks
	MaxRetries     int           // Max retries per provider (default: 3)
	BaseDelay      time.Duration // Base delay for exponential backoff (default: 1s)
	MaxDelay       time.Duration // Max delay between retries (default: 30s)
}

// NewFactory creates a new Factory instance.
func NewFactory(reg *registry.Registry, secretsResolver *secrets.Resolver) *Factory {
	return &Factory{
		registry: reg,
		matcher:  registry.NewMatcher(reg),
		secrets:  secretsResolver,
	}
}

// NewClient creates a client for the specified provider and model.
// If EnableFallback is true, it wraps the client in a ResilientClient
// with fallback providers that have similar capabilities.
func (f *Factory) NewClient(
	ctx context.Context,
	provider, modelID string,
	config FactoryConfig,
) (llm.Client, error) {
	return f.NewClientWithWorkspace(ctx, provider, modelID, "", config)
}

// NewClientWithWorkspace creates a client for the specified provider and model,
// using workspace-level credentials if available.
func (f *Factory) NewClientWithWorkspace(
	ctx context.Context,
	provider, modelID, workspaceID string,
	config FactoryConfig,
) (llm.Client, error) {
	// Create the primary client
	primary, err := f.createProviderClient(ctx, provider, modelID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("factory: failed to create primary client: %w", err)
	}

	// If fallback is not enabled, return raw client
	if !config.EnableFallback {
		return primary, nil
	}

	// Build fallback clients
	var fallbacks []llm.Client
	model, found := f.registry.Get(modelID)
	if found {
		fallbacks = f.createFallbackClients(ctx, model.Capabilities, []string{modelID}, workspaceID)
	}

	// Wrap in ResilientClient for retry and fallback support
	resilientConfig := llm.ResilientConfig{
		MaxRetries: config.MaxRetries,
		BaseDelay:  config.BaseDelay,
		MaxDelay:   config.MaxDelay,
	}

	return llm.NewResilientClient(primary, fallbacks, resilientConfig), nil
}

// createProviderClient creates a raw client for a specific provider/model.
func (f *Factory) createProviderClient(
	ctx context.Context,
	provider, modelID, workspaceID string,
) (llm.Client, error) {
	// Pass workspaceID to resolver for workspace-level credentials
	creds, err := f.secrets.Resolve(ctx, provider, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve credentials for %s: %w", provider, err)
	}

	switch provider {
	case "openai":
		return openai.New(creds.APIKey, creds.BaseURL, modelID), nil

	case "anthropic":
		return anthropic.New(creds.APIKey, modelID), nil

	case "gemini":
		return gemini.New(ctx, creds.APIKey, modelID)

	case "bedrock":
		region := creds.Region
		if region == "" {
			region = "us-east-1"
		}
		return bedrock.New(ctx, region, creds.AccessKey, creds.SecretKey, modelID)

	case "azure":
		return azure.New(creds.Endpoint, creds.APIKey, modelID)

	case "ollama":
		baseURL := creds.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		return openai.New("ollama", baseURL+"/v1", modelID), nil

	case "qwen":
		// If credentials are from workspace/platform, call DashScope directly
		if creds.IsFromWorkspace() || creds.Source == secrets.SourcePlatform || creds.Source == secrets.SourceVault {
			// DashScope International OpenAI-compatible endpoint
			baseURL := creds.BaseURL
			if baseURL == "" {
				baseURL = "https://dashscope-intl.aliyuncs.com/compatible-mode/v1"
			}
			return openai.New(creds.APIKey, baseURL, modelID), nil
		}
		// Otherwise, route through LLM gateway (centralized credentials)
		baseURL := creds.BaseURL
		if baseURL == "" {
			baseURL = "http://passflow-llm-gateway"
		}
		return openai.New(creds.APIKey, baseURL+"/v1", modelID), nil

	case "gateway":
		// LLM Gateway is OpenAI-compatible
		baseURL := creds.BaseURL
		if baseURL == "" {
			baseURL = "http://passflow-llm-gateway"
		}
		return openai.New(creds.APIKey, baseURL+"/v1", modelID), nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// createFallbackClients finds and creates fallback clients based on capabilities.
func (f *Factory) createFallbackClients(
	ctx context.Context,
	required registry.Capabilities,
	exclude []string,
	workspaceID string,
) []llm.Client {
	fallbackModels := f.matcher.FindFallbacks(required, exclude)

	var clients []llm.Client
	for _, model := range fallbackModels {
		client, err := f.createProviderClient(ctx, model.Provider, model.ID, workspaceID)
		if err != nil {
			// Skip fallbacks we can't create (missing credentials, etc.)
			continue
		}
		clients = append(clients, client)
	}

	return clients
}

// ====================
// Backward Compatibility
// ====================

// New creates the appropriate LLM client based on the provider string.
// This function maintains backward compatibility with existing code.
// For full provider support, use NewSimple or NewWithFallback instead.
func New(provider, modelID string, cfg *config.Config) (llm.Client, error) {
	switch provider {
	case "openai":
		return openai.New(cfg.OpenAIAPIKey, cfg.OpenAIBaseURL, modelID), nil
	case "anthropic":
		return anthropic.New(cfg.AnthropicAPIKey, modelID), nil
	case "ollama":
		baseURL := cfg.OllamaBaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		return openai.New("ollama", baseURL+"/v1", modelID), nil
	case "qwen":
		// Qwen/DashScope - route through LLM gateway
		baseURL := cfg.LLMGatewayURL
		if baseURL == "" {
			baseURL = "http://passflow-llm-gateway"
		}
		return openai.New(cfg.ServiceToken, baseURL+"/v1", modelID), nil
	case "gateway":
		baseURL := cfg.LLMGatewayURL
		if baseURL == "" {
			baseURL = "http://passflow-llm-gateway"
		}
		return openai.New(cfg.ServiceToken, baseURL+"/v1", modelID), nil
	default:
		return nil, fmt.Errorf("unsupported model provider: %q", provider)
	}
}

// NewSimple creates a client using environment variables for credentials.
// This is a convenience function for simple use cases.
func NewSimple(ctx context.Context, provider, modelID string) (llm.Client, error) {
	reg := registry.New()
	resolver := secrets.NewResolver(nil)
	factory := NewFactory(reg, resolver)
	return factory.NewClient(ctx, provider, modelID, FactoryConfig{})
}

// NewWithFallback creates a client with fallback support using env vars.
func NewWithFallback(
	ctx context.Context,
	provider, modelID string,
	maxRetries int,
) (llm.Client, error) {
	reg := registry.New()
	resolver := secrets.NewResolver(nil)
	factory := NewFactory(reg, resolver)
	return factory.NewClient(ctx, provider, modelID, FactoryConfig{
		EnableFallback: true,
		MaxRetries:     maxRetries,
	})
}

// DefaultFactory creates a factory with default registry and env-based secrets.
func DefaultFactory() *Factory {
	return NewFactory(registry.New(), secrets.NewResolver(nil))
}

// FromEnv creates a provider client directly from environment variables.
// Convenience function for scripts and simple applications.
func FromEnv(ctx context.Context, provider, modelID string) (llm.Client, error) {
	return NewSimple(ctx, provider, modelID)
}

// MustFromEnv creates a provider client from env vars and panics on error.
func MustFromEnv(ctx context.Context, provider, modelID string) llm.Client {
	client, err := FromEnv(ctx, provider, modelID)
	if err != nil {
		panic(fmt.Sprintf("factory: failed to create client: %v", err))
	}
	return client
}

// GetOllamaBaseURL returns the Ollama base URL from environment.
func GetOllamaBaseURL() string {
	if url := os.Getenv("OLLAMA_BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:11434"
}
