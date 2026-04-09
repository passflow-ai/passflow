package registry

import (
	"testing"
)

func TestCapabilities_Satisfies(t *testing.T) {
	full := Capabilities{
		ToolCalling:   true,
		Vision:        true,
		ContextWindow: 128000,
		Streaming:     true,
		JSONMode:      true,
		FunctionStyle: "openai",
	}

	tests := []struct {
		name     string
		required Capabilities
		want     bool
	}{
		{
			name:     "empty requirements satisfied",
			required: Capabilities{},
			want:     true,
		},
		{
			name:     "tool calling required and present",
			required: Capabilities{ToolCalling: true},
			want:     true,
		},
		{
			name:     "vision required and present",
			required: Capabilities{Vision: true},
			want:     true,
		},
		{
			name:     "context window sufficient",
			required: Capabilities{ContextWindow: 64000},
			want:     true,
		},
		{
			name:     "context window insufficient",
			required: Capabilities{ContextWindow: 256000},
			want:     false,
		},
		{
			name:     "streaming required and present",
			required: Capabilities{Streaming: true},
			want:     true,
		},
		{
			name:     "json mode required and present",
			required: Capabilities{JSONMode: true},
			want:     true,
		},
		{
			name:     "tool calling required but missing",
			required: Capabilities{ToolCalling: true},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := full.Satisfies(tt.required); got != tt.want {
				t.Errorf("Satisfies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCapabilities_Satisfies_Missing(t *testing.T) {
	limited := Capabilities{
		ToolCalling:   false,
		Vision:        false,
		ContextWindow: 8000,
		Streaming:     false,
		JSONMode:      false,
	}

	tests := []struct {
		name     string
		required Capabilities
		want     bool
	}{
		{
			name:     "tool calling required but missing",
			required: Capabilities{ToolCalling: true},
			want:     false,
		},
		{
			name:     "vision required but missing",
			required: Capabilities{Vision: true},
			want:     false,
		},
		{
			name:     "streaming required but missing",
			required: Capabilities{Streaming: true},
			want:     false,
		},
		{
			name:     "json mode required but missing",
			required: Capabilities{JSONMode: true},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := limited.Satisfies(tt.required); got != tt.want {
				t.Errorf("Satisfies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_String(t *testing.T) {
	tests := []struct {
		name     string
		model    Model
		expected string
	}{
		{
			name: "openai model",
			model: Model{
				ID:       "gpt-4o",
				Provider: "openai",
			},
			expected: "openai/gpt-4o",
		},
		{
			name: "anthropic model",
			model: Model{
				ID:       "claude-3-opus",
				Provider: "anthropic",
			},
			expected: "anthropic/claude-3-opus",
		},
		{
			name: "google model",
			model: Model{
				ID:       "gemini-pro",
				Provider: "google",
			},
			expected: "google/gemini-pro",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestModel_WithCapabilities(t *testing.T) {
	m := Model{
		ID:       "gpt-4o",
		Provider: "openai",
		Capabilities: Capabilities{
			ToolCalling:   true,
			Vision:        true,
			ContextWindow: 128000,
			Streaming:     true,
			JSONMode:      true,
			FunctionStyle: "openai",
		},
		Aliases: []string{"gpt4o", "gpt-4-omni"},
	}

	if m.ID != "gpt-4o" {
		t.Errorf("ID = %q, want %q", m.ID, "gpt-4o")
	}
	if m.Provider != "openai" {
		t.Errorf("Provider = %q, want %q", m.Provider, "openai")
	}
	if !m.Capabilities.ToolCalling {
		t.Error("ToolCalling should be true")
	}
	if len(m.Aliases) != 2 {
		t.Errorf("Aliases length = %d, want %d", len(m.Aliases), 2)
	}
}
