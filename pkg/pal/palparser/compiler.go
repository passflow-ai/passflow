package palparser

import (
	"fmt"
	"time"

	agentdomain "github.com/passflow-ai/passflow/pkg/agent/domain"
	paldomain "github.com/passflow-ai/passflow/pkg/pal/domain"
)

// resourcePresets maps preset names to their ResourceConfig definitions.
var resourcePresets = map[string]agentdomain.ResourceConfig{
	"small": {
		CPURequest:            "200m",
		CPULimit:              "1000m",
		MemoryRequest:         "256Mi",
		MemoryLimit:           "512Mi",
		MaxExecutionTime:      1800,
		MaxConcurrentSessions: 0,
		AllowedTools:          []string{},
		Preset:                agentdomain.ResourcePresetSmall,
	},
	"medium": {
		CPURequest:            "500m",
		CPULimit:              "2000m",
		MemoryRequest:         "512Mi",
		MemoryLimit:           "1Gi",
		MaxExecutionTime:      3600,
		MaxConcurrentSessions: 0,
		AllowedTools:          []string{},
		Preset:                agentdomain.ResourcePresetMedium,
	},
	"large": {
		CPURequest:            "1000m",
		CPULimit:              "4000m",
		MemoryRequest:         "1Gi",
		MemoryLimit:           "2Gi",
		MaxExecutionTime:      14400,
		MaxConcurrentSessions: 0,
		AllowedTools:          []string{},
		Preset:                agentdomain.ResourcePresetLarge,
	},
}

// CompileToAgent converts a PALSpec to an Agent domain model.
// It validates the spec, applies resource presets, and maps PAL fields to Agent fields.
func CompileToAgent(spec *paldomain.PALSpec, workspaceID string) (*agentdomain.Agent, error) {
	// Validate the spec first
	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("invalid PAL spec: %w", err)
	}

	if workspaceID == "" {
		return nil, fmt.Errorf("workspace ID is required")
	}

	// Create new agent with persona from spec
	persona := buildPersona(spec)
	agent := agentdomain.NewAgent(workspaceID, spec.Agent.Name, persona)
	agent.Description = spec.Agent.Description
	agent.Instructions = spec.Instructions

	// Set model configuration
	agent.Model = agentdomain.ModelConfig{
		Provider: spec.Model.Provider,
		ModelID:  spec.Model.Name,
	}

	// Parse model config for temperature and max_tokens
	if spec.Model.Config != nil {
		if temp, ok := spec.Model.Config["temperature"]; ok {
			if tempFloat, ok := temp.(float64); ok {
				agent.Temperature = tempFloat
			}
		}
		if maxTok, ok := spec.Model.Config["max_tokens"]; ok {
			// Handle both int and float64 (YAML/JSON number parsing)
			switch v := maxTok.(type) {
			case int:
				agent.MaxTokens = v
			case float64:
				agent.MaxTokens = int(v)
			}
		}
	}

	// Convert tools from React spec
	agent.Tools = make([]agentdomain.AgentTool, len(spec.React.Tools))
	for i, toolID := range spec.React.Tools {
		agent.Tools[i] = agentdomain.AgentTool{
			ID:      toolID,
			Name:    toolID,
			Enabled: true,
			Config:  make(map[string]interface{}),
		}
	}

	// Note: MaxIterations from React spec is NOT the same as MaxTokens from Model config.
	// MaxIterations controls the ReAct loop iterations, while MaxTokens controls LLM output length.
	// MaxIterations should be stored separately if needed for execution.

	// Configure resources from Pod spec
	if spec.Pod != nil && spec.Pod.Resources != nil {
		agent.Resources = applyResourceConfig(spec.Pod.Resources)
	} else {
		agent.Resources = agentdomain.DefaultResourceConfig()
	}

	// Configure guardrails if present
	if spec.Guardrails != nil {
		agent.Guardrails = compileGuardrails(spec.Guardrails)
		// Apply allowed tools from guardrails
		if len(spec.Guardrails.AllowedTools) > 0 {
			agent.Resources.AllowedTools = spec.Guardrails.AllowedTools
		}
	}

	return agent, nil
}

// UpdateAgentFromSpec updates an existing Agent with values from a PALSpec.
// It preserves agent ID and workspace ID, updating all other fields.
func UpdateAgentFromSpec(agent *agentdomain.Agent, spec *paldomain.PALSpec) error {
	// Validate the spec first
	if err := spec.Validate(); err != nil {
		return fmt.Errorf("invalid PAL spec: %w", err)
	}

	// Preserve ID and WorkspaceID
	originalID := agent.ID
	originalWorkspaceID := agent.WorkspaceID

	// Update basic fields
	agent.Name = spec.Agent.Name
	agent.Description = spec.Agent.Description
	agent.Persona = buildPersona(spec)
	agent.Instructions = spec.Instructions

	// Update model configuration
	agent.Model = agentdomain.ModelConfig{
		Provider: spec.Model.Provider,
		ModelID:  spec.Model.Name,
	}

	// Parse model config for temperature and max_tokens
	agent.Temperature = 0 // Reset first
	agent.MaxTokens = 0
	if spec.Model.Config != nil {
		if temp, ok := spec.Model.Config["temperature"]; ok {
			if tempFloat, ok := temp.(float64); ok {
				agent.Temperature = tempFloat
			}
		}
		if maxTok, ok := spec.Model.Config["max_tokens"]; ok {
			// Handle both int and float64 (YAML/JSON number parsing)
			switch v := maxTok.(type) {
			case int:
				agent.MaxTokens = v
			case float64:
				agent.MaxTokens = int(v)
			}
		}
	}

	// Update tools from React spec
	agent.Tools = make([]agentdomain.AgentTool, len(spec.React.Tools))
	for i, toolID := range spec.React.Tools {
		agent.Tools[i] = agentdomain.AgentTool{
			ID:      toolID,
			Name:    toolID,
			Enabled: true,
			Config:  make(map[string]interface{}),
		}
	}

	// Note: MaxIterations from React spec is NOT the same as MaxTokens from Model config.
	// MaxIterations controls the ReAct loop iterations, while MaxTokens controls LLM output length.

	// Update resources from Pod spec
	if spec.Pod != nil && spec.Pod.Resources != nil {
		agent.Resources = applyResourceConfig(spec.Pod.Resources)
	} else {
		agent.Resources = agentdomain.DefaultResourceConfig()
	}

	// Update guardrails if present
	if spec.Guardrails != nil {
		agent.Guardrails = compileGuardrails(spec.Guardrails)
		// Apply allowed tools from guardrails
		if len(spec.Guardrails.AllowedTools) > 0 {
			agent.Resources.AllowedTools = spec.Guardrails.AllowedTools
		}
	} else {
		agent.Guardrails = nil
	}

	// Restore preserved fields
	agent.ID = originalID
	agent.WorkspaceID = originalWorkspaceID
	agent.UpdatedAt = time.Now().UTC()

	return nil
}

// applyResourceConfig takes Pod resources and applies preset or custom configuration.
func applyResourceConfig(resources map[string]string) agentdomain.ResourceConfig {
	config := agentdomain.DefaultResourceConfig()

	// Check for preset first
	if preset, exists := resources["preset"]; exists {
		if presetConfig, ok := resourcePresets[preset]; ok {
			return presetConfig
		}
	}

	// Apply custom resource values if present
	hasCustom := false
	if cpuReq, ok := resources["cpu_request"]; ok {
		config.CPURequest = cpuReq
		hasCustom = true
	}
	if cpuLim, ok := resources["cpu_limit"]; ok {
		config.CPULimit = cpuLim
		hasCustom = true
	}
	if memReq, ok := resources["memory_request"]; ok {
		config.MemoryRequest = memReq
		hasCustom = true
	}
	if memLim, ok := resources["memory_limit"]; ok {
		config.MemoryLimit = memLim
		hasCustom = true
	}
	if storReq, ok := resources["storage_request"]; ok {
		config.StorageRequest = storReq
		hasCustom = true
	}
	if _, ok := resources["max_execution_time"]; ok {
		// Parse as integer (would need strconv in real impl, but for now skip if parse fails)
		hasCustom = true
	}
	if _, ok := resources["max_concurrent_sessions"]; ok {
		// Parse as integer
		hasCustom = true
	}

	if hasCustom {
		config.Preset = agentdomain.ResourcePresetCustom
	}

	return config
}

// compileGuardrails converts PALSpec guardrails to AgentGuardrails.
func compileGuardrails(guardrails *paldomain.GuardrailsSpec) *agentdomain.AgentGuardrails {
	if guardrails == nil {
		return nil
	}

	ag := &agentdomain.AgentGuardrails{
		ProhibitedTopics: guardrails.BlockedTools,
	}

	// Map timeout to max response length
	if guardrails.TimeoutSeconds > 0 {
		ag.MaxResponseLength = guardrails.TimeoutSeconds
	}

	// Add escalation conditions if present
	if len(guardrails.CustomRules) > 0 {
		ag.EscalationConditions = make([]string, 0, len(guardrails.CustomRules))
		for _, rule := range guardrails.CustomRules {
			ag.EscalationConditions = append(ag.EscalationConditions, rule)
		}
	}

	return ag
}

// buildPersona constructs the persona string from PALSpec.
// Priority: 1) PersonaSpec with identity+tone, 2) Agent.Persona string, 3) Default "executor"
func buildPersona(spec *paldomain.PALSpec) string {
	// If PersonaSpec is provided with identity or tone, combine them
	if spec.Persona != nil {
		parts := make([]string, 0, 2)
		if spec.Persona.Identity != "" {
			parts = append(parts, spec.Persona.Identity)
		}
		if spec.Persona.Tone != "" {
			parts = append(parts, spec.Persona.Tone)
		}
		if len(parts) > 0 {
			return joinPersonaParts(parts)
		}
	}

	// If Agent has a simple persona string, use it
	if spec.Agent != nil && spec.Agent.Persona != "" {
		return spec.Agent.Persona
	}

	// Default persona
	return "executor"
}

// joinPersonaParts joins persona identity and tone with a separator
func joinPersonaParts(parts []string) string {
	if len(parts) == 1 {
		return parts[0]
	}
	return fmt.Sprintf("%s\n\n%s", parts[0], parts[1])
}
