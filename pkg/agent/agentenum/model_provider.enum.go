package agentenum

// ModelProvider represents the AI model provider.
type ModelProvider string

const (
	ProviderAnthropic ModelProvider = "anthropic"
	ProviderGateway   ModelProvider = "gateway"
	ProviderOpenAI    ModelProvider = "openai"
	ProviderGoogle    ModelProvider = "google"
	ProviderMistral   ModelProvider = "mistral"
	ProviderLocal     ModelProvider = "local"
	ProviderOllama    ModelProvider = "ollama"
	ProviderQwen      ModelProvider = "qwen"
)

// IsValid checks if the provider is a valid model provider.
func (p ModelProvider) IsValid() bool {
	switch p {
	case ProviderAnthropic, ProviderGateway, ProviderOpenAI, ProviderGoogle, ProviderMistral, ProviderLocal, ProviderOllama, ProviderQwen:
		return true
	}
	return false
}

// String returns the string representation of the provider.
func (p ModelProvider) String() string {
	return string(p)
}

// DisplayName returns a human-readable name for the provider.
func (p ModelProvider) DisplayName() string {
	switch p {
	case ProviderAnthropic:
		return "Anthropic"
	case ProviderGateway:
		return "Passflow Gateway"
	case ProviderOpenAI:
		return "OpenAI"
	case ProviderGoogle:
		return "Google AI"
	case ProviderMistral:
		return "Mistral AI"
	case ProviderLocal:
		return "Local Model"
	case ProviderOllama:
		return "Ollama"
	case ProviderQwen:
		return "Qwen"
	}
	return "Unknown"
}

// DefaultModel returns the default model ID for the provider.
func (p ModelProvider) DefaultModel() string {
	switch p {
	case ProviderAnthropic:
		return "claude-opus-4-5-20251101"
	case ProviderGateway:
		return "auto"
	case ProviderOpenAI:
		return "gpt-4o"
	case ProviderGoogle:
		return "gemini-2.0-flash"
	case ProviderMistral:
		return "mistral-large-latest"
	case ProviderLocal:
		return "llama-3.2"
	case ProviderOllama:
		return "qwen2.5:7b"
	case ProviderQwen:
		return "qwen3.6-plus"
	}
	return ""
}

// AllProviders returns all valid model providers.
func AllProviders() []ModelProvider {
	return []ModelProvider{
		ProviderAnthropic,
		ProviderGateway,
		ProviderOpenAI,
		ProviderGoogle,
		ProviderMistral,
		ProviderLocal,
		ProviderOllama,
		ProviderQwen,
	}
}
