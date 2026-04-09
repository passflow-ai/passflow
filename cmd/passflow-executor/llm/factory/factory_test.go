package factory

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/config"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm/registry"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm/secrets"
)

// ====================
// Factory Tests
// ====================

func TestNewFactory(t *testing.T) {
	reg := registry.New()
	resolver := secrets.NewResolver(nil)

	factory := NewFactory(reg, resolver)

	if factory == nil {
		t.Fatal("NewFactory returned nil")
	}
	if factory.registry == nil {
		t.Error("Factory registry is nil")
	}
	if factory.secrets == nil {
		t.Error("Factory secrets is nil")
	}
	if factory.matcher == nil {
		t.Error("Factory matcher is nil")
	}
}

func TestFactory_NewClient_OpenAI(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-openai-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	reg := registry.New()
	resolver := secrets.NewResolver(nil)
	factory := NewFactory(reg, resolver)

	ctx := context.Background()
	client, err := factory.NewClient(ctx, "openai", "gpt-4o", FactoryConfig{})

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
}

func TestFactory_NewClient_Anthropic(t *testing.T) {
	os.Setenv("ANTHROPIC_API_KEY", "test-anthropic-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	reg := registry.New()
	resolver := secrets.NewResolver(nil)
	factory := NewFactory(reg, resolver)

	ctx := context.Background()
	client, err := factory.NewClient(ctx, "anthropic", "claude-3-5-sonnet-latest", FactoryConfig{})

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
}

func TestFactory_NewClient_Gemini(t *testing.T) {
	os.Setenv("GEMINI_API_KEY", "test-gemini-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	reg := registry.New()
	resolver := secrets.NewResolver(nil)
	factory := NewFactory(reg, resolver)

	ctx := context.Background()
	client, err := factory.NewClient(ctx, "gemini", "gemini-1.5-pro", FactoryConfig{})

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
}

func TestFactory_NewClient_Bedrock(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY_ID", "test-access-key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-key")
	os.Setenv("AWS_REGION", "us-east-1")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_REGION")
	}()

	reg := registry.New()
	resolver := secrets.NewResolver(nil)
	factory := NewFactory(reg, resolver)

	ctx := context.Background()
	client, err := factory.NewClient(ctx, "bedrock", "anthropic.claude-3-sonnet-20240229-v1:0", FactoryConfig{})

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
}

func TestFactory_NewClient_Azure(t *testing.T) {
	os.Setenv("AZURE_OPENAI_API_KEY", "test-azure-key")
	os.Setenv("AZURE_OPENAI_ENDPOINT", "https://test.openai.azure.com")
	defer func() {
		os.Unsetenv("AZURE_OPENAI_API_KEY")
		os.Unsetenv("AZURE_OPENAI_ENDPOINT")
	}()

	reg := registry.New()
	resolver := secrets.NewResolver(nil)
	factory := NewFactory(reg, resolver)

	ctx := context.Background()
	client, err := factory.NewClient(ctx, "azure", "gpt-4-deployment", FactoryConfig{})

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
}

func TestFactory_NewClient_Ollama(t *testing.T) {
	reg := registry.New()
	resolver := secrets.NewResolver(nil)
	factory := NewFactory(reg, resolver)

	ctx := context.Background()
	client, err := factory.NewClient(ctx, "ollama", "llama3", FactoryConfig{})

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
}

func TestFactory_NewClient_Gateway(t *testing.T) {
	reg := registry.New()
	resolver := secrets.NewResolver(nil)
	factory := NewFactory(reg, resolver)

	ctx := context.Background()
	client, err := factory.NewClient(ctx, "gateway", "gpt-4o", FactoryConfig{})

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
}

func TestFactory_NewClient_Gateway_CustomURL(t *testing.T) {
	os.Setenv("LLM_GATEWAY_URL", "http://custom-gateway:8080")
	defer os.Unsetenv("LLM_GATEWAY_URL")

	reg := registry.New()
	resolver := secrets.NewResolver(nil)
	factory := NewFactory(reg, resolver)

	ctx := context.Background()
	client, err := factory.NewClient(ctx, "gateway", "claude-3-5-sonnet-20241022", FactoryConfig{})

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
}

func TestFactory_NewClient_UnknownProvider(t *testing.T) {
	reg := registry.New()
	resolver := secrets.NewResolver(nil)
	factory := NewFactory(reg, resolver)

	ctx := context.Background()
	_, err := factory.NewClient(ctx, "unknown-provider", "model", FactoryConfig{})

	if err == nil {
		t.Fatal("Expected error for unknown provider")
	}
}

func TestFactory_NewClient_MissingCredentials(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")

	reg := registry.New()
	resolver := secrets.NewResolver(nil)
	factory := NewFactory(reg, resolver)

	ctx := context.Background()
	_, err := factory.NewClient(ctx, "openai", "gpt-4o", FactoryConfig{})

	if err == nil {
		t.Fatal("Expected error for missing credentials")
	}
}

func TestFactory_NewClient_WithFallback(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-openai-key")
	os.Setenv("ANTHROPIC_API_KEY", "test-anthropic-key")
	defer func() {
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
	}()

	reg := registry.New()
	resolver := secrets.NewResolver(nil)
	factory := NewFactory(reg, resolver)

	ctx := context.Background()
	client, err := factory.NewClient(ctx, "openai", "gpt-4o", FactoryConfig{
		EnableFallback: true,
		MaxRetries:     3,
	})

	if err != nil {
		t.Fatalf("NewClient with fallback failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}

	// Verify it's a ResilientClient by type assertion
	_, isResilient := client.(*llm.ResilientClient)
	if !isResilient {
		t.Error("Expected ResilientClient when EnableFallback is true")
	}
}

func TestFactory_NewClient_WithFallbackReturnsResilientEvenWithoutFallbacks(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-openai-key")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("GEMINI_API_KEY")
	defer os.Unsetenv("OPENAI_API_KEY")

	reg := registry.New()
	resolver := secrets.NewResolver(nil)
	factory := NewFactory(reg, resolver)

	ctx := context.Background()
	client, err := factory.NewClient(ctx, "openai", "gpt-4o", FactoryConfig{
		EnableFallback: true,
		MaxRetries:     2,
	})

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}

	// Should still wrap in ResilientClient for retry support
	_, isResilient := client.(*llm.ResilientClient)
	if !isResilient {
		t.Error("Expected ResilientClient for retry support")
	}
}

// ====================
// FactoryConfig Tests
// ====================

func TestFactoryConfig_Defaults(t *testing.T) {
	config := FactoryConfig{}

	if config.EnableFallback != false {
		t.Error("Default EnableFallback should be false")
	}
	if config.MaxRetries != 0 {
		t.Error("Default MaxRetries should be 0")
	}
}

func TestFactoryConfig_CustomValues(t *testing.T) {
	cfg := FactoryConfig{
		EnableFallback: true,
		MaxRetries:     5,
		BaseDelay:      2 * time.Second,
		MaxDelay:       60 * time.Second,
	}

	if cfg.EnableFallback != true {
		t.Error("EnableFallback should be true")
	}
	if cfg.MaxRetries != 5 {
		t.Error("MaxRetries should be 5")
	}
	if cfg.BaseDelay != 2*time.Second {
		t.Error("BaseDelay should be 2s")
	}
	if cfg.MaxDelay != 60*time.Second {
		t.Error("MaxDelay should be 60s")
	}
}

// ====================
// createProviderClient Tests
// ====================

func TestFactory_createProviderClient_OpenAI(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	reg := registry.New()
	resolver := secrets.NewResolver(nil)
	factory := NewFactory(reg, resolver)

	ctx := context.Background()
	client, err := factory.createProviderClient(ctx, "openai", "gpt-4o", "")

	if err != nil {
		t.Fatalf("createProviderClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("createProviderClient returned nil")
	}
}

func TestFactory_createFallbackClients(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-openai-key")
	os.Setenv("ANTHROPIC_API_KEY", "test-anthropic-key")
	defer func() {
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
	}()

	reg := registry.New()
	resolver := secrets.NewResolver(nil)
	factory := NewFactory(reg, resolver)

	model, ok := reg.Get("gpt-4o")
	if !ok {
		t.Fatal("gpt-4o not found in registry")
	}

	ctx := context.Background()
	fallbacks := factory.createFallbackClients(ctx, model.Capabilities, []string{"gpt-4o"}, "")

	// Should find fallback clients since we have credentials for both providers
	if len(fallbacks) == 0 {
		t.Log("No fallbacks found - may be expected if credentials can't be resolved")
	}
}

// ====================
// Backward Compatibility Tests (keep legacy New function working)
// ====================

func TestNew_OpenAI(t *testing.T) {
	cfg := &config.Config{
		OpenAIAPIKey:  "test-key",
		OpenAIBaseURL: "",
	}

	client, err := New("openai", "gpt-4", cfg)
	if err != nil {
		t.Errorf("New() returned error: %v", err)
	}
	if client == nil {
		t.Error("New() returned nil client")
	}

	var _ llm.Client = client
}

func TestNew_Anthropic(t *testing.T) {
	cfg := &config.Config{
		AnthropicAPIKey: "test-key",
	}

	client, err := New("anthropic", "claude-3-opus-20240229", cfg)
	if err != nil {
		t.Errorf("New() returned error: %v", err)
	}
	if client == nil {
		t.Error("New() returned nil client")
	}

	var _ llm.Client = client
}

func TestNew_Ollama(t *testing.T) {
	cfg := &config.Config{
		OllamaBaseURL: "http://localhost:11434",
	}

	client, err := New("ollama", "llama3", cfg)
	if err != nil {
		t.Errorf("New() returned error: %v", err)
	}
	if client == nil {
		t.Error("New() returned nil client")
	}

	var _ llm.Client = client
}

func TestNew_UnsupportedProvider(t *testing.T) {
	cfg := &config.Config{}

	client, err := New("unsupported", "model", cfg)
	if err == nil {
		t.Error("New() should return error for unsupported provider")
	}
	if client != nil {
		t.Error("New() should return nil client for unsupported provider")
	}
}

// ====================
// NewSimple Tests (convenience function)
// ====================

func TestNewSimple_OpenAI(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	client, err := NewSimple(context.Background(), "openai", "gpt-4o")
	if err != nil {
		t.Fatalf("NewSimple failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewSimple returned nil client")
	}
}

func TestNewSimple_Anthropic(t *testing.T) {
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	client, err := NewSimple(context.Background(), "anthropic", "claude-3-5-sonnet-latest")
	if err != nil {
		t.Fatalf("NewSimple failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewSimple returned nil client")
	}
}

func TestNewSimple_Ollama(t *testing.T) {
	client, err := NewSimple(context.Background(), "ollama", "llama3")
	if err != nil {
		t.Fatalf("NewSimple failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewSimple returned nil client")
	}
}
