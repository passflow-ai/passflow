package domain

import (
	"errors"
	"fmt"

	"github.com/jaak-ai/passflow-api/src/usecase/agent/agentenum"
)

// Validation errors
var (
	ErrAgentNameRequired      = errors.New("agent name is required")
	ErrModelRequired          = errors.New("model is required")
	ErrModelProviderRequired  = errors.New("model provider is required")
	ErrModelProviderInvalid   = errors.New("invalid model provider")
	ErrModelNameRequired      = errors.New("model name is required")
	ErrReactSpecRequired      = errors.New("react spec is required")
	ErrReactToolsRequired     = errors.New("react tools are required")
)

// PALSpec defines the structure of a Passflow Agent Language specification
type PALSpec struct {
	Agent        *AgentSpec      `json:"agent" yaml:"agent"`
	Model        *ModelSpec      `json:"model" yaml:"model"`
	Persona      *PersonaSpec    `json:"persona,omitempty" yaml:"persona,omitempty"`         // Agent's persona with identity and tone
	Instructions string          `json:"instructions,omitempty" yaml:"instructions,omitempty"` // Agent's instructions/system prompt
	Triggers     []*TriggerSpec  `json:"triggers,omitempty" yaml:"triggers,omitempty"`
	React        *ReactSpec      `json:"react" yaml:"react"`
	Guardrails   *GuardrailsSpec `json:"guardrails,omitempty" yaml:"guardrails,omitempty"`
	Approvals    *ApprovalSpec   `json:"approvals,omitempty" yaml:"approvals,omitempty"`
	Output       *OutputSpec     `json:"output,omitempty" yaml:"output,omitempty"`
	Pod          *PodSpec        `json:"pod,omitempty" yaml:"pod,omitempty"`
}

// AgentSpec defines agent configuration
type AgentSpec struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Persona     string `json:"persona,omitempty" yaml:"persona,omitempty"` // Simple string persona (alternative to PersonaSpec)
}

// PersonaSpec defines the agent's persona configuration
type PersonaSpec struct {
	Identity string `json:"identity,omitempty" yaml:"identity,omitempty"` // Who the agent is
	Tone     string `json:"tone,omitempty" yaml:"tone,omitempty"`         // How the agent communicates
}

// TriggerSpec defines how an agent can be triggered
type TriggerSpec struct {
	Type        string            `json:"type" yaml:"type"`                   // cron, slack, email, webhook, manual
	Schedule    string            `json:"schedule,omitempty" yaml:"schedule"` // for cron triggers
	Endpoint    string            `json:"endpoint,omitempty" yaml:"endpoint"` // for webhook triggers
	Channel     string            `json:"channel,omitempty" yaml:"channel"`   // for slack triggers
	Filters     map[string]string `json:"filters,omitempty" yaml:"filters"`
}

// ModelSpec defines the LLM model to use
type ModelSpec struct {
	Provider string            `json:"provider" yaml:"provider"`           // anthropic, openai, bedrock, etc.
	Name     string            `json:"name" yaml:"name"`                   // model name like claude-3-opus
	Config   map[string]interface{} `json:"config,omitempty" yaml:"config"` // temperature, max_tokens, etc.
}

// ReactSpec defines the ReAct loop configuration
type ReactSpec struct {
	Tools          []string `json:"tools" yaml:"tools"`                      // Available MCP tools
	MaxIterations  int      `json:"max_iterations,omitempty" yaml:"max_iterations"`
	StopSequences  []string `json:"stop_sequences,omitempty" yaml:"stop_sequences"`
}

// GuardrailsSpec defines safety and compliance rules
type GuardrailsSpec struct {
	AllowedTools   []string          `json:"allowed_tools,omitempty" yaml:"allowed_tools"`
	BlockedTools   []string          `json:"blocked_tools,omitempty" yaml:"blocked_tools"`
	RateLimits     map[string]int    `json:"rate_limits,omitempty" yaml:"rate_limits"`
	TimeoutSeconds int               `json:"timeout_seconds,omitempty" yaml:"timeout_seconds"`
	CustomRules    map[string]string `json:"custom_rules,omitempty" yaml:"custom_rules"`
}

// ApprovalSpec defines approval workflows
type ApprovalSpec struct {
	Required      bool     `json:"required" yaml:"required"`
	Approvers     []string `json:"approvers,omitempty" yaml:"approvers"`
	TimeoutHours  int      `json:"timeout_hours,omitempty" yaml:"timeout_hours"`
}

// OutputSpec defines output handling
type OutputSpec struct {
	Format      string            `json:"format,omitempty" yaml:"format"` // json, markdown, plaintext
	Destination string            `json:"destination,omitempty" yaml:"destination"`
	Transform   map[string]string `json:"transform,omitempty" yaml:"transform"`
}

// PodSpec defines Kubernetes pod configuration for the executor
type PodSpec struct {
	Image       string            `json:"image,omitempty" yaml:"image"`
	Resources   map[string]string `json:"resources,omitempty" yaml:"resources"`
	Environment map[string]string `json:"environment,omitempty" yaml:"environment"`
}

// Validate validates the PALSpec
func (p *PALSpec) Validate() error {
	if p.Agent == nil {
		return ErrAgentNameRequired
	}

	if err := p.Agent.Validate(); err != nil {
		return err
	}

	if err := ModelValidate(p.Model); err != nil {
		return err
	}

	if err := ReactValidate(p.React); err != nil {
		return err
	}

	return nil
}

// Validate validates the AgentSpec
func (a *AgentSpec) Validate() error {
	if a.Name == "" {
		return ErrAgentNameRequired
	}
	return nil
}

// ModelValidate validates the ModelSpec
func ModelValidate(m *ModelSpec) error {
	if m == nil {
		return ErrModelRequired
	}
	if m.Provider == "" {
		return ErrModelProviderRequired
	}
	// Validate provider is a known provider
	provider := agentenum.ModelProvider(m.Provider)
	if !provider.IsValid() {
		validProviders := agentenum.AllProviders()
		providerNames := make([]string, len(validProviders))
		for i, p := range validProviders {
			providerNames[i] = string(p)
		}
		return fmt.Errorf("invalid model provider '%s': must be one of %v", m.Provider, providerNames)
	}
	if m.Name == "" {
		return ErrModelNameRequired
	}
	return nil
}

// ReactValidate validates the ReactSpec
func ReactValidate(r *ReactSpec) error {
	if r == nil {
		return ErrReactSpecRequired
	}
	if len(r.Tools) == 0 {
		return ErrReactToolsRequired
	}
	return nil
}

// contains is a helper function to check if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
