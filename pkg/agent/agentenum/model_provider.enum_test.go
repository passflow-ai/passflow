package agentenum

import "testing"

func TestModelProvider_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		provider ModelProvider
		want     bool
	}{
		{"anthropic is valid", ProviderAnthropic, true},
		{"gateway is valid", ProviderGateway, true},
		{"openai is valid", ProviderOpenAI, true},
		{"google is valid", ProviderGoogle, true},
		{"mistral is valid", ProviderMistral, true},
		{"local is valid", ProviderLocal, true},
		{"ollama is valid", ProviderOllama, true},
		{"qwen is valid", ProviderQwen, true},
		{"invalid provider", ModelProvider("invalid"), false},
		{"empty provider", ModelProvider(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.provider.IsValid(); got != tt.want {
				t.Errorf("ModelProvider.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModelProvider_String(t *testing.T) {
	tests := []struct {
		name     string
		provider ModelProvider
		want     string
	}{
		{"anthropic", ProviderAnthropic, "anthropic"},
		{"gateway", ProviderGateway, "gateway"},
		{"openai", ProviderOpenAI, "openai"},
		{"google", ProviderGoogle, "google"},
		{"mistral", ProviderMistral, "mistral"},
		{"local", ProviderLocal, "local"},
		{"ollama", ProviderOllama, "ollama"},
		{"qwen", ProviderQwen, "qwen"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.provider.String(); got != tt.want {
				t.Errorf("ModelProvider.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModelProvider_DisplayName(t *testing.T) {
	tests := []struct {
		name     string
		provider ModelProvider
		want     string
	}{
		{"anthropic", ProviderAnthropic, "Anthropic"},
		{"gateway", ProviderGateway, "Passflow Gateway"},
		{"openai", ProviderOpenAI, "OpenAI"},
		{"google", ProviderGoogle, "Google AI"},
		{"mistral", ProviderMistral, "Mistral AI"},
		{"local", ProviderLocal, "Local Model"},
		{"ollama", ProviderOllama, "Ollama"},
		{"qwen", ProviderQwen, "Qwen"},
		{"invalid", ModelProvider("invalid"), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.provider.DisplayName(); got != tt.want {
				t.Errorf("ModelProvider.DisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModelProvider_DefaultModel(t *testing.T) {
	tests := []struct {
		name     string
		provider ModelProvider
		want     string
	}{
		{"anthropic", ProviderAnthropic, "claude-opus-4-5-20251101"},
		{"gateway", ProviderGateway, "auto"},
		{"openai", ProviderOpenAI, "gpt-4o"},
		{"google", ProviderGoogle, "gemini-2.0-flash"},
		{"mistral", ProviderMistral, "mistral-large-latest"},
		{"local", ProviderLocal, "llama-3.2"},
		{"ollama", ProviderOllama, "qwen2.5:7b"},
		{"qwen", ProviderQwen, "qwen3.6-plus"},
		{"invalid", ModelProvider("invalid"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.provider.DefaultModel(); got != tt.want {
				t.Errorf("ModelProvider.DefaultModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllProviders(t *testing.T) {
	providers := AllProviders()
	if len(providers) != 8 {
		t.Errorf("AllProviders() returned %d providers, want 8", len(providers))
	}

	expected := map[ModelProvider]bool{
		ProviderAnthropic: true,
		ProviderGateway:   true,
		ProviderOpenAI:    true,
		ProviderGoogle:    true,
		ProviderMistral:   true,
		ProviderLocal:     true,
		ProviderOllama:    true,
		ProviderQwen:      true,
	}

	for _, p := range providers {
		if !expected[p] {
			t.Errorf("Unexpected provider in AllProviders(): %v", p)
		}
	}
}
