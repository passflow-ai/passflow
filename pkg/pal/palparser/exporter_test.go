package palparser

import (
	"encoding/json"
	"testing"
	"time"

	agentdomain "github.com/passflow-ai/passflow/pkg/agent/domain"
	"github.com/passflow-ai/passflow/pkg/pal/domain"
	"gopkg.in/yaml.v3"
)

func TestExportToSpec_BasicFields(t *testing.T) {
	now := time.Now().UTC()
	agent := &agentdomain.Agent{
		ID:             "agent-123",
		WorkspaceID:    "workspace-123",
		Name:           "test-agent",
		Description:    "Test description",
		Persona:        "executor",
		Model: agentdomain.ModelConfig{
			Provider: "anthropic",
			ModelID:  "claude-3-opus",
		},
		Temperature: 0.7,
		MaxTokens:   10,
		Tools: []agentdomain.AgentTool{
			{ID: "tool1", Name: "tool1", Enabled: true, Config: make(map[string]interface{})},
			{ID: "tool2", Name: "tool2", Enabled: true, Config: make(map[string]interface{})},
		},
		Resources: agentdomain.ResourceConfig{
			CPURequest:            "200m",
			CPULimit:              "1000m",
			MemoryRequest:         "256Mi",
			MemoryLimit:           "512Mi",
			MaxExecutionTime:      1800,
			MaxConcurrentSessions: 0,
			AllowedTools:          []string{},
			Preset:                agentdomain.ResourcePresetSmall,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	spec, err := ExportToSpec(agent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if spec == nil {
		t.Fatal("expected spec to be non-nil")
	}

	if spec.Agent.Name != "test-agent" {
		t.Errorf("expected agent name 'test-agent', got %q", spec.Agent.Name)
	}

	if spec.Agent.Description != "Test description" {
		t.Errorf("expected description 'Test description', got %q", spec.Agent.Description)
	}

	if spec.Model.Provider != "anthropic" {
		t.Errorf("expected model provider 'anthropic', got %q", spec.Model.Provider)
	}

	if spec.Model.Name != "claude-3-opus" {
		t.Errorf("expected model name 'claude-3-opus', got %q", spec.Model.Name)
	}

	if spec.Model.Config["temperature"] != 0.7 {
		t.Errorf("expected temperature 0.7, got %v", spec.Model.Config["temperature"])
	}

	if len(spec.React.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(spec.React.Tools))
	}

	if spec.React.Tools[0] != "tool1" || spec.React.Tools[1] != "tool2" {
		t.Errorf("expected tools [tool1, tool2], got %v", spec.React.Tools)
	}

	if spec.React.MaxIterations != 10 {
		t.Errorf("expected max iterations 10, got %d", spec.React.MaxIterations)
	}

	if spec.Pod == nil {
		t.Fatal("expected Pod spec to be non-nil")
	}

	if spec.Pod.Resources["preset"] != "small" {
		t.Errorf("expected preset 'small', got %q", spec.Pod.Resources["preset"])
	}
}

func TestExportToSpec_GuardrailsConversion(t *testing.T) {
	now := time.Now().UTC()
	agent := &agentdomain.Agent{
		ID:          "agent-123",
		WorkspaceID: "workspace-123",
		Name:        "test-agent",
		Persona:     "executor",
		Model: agentdomain.ModelConfig{
			Provider: "anthropic",
			ModelID:  "claude-3-opus",
		},
		Tools: []agentdomain.AgentTool{
			{ID: "tool1", Name: "tool1", Enabled: true, Config: make(map[string]interface{})},
		},
		Guardrails: &agentdomain.AgentGuardrails{
			ProhibitedTopics:     []string{"tool-blocked"},
			MaxResponseLength:    300,
			EscalationConditions: []string{"rule1", "rule2"},
		},
		Resources: agentdomain.DefaultResourceConfig(),
		CreatedAt: now,
		UpdatedAt: now,
	}

	spec, err := ExportToSpec(agent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if spec.Guardrails == nil {
		t.Fatal("expected guardrails to be non-nil")
	}

	if len(spec.Guardrails.BlockedTools) != 1 || spec.Guardrails.BlockedTools[0] != "tool-blocked" {
		t.Errorf("expected blocked tools [tool-blocked], got %v", spec.Guardrails.BlockedTools)
	}

	if spec.Guardrails.TimeoutSeconds != 300 {
		t.Errorf("expected timeout 300, got %d", spec.Guardrails.TimeoutSeconds)
	}

	if len(spec.Guardrails.CustomRules) != 2 {
		t.Errorf("expected 2 custom rules, got %d", len(spec.Guardrails.CustomRules))
	}
}

func TestExportToSpec_ResourcePresets(t *testing.T) {
	tests := []struct {
		name            string
		preset          agentdomain.ResourcePreset
		expectedPreset  string
		expectedCPU     string
		expectedMemory  string
	}{
		{
			name:           "small preset",
			preset:         agentdomain.ResourcePresetSmall,
			expectedPreset: "small",
			expectedCPU:    "200m",
			expectedMemory: "256Mi",
		},
		{
			name:           "medium preset",
			preset:         agentdomain.ResourcePresetMedium,
			expectedPreset: "medium",
			expectedCPU:    "500m",
			expectedMemory: "512Mi",
		},
		{
			name:           "large preset",
			preset:         agentdomain.ResourcePresetLarge,
			expectedPreset: "large",
			expectedCPU:    "1000m",
			expectedMemory: "1Gi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now().UTC()
			agent := &agentdomain.Agent{
				ID:          "agent-123",
				WorkspaceID: "workspace-123",
				Name:        "test-agent",
				Persona:     "executor",
				Model: agentdomain.ModelConfig{
					Provider: "anthropic",
					ModelID:  "claude-3-opus",
				},
				Tools: []agentdomain.AgentTool{
					{ID: "tool1", Name: "tool1", Enabled: true, Config: make(map[string]interface{})},
				},
				Resources: agentdomain.PresetConfigs[tt.preset],
				CreatedAt: now,
				UpdatedAt: now,
			}

			spec, err := ExportToSpec(agent)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if spec.Pod == nil {
				t.Fatal("expected Pod spec to be non-nil")
			}

			if spec.Pod.Resources["preset"] != tt.expectedPreset {
				t.Errorf("expected preset %q, got %q", tt.expectedPreset, spec.Pod.Resources["preset"])
			}

			if spec.Pod.Resources["cpu_request"] != tt.expectedCPU {
				t.Errorf("expected CPU request %q, got %q", tt.expectedCPU, spec.Pod.Resources["cpu_request"])
			}

			if spec.Pod.Resources["memory_request"] != tt.expectedMemory {
				t.Errorf("expected memory request %q, got %q", tt.expectedMemory, spec.Pod.Resources["memory_request"])
			}
		})
	}
}

func TestExportToSpec_CustomResources(t *testing.T) {
	now := time.Now().UTC()
	agent := &agentdomain.Agent{
		ID:          "agent-123",
		WorkspaceID: "workspace-123",
		Name:        "test-agent",
		Persona:     "executor",
		Model: agentdomain.ModelConfig{
			Provider: "anthropic",
			ModelID:  "claude-3-opus",
		},
		Tools: []agentdomain.AgentTool{
			{ID: "tool1", Name: "tool1", Enabled: true, Config: make(map[string]interface{})},
		},
		Resources: agentdomain.ResourceConfig{
			CPURequest:            "750m",
			CPULimit:              "2000m",
			MemoryRequest:         "768Mi",
			MemoryLimit:           "1536Mi",
			MaxExecutionTime:      7200,
			MaxConcurrentSessions: 5,
			AllowedTools:          []string{},
			Preset:                agentdomain.ResourcePresetCustom,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	spec, err := ExportToSpec(agent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if spec.Pod == nil {
		t.Fatal("expected Pod spec to be non-nil")
	}

	if spec.Pod.Resources["cpu_request"] != "750m" {
		t.Errorf("expected CPU request '750m', got %q", spec.Pod.Resources["cpu_request"])
	}

	if spec.Pod.Resources["cpu_limit"] != "2000m" {
		t.Errorf("expected CPU limit '2000m', got %q", spec.Pod.Resources["cpu_limit"])
	}

	if spec.Pod.Resources["memory_request"] != "768Mi" {
		t.Errorf("expected memory request '768Mi', got %q", spec.Pod.Resources["memory_request"])
	}

	if spec.Pod.Resources["memory_limit"] != "1536Mi" {
		t.Errorf("expected memory limit '1536Mi', got %q", spec.Pod.Resources["memory_limit"])
	}

	if spec.Pod.Resources["max_execution_time"] != "7200" {
		t.Errorf("expected max_execution_time '7200', got %q", spec.Pod.Resources["max_execution_time"])
	}

	if spec.Pod.Resources["max_concurrent_sessions"] != "5" {
		t.Errorf("expected max_concurrent_sessions '5', got %q", spec.Pod.Resources["max_concurrent_sessions"])
	}
}

func TestExportToSpec_NoGuardrails(t *testing.T) {
	now := time.Now().UTC()
	agent := &agentdomain.Agent{
		ID:          "agent-123",
		WorkspaceID: "workspace-123",
		Name:        "test-agent",
		Persona:     "executor",
		Model: agentdomain.ModelConfig{
			Provider: "anthropic",
			ModelID:  "claude-3-opus",
		},
		Tools: []agentdomain.AgentTool{
			{ID: "tool1", Name: "tool1", Enabled: true, Config: make(map[string]interface{})},
		},
		Guardrails: nil,
		Resources:  agentdomain.DefaultResourceConfig(),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	spec, err := ExportToSpec(agent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if spec.Guardrails != nil {
		t.Errorf("expected guardrails to be nil when not set, got %v", spec.Guardrails)
	}
}

func TestExportToYAML_ValidOutput(t *testing.T) {
	now := time.Now().UTC()
	agent := &agentdomain.Agent{
		ID:          "agent-123",
		WorkspaceID: "workspace-123",
		Name:        "test-agent",
		Description: "Test description",
		Persona:     "executor",
		Model: agentdomain.ModelConfig{
			Provider: "anthropic",
			ModelID:  "claude-3-opus",
		},
		Temperature: 0.7,
		MaxTokens:   10,
		Tools: []agentdomain.AgentTool{
			{ID: "tool1", Name: "tool1", Enabled: true, Config: make(map[string]interface{})},
		},
		Resources: agentdomain.ResourceConfig{
			CPURequest:            "200m",
			CPULimit:              "1000m",
			MemoryRequest:         "256Mi",
			MemoryLimit:           "512Mi",
			MaxExecutionTime:      1800,
			MaxConcurrentSessions: 0,
			AllowedTools:          []string{},
			Preset:                agentdomain.ResourcePresetSmall,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	yamlBytes, err := ExportToYAML(agent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(yamlBytes) == 0 {
		t.Fatal("expected non-empty YAML output")
	}

	// Validate YAML by unmarshalling
	var spec domain.PALSpec
	err = yaml.Unmarshal(yamlBytes, &spec)
	if err != nil {
		t.Fatalf("expected valid YAML, got error: %v", err)
	}

	if spec.Agent.Name != "test-agent" {
		t.Errorf("expected agent name 'test-agent', got %q", spec.Agent.Name)
	}

	if spec.Model.Provider != "anthropic" {
		t.Errorf("expected model provider 'anthropic', got %q", spec.Model.Provider)
	}
}

func TestExportToJSON_ValidOutput(t *testing.T) {
	now := time.Now().UTC()
	agent := &agentdomain.Agent{
		ID:          "agent-123",
		WorkspaceID: "workspace-123",
		Name:        "test-agent",
		Description: "Test description",
		Persona:     "executor",
		Model: agentdomain.ModelConfig{
			Provider: "anthropic",
			ModelID:  "claude-3-opus",
		},
		Temperature: 0.7,
		MaxTokens:   10,
		Tools: []agentdomain.AgentTool{
			{ID: "tool1", Name: "tool1", Enabled: true, Config: make(map[string]interface{})},
		},
		Resources: agentdomain.ResourceConfig{
			CPURequest:            "200m",
			CPULimit:              "1000m",
			MemoryRequest:         "256Mi",
			MemoryLimit:           "512Mi",
			MaxExecutionTime:      1800,
			MaxConcurrentSessions: 0,
			AllowedTools:          []string{},
			Preset:                agentdomain.ResourcePresetSmall,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	jsonBytes, err := ExportToJSON(agent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(jsonBytes) == 0 {
		t.Fatal("expected non-empty JSON output")
	}

	// Validate JSON by unmarshalling
	var spec domain.PALSpec
	err = json.Unmarshal(jsonBytes, &spec)
	if err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	if spec.Agent.Name != "test-agent" {
		t.Errorf("expected agent name 'test-agent', got %q", spec.Agent.Name)
	}

	if spec.Model.Provider != "anthropic" {
		t.Errorf("expected model provider 'anthropic', got %q", spec.Model.Provider)
	}
}

func TestExportToSpec_WithAllowedTools(t *testing.T) {
	now := time.Now().UTC()
	agent := &agentdomain.Agent{
		ID:          "agent-123",
		WorkspaceID: "workspace-123",
		Name:        "test-agent",
		Persona:     "executor",
		Model: agentdomain.ModelConfig{
			Provider: "anthropic",
			ModelID:  "claude-3-opus",
		},
		Tools: []agentdomain.AgentTool{
			{ID: "tool1", Name: "tool1", Enabled: true, Config: make(map[string]interface{})},
			{ID: "tool2", Name: "tool2", Enabled: true, Config: make(map[string]interface{})},
			{ID: "tool3", Name: "tool3", Enabled: true, Config: make(map[string]interface{})},
		},
		Resources: agentdomain.ResourceConfig{
			CPURequest:            "200m",
			CPULimit:              "1000m",
			MemoryRequest:         "256Mi",
			MemoryLimit:           "512Mi",
			MaxExecutionTime:      1800,
			MaxConcurrentSessions: 0,
			AllowedTools:          []string{"tool1", "tool2"},
			Preset:                agentdomain.ResourcePresetSmall,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	spec, err := ExportToSpec(agent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if spec.Guardrails == nil {
		t.Fatal("expected guardrails to be created when allowed tools exist")
	}

	if len(spec.Guardrails.AllowedTools) != 2 {
		t.Errorf("expected 2 allowed tools, got %d", len(spec.Guardrails.AllowedTools))
	}

	if spec.Guardrails.AllowedTools[0] != "tool1" || spec.Guardrails.AllowedTools[1] != "tool2" {
		t.Errorf("expected allowed tools [tool1, tool2], got %v", spec.Guardrails.AllowedTools)
	}
}

func TestExportToSpec_RoundTrip(t *testing.T) {
	// Create an agent
	now := time.Now().UTC()
	originalAgent := &agentdomain.Agent{
		ID:          "agent-123",
		WorkspaceID: "workspace-123",
		Name:        "test-agent",
		Description: "Test description",
		Persona:     "executor",
		Model: agentdomain.ModelConfig{
			Provider: "anthropic",
			ModelID:  "claude-3-opus",
		},
		Temperature: 0.8,
		MaxTokens:   15,
		Tools: []agentdomain.AgentTool{
			{ID: "tool1", Name: "tool1", Enabled: true, Config: make(map[string]interface{})},
			{ID: "tool2", Name: "tool2", Enabled: true, Config: make(map[string]interface{})},
		},
		Resources: agentdomain.ResourceConfig{
			CPURequest:            "500m",
			CPULimit:              "2000m",
			MemoryRequest:         "512Mi",
			MemoryLimit:           "1Gi",
			MaxExecutionTime:      3600,
			MaxConcurrentSessions: 0,
			AllowedTools:          []string{},
			Preset:                agentdomain.ResourcePresetMedium,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Export to spec
	spec, err := ExportToSpec(originalAgent)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	// Compile back to agent
	newAgent, err := CompileToAgent(spec, "workspace-123")
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	// Verify key fields match
	if newAgent.Name != originalAgent.Name {
		t.Errorf("name mismatch: expected %q, got %q", originalAgent.Name, newAgent.Name)
	}

	if newAgent.Description != originalAgent.Description {
		t.Errorf("description mismatch: expected %q, got %q", originalAgent.Description, newAgent.Description)
	}

	if newAgent.Model.Provider != originalAgent.Model.Provider {
		t.Errorf("provider mismatch: expected %q, got %q", originalAgent.Model.Provider, newAgent.Model.Provider)
	}

	if newAgent.Model.ModelID != originalAgent.Model.ModelID {
		t.Errorf("model ID mismatch: expected %q, got %q", originalAgent.Model.ModelID, newAgent.Model.ModelID)
	}

	if newAgent.Temperature != originalAgent.Temperature {
		t.Errorf("temperature mismatch: expected %f, got %f", originalAgent.Temperature, newAgent.Temperature)
	}

	if len(newAgent.Tools) != len(originalAgent.Tools) {
		t.Errorf("tools count mismatch: expected %d, got %d", len(originalAgent.Tools), len(newAgent.Tools))
	}

	if newAgent.Resources.Preset != originalAgent.Resources.Preset {
		t.Errorf("preset mismatch: expected %q, got %q", originalAgent.Resources.Preset, newAgent.Resources.Preset)
	}
}

func TestExportToSpec_WithModelConfig(t *testing.T) {
	now := time.Now().UTC()
	agent := &agentdomain.Agent{
		ID:          "agent-123",
		WorkspaceID: "workspace-123",
		Name:        "test-agent",
		Persona:     "executor",
		Model: agentdomain.ModelConfig{
			Provider: "openai",
			ModelID:  "gpt-4o",
		},
		Temperature: 0.5,
		MaxTokens:   20,
		Tools: []agentdomain.AgentTool{
			{ID: "tool1", Name: "tool1", Enabled: true, Config: make(map[string]interface{})},
		},
		Resources: agentdomain.DefaultResourceConfig(),
		CreatedAt: now,
		UpdatedAt: now,
	}

	spec, err := ExportToSpec(agent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if spec.Model.Config["temperature"] != 0.5 {
		t.Errorf("expected temperature 0.5 in config, got %v", spec.Model.Config["temperature"])
	}

	if spec.Model.Config["max_tokens"] != 20 {
		t.Errorf("expected max_tokens 20 in config, got %v", spec.Model.Config["max_tokens"])
	}
}

func TestExportToSpec_NilAgent(t *testing.T) {
	spec, err := ExportToSpec(nil)
	if err == nil {
		t.Fatal("expected error for nil agent, got nil")
	}

	if spec != nil {
		t.Errorf("expected spec to be nil on error, got %v", spec)
	}
}
