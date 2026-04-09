package fallback

import (
	"errors"
	"net/http"
	"testing"
)

func TestClassifyError_RateLimit_ByStatusCode(t *testing.T) {
	class := ClassifyError(nil, http.StatusTooManyRequests)
	if class != ErrorClassRateLimit {
		t.Errorf("expected %s, got %s", ErrorClassRateLimit, class)
	}
}

func TestClassifyError_RateLimit_ByMessage(t *testing.T) {
	err := errors.New("rate limit exceeded")
	class := ClassifyError(err, 0)
	if class != ErrorClassRateLimit {
		t.Errorf("expected %s, got %s", ErrorClassRateLimit, class)
	}
}

func TestClassifyError_ProviderDown_5xx(t *testing.T) {
	class := ClassifyError(nil, http.StatusInternalServerError)
	if class != ErrorClassProviderDown {
		t.Errorf("expected %s, got %s", ErrorClassProviderDown, class)
	}
}

func TestClassifyError_ProviderDown_Timeout(t *testing.T) {
	err := errors.New("connection timeout")
	class := ClassifyError(err, 0)
	if class != ErrorClassProviderDown {
		t.Errorf("expected %s, got %s", ErrorClassProviderDown, class)
	}
}

func TestClassifyError_InvalidAuth(t *testing.T) {
	class := ClassifyError(nil, http.StatusUnauthorized)
	if class != ErrorClassInvalidAuth {
		t.Errorf("expected %s, got %s", ErrorClassInvalidAuth, class)
	}
}

func TestClassifyError_PolicyViolation(t *testing.T) {
	err := errors.New("content policy violation")
	class := ClassifyError(err, 0)
	if class != ErrorClassPolicyViolation {
		t.Errorf("expected %s, got %s", ErrorClassPolicyViolation, class)
	}
}

func TestClassifyError_BadRequest(t *testing.T) {
	class := ClassifyError(nil, http.StatusBadRequest)
	if class != ErrorClassBadRequest {
		t.Errorf("expected %s, got %s", ErrorClassBadRequest, class)
	}
}

func TestIsRetryable_RateLimit(t *testing.T) {
	if !IsRetryable(ErrorClassRateLimit) {
		t.Error("rate_limit should be retryable")
	}
}

func TestIsRetryable_ProviderDown(t *testing.T) {
	if !IsRetryable(ErrorClassProviderDown) {
		t.Error("provider_down should be retryable")
	}
}

func TestIsRetryable_InvalidAuth(t *testing.T) {
	if IsRetryable(ErrorClassInvalidAuth) {
		t.Error("invalid_auth should NOT be retryable")
	}
}

func TestAllowsCrossProviderFallback_ProviderDown(t *testing.T) {
	if !AllowsCrossProviderFallback(ErrorClassProviderDown) {
		t.Error("provider_down should allow cross-provider fallback")
	}
}

func TestAllowsCrossProviderFallback_InvalidAuth(t *testing.T) {
	if AllowsCrossProviderFallback(ErrorClassInvalidAuth) {
		t.Error("invalid_auth should NOT allow cross-provider fallback")
	}
}
