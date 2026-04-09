package redactor

import "testing"

func TestRedactString(t *testing.T) {
	input := `Authorization: Bearer abcdef123456 token=secret-value`
	got := RedactString(input)
	if got == input {
		t.Fatal("expected redaction")
	}
}

func TestRedactAnyMap(t *testing.T) {
	payload := map[string]any{
		"api_key": "secret",
		"nested": map[string]any{
			"token": "abc",
		},
		"message": "Bearer token123",
	}
	got := RedactAnyMap(payload)
	if got["api_key"] != RedactedValue {
		t.Fatalf("expected api_key redacted, got %v", got["api_key"])
	}
	nested := got["nested"].(map[string]any)
	if nested["token"] != RedactedValue {
		t.Fatalf("expected nested token redacted, got %v", nested["token"])
	}
}
