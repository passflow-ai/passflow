package fallback

import (
	"net/http"
	"strings"
)

// ErrorClass categorizes errors for fallback decisions.
type ErrorClass string

const (
	ErrorClassRateLimit       ErrorClass = "rate_limit"       // 429, quota exceeded
	ErrorClassProviderDown    ErrorClass = "provider_down"    // 5xx, timeout
	ErrorClassInvalidAuth     ErrorClass = "invalid_auth"     // 401, 403
	ErrorClassPolicyViolation ErrorClass = "policy_violation" // content blocked
	ErrorClassBadRequest      ErrorClass = "bad_request"      // 400, invalid input
	ErrorClassUnknown         ErrorClass = "unknown"
)

// ClassifyError categorizes an error for fallback decisions.
func ClassifyError(err error, statusCode int) ErrorClass {
	if err == nil && statusCode == 0 {
		return ErrorClassUnknown
	}

	// Check status code first
	switch {
	case statusCode == http.StatusTooManyRequests:
		return ErrorClassRateLimit
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		return ErrorClassInvalidAuth
	case statusCode >= 500 && statusCode < 600:
		return ErrorClassProviderDown
	case statusCode == http.StatusBadRequest:
		return ErrorClassBadRequest
	}

	// Check error message patterns
	if err != nil {
		errStr := strings.ToLower(err.Error())

		if strings.Contains(errStr, "rate limit") ||
			strings.Contains(errStr, "quota exceeded") ||
			strings.Contains(errStr, "too many requests") {
			return ErrorClassRateLimit
		}

		if strings.Contains(errStr, "timeout") ||
			strings.Contains(errStr, "connection refused") ||
			strings.Contains(errStr, "service unavailable") {
			return ErrorClassProviderDown
		}

		if strings.Contains(errStr, "invalid api key") ||
			strings.Contains(errStr, "unauthorized") ||
			strings.Contains(errStr, "authentication") {
			return ErrorClassInvalidAuth
		}

		if strings.Contains(errStr, "content policy") ||
			strings.Contains(errStr, "safety") ||
			strings.Contains(errStr, "blocked") {
			return ErrorClassPolicyViolation
		}
	}

	return ErrorClassUnknown
}

// IsRetryable returns true if the error class allows retry.
func IsRetryable(class ErrorClass) bool {
	switch class {
	case ErrorClassRateLimit, ErrorClassProviderDown:
		return true
	default:
		return false
	}
}

// AllowsCrossProviderFallback returns true if cross-provider fallback is allowed.
func AllowsCrossProviderFallback(class ErrorClass) bool {
	switch class {
	case ErrorClassProviderDown:
		return true
	case ErrorClassRateLimit:
		return true // Can try different provider
	default:
		return false
	}
}
