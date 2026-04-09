package fallback

import (
	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm/secrets"
)

// FallbackAction represents what action to take on error.
type FallbackAction int

const (
	ActionStop          FallbackAction = iota // Stop, no more retries
	ActionRetrySame                           // Retry same provider/credential
	ActionNextCredential                      // Try next credential for same provider
	ActionNextProvider                        // Try next provider in fallback chain
)

// Strategy defines fallback behavior.
type Strategy struct {
	MaxRetries         int
	MaxProviderRetries int
}

// DefaultStrategy returns the default fallback strategy.
func DefaultStrategy() *Strategy {
	return &Strategy{
		MaxRetries:         3,
		MaxProviderRetries: 2,
	}
}

// DetermineAction decides what to do based on error and attempt count.
func (s *Strategy) DetermineAction(
	errClass ErrorClass,
	attempt int,
	providerAttempt int,
	fallbackChain []secrets.FallbackEntry,
	currentIndex int,
) FallbackAction {
	if attempt >= s.MaxRetries {
		return ActionStop
	}

	switch errClass {
	case ErrorClassRateLimit:
		// Rate limit: try next credential, then next provider
		if providerAttempt < s.MaxProviderRetries {
			return ActionNextCredential
		}
		if currentIndex < len(fallbackChain)-1 {
			return ActionNextProvider
		}
		return ActionStop

	case ErrorClassProviderDown:
		// Provider down: skip to next provider immediately
		if currentIndex < len(fallbackChain)-1 {
			return ActionNextProvider
		}
		return ActionStop

	case ErrorClassInvalidAuth:
		// Auth error: do not retry blindly
		return ActionStop

	case ErrorClassPolicyViolation:
		// Policy violation: no cross-provider fallback
		return ActionStop

	case ErrorClassBadRequest:
		// Bad request: client error, no retry
		return ActionStop

	default:
		// Unknown: retry same once
		if attempt == 0 {
			return ActionRetrySame
		}
		return ActionStop
	}
}
