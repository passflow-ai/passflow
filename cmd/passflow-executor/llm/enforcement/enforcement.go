package enforcement

import "fmt"

// ProviderConfig holds the data needed for enforcement checks.
type ProviderConfig struct {
	PlatformEnabled  bool
	WorkspaceEnabled bool
	ModelAllowed     bool
	HasAPIKey        bool
	SpendingLimit    float64 // per-provider monthly limit in USD, 0 = unlimited
	CurrentSpend     float64 // current month spend for this provider
	GlobalLimit      float64 // workspace global monthly limit in USD, 0 = unlimited
	GlobalSpend      float64 // current month total spend across all providers
}

// Check validates that an agent execution is allowed.
// Returns nil if allowed, error with clear message if not.
func Check(cfg ProviderConfig, provider, model string) error {
	if !cfg.PlatformEnabled {
		return fmt.Errorf("provider %q is not enabled on this platform", provider)
	}
	if !cfg.WorkspaceEnabled {
		return fmt.Errorf("provider %q is not configured in this workspace", provider)
	}
	if !cfg.ModelAllowed {
		return fmt.Errorf("model %q is not allowed in this workspace", model)
	}
	if !cfg.HasAPIKey {
		return fmt.Errorf("no API key configured for provider %q in this workspace", provider)
	}
	if cfg.SpendingLimit > 0 && cfg.CurrentSpend >= cfg.SpendingLimit {
		return fmt.Errorf("spending limit reached for provider %q (%.2f/%.2f USD)", provider, cfg.CurrentSpend, cfg.SpendingLimit)
	}
	if cfg.GlobalLimit > 0 && cfg.GlobalSpend >= cfg.GlobalLimit {
		return fmt.Errorf("workspace global spending limit reached (%.2f/%.2f USD)", cfg.GlobalSpend, cfg.GlobalLimit)
	}
	return nil
}
