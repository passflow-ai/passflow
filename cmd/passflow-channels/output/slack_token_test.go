package output

import (
	"strings"
	"testing"
)

// TestRedactConfig_TokenIsRedacted verifies that redactConfig masks token values.
func TestRedactConfig_TokenIsRedacted(t *testing.T) {
	cfg := map[string]interface{}{
		"token":   "xoxb-secret-slack-token",
		"channel": "#general",
	}
	redacted := redactConfig(cfg)
	tokenVal, _ := redacted["token"].(string)
	if tokenVal == "xoxb-secret-slack-token" {
		t.Error("expected token to be redacted in log output, but it was returned as-is")
	}
	if !strings.Contains(tokenVal, "***") {
		t.Errorf("expected redacted token to contain ***, got %q", tokenVal)
	}
}

func TestRedactConfig_AuthTokenIsRedacted(t *testing.T) {
	cfg := map[string]interface{}{
		"auth_token": "twilio-secret-auth-token",
		"from":       "+15551234567",
	}
	redacted := redactConfig(cfg)
	tokenVal, _ := redacted["auth_token"].(string)
	if tokenVal == "twilio-secret-auth-token" {
		t.Error("expected auth_token to be redacted in log output, but it was returned as-is")
	}
	if !strings.Contains(tokenVal, "***") {
		t.Errorf("expected redacted auth_token to contain ***, got %q", tokenVal)
	}
}

// TestRedactConfig_NonTokenFieldPreserved verifies that non-sensitive fields
// are not modified by redactConfig.
func TestRedactConfig_NonTokenFieldPreserved(t *testing.T) {
	cfg := map[string]interface{}{
		"channel": "#alerts",
		"url":     "https://hooks.example.com",
	}
	redacted := redactConfig(cfg)
	if ch, _ := redacted["channel"].(string); ch != "#alerts" {
		t.Errorf("non-token field 'channel' should be preserved, got %q", ch)
	}
	if u, _ := redacted["url"].(string); u != "https://hooks.example.com" {
		t.Errorf("non-token field 'url' should be preserved, got %q", u)
	}
}

// TestRedactConfig_OriginalMapNotMutated verifies that redactConfig does not
// modify the original config map (returns a copy).
func TestRedactConfig_OriginalMapNotMutated(t *testing.T) {
	original := "xoxb-real-token"
	cfg := map[string]interface{}{
		"token": original,
	}
	_ = redactConfig(cfg)
	if cfg["token"] != original {
		t.Error("redactConfig must not mutate the original config map")
	}
}

// TestRedactConfig_EmptyConfig verifies that an empty config does not panic.
func TestRedactConfig_EmptyConfig(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("redactConfig panicked on empty config: %v", r)
		}
	}()
	_ = redactConfig(map[string]interface{}{})
}
