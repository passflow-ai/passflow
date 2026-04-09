package palparser

import (
	"encoding/json"
	"fmt"
	"strconv"

	agentdomain "github.com/passflow-ai/passflow/pkg/agent/domain"
	paldomain "github.com/passflow-ai/passflow/pkg/pal/domain"
	"gopkg.in/yaml.v3"
)

// ExportToSpec converts an Agent domain model back to a PALSpec.
// This is the reverse operation of CompileToAgent, enabling export/backup of agents.
func ExportToSpec(agent *agentdomain.Agent) (*paldomain.PALSpec, error) {
	if agent == nil {
		return nil, fmt.Errorf("agent is required")
	}

	spec := &paldomain.PALSpec{
		Agent: &paldomain.AgentSpec{
			Name:        agent.Name,
			Description: agent.Description,
		},
		Model: &paldomain.ModelSpec{
			Provider: agent.Model.Provider,
			Name:     agent.Model.ModelID,
			Config:   make(map[string]interface{}),
		},
		React: &paldomain.ReactSpec{
			Tools:         extractToolIDs(agent.Tools),
			MaxIterations: agent.MaxTokens,
		},
		Pod: exportResourceConfig(agent.Resources),
	}

	// Add model configuration if temperature or max_tokens are set
	if agent.Temperature > 0 {
		spec.Model.Config["temperature"] = agent.Temperature
	}
	if agent.MaxTokens > 0 {
		spec.Model.Config["max_tokens"] = agent.MaxTokens
	}

	// Export guardrails if present or if there are allowed tools
	if agent.Guardrails != nil || len(agent.Resources.AllowedTools) > 0 {
		spec.Guardrails = exportGuardrails(agent.Guardrails, agent.Resources.AllowedTools)
	}

	return spec, nil
}

// ExportToYAML converts an Agent to PALSpec and serializes it to YAML bytes.
func ExportToYAML(agent *agentdomain.Agent) ([]byte, error) {
	spec, err := ExportToSpec(agent)
	if err != nil {
		return nil, err
	}

	yamlBytes, err := yaml.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal spec to YAML: %w", err)
	}

	return yamlBytes, nil
}

// ExportToJSON converts an Agent to PALSpec and serializes it to JSON bytes.
func ExportToJSON(agent *agentdomain.Agent) ([]byte, error) {
	spec, err := ExportToSpec(agent)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal spec to JSON: %w", err)
	}

	return jsonBytes, nil
}

// extractToolIDs extracts tool IDs from a slice of AgentTools.
func extractToolIDs(tools []agentdomain.AgentTool) []string {
	if len(tools) == 0 {
		return []string{}
	}

	ids := make([]string, 0, len(tools))
	for _, tool := range tools {
		if tool.Enabled {
			ids = append(ids, tool.ID)
		}
	}
	return ids
}

// exportResourceConfig converts ResourceConfig back to Pod spec resources map.
// Attempts to match against known presets first, otherwise exports as custom resources.
func exportResourceConfig(config agentdomain.ResourceConfig) *paldomain.PodSpec {
	resources := make(map[string]string)

	// Try to match against known presets
	if matchesPreset(&config, agentdomain.ResourcePresetSmall) {
		resources["preset"] = "small"
		// Also include the individual fields for clarity
		resources["cpu_request"] = config.CPURequest
		resources["cpu_limit"] = config.CPULimit
		resources["memory_request"] = config.MemoryRequest
		resources["memory_limit"] = config.MemoryLimit
	} else if matchesPreset(&config, agentdomain.ResourcePresetMedium) {
		resources["preset"] = "medium"
		resources["cpu_request"] = config.CPURequest
		resources["cpu_limit"] = config.CPULimit
		resources["memory_request"] = config.MemoryRequest
		resources["memory_limit"] = config.MemoryLimit
	} else if matchesPreset(&config, agentdomain.ResourcePresetLarge) {
		resources["preset"] = "large"
		resources["cpu_request"] = config.CPURequest
		resources["cpu_limit"] = config.CPULimit
		resources["memory_request"] = config.MemoryRequest
		resources["memory_limit"] = config.MemoryLimit
	} else {
		// Export as custom resources
		resources["cpu_request"] = config.CPURequest
		resources["cpu_limit"] = config.CPULimit
		resources["memory_request"] = config.MemoryRequest
		resources["memory_limit"] = config.MemoryLimit

		if config.StorageRequest != "" {
			resources["storage_request"] = config.StorageRequest
		}

		if config.MaxExecutionTime > 0 {
			resources["max_execution_time"] = strconv.Itoa(config.MaxExecutionTime)
		}

		if config.MaxConcurrentSessions > 0 {
			resources["max_concurrent_sessions"] = strconv.Itoa(config.MaxConcurrentSessions)
		}
	}

	return &paldomain.PodSpec{
		Resources: resources,
	}
}

// matchesPreset checks if a ResourceConfig matches a given preset.
func matchesPreset(config *agentdomain.ResourceConfig, preset agentdomain.ResourcePreset) bool {
	presetConfig, ok := agentdomain.PresetConfigs[preset]
	if !ok {
		return false
	}

	// Compare key resource fields
	return config.CPURequest == presetConfig.CPURequest &&
		config.CPULimit == presetConfig.CPULimit &&
		config.MemoryRequest == presetConfig.MemoryRequest &&
		config.MemoryLimit == presetConfig.MemoryLimit &&
		config.MaxExecutionTime == presetConfig.MaxExecutionTime
}

// exportGuardrails converts AgentGuardrails and allowed tools back to GuardrailsSpec.
func exportGuardrails(guardrails *agentdomain.AgentGuardrails, allowedTools []string) *paldomain.GuardrailsSpec {
	spec := &paldomain.GuardrailsSpec{
		AllowedTools:   allowedTools,
		BlockedTools:   []string{},
		RateLimits:     make(map[string]int),
		CustomRules:    make(map[string]string),
	}

	if guardrails != nil {
		spec.BlockedTools = guardrails.ProhibitedTopics
		spec.TimeoutSeconds = guardrails.MaxResponseLength

		// Map escalation conditions to custom rules
		if len(guardrails.EscalationConditions) > 0 {
			for i, condition := range guardrails.EscalationConditions {
				key := fmt.Sprintf("escalation_%d", i+1)
				spec.CustomRules[key] = condition
			}
		}
	}

	return spec
}
