package setup

// Configuration holds all application configuration.
type Configuration struct {
	Server       ServerConfiguration       `cfg:"server"`
	Database     DatabaseConfiguration     `cfg:"database"`
	Redis        RedisConfiguration        `cfg:"redis"`
	JWT          JWTConfiguration          `cfg:"jwt"`
	AI           AIConfiguration           `cfg:"ai"`
	Integrations IntegrationsConfiguration `cfg:"integrations"`
	Auth         AuthConfiguration         `cfg:"auth"`
	Service      ServiceConfiguration      `cfg:"service"`
	Email        EmailConfiguration        `cfg:"email"`
	Vault        VaultConfiguration        `cfg:"vault"`
	Kubernetes   KubernetesConfiguration   `cfg:"kubernetes"`
}

// AuthConfiguration holds OAuth provider settings.
type AuthConfiguration struct {
	Google    OAuthProviderConfiguration `cfg:"google"`
	Microsoft OAuthProviderConfiguration `cfg:"microsoft"`
	FrontendURL string                  `cfg:"frontend_url"`
}

// OAuthProviderConfiguration holds OAuth2 client settings for a provider.
type OAuthProviderConfiguration struct {
	ClientID     string `cfg:"client_id"`
	ClientSecret string `cfg:"client_secret"`
	RedirectURL  string `cfg:"redirect_url"`
}

// IntegrationsConfiguration holds integrations security settings and OAuth credentials.
type IntegrationsConfiguration struct {
	EncryptionKey string                       `cfg:"encryption_key"`
	GitHub        IntegrationOAuthConfiguration `cfg:"github"`
	Linear        IntegrationOAuthConfiguration `cfg:"linear"`
	Slack         IntegrationOAuthConfiguration `cfg:"slack"`
}

// IntegrationOAuthConfiguration holds OAuth2 client credentials for an integration provider.
type IntegrationOAuthConfiguration struct {
	ClientID     string `cfg:"client_id"`
	ClientSecret string `cfg:"client_secret"`
}

// ServiceConfiguration holds internal service-to-service auth settings.
type ServiceConfiguration struct {
	Token string `cfg:"token"`
}

// ServerConfiguration holds HTTP server settings.
type ServerConfiguration struct {
	Addr      string              `cfg:"addr"`
	AppName   string              `cfg:"app_name"`
	BodyLimit int                 `cfg:"body_limit"`
	UseSSL    bool                `cfg:"use_ssl"`
	CertSSL   string              `cfg:"cert_ssl"`
	CertKey   string              `cfg:"cert_key"`
	Logger    LoggerConfiguration `cfg:"logger"`
}

// LoggerConfiguration holds logging settings.
type LoggerConfiguration struct {
	FolderPath string `cfg:"folder_path"`
	WriteOnLog bool   `cfg:"write_on_log"`
}

// DatabaseConfiguration holds MongoDB connection settings.
type DatabaseConfiguration struct {
	URI    string `cfg:"uri"`
	DBName string `cfg:"db_name"`
}

// RedisConfiguration holds Redis connection settings.
type RedisConfiguration struct {
	URL      string `cfg:"url"`
	Password string `cfg:"password"`
	DB       int    `cfg:"db"`
}

// JWTConfiguration holds JWT authentication settings.
type JWTConfiguration struct {
	Signature         string `cfg:"signature"`
	AccessExpiration  int    `cfg:"access_expiration"`
	RefreshExpiration int    `cfg:"refresh_expiration"`
}

// AIConfiguration holds AI provider settings.
type AIConfiguration struct {
	OpenAI    OpenAIConfiguration    `cfg:"openai"`
	Anthropic AnthropicConfiguration `cfg:"anthropic"`
}

// OpenAIConfiguration holds OpenAI API settings.
type OpenAIConfiguration struct {
	APIKey       string `cfg:"api_key"`
	Model        string `cfg:"model"`
	MaxTokens    int    `cfg:"max_tokens"`
	Temperature  float64 `cfg:"temperature"`
	Organization string `cfg:"organization"`
}

// AnthropicConfiguration holds Anthropic API settings.
type AnthropicConfiguration struct {
	APIKey      string  `cfg:"api_key"`
	Model       string  `cfg:"model"`
	MaxTokens   int     `cfg:"max_tokens"`
	Temperature float64 `cfg:"temperature"`
}

// EmailConfiguration holds email provider settings.
type EmailConfiguration struct {
	SendGrid SendGridConfiguration `cfg:"sendgrid"`
	Resend   ResendConfiguration   `cfg:"resend"`
}

// SendGridConfiguration holds SendGrid API settings.
type SendGridConfiguration struct {
	APIKey    string `cfg:"api_key"`
	FromEmail string `cfg:"from_email"`
	FromName  string `cfg:"from_name"`
}

// ResendConfiguration holds Resend API settings.
type ResendConfiguration struct {
	APIKey    string `cfg:"api_key"`
	FromEmail string `cfg:"from_email"`
	FromName  string `cfg:"from_name"`
	BaseURL   string `cfg:"base_url"`
}

// VaultConfiguration holds OCI Vault connection settings.
type VaultConfiguration struct {
	Enabled       bool   `cfg:"enabled"`
	CompartmentID string `cfg:"compartment_id"`
	VaultID       string `cfg:"vault_id"`
	KeyID         string `cfg:"key_id"`
}

// KubernetesConfiguration holds Kubernetes orchestration settings.
type KubernetesConfiguration struct {
	// Enabled controls whether isolated pod execution is available.
	Enabled bool `cfg:"enabled"`
	// DefaultNamespace is the namespace used when workspace has no dedicated namespace.
	DefaultNamespace string `cfg:"default_namespace"`
	// AgentImage is the Docker image used for agent pods.
	AgentImage string `cfg:"agent_image"`
}
