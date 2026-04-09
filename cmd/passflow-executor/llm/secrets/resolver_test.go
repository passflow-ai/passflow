package secrets

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestResolver_Resolve_EnvVars(t *testing.T) {
	// Set up test env vars
	os.Setenv("OPENAI_API_KEY", "test-openai-key")
	os.Setenv("ANTHROPIC_API_KEY", "test-anthropic-key")
	os.Setenv("GEMINI_API_KEY", "test-gemini-key")
	defer func() {
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
		os.Unsetenv("GEMINI_API_KEY")
	}()

	// No OCI vault configured
	r := NewResolver(nil)

	tests := []struct {
		provider string
		wantKey  string
	}{
		{"openai", "test-openai-key"},
		{"anthropic", "test-anthropic-key"},
		{"gemini", "test-gemini-key"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			creds, err := r.Resolve(context.Background(), tt.provider)
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}
			if creds.APIKey != tt.wantKey {
				t.Errorf("APIKey = %q, want %q", creds.APIKey, tt.wantKey)
			}
		})
	}
}

func TestResolver_Resolve_UnknownProvider(t *testing.T) {
	r := NewResolver(nil)
	_, err := r.Resolve(context.Background(), "unknown")
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestCredentials_Validate(t *testing.T) {
	tests := []struct {
		name    string
		creds   Credentials
		wantErr bool
	}{
		{"valid with api key", Credentials{APIKey: "key"}, false},
		{"valid with access key", Credentials{AccessKey: "key"}, false},
		{"empty", Credentials{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.creds.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolver_Resolve_Ollama(t *testing.T) {
	r := NewResolver(nil)
	creds, err := r.Resolve(context.Background(), "ollama")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if creds.APIKey != "ollama" {
		t.Errorf("APIKey = %q, want %q", creds.APIKey, "ollama")
	}
}

func TestResolver_Resolve_Bedrock(t *testing.T) {
	// Set up AWS env vars
	os.Setenv("AWS_ACCESS_KEY_ID", "test-access-key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-key")
	os.Setenv("AWS_REGION", "us-west-2")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_REGION")
	}()

	r := NewResolver(nil)
	creds, err := r.Resolve(context.Background(), "bedrock")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if creds.AccessKey != "test-access-key" {
		t.Errorf("AccessKey = %q, want %q", creds.AccessKey, "test-access-key")
	}
	if creds.SecretKey != "test-secret-key" {
		t.Errorf("SecretKey = %q, want %q", creds.SecretKey, "test-secret-key")
	}
	if creds.Region != "us-west-2" {
		t.Errorf("Region = %q, want %q", creds.Region, "us-west-2")
	}
}

func TestResolver_Resolve_Azure(t *testing.T) {
	// Set up Azure env vars
	os.Setenv("AZURE_OPENAI_API_KEY", "test-azure-key")
	os.Setenv("AZURE_OPENAI_ENDPOINT", "https://test.openai.azure.com")
	defer func() {
		os.Unsetenv("AZURE_OPENAI_API_KEY")
		os.Unsetenv("AZURE_OPENAI_ENDPOINT")
	}()

	r := NewResolver(nil)
	creds, err := r.Resolve(context.Background(), "azure")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if creds.APIKey != "test-azure-key" {
		t.Errorf("APIKey = %q, want %q", creds.APIKey, "test-azure-key")
	}
	if creds.Endpoint != "https://test.openai.azure.com" {
		t.Errorf("Endpoint = %q, want %q", creds.Endpoint, "https://test.openai.azure.com")
	}
}

func TestResolver_Resolve_MissingEnvVar(t *testing.T) {
	// Ensure env vars are not set
	os.Unsetenv("OPENAI_API_KEY")

	r := NewResolver(nil)
	_, err := r.Resolve(context.Background(), "openai")
	if err == nil {
		t.Error("expected error for missing env var")
	}
}

func TestResolver_Resolve_OpenAIWithBaseURL(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("OPENAI_BASE_URL", "https://custom.openai.com")
	defer func() {
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("OPENAI_BASE_URL")
	}()

	r := NewResolver(nil)
	creds, err := r.Resolve(context.Background(), "openai")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if creds.BaseURL != "https://custom.openai.com" {
		t.Errorf("BaseURL = %q, want %q", creds.BaseURL, "https://custom.openai.com")
	}
}

func TestResolver_Resolve_OllamaWithBaseURL(t *testing.T) {
	os.Setenv("OLLAMA_BASE_URL", "http://localhost:11434")
	defer os.Unsetenv("OLLAMA_BASE_URL")

	r := NewResolver(nil)
	creds, err := r.Resolve(context.Background(), "ollama")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if creds.BaseURL != "http://localhost:11434" {
		t.Errorf("BaseURL = %q, want %q", creds.BaseURL, "http://localhost:11434")
	}
}

// --- Workspace credential tests ---

type mockWorkspaceSource struct {
	creds map[string]string
	err   error
}

func (m *mockWorkspaceSource) GetCredentials(ctx context.Context, workspaceID, provider string) (map[string]string, error) {
	return m.creds, m.err
}

func TestResolver_Resolve_WorkspaceCredentials(t *testing.T) {
	ws := &mockWorkspaceSource{
		creds: map[string]string{"apiKey": "ws-key-123"},
	}
	resolver := NewResolver(nil, ws)

	creds, err := resolver.Resolve(context.Background(), "anthropic", "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.APIKey != "ws-key-123" {
		t.Errorf("expected ws-key-123, got %s", creds.APIKey)
	}
}

func TestResolver_Resolve_WorkspaceFallsToEnv(t *testing.T) {
	ws := &mockWorkspaceSource{err: fmt.Errorf("not found")}
	resolver := NewResolver(nil, ws)

	t.Setenv("ANTHROPIC_API_KEY", "env-key-456")
	creds, err := resolver.Resolve(context.Background(), "anthropic", "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.APIKey != "env-key-456" {
		t.Errorf("expected env-key-456, got %s", creds.APIKey)
	}
}

func TestResolver_Resolve_WorkspaceBedrockCredentials(t *testing.T) {
	ws := &mockWorkspaceSource{
		creds: map[string]string{
			"accessKey": "AKIA-test",
			"secretKey": "secret-test",
			"region":    "eu-west-1",
		},
	}
	resolver := NewResolver(nil, ws)

	creds, err := resolver.Resolve(context.Background(), "bedrock", "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.AccessKey != "AKIA-test" {
		t.Errorf("expected AKIA-test, got %s", creds.AccessKey)
	}
	if creds.Region != "eu-west-1" {
		t.Errorf("expected eu-west-1, got %s", creds.Region)
	}
}

func TestResolver_Resolve_NoWorkspaceIDFallsToEnv(t *testing.T) {
	ws := &mockWorkspaceSource{
		creds: map[string]string{"apiKey": "ws-key-should-not-use"},
	}
	resolver := NewResolver(nil, ws)

	t.Setenv("ANTHROPIC_API_KEY", "env-key-789")
	// Call without workspaceID - should skip workspace source entirely
	creds, err := resolver.Resolve(context.Background(), "anthropic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.APIKey != "env-key-789" {
		t.Errorf("expected env-key-789, got %s", creds.APIKey)
	}
}

func TestResolver_Resolve_EmptyWorkspaceIDFallsToEnv(t *testing.T) {
	ws := &mockWorkspaceSource{
		creds: map[string]string{"apiKey": "ws-key-should-not-use"},
	}
	resolver := NewResolver(nil, ws)

	t.Setenv("OPENAI_API_KEY", "env-key-empty-ws")
	// Call with empty workspaceID - should skip workspace source
	creds, err := resolver.Resolve(context.Background(), "openai", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.APIKey != "env-key-empty-ws" {
		t.Errorf("expected env-key-empty-ws, got %s", creds.APIKey)
	}
}

func TestResolver_Resolve_WorkspaceEmptyCredsMap(t *testing.T) {
	ws := &mockWorkspaceSource{
		creds: map[string]string{},
	}
	resolver := NewResolver(nil, ws)

	t.Setenv("GEMINI_API_KEY", "env-fallback")
	creds, err := resolver.Resolve(context.Background(), "gemini", "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.APIKey != "env-fallback" {
		t.Errorf("expected env-fallback, got %s", creds.APIKey)
	}
}

// --- Gateway provider tests ---

func TestResolver_Resolve_Gateway_DefaultURL(t *testing.T) {
	// Ensure env var is not set
	os.Unsetenv("LLM_GATEWAY_URL")

	r := NewResolver(nil)
	creds, err := r.Resolve(context.Background(), "gateway")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if creds.APIKey != "gateway" {
		t.Errorf("APIKey = %q, want %q", creds.APIKey, "gateway")
	}
	if creds.BaseURL != "http://passflow-llm-gateway" {
		t.Errorf("BaseURL = %q, want %q", creds.BaseURL, "http://passflow-llm-gateway")
	}
}

func TestResolver_Resolve_Gateway_CustomURL(t *testing.T) {
	os.Setenv("LLM_GATEWAY_URL", "http://custom-gateway:8080")
	defer os.Unsetenv("LLM_GATEWAY_URL")

	r := NewResolver(nil)
	creds, err := r.Resolve(context.Background(), "gateway")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if creds.BaseURL != "http://custom-gateway:8080" {
		t.Errorf("BaseURL = %q, want %q", creds.BaseURL, "http://custom-gateway:8080")
	}
}
