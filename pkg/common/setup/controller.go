package setup

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/ini.v1"
)

// ISetupController defines the interface for configuration loading.
type ISetupController interface {
	LoadConfiguration(pathList ...string) *Configuration
}

type setupController struct {
	once sync.Once
}

// NewSetupController creates a new setup controller.
func NewSetupController() ISetupController {
	return &setupController{}
}

// LoadConfiguration loads the configuration from file and environment variables.
func (c *setupController) LoadConfiguration(pathList ...string) *Configuration {
	instance := &Configuration{}

	c.once.Do(func() {
		path := c.resolveConfigPath(pathList)
		c.loadFromFile(instance, path)
		c.applyEnvironmentOverrides(instance)
		c.applyDefaults(instance)
	})

	return instance
}

func (c *setupController) resolveConfigPath(pathList []string) string {
	if len(pathList) > 0 {
		return pathList[0]
	}

	configPaths := []string{
		"/app/config.ini",
		"config.ini",
		"config/config.ini",
	}

	for _, p := range configPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return "config.ini"
}

func (c *setupController) loadFromFile(instance *Configuration, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("Config file not found at %s, using defaults and environment variables\n", path)
		return
	}

	fmt.Printf("Loading configuration from: %s\n", path)

	cfg, err := ini.Load(path)
	if err != nil {
		fmt.Printf("Warning: Failed to load config file: %v\n", err)
		return
	}

	c.loadServerSection(instance, cfg)
	c.loadDatabaseSection(instance, cfg)
	c.loadRedisSection(instance, cfg)
	c.loadJWTSection(instance, cfg)
	c.loadAISection(instance, cfg)
}

func (c *setupController) loadServerSection(instance *Configuration, cfg *ini.File) {
	section := cfg.Section("server")
	instance.Server.Addr = section.Key("addr").MustString(":8080")
	instance.Server.AppName = section.Key("app_name").MustString("Passflow API")
	instance.Server.BodyLimit = section.Key("body_limit").MustInt(10 * 1024 * 1024)
	instance.Server.UseSSL = section.Key("use_ssl").MustBool(false)
	instance.Server.CertSSL = section.Key("cert_ssl").String()
	instance.Server.CertKey = section.Key("cert_key").String()

	loggerSection := cfg.Section("server.logger")
	instance.Server.Logger.FolderPath = loggerSection.Key("folder_path").MustString("./logs")
	instance.Server.Logger.WriteOnLog = loggerSection.Key("write_on_log").MustBool(false)
}

func (c *setupController) loadDatabaseSection(instance *Configuration, cfg *ini.File) {
	section := cfg.Section("database")
	instance.Database.URI = section.Key("uri").String()
	instance.Database.DBName = section.Key("db_name").MustString("passflow")
}

func (c *setupController) loadRedisSection(instance *Configuration, cfg *ini.File) {
	section := cfg.Section("redis")
	instance.Redis.URL = section.Key("url").String()
	instance.Redis.Password = section.Key("password").String()
	instance.Redis.DB = section.Key("db").MustInt(0)
}

func (c *setupController) loadJWTSection(instance *Configuration, cfg *ini.File) {
	section := cfg.Section("jwt")
	instance.JWT.Signature = section.Key("signature").String()
	instance.JWT.AccessExpiration = section.Key("access_expiration").MustInt(3600)
	instance.JWT.RefreshExpiration = section.Key("refresh_expiration").MustInt(604800)
}

func (c *setupController) loadAISection(instance *Configuration, cfg *ini.File) {
	openaiSection := cfg.Section("ai.openai")
	instance.AI.OpenAI.APIKey = openaiSection.Key("api_key").String()
	instance.AI.OpenAI.Model = openaiSection.Key("model").MustString("gpt-4")
	instance.AI.OpenAI.MaxTokens = openaiSection.Key("max_tokens").MustInt(4096)
	instance.AI.OpenAI.Temperature = openaiSection.Key("temperature").MustFloat64(0.7)
	instance.AI.OpenAI.Organization = openaiSection.Key("organization").String()

	anthropicSection := cfg.Section("ai.anthropic")
	instance.AI.Anthropic.APIKey = anthropicSection.Key("api_key").String()
	instance.AI.Anthropic.Model = anthropicSection.Key("model").MustString("claude-3-opus")
	instance.AI.Anthropic.MaxTokens = anthropicSection.Key("max_tokens").MustInt(4096)
	instance.AI.Anthropic.Temperature = anthropicSection.Key("temperature").MustFloat64(0.7)
}

func (c *setupController) applyEnvironmentOverrides(instance *Configuration) {
	if addr := os.Getenv("SERVER_ADDR"); addr != "" {
		instance.Server.Addr = addr
	}

	if dbURI := os.Getenv("DATABASE_URI"); dbURI != "" {
		instance.Database.URI = dbURI
	} else if mongoURI := os.Getenv("MONGODB_URI"); mongoURI != "" {
		instance.Database.URI = mongoURI
	}

	if dbName := os.Getenv("DATABASE_NAME"); dbName != "" {
		instance.Database.DBName = dbName
	}

	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		instance.Redis.URL = redisURL
	}

	if jwtSignature := os.Getenv("JWT_SIGNATURE"); jwtSignature != "" {
		instance.JWT.Signature = jwtSignature
	} else if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		instance.JWT.Signature = jwtSecret
	}

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		instance.AI.OpenAI.APIKey = openaiKey
	}

	if anthropicKey := os.Getenv("ANTHROPIC_API_KEY"); anthropicKey != "" {
		instance.AI.Anthropic.APIKey = anthropicKey
	}

	// INTEGRATIONS_ENCRYPTION_KEY is the primary env var for provider/credential encryption.
	// CHRONOS_ENCRYPTION_KEY is the Chronos-specific alias (takes precedence).
	// CHRONOS_JWT_SECRET / JWT_SECRET are used as fallback so existing deployments
	// automatically get encryption without extra configuration.
	if encryptionKey := os.Getenv("CHRONOS_ENCRYPTION_KEY"); encryptionKey != "" {
		instance.Integrations.EncryptionKey = encryptionKey
	} else if encryptionKey := os.Getenv("INTEGRATIONS_ENCRYPTION_KEY"); encryptionKey != "" {
		instance.Integrations.EncryptionKey = encryptionKey
	} else if jwtSecret := os.Getenv("CHRONOS_JWT_SECRET"); jwtSecret != "" {
		instance.Integrations.EncryptionKey = jwtSecret
	}

	// Auth configuration
	if frontendURL := os.Getenv("AUTH_FRONTEND_URL"); frontendURL != "" {
		instance.Auth.FrontendURL = frontendURL
	}

	// Google OAuth
	if googleClientID := os.Getenv("AUTH_GOOGLE_CLIENT_ID"); googleClientID != "" {
		instance.Auth.Google.ClientID = googleClientID
	} else if googleClientID := os.Getenv("GOOGLE_CLIENT_ID"); googleClientID != "" {
		instance.Auth.Google.ClientID = googleClientID
	}
	if googleClientSecret := os.Getenv("AUTH_GOOGLE_CLIENT_SECRET"); googleClientSecret != "" {
		instance.Auth.Google.ClientSecret = googleClientSecret
	} else if googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET"); googleClientSecret != "" {
		instance.Auth.Google.ClientSecret = googleClientSecret
	}
	if googleRedirectURL := os.Getenv("AUTH_GOOGLE_REDIRECT_URL"); googleRedirectURL != "" {
		instance.Auth.Google.RedirectURL = googleRedirectURL
	}

	// Microsoft OAuth
	if msClientID := os.Getenv("AUTH_MICROSOFT_CLIENT_ID"); msClientID != "" {
		instance.Auth.Microsoft.ClientID = msClientID
	}
	if msClientSecret := os.Getenv("AUTH_MICROSOFT_CLIENT_SECRET"); msClientSecret != "" {
		instance.Auth.Microsoft.ClientSecret = msClientSecret
	}
	if msRedirectURL := os.Getenv("AUTH_MICROSOFT_REDIRECT_URL"); msRedirectURL != "" {
		instance.Auth.Microsoft.RedirectURL = msRedirectURL
	}

	// Resend email
	if resendKey := os.Getenv("RESEND_API_KEY"); resendKey != "" {
		instance.Email.Resend.APIKey = resendKey
	}
	if resendFrom := os.Getenv("RESEND_FROM_EMAIL"); resendFrom != "" {
		instance.Email.Resend.FromEmail = resendFrom
	}
	if resendName := os.Getenv("RESEND_FROM_NAME"); resendName != "" {
		instance.Email.Resend.FromName = resendName
	}
	if resendBase := os.Getenv("RESEND_BASE_URL"); resendBase != "" {
		instance.Email.Resend.BaseURL = resendBase
	}

	// Integration OAuth providers (GitHub, Linear, Slack)
	if githubClientID := os.Getenv("INTEGRATIONS_GITHUB_CLIENT_ID"); githubClientID != "" {
		instance.Integrations.GitHub.ClientID = githubClientID
	}
	if githubClientSecret := os.Getenv("INTEGRATIONS_GITHUB_CLIENT_SECRET"); githubClientSecret != "" {
		instance.Integrations.GitHub.ClientSecret = githubClientSecret
	}
	if linearClientID := os.Getenv("INTEGRATIONS_LINEAR_CLIENT_ID"); linearClientID != "" {
		instance.Integrations.Linear.ClientID = linearClientID
	}
	if linearClientSecret := os.Getenv("INTEGRATIONS_LINEAR_CLIENT_SECRET"); linearClientSecret != "" {
		instance.Integrations.Linear.ClientSecret = linearClientSecret
	}
	if slackClientID := os.Getenv("INTEGRATIONS_SLACK_CLIENT_ID"); slackClientID != "" {
		instance.Integrations.Slack.ClientID = slackClientID
	}
	if slackClientSecret := os.Getenv("INTEGRATIONS_SLACK_CLIENT_SECRET"); slackClientSecret != "" {
		instance.Integrations.Slack.ClientSecret = slackClientSecret
	}

	// Service token
	if serviceToken := os.Getenv("SERVICE_TOKEN"); serviceToken != "" {
		instance.Service.Token = serviceToken
	}
}

func (c *setupController) applyDefaults(instance *Configuration) {
	if instance.Server.Addr == "" {
		instance.Server.Addr = ":8080"
	}

	if instance.Server.AppName == "" {
		instance.Server.AppName = "Passflow API"
	}

	if instance.Server.BodyLimit == 0 {
		instance.Server.BodyLimit = 10 * 1024 * 1024
	}

	// Database defaults
	if instance.Database.DBName == "" {
		instance.Database.DBName = "passflow"
	}

	if instance.JWT.AccessExpiration == 0 {
		instance.JWT.AccessExpiration = 3600
	}

	if instance.JWT.RefreshExpiration == 0 {
		instance.JWT.RefreshExpiration = 604800
	}

	if instance.Integrations.EncryptionKey == "" {
		// Default key for development only (32 bytes for AES-256)
		instance.Integrations.EncryptionKey = "passflow-dev-key-0123456789abcde"
	}

	// Resend email defaults
	if instance.Email.Resend.FromEmail == "" {
		instance.Email.Resend.FromEmail = "noreply@example.com"
	}
	if instance.Email.Resend.FromName == "" {
		instance.Email.Resend.FromName = "Passflow"
	}
	if instance.Email.Resend.BaseURL == "" {
		instance.Email.Resend.BaseURL = "http://localhost:3000"
	}
}
