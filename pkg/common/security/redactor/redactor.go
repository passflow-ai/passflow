package redactor

import (
	"regexp"
	"strings"
)

const RedactedValue = "[REDACTED]"

var sensitiveKeyPatterns = []string{
	"password",
	"secret",
	"token",
	"api_key",
	"apikey",
	"authorization",
	"cookie",
	"access_key",
	"private_key",
	"client_secret",
	"webhook_url",
}

var sensitiveValuePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bbearer\s+[a-z0-9\-\._~\+/]+=*`),
	regexp.MustCompile(`(?i)\bsk-[a-z0-9]{8,}\b`),
	regexp.MustCompile(`(?i)\b(api[_-]?key|token|secret|password)\s*[:=]\s*["']?[^"',\s]+`),
	regexp.MustCompile(`https://[^?\s]+[?&](token|key|secret)=[^&\s]+`),
}

func RedactString(input string) string {
	out := input
	for _, pattern := range sensitiveValuePatterns {
		out = pattern.ReplaceAllString(out, RedactedValue)
	}
	return out
}

func RedactStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		if isSensitiveKey(key) {
			out[key] = RedactedValue
			continue
		}
		out[key] = RedactString(value)
	}
	return out
}

func RedactAnyMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = redactValue(key, value)
	}
	return out
}

func redactValue(key string, value any) any {
	if isSensitiveKey(key) {
		return RedactedValue
	}
	switch typed := value.(type) {
	case string:
		return RedactString(typed)
	case map[string]any:
		return RedactAnyMap(typed)
	case map[string]string:
		redacted := RedactStringMap(typed)
		out := make(map[string]any, len(redacted))
		for k, v := range redacted {
			out[k] = v
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = redactValue(key, item)
		}
		return out
	case []string:
		out := make([]string, len(typed))
		for i, item := range typed {
			out[i] = RedactString(item)
		}
		return out
	default:
		return value
	}
}

func isSensitiveKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, pattern := range sensitiveKeyPatterns {
		if strings.Contains(key, pattern) {
			return true
		}
	}
	return false
}
