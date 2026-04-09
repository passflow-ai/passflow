package palparser

import (
	"testing"

	agentdomain "github.com/jaak-ai/passflow-api/src/usecase/agent/domain"
	"github.com/jaak-ai/passflow-api/src/usecase/pal/domain"
)

func TestCompileToAgent_BasicFields(t *testing.T) {
	spec := &domain.PALSpec{
		Agent: &domain.AgentSpec{
			Name:        "test-agent",
			Description: "Test description",
		},
		Model: &domain.ModelSpec{
			Provider: "anthropic",
			Name:     "claude-3-opus",
			Config: map[string]interface{}{
				"temperature": 0.7,
			},
		},
		React: &domain.ReactSpec{
			Tools:         []string{"tool1", "tool2"},
			MaxIterations: 10,
		},
	}

	agent, err := CompileToAgent(spec, "workspace-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if agent == nil {
		t.Fatal("expected agent to be non-nil")
	}

	if agent.Name != "test-agent" {
		t.Errorf("expected agent name 'test-agent', got %q", agent.Name)
	}

	if agent.Description != "Test description" {
		t.Errorf("expected description 'Test description', got %q", agent.Description)
	}

	if agent.WorkspaceID != "workspace-123" {
		t.Errorf("expected workspace ID 'workspace-123', got %q", agent.WorkspaceID)
	}

	if agent.Model.Provider != "anthropic" {
		t.Errorf("expected model provider 'anthropic', got %q", agent.Model.Provider)
	}

	if agent.Model.ModelID != "claude-3-opus" {
		t.Errorf("expected model ID 'claude-3-opus', got %q", agent.Model.ModelID)
	}

	if len(agent.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(agent.Tools))
	}

	if agent.Tools[0].ID != "tool1" || agent.Tools[0].Name != "tool1" {
		t.Errorf("expected first tool to be 'tool1', got ID=%q, Name=%q", agent.Tools[0].ID, agent.Tools[0].Name)
	}

	if agent.Tools[1].ID != "tool2" || agent.Tools[1].Name != "tool2" {
		t.Errorf("expected second tool to be 'tool2', got ID=%q, Name=%q", agent.Tools[1].ID, agent.Tools[1].Name)
	}

	if agent.Temperature != 0.7 {
		t.Errorf("expected temperature 0.7, got %f", agent.Temperature)
	}

	if agent.Persona == "" {
		t.Error("expected persona to be set to default")
	}
}

func TestCompileToAgent_ResourcePresets(t *testing.T) {
	tests := []struct {
		name           string
		resources      map[string]string
		expectedPreset agentdomain.ResourcePreset
		expectedCPU    string
		expectedMem    string
	}{
		{
			name:           "small preset",
			resources:      map[string]string{"preset": "small"},
			expectedPreset: agentdomain.ResourcePresetSmall,
			expectedCPU:    "200m",
			expectedMem:    "256Mi",
		},
		{
			name:           "medium preset",
			resources:      map[string]string{"preset": "medium"},
			expectedPreset: agentdomain.ResourcePresetMedium,
			expectedCPU:    "500m",
			expectedMem:    "512Mi",
		},
		{
			name:           "large preset",
			resources:      map[string]string{"preset": "large"},
			expectedPreset: agentdomain.ResourcePresetLarge,
			expectedCPU:    "1000m",
			expectedMem:    "1Gi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &domain.PALSpec{
				Agent: &domain.AgentSpec{
					Name: "test-agent",
				},
				Model: &domain.ModelSpec{
					Provider: "anthropic",
					Name:     "claude-3-opus",
				},
				React: &domain.ReactSpec{
					Tools: []string{"tool1"},
				},
				Pod: &domain.PodSpec{
					Resources: tt.resources,
				},
			}

			agent, err := CompileToAgent(spec, "workspace-123")
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if agent.Resources.Preset != tt.expectedPreset {
				t.Errorf("expected preset %q, got %q", tt.expectedPreset, agent.Resources.Preset)
			}

			if agent.Resources.CPURequest != tt.expectedCPU {
				t.Errorf("expected CPU request %q, got %q", tt.expectedCPU, agent.Resources.CPURequest)
			}

			if agent.Resources.MemoryRequest != tt.expectedMem {
				t.Errorf("expected memory request %q, got %q", tt.expectedMem, agent.Resources.MemoryRequest)
			}
		})
	}
}

func TestCompileToAgent_CustomResources(t *testing.T) {
	spec := &domain.PALSpec{
		Agent: &domain.AgentSpec{
			Name: "test-agent",
		},
		Model: &domain.ModelSpec{
			Provider: "anthropic",
			Name:     "claude-3-opus",
		},
		React: &domain.ReactSpec{
			Tools: []string{"tool1"},
		},
		Pod: &domain.PodSpec{
			Resources: map[string]string{
				"cpu_request":    "750m",
				"cpu_limit":      "2000m",
				"memory_request": "768Mi",
				"memory_limit":   "1536Mi",
			},
		},
	}

	agent, err := CompileToAgent(spec, "workspace-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if agent.Resources.CPURequest != "750m" {
		t.Errorf("expected CPU request '750m', got %q", agent.Resources.CPURequest)
	}

	if agent.Resources.CPULimit != "2000m" {
		t.Errorf("expected CPU limit '2000m', got %q", agent.Resources.CPULimit)
	}

	if agent.Resources.MemoryRequest != "768Mi" {
		t.Errorf("expected memory request '768Mi', got %q", agent.Resources.MemoryRequest)
	}

	if agent.Resources.MemoryLimit != "1536Mi" {
		t.Errorf("expected memory limit '1536Mi', got %q", agent.Resources.MemoryLimit)
	}

	if agent.Resources.Preset != agentdomain.ResourcePresetCustom {
		t.Errorf("expected preset 'custom', got %q", agent.Resources.Preset)
	}
}

func TestCompileToAgent_Guardrails(t *testing.T) {
	spec := &domain.PALSpec{
		Agent: &domain.AgentSpec{
			Name: "test-agent",
		},
		Model: &domain.ModelSpec{
			Provider: "anthropic",
			Name:     "claude-3-opus",
		},
		React: &domain.ReactSpec{
			Tools: []string{"tool1", "tool2"},
		},
		Guardrails: &domain.GuardrailsSpec{
			AllowedTools:   []string{"tool1"},
			BlockedTools:   []string{"tool3"},
			TimeoutSeconds: 300,
			RateLimits: map[string]int{
				"requests_per_minute": 60,
			},
		},
	}

	agent, err := CompileToAgent(spec, "workspace-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if agent.Guardrails == nil {
		t.Fatal("expected guardrails to be set")
	}

	if len(agent.Resources.AllowedTools) != 1 || agent.Resources.AllowedTools[0] != "tool1" {
		t.Errorf("expected allowed tools to contain 'tool1'")
	}

	if agent.Guardrails.MaxResponseLength != 300 {
		t.Errorf("expected max response length 300, got %d", agent.Guardrails.MaxResponseLength)
	}

	if len(agent.Guardrails.ProhibitedTopics) != 1 || agent.Guardrails.ProhibitedTopics[0] != "tool3" {
		t.Errorf("expected prohibited topics to contain 'tool3'")
	}
}

func TestCompileToAgent_WithoutGuardrails(t *testing.T) {
	spec := &domain.PALSpec{
		Agent: &domain.AgentSpec{
			Name: "test-agent",
		},
		Model: &domain.ModelSpec{
			Provider: "anthropic",
			Name:     "claude-3-opus",
		},
		React: &domain.ReactSpec{
			Tools: []string{"tool1"},
		},
	}

	agent, err := CompileToAgent(spec, "workspace-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if agent.Guardrails != nil {
		t.Errorf("expected guardrails to be nil when not specified")
	}
}

func TestUpdateAgentFromSpec_UpdatesFields(t *testing.T) {
	agent := agentdomain.NewAgent("workspace-123", "old-name", "persona")
	agent.ID = "agent-123"

	spec := &domain.PALSpec{
		Agent: &domain.AgentSpec{
			Name:        "updated-agent",
			Description: "Updated description",
		},
		Model: &domain.ModelSpec{
			Provider: "openai",
			Name:     "gpt-4o",
		},
		React: &domain.ReactSpec{
			Tools: []string{"tool1", "tool2"},
		},
		Pod: &domain.PodSpec{
			Resources: map[string]string{
				"preset": "medium",
			},
		},
	}

	err := UpdateAgentFromSpec(agent, spec)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if agent.Name != "updated-agent" {
		t.Errorf("expected updated name 'updated-agent', got %q", agent.Name)
	}

	if agent.Description != "Updated description" {
		t.Errorf("expected updated description, got %q", agent.Description)
	}

	if agent.Model.Provider != "openai" {
		t.Errorf("expected updated provider 'openai', got %q", agent.Model.Provider)
	}

	if agent.Model.ModelID != "gpt-4o" {
		t.Errorf("expected updated model ID 'gpt-4o', got %q", agent.Model.ModelID)
	}

	if len(agent.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(agent.Tools))
	}

	if agent.Resources.Preset != agentdomain.ResourcePresetMedium {
		t.Errorf("expected medium preset, got %q", agent.Resources.Preset)
	}

	if agent.ID != "agent-123" {
		t.Errorf("expected ID to remain 'agent-123', got %q", agent.ID)
	}
}

func TestUpdateAgentFromSpec_PreservesWorkspaceID(t *testing.T) {
	agent := agentdomain.NewAgent("original-workspace", "agent-name", "persona")

	spec := &domain.PALSpec{
		Agent: &domain.AgentSpec{
			Name: "agent-name",
		},
		Model: &domain.ModelSpec{
			Provider: "anthropic",
			Name:     "claude-3-opus",
		},
		React: &domain.ReactSpec{
			Tools: []string{"tool1"},
		},
	}

	err := UpdateAgentFromSpec(agent, spec)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if agent.WorkspaceID != "original-workspace" {
		t.Errorf("expected workspace ID to be preserved, got %q", agent.WorkspaceID)
	}
}

func TestCompileToAgent_MaxIterationsDoesNotOverwriteMaxTokens(t *testing.T) {
	// MaxIterations (react loop iterations) should NOT overwrite MaxTokens (LLM output limit)
	// They are two different concepts:
	// - MaxIterations: how many tool-use cycles the ReAct loop can perform
	// - MaxTokens: maximum tokens in the LLM response
	spec := &domain.PALSpec{
		Agent: &domain.AgentSpec{
			Name: "test-agent",
		},
		Model: &domain.ModelSpec{
			Provider: "anthropic",
			Name:     "claude-3-opus",
			Config: map[string]interface{}{
				"max_tokens": 4096, // This is the LLM max tokens
			},
		},
		React: &domain.ReactSpec{
			Tools:         []string{"tool1"},
			MaxIterations: 25, // This is the react loop iterations, NOT max_tokens
		},
	}

	agent, err := CompileToAgent(spec, "workspace-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// MaxTokens should be 4096 from model config, NOT 25 from MaxIterations
	if agent.MaxTokens != 4096 {
		t.Errorf("expected maxTokens to be 4096 from model config, got %d", agent.MaxTokens)
	}
}

func TestCompileToAgent_InvalidSpec(t *testing.T) {
	spec := &domain.PALSpec{
		Agent: &domain.AgentSpec{
			Name: "", // Invalid: empty name
		},
		Model: &domain.ModelSpec{
			Provider: "anthropic",
			Name:     "claude-3-opus",
		},
		React: &domain.ReactSpec{
			Tools: []string{"tool1"},
		},
	}

	agent, err := CompileToAgent(spec, "workspace-123")
	if err == nil {
		t.Fatal("expected error for invalid spec, got nil")
	}

	if agent != nil {
		t.Errorf("expected agent to be nil on error, got %v", agent)
	}
}

func TestCompileToAgent_DefaultPersona(t *testing.T) {
	spec := &domain.PALSpec{
		Agent: &domain.AgentSpec{
			Name: "test-agent",
		},
		Model: &domain.ModelSpec{
			Provider: "anthropic",
			Name:     "claude-3-opus",
		},
		React: &domain.ReactSpec{
			Tools: []string{"tool1"},
		},
	}

	agent, err := CompileToAgent(spec, "workspace-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if agent.Persona != "executor" {
		t.Errorf("expected default persona 'executor', got %q", agent.Persona)
	}
}

func TestCompileToAgent_ToolsEnabled(t *testing.T) {
	spec := &domain.PALSpec{
		Agent: &domain.AgentSpec{
			Name: "test-agent",
		},
		Model: &domain.ModelSpec{
			Provider: "anthropic",
			Name:     "claude-3-opus",
		},
		React: &domain.ReactSpec{
			Tools: []string{"tool1", "tool2"},
		},
	}

	agent, err := CompileToAgent(spec, "workspace-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for i, tool := range agent.Tools {
		if !tool.Enabled {
			t.Errorf("tool %d expected to be enabled, got enabled=%v", i, tool.Enabled)
		}
	}
}

func TestCompileToAgent_PersonaSpec(t *testing.T) {
	spec := &domain.PALSpec{
		Agent: &domain.AgentSpec{
			Name: "test-agent",
		},
		Model: &domain.ModelSpec{
			Provider: "anthropic",
			Name:     "claude-3-opus",
		},
		Persona: &domain.PersonaSpec{
			Identity: "A senior SRE on-call engineer",
			Tone:     "Technical and direct. Speaks with data.",
		},
		React: &domain.ReactSpec{
			Tools: []string{"tool1"},
		},
	}

	agent, err := CompileToAgent(spec, "workspace-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedPersona := "A senior SRE on-call engineer\n\nTechnical and direct. Speaks with data."
	if agent.Persona != expectedPersona {
		t.Errorf("expected persona to combine identity and tone:\nexpected: %q\ngot: %q", expectedPersona, agent.Persona)
	}
}

func TestCompileToAgent_PersonaFromAgentSpec(t *testing.T) {
	// If only Agent.Persona is set (simple string), use that
	spec := &domain.PALSpec{
		Agent: &domain.AgentSpec{
			Name:    "test-agent",
			Persona: "A helpful assistant",
		},
		Model: &domain.ModelSpec{
			Provider: "anthropic",
			Name:     "claude-3-opus",
		},
		React: &domain.ReactSpec{
			Tools: []string{"tool1"},
		},
	}

	agent, err := CompileToAgent(spec, "workspace-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if agent.Persona != "A helpful assistant" {
		t.Errorf("expected persona from Agent.Persona, got %q", agent.Persona)
	}
}

func TestCompileToAgent_Instructions(t *testing.T) {
	instructions := `You are the coordinator agent of JAAK's SRE system.

OBJECTIVE: Coordinate specialized agents to maintain service availability.

PROHIBITED:
- Taking destructive actions without human approval.`

	spec := &domain.PALSpec{
		Agent: &domain.AgentSpec{
			Name: "test-agent",
		},
		Model: &domain.ModelSpec{
			Provider: "anthropic",
			Name:     "claude-3-opus",
		},
		Instructions: instructions,
		React: &domain.ReactSpec{
			Tools: []string{"tool1"},
		},
	}

	agent, err := CompileToAgent(spec, "workspace-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if agent.Instructions != instructions {
		t.Errorf("expected instructions to be preserved:\nexpected: %q\ngot: %q", instructions, agent.Instructions)
	}
}
