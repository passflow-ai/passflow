package registry

import "fmt"

// Capabilities describes what a model can do.
type Capabilities struct {
	ToolCalling   bool   `json:"tool_calling"`
	Vision        bool   `json:"vision"`
	ContextWindow int    `json:"context_window"`
	Streaming     bool   `json:"streaming"`
	JSONMode      bool   `json:"json_mode"`
	FunctionStyle string `json:"function_style"` // openai | anthropic | gemini
}

// Satisfies returns true if this capabilities meets the required capabilities.
func (c Capabilities) Satisfies(required Capabilities) bool {
	if required.ToolCalling && !c.ToolCalling {
		return false
	}
	if required.Vision && !c.Vision {
		return false
	}
	if required.ContextWindow > 0 && c.ContextWindow < required.ContextWindow {
		return false
	}
	if required.Streaming && !c.Streaming {
		return false
	}
	if required.JSONMode && !c.JSONMode {
		return false
	}
	return true
}

// Model represents an LLM model with its capabilities.
type Model struct {
	ID           string       `json:"id"`
	Provider     string       `json:"provider"`
	Capabilities Capabilities `json:"capabilities"`
	Aliases      []string     `json:"aliases"`
}

// String returns provider/model format.
func (m Model) String() string {
	return fmt.Sprintf("%s/%s", m.Provider, m.ID)
}
