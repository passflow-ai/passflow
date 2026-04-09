package secrets

import (
	"context"
	"fmt"
	"os"
)

// CredentialSource indicates where credentials came from.
type CredentialSource string

const (
	SourceWorkspace CredentialSource = "workspace"
	SourcePlatform  CredentialSource = "platform"
	SourceVault     CredentialSource = "vault"
	SourceEnv       CredentialSource = "env"
)

// Credentials holds provider credentials.
type Credentials struct {
	APIKey     string
	BaseURL    string
	AccessKey  string // AWS
	SecretKey  string // AWS
	Region     string // AWS/Azure
	Endpoint   string // Azure
	Deployment string // Azure
	ProjectID  string // GCP
	Source     CredentialSource
}

// IsFromWorkspace returns true if credentials came from workspace config.
func (c *Credentials) IsFromWorkspace() bool {
	return c.Source == SourceWorkspace
}

// Validate checks if credentials are sufficient.
func (c *Credentials) Validate() error {
	if c.APIKey == "" && c.AccessKey == "" {
		return fmt.Errorf("no credentials provided")
	}
	return nil
}

// WorkspaceCredentialSource reads provider credentials for a workspace.
type WorkspaceCredentialSource interface {
	GetCredentials(ctx context.Context, workspaceID, provider string) (map[string]string, error)
}

// PlatformCredentialSource reads provider credentials at the enterprise/organization level.
type PlatformCredentialSource interface {
	GetCredentials(ctx context.Context, orgID, provider string) (map[string]string, error)
}

// Resolver resolves credentials from workspace config, platform config, OCI Vault, or environment variables.
type Resolver struct {
	vault     *OCIVaultClient
	workspace WorkspaceCredentialSource
	platform  PlatformCredentialSource
}

// NewResolver creates a new secrets resolver.
// An optional WorkspaceCredentialSource can be provided for workspace-level credentials.
func NewResolver(vault *OCIVaultClient, workspace ...WorkspaceCredentialSource) *Resolver {
	r := &Resolver{vault: vault}
	if len(workspace) > 0 {
		r.workspace = workspace[0]
	}
	return r
}

// SetPlatformSource sets the platform credential source for enterprise-level credentials.
func (r *Resolver) SetPlatformSource(platform PlatformCredentialSource) {
	r.platform = platform
}

// Resolve gets credentials for a provider.
// Priority: Workspace credentials > Platform/enterprise credentials > OCI Vault > Environment Variables
// An optional workspaceID can be provided to look up workspace-level credentials first.
// An optional orgID can be provided (after workspaceID) to look up platform-level credentials.
func (r *Resolver) Resolve(ctx context.Context, provider string, workspaceID ...string) (*Credentials, error) {
	wsID := ""
	orgID := ""
	if len(workspaceID) > 0 {
		wsID = workspaceID[0]
	}
	if len(workspaceID) > 1 {
		orgID = workspaceID[1]
	}

	// Try workspace credentials first
	if r.workspace != nil && wsID != "" {
		creds, err := r.resolveFromWorkspace(ctx, wsID, provider)
		if err == nil {
			return creds, nil
		}
		// Fall through to platform/vault/env on workspace error
	}

	// Try platform/enterprise credentials
	if r.platform != nil && orgID != "" {
		creds, err := r.resolveFromPlatform(ctx, orgID, provider)
		if err == nil {
			return creds, nil
		}
		// Fall through to vault/env on platform error
	}

	// Try OCI Vault if configured
	if r.vault != nil && r.vault.IsConfigured() {
		creds, err := r.resolveFromVault(ctx, provider)
		if err == nil {
			return creds, nil
		}
		// Fall through to env vars on vault error
	}

	return r.resolveFromEnv(provider)
}

func (r *Resolver) resolveFromWorkspace(ctx context.Context, workspaceID, provider string) (*Credentials, error) {
	m, err := r.workspace.GetCredentials(ctx, workspaceID, provider)
	if err != nil || len(m) == 0 {
		return nil, fmt.Errorf("no workspace credentials for %s", provider)
	}

	creds := &Credentials{
		APIKey:    m["apiKey"],
		BaseURL:   m["baseUrl"],
		AccessKey: m["accessKey"],
		SecretKey: m["secretKey"],
		Region:    m["region"],
		Endpoint:  m["endpoint"],
		Source:    SourceWorkspace,
	}
	if err := creds.Validate(); err != nil {
		return nil, err
	}
	return creds, nil
}

func (r *Resolver) resolveFromPlatform(ctx context.Context, orgID, provider string) (*Credentials, error) {
	m, err := r.platform.GetCredentials(ctx, orgID, provider)
	if err != nil || len(m) == 0 {
		return nil, fmt.Errorf("no platform credentials for %s", provider)
	}

	creds := &Credentials{
		APIKey:    m["apiKey"],
		BaseURL:   m["baseUrl"],
		AccessKey: m["accessKey"],
		SecretKey: m["secretKey"],
		Region:    m["region"],
		Endpoint:  m["endpoint"],
		Source:    SourcePlatform,
	}
	if err := creds.Validate(); err != nil {
		return nil, err
	}
	return creds, nil
}

func (r *Resolver) resolveFromVault(ctx context.Context, provider string) (*Credentials, error) {
	switch provider {
	case "openai":
		apiKey, err := r.vault.GetSecret(ctx, "passflow/llm/openai/api-key")
		if err != nil {
			return nil, err
		}
		return &Credentials{APIKey: apiKey, Source: SourceVault}, nil

	case "anthropic":
		apiKey, err := r.vault.GetSecret(ctx, "passflow/llm/anthropic/api-key")
		if err != nil {
			return nil, err
		}
		return &Credentials{APIKey: apiKey, Source: SourceVault}, nil

	case "gemini":
		apiKey, err := r.vault.GetSecret(ctx, "passflow/llm/gemini/api-key")
		if err != nil {
			return nil, err
		}
		return &Credentials{APIKey: apiKey, Source: SourceVault}, nil

	case "bedrock":
		accessKey, err := r.vault.GetSecret(ctx, "passflow/llm/bedrock/access-key")
		if err != nil {
			return nil, err
		}
		secretKey, err := r.vault.GetSecret(ctx, "passflow/llm/bedrock/secret-key")
		if err != nil {
			return nil, err
		}
		return &Credentials{AccessKey: accessKey, SecretKey: secretKey, Source: SourceVault}, nil

	case "azure":
		apiKey, err := r.vault.GetSecret(ctx, "passflow/llm/azure/api-key")
		if err != nil {
			return nil, err
		}
		endpoint, _ := r.vault.GetSecret(ctx, "passflow/llm/azure/endpoint")
		return &Credentials{APIKey: apiKey, Endpoint: endpoint, Source: SourceVault}, nil

	case "ollama":
		return &Credentials{APIKey: "ollama", Source: SourceVault}, nil

	case "qwen":
		apiKey, err := r.vault.GetSecret(ctx, "passflow/llm/qwen/api-key")
		if err != nil {
			return nil, err
		}
		return &Credentials{APIKey: apiKey, Source: SourceVault}, nil

	case "gateway":
		// LLM Gateway uses service token for auth, URL from vault or default
		baseURL, _ := r.vault.GetSecret(ctx, "passflow/llm/gateway/base-url")
		if baseURL == "" {
			baseURL = "http://passflow-llm-gateway"
		}
		// Try to get service token from vault, fallback to env
		apiKey, _ := r.vault.GetSecret(ctx, "passflow/llm/gateway/service-token")
		if apiKey == "" {
			apiKey = os.Getenv("SERVICE_TOKEN")
		}
		if apiKey == "" {
			apiKey = "gateway" // Fallback for development
		}
		return &Credentials{APIKey: apiKey, BaseURL: baseURL, Source: SourceVault}, nil

	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

func (r *Resolver) resolveFromEnv(provider string) (*Credentials, error) {
	switch provider {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY not set")
		}
		return &Credentials{
			APIKey:  apiKey,
			BaseURL: os.Getenv("OPENAI_BASE_URL"),
			Source:  SourceEnv,
		}, nil

	case "anthropic":
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY not set")
		}
		return &Credentials{APIKey: apiKey, Source: SourceEnv}, nil

	case "gemini":
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY not set")
		}
		return &Credentials{APIKey: apiKey, Source: SourceEnv}, nil

	case "bedrock":
		accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		if accessKey == "" || secretKey == "" {
			return nil, fmt.Errorf("AWS credentials not set")
		}
		return &Credentials{
			AccessKey: accessKey,
			SecretKey: secretKey,
			Region:    os.Getenv("AWS_REGION"),
			Source:    SourceEnv,
		}, nil

	case "azure":
		apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("AZURE_OPENAI_API_KEY not set")
		}
		return &Credentials{
			APIKey:   apiKey,
			Endpoint: os.Getenv("AZURE_OPENAI_ENDPOINT"),
			Source:   SourceEnv,
		}, nil

	case "ollama":
		return &Credentials{
			APIKey:  "ollama",
			BaseURL: os.Getenv("OLLAMA_BASE_URL"),
			Source:  SourceEnv,
		}, nil

	case "qwen":
		// Qwen routes through LLM gateway - use SERVICE_TOKEN for auth
		baseURL := os.Getenv("LLM_GATEWAY_URL")
		if baseURL == "" {
			baseURL = "http://passflow-llm-gateway"
		}
		apiKey := os.Getenv("SERVICE_TOKEN")
		if apiKey == "" {
			apiKey = os.Getenv("CHRONOS_SERVICE_TOKEN")
		}
		if apiKey == "" {
			apiKey = "gateway" // Fallback for development
		}
		return &Credentials{
			APIKey:  apiKey,
			BaseURL: baseURL,
			Source:  SourceEnv,
		}, nil

	case "gateway":
		baseURL := os.Getenv("LLM_GATEWAY_URL")
		if baseURL == "" {
			baseURL = "http://passflow-llm-gateway"
		}
		// Use SERVICE_TOKEN for internal service-to-service auth
		apiKey := os.Getenv("SERVICE_TOKEN")
		if apiKey == "" {
			apiKey = "gateway" // Fallback for development
		}
		return &Credentials{
			APIKey:  apiKey,
			BaseURL: baseURL,
			Source:  SourceEnv,
		}, nil

	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}
