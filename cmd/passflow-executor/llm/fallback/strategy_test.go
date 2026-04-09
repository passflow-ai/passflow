package fallback

import (
	"testing"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm/secrets"
)

func TestDefaultStrategy(t *testing.T) {
	s := DefaultStrategy()
	if s.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", s.MaxRetries)
	}
	if s.MaxProviderRetries != 2 {
		t.Errorf("expected MaxProviderRetries=2, got %d", s.MaxProviderRetries)
	}
}

func TestDetermineAction_MaxRetriesExceeded(t *testing.T) {
	s := DefaultStrategy()
	action := s.DetermineAction(ErrorClassRateLimit, 3, 0, nil, 0)
	if action != ActionStop {
		t.Errorf("expected ActionStop when max retries exceeded, got %v", action)
	}
}

func TestDetermineAction_RateLimit_NextCredential(t *testing.T) {
	s := DefaultStrategy()
	fallbacks := []secrets.FallbackEntry{{Provider: "openai"}}
	action := s.DetermineAction(ErrorClassRateLimit, 0, 0, fallbacks, 0)
	if action != ActionNextCredential {
		t.Errorf("expected ActionNextCredential for rate_limit on first attempt, got %v", action)
	}
}

func TestDetermineAction_RateLimit_NextProvider(t *testing.T) {
	s := DefaultStrategy()
	fallbacks := []secrets.FallbackEntry{{Provider: "openai"}, {Provider: "gemini"}}
	action := s.DetermineAction(ErrorClassRateLimit, 1, 2, fallbacks, 0)
	if action != ActionNextProvider {
		t.Errorf("expected ActionNextProvider when provider retries exhausted, got %v", action)
	}
}

func TestDetermineAction_ProviderDown_NextProvider(t *testing.T) {
	s := DefaultStrategy()
	fallbacks := []secrets.FallbackEntry{{Provider: "openai"}, {Provider: "gemini"}}
	action := s.DetermineAction(ErrorClassProviderDown, 0, 0, fallbacks, 0)
	if action != ActionNextProvider {
		t.Errorf("expected ActionNextProvider for provider_down, got %v", action)
	}
}

func TestDetermineAction_InvalidAuth_Stop(t *testing.T) {
	s := DefaultStrategy()
	action := s.DetermineAction(ErrorClassInvalidAuth, 0, 0, nil, 0)
	if action != ActionStop {
		t.Errorf("expected ActionStop for invalid_auth, got %v", action)
	}
}

func TestDetermineAction_PolicyViolation_Stop(t *testing.T) {
	s := DefaultStrategy()
	action := s.DetermineAction(ErrorClassPolicyViolation, 0, 0, nil, 0)
	if action != ActionStop {
		t.Errorf("expected ActionStop for policy_violation, got %v", action)
	}
}

func TestDetermineAction_Unknown_RetrySameOnce(t *testing.T) {
	s := DefaultStrategy()
	action := s.DetermineAction(ErrorClassUnknown, 0, 0, nil, 0)
	if action != ActionRetrySame {
		t.Errorf("expected ActionRetrySame for unknown on first attempt, got %v", action)
	}

	action = s.DetermineAction(ErrorClassUnknown, 1, 0, nil, 0)
	if action != ActionStop {
		t.Errorf("expected ActionStop for unknown on second attempt, got %v", action)
	}
}
