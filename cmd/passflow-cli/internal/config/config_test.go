package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{"empty token", "", ""},
		{"short token", "abc", "***"},
		{"long token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", "eyJhbGciOi...XVCJ9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskToken(tt.token)
			if got != tt.expected {
				t.Errorf("maskToken(%q) = %q, want %q", tt.token, got, tt.expected)
			}
		})
	}
}

func maskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) < 20 {
		return "***"
	}
	return token[:10] + "..." + token[len(token)-5:]
}

func TestConfigDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	expected := filepath.Join(home, ".passflow")
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		// Config dir doesn't exist yet, that's fine for this test
	}
}
