package enforcement

import (
	"strings"
	"testing"
)

// validConfig returns a ProviderConfig where all checks pass.
func validConfig() ProviderConfig {
	return ProviderConfig{
		PlatformEnabled:  true,
		WorkspaceEnabled: true,
		ModelAllowed:     true,
		HasAPIKey:        true,
		SpendingLimit:    0,
		CurrentSpend:     0,
		GlobalLimit:      0,
		GlobalSpend:      0,
	}
}

func TestCheck_AllValid(t *testing.T) {
	err := Check(validConfig(), "anthropic", "claude-opus-4-6")
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestCheck_PlatformDisabled(t *testing.T) {
	cfg := validConfig()
	cfg.PlatformEnabled = false

	err := Check(cfg, "anthropic", "claude-opus-4-6")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not enabled on this platform") {
		t.Errorf("expected error about platform not enabled, got: %v", err)
	}
	if !strings.Contains(err.Error(), "anthropic") {
		t.Errorf("expected error to mention provider name, got: %v", err)
	}
}

func TestCheck_WorkspaceDisabled(t *testing.T) {
	cfg := validConfig()
	cfg.WorkspaceEnabled = false

	err := Check(cfg, "openai", "gpt-4o")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not configured in this workspace") {
		t.Errorf("expected error about workspace not configured, got: %v", err)
	}
	if !strings.Contains(err.Error(), "openai") {
		t.Errorf("expected error to mention provider name, got: %v", err)
	}
}

func TestCheck_ModelNotAllowed(t *testing.T) {
	cfg := validConfig()
	cfg.ModelAllowed = false

	err := Check(cfg, "anthropic", "claude-opus-4-6")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not allowed in this workspace") {
		t.Errorf("expected error about model not allowed, got: %v", err)
	}
	if !strings.Contains(err.Error(), "claude-opus-4-6") {
		t.Errorf("expected error to mention model name, got: %v", err)
	}
}

func TestCheck_NoAPIKey(t *testing.T) {
	cfg := validConfig()
	cfg.HasAPIKey = false

	err := Check(cfg, "gemini", "gemini-2.0-flash")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no API key configured") {
		t.Errorf("expected error about missing API key, got: %v", err)
	}
	if !strings.Contains(err.Error(), "gemini") {
		t.Errorf("expected error to mention provider name, got: %v", err)
	}
}

func TestCheck_ProviderSpendingLimitExceeded(t *testing.T) {
	cfg := validConfig()
	cfg.SpendingLimit = 100.0
	cfg.CurrentSpend = 100.0

	err := Check(cfg, "openai", "gpt-4o")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "spending limit reached") {
		t.Errorf("expected error about spending limit, got: %v", err)
	}
	if !strings.Contains(err.Error(), "100.00/100.00") {
		t.Errorf("expected error to show spend amounts, got: %v", err)
	}
}

func TestCheck_ProviderSpendingLimitNotExceeded(t *testing.T) {
	cfg := validConfig()
	cfg.SpendingLimit = 100.0
	cfg.CurrentSpend = 50.0

	err := Check(cfg, "openai", "gpt-4o")
	if err != nil {
		t.Errorf("expected nil error when under spending limit, got: %v", err)
	}
}

func TestCheck_GlobalSpendingLimitExceeded(t *testing.T) {
	cfg := validConfig()
	cfg.GlobalLimit = 500.0
	cfg.GlobalSpend = 500.0

	err := Check(cfg, "anthropic", "claude-opus-4-6")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "workspace global spending limit reached") {
		t.Errorf("expected error about global spending limit, got: %v", err)
	}
	if !strings.Contains(err.Error(), "500.00/500.00") {
		t.Errorf("expected error to show global spend amounts, got: %v", err)
	}
}

func TestCheck_UnlimitedSpending(t *testing.T) {
	cfg := validConfig()
	cfg.SpendingLimit = 0
	cfg.CurrentSpend = 999999.99
	cfg.GlobalLimit = 0
	cfg.GlobalSpend = 999999.99

	err := Check(cfg, "openai", "gpt-4o")
	if err != nil {
		t.Errorf("expected nil error when limits are 0 (unlimited), got: %v", err)
	}
}
