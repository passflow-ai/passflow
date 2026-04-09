package domain

import (
	"encoding/json"
)

// AgentCard represents the A2A (Agent-to-Agent) discovery metadata for an agent.
// This follows the Google A2A protocol for standardized agent communication.
type AgentCard struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Version      string          `json:"version"`
	Capabilities []string        `json:"capabilities"`
	InputSchema  json.RawMessage `json:"input_schema,omitempty"`
	OutputSchema json.RawMessage `json:"output_schema,omitempty"`
	AuthRequired bool            `json:"auth_required"`
	Endpoint     string          `json:"endpoint,omitempty"`
}

// ToAgentCard converts an Agent to an AgentCard for A2A discovery.
func (a *Agent) ToAgentCard(baseURL string) *AgentCard {
	// Determine capabilities based on agent configuration
	capabilities := []string{"execute", "chat"}

	if a.Heartbeat.Enabled {
		capabilities = append(capabilities, "heartbeat", "monitoring")
	}

	if len(a.Integrations) > 0 {
		capabilities = append(capabilities, "integrations")
	}

	// Version is derived from model config or defaults to "1.0"
	version := "1.0"
	if a.Model.ModelID != "" {
		version = a.Model.Provider + "/" + a.Model.ModelID
	}

	return &AgentCard{
		Name:         a.Name,
		Description:  a.Persona,
		Version:      version,
		Capabilities: capabilities,
		AuthRequired: true,
		Endpoint:     baseURL + "/api/v1/agents/" + a.ID + "/execute",
	}
}

// AgentCardsResponse represents a collection of agent cards.
type AgentCardsResponse struct {
	Agents  []*AgentCard `json:"agents"`
	Count   int          `json:"count"`
	Version string       `json:"version"`
}
