package config

import (
	"net/url"
	"os"
	"strings"
)

type Config struct {
	// Server
	Addr string // CHANNELS_ADDR (default: :8081)

	// Redis — rules cache + event bus
	RedisURL  string // REDIS_URL
	StreamKey string // PASSFLOW_STREAM_KEY (same as agent-executor)

	// passflow-api
	APIURL       string // PASSFLOW_API_URL
	ServiceToken string // PASSFLOW_SERVICE_TOKEN

	// Slack Events API
	SlackSigningSecret string // SLACK_SIGNING_SECRET

	// Email (IMAP input)
	IMAPHost         string // IMAP_HOST
	IMAPPort         string // IMAP_PORT  (default: 993)
	IMAPUser         string // IMAP_USER
	IMAPPassword     string // IMAP_PASSWORD
	IMAPMailbox      string // IMAP_MAILBOX (default: INBOX)
	IMAPPollSec      int    // IMAP_POLL_SECONDS (default: 60)
	EmailWorkspaceID string // EMAIL_WORKSPACE_ID — required; events without this are dropped

	// Email (SMTP output)
	SMTPHost     string // SMTP_HOST
	SMTPPort     string // SMTP_PORT (default: 587)
	SMTPUser     string // SMTP_USER
	SMTPPassword string // SMTP_PASSWORD
	SMTPFrom     string // SMTP_FROM
}

func Load() *Config {
	return &Config{
		Addr:               getEnv("CHANNELS_ADDR", ":8081"),
		RedisURL:           parseRedisAddr(os.Getenv("REDIS_URL")),
		StreamKey:          getEnv("PASSFLOW_STREAM_KEY", "passflow.tasks.queue"),
		APIURL:             getEnv("PASSFLOW_API_URL", "http://passflow-api:8080"),
		ServiceToken:       getEnv("PASSFLOW_SERVICE_TOKEN", ""),
		SlackSigningSecret: getEnv("SLACK_SIGNING_SECRET", ""),
		IMAPHost:           getEnv("IMAP_HOST", ""),
		IMAPPort:           getEnv("IMAP_PORT", "993"),
		IMAPUser:           getEnv("IMAP_USER", ""),
		IMAPPassword:       getEnv("IMAP_PASSWORD", ""),
		IMAPMailbox:        getEnv("IMAP_MAILBOX", "INBOX"),
		IMAPPollSec:        getEnvInt("IMAP_POLL_SECONDS", 60),
		EmailWorkspaceID:   getEnv("EMAIL_WORKSPACE_ID", ""),
		SMTPHost:           getEnv("SMTP_HOST", ""),
		SMTPPort:           getEnv("SMTP_PORT", "587"),
		SMTPUser:           getEnv("SMTP_USER", ""),
		SMTPPassword:       getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:           getEnv("SMTP_FROM", ""),
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

// parseRedisAddr extracts host:port from a Redis URL.
// Supports both "redis://host:port" and plain "host:port" formats.
func parseRedisAddr(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "localhost:6379"
	}

	// If it has a scheme, parse as URL
	if strings.HasPrefix(rawURL, "redis://") || strings.HasPrefix(rawURL, "rediss://") {
		if parsed, err := url.Parse(rawURL); err == nil && parsed.Host != "" {
			return parsed.Host
		}
	}

	// Otherwise return as-is (assumed to be host:port)
	return rawURL
}
