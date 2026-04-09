package registry

import "sync"

// Registry holds all known models.
type Registry struct {
	mu      sync.RWMutex
	models  map[string]Model
	aliases map[string]string
}

// New creates a registry with default models.
func New() *Registry {
	r := &Registry{
		models:  make(map[string]Model),
		aliases: make(map[string]string),
	}
	r.registerDefaults()
	return r
}

// Get retrieves a model by ID or alias.
func (r *Registry) Get(idOrAlias string) (Model, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if m, ok := r.models[idOrAlias]; ok {
		return m, true
	}
	if id, ok := r.aliases[idOrAlias]; ok {
		return r.models[id], true
	}
	return Model{}, false
}

// Register adds a model to the registry.
func (r *Registry) Register(m Model) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.models[m.ID] = m
	for _, alias := range m.Aliases {
		r.aliases[alias] = m.ID
	}
}

// ListByProvider returns all models for a provider.
func (r *Registry) ListByProvider(provider string) []Model {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Model
	for _, m := range r.models {
		if m.Provider == provider {
			result = append(result, m)
		}
	}
	return result
}

// All returns all registered models.
func (r *Registry) All() []Model {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Model, 0, len(r.models))
	for _, m := range r.models {
		result = append(result, m)
	}
	return result
}

func (r *Registry) registerDefaults() {
	defaults := []Model{
		// OpenAI
		{
			ID:       "gpt-4o",
			Provider: "openai",
			Aliases:  []string{"gpt4o"},
			Capabilities: Capabilities{
				ToolCalling:   true,
				Vision:        true,
				ContextWindow: 128000,
				Streaming:     true,
				JSONMode:      true,
				FunctionStyle: "openai",
			},
		},
		{
			ID:       "gpt-4o-mini",
			Provider: "openai",
			Aliases:  []string{"gpt4o-mini"},
			Capabilities: Capabilities{
				ToolCalling:   true,
				Vision:        true,
				ContextWindow: 128000,
				Streaming:     true,
				JSONMode:      true,
				FunctionStyle: "openai",
			},
		},
		{
			ID:       "gpt-4-turbo",
			Provider: "openai",
			Capabilities: Capabilities{
				ToolCalling:   true,
				Vision:        true,
				ContextWindow: 128000,
				Streaming:     true,
				JSONMode:      true,
				FunctionStyle: "openai",
			},
		},
		// Anthropic
		{
			ID:       "claude-3-5-sonnet-latest",
			Provider: "anthropic",
			Aliases:  []string{"claude-3-5-sonnet", "claude-sonnet"},
			Capabilities: Capabilities{
				ToolCalling:   true,
				Vision:        true,
				ContextWindow: 200000,
				Streaming:     true,
				JSONMode:      true,
				FunctionStyle: "anthropic",
			},
		},
		{
			ID:       "claude-3-5-haiku-latest",
			Provider: "anthropic",
			Aliases:  []string{"claude-3-5-haiku", "claude-haiku"},
			Capabilities: Capabilities{
				ToolCalling:   true,
				Vision:        true,
				ContextWindow: 200000,
				Streaming:     true,
				JSONMode:      true,
				FunctionStyle: "anthropic",
			},
		},
		{
			ID:       "claude-3-opus-latest",
			Provider: "anthropic",
			Aliases:  []string{"claude-3-opus", "claude-opus"},
			Capabilities: Capabilities{
				ToolCalling:   true,
				Vision:        true,
				ContextWindow: 200000,
				Streaming:     true,
				JSONMode:      true,
				FunctionStyle: "anthropic",
			},
		},
		// Gemini
		{
			ID:       "gemini-1.5-pro",
			Provider: "gemini",
			Aliases:  []string{"gemini-pro"},
			Capabilities: Capabilities{
				ToolCalling:   true,
				Vision:        true,
				ContextWindow: 1000000,
				Streaming:     true,
				JSONMode:      true,
				FunctionStyle: "gemini",
			},
		},
		{
			ID:       "gemini-1.5-flash",
			Provider: "gemini",
			Aliases:  []string{"gemini-flash"},
			Capabilities: Capabilities{
				ToolCalling:   true,
				Vision:        true,
				ContextWindow: 1000000,
				Streaming:     true,
				JSONMode:      true,
				FunctionStyle: "gemini",
			},
		},
		// Ollama (local)
		{
			ID:       "llama3",
			Provider: "ollama",
			Capabilities: Capabilities{
				ToolCalling:   true,
				Vision:        false,
				ContextWindow: 8000,
				Streaming:     true,
				JSONMode:      false,
				FunctionStyle: "openai",
			},
		},
		{
			ID:       "mistral",
			Provider: "ollama",
			Capabilities: Capabilities{
				ToolCalling:   true,
				Vision:        false,
				ContextWindow: 32000,
				Streaming:     true,
				JSONMode:      false,
				FunctionStyle: "openai",
			},
		},
	}

	for _, m := range defaults {
		r.Register(m)
	}
}
