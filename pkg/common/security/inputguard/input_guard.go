package inputguard

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const (
	MaxFreeTextLen    = 8000
	MaxStringLen      = 8000
	MaxListItems      = 50
	MaxMapEntries     = 100
	MaxNestedDepth    = 6
	MaxAggregateNodes = 500
)

var suspiciousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bignore\s+(all\s+)?previous\s+instructions\b`),
	regexp.MustCompile(`(?i)\bignore\s+(the\s+)?system\s+prompt\b`),
	regexp.MustCompile(`(?i)\breveal\s+(your\s+)?(system\s+prompt|hidden\s+prompt|developer\s+message)\b`),
	regexp.MustCompile(`(?i)\bdeveloper\s+message\b`),
	regexp.MustCompile(`(?i)\bsystem\s+prompt\b`),
	regexp.MustCompile(`(?i)<\s*/?\s*system\s*>`),
	regexp.MustCompile(`(?i)<\s*/?\s*assistant\s*>`),
	regexp.MustCompile(`(?i)\bjailbreak\b`),
	regexp.MustCompile(`(?i)\bdo not follow your rules\b`),
}

func NormalizeText(input string) string {
	input = strings.TrimSpace(input)
	var b strings.Builder
	b.Grow(len(input))
	for _, r := range input {
		switch {
		case r == '\n' || r == '\r' || r == '\t':
			b.WriteRune(r)
		case unicode.IsControl(r):
			continue
		default:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

func ValidateFreeText(label, input string, required bool) (string, error) {
	normalized := NormalizeText(input)
	if required && normalized == "" {
		return "", fmt.Errorf("%s is required", label)
	}
	if normalized == "" {
		return "", nil
	}
	if len(normalized) > MaxFreeTextLen {
		return "", fmt.Errorf("%s exceeds max length", label)
	}
	if looksSuspicious(normalized) {
		return "", fmt.Errorf("%s contains unsafe prompt-like instructions", label)
	}
	return normalized, nil
}

func ValidateShortText(label, input string) (string, error) {
	normalized := NormalizeText(input)
	if normalized == "" {
		return "", nil
	}
	if len(normalized) > MaxStringLen {
		return "", fmt.Errorf("%s exceeds max length", label)
	}
	if looksSuspicious(normalized) {
		return "", fmt.Errorf("%s contains unsafe prompt-like instructions", label)
	}
	return normalized, nil
}

func ValidateStringSlice(label string, values []string) ([]string, error) {
	if values == nil {
		return nil, nil
	}
	if len(values) > MaxListItems {
		return nil, fmt.Errorf("%s exceeds max item count", label)
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		normalized, err := ValidateShortText(label, value)
		if err != nil {
			return nil, err
		}
		if normalized == "" {
			continue
		}
		out = append(out, normalized)
	}
	return out, nil
}

func ValidateGenericPayload(label string, payload map[string]any) error {
	if len(payload) > MaxMapEntries {
		return fmt.Errorf("%s exceeds max field count", label)
	}
	count := 0
	return walkValue(label, payload, 0, &count)
}

func ValidateAnyValue(label string, value any) error {
	count := 0
	return walkValue(label, value, 0, &count)
}

func walkValue(label string, value any, depth int, count *int) error {
	if depth > MaxNestedDepth {
		return fmt.Errorf("%s exceeds max nesting depth", label)
	}
	*count++
	if *count > MaxAggregateNodes {
		return fmt.Errorf("%s exceeds max payload complexity", label)
	}

	switch typed := value.(type) {
	case string:
		_, err := ValidateShortText(label, typed)
		return err
	case []string:
		_, err := ValidateStringSlice(label, typed)
		return err
	case []any:
		if len(typed) > MaxListItems {
			return fmt.Errorf("%s exceeds max item count", label)
		}
		for _, item := range typed {
			if err := walkValue(label, item, depth+1, count); err != nil {
				return err
			}
		}
	case map[string]any:
		if len(typed) > MaxMapEntries {
			return fmt.Errorf("%s exceeds max field count", label)
		}
		for key, item := range typed {
			if _, err := ValidateShortText(label+"."+key, key); err != nil {
				return err
			}
			if err := walkValue(label+"."+key, item, depth+1, count); err != nil {
				return err
			}
		}
	}
	return nil
}

func looksSuspicious(input string) bool {
	for _, pattern := range suspiciousPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}
