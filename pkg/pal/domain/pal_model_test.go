package domain

import (
	"testing"
)

func TestPALSpec_Validate_RequiresName(t *testing.T) {
	spec := &PALSpec{
		Agent: &AgentSpec{
			Name: "",
		},
	}

	err := spec.Validate()
	if err == nil {
		t.Fatal("expected error for missing agent name, got nil")
	}
	if err != ErrAgentNameRequired {
		t.Fatalf("expected ErrAgentNameRequired, got %v", err)
	}
}

func TestPALSpec_Validate_RequiresModel(t *testing.T) {
	spec := &PALSpec{
		Agent: &AgentSpec{
			Name: "test-agent",
		},
		Model: nil,
	}

	err := spec.Validate()
	if err == nil {
		t.Fatal("expected error for missing model, got nil")
	}
	if err != ErrModelRequired {
		t.Fatalf("expected ErrModelRequired, got %v", err)
	}
}

func TestPALSpec_Validate_RequiresReactTools(t *testing.T) {
	spec := &PALSpec{
		Agent: &AgentSpec{
			Name: "test-agent",
		},
		Model: &ModelSpec{
			Provider: "anthropic",
			Name:     "claude-3-opus",
		},
		React: nil,
	}

	err := spec.Validate()
	if err == nil {
		t.Fatal("expected error for missing react spec, got nil")
	}
	if err != ErrReactSpecRequired {
		t.Fatalf("expected ErrReactSpecRequired, got %v", err)
	}
}

func TestPALSpec_Validate_ValidSpec(t *testing.T) {
	spec := &PALSpec{
		Agent: &AgentSpec{
			Name: "test-agent",
		},
		Model: &ModelSpec{
			Provider: "anthropic",
			Name:     "claude-3-opus",
		},
		React: &ReactSpec{
			Tools: []string{"tool1", "tool2"},
		},
	}

	err := spec.Validate()
	if err != nil {
		t.Fatalf("expected no error for valid spec, got %v", err)
	}
}

func TestAgentSpec_Validate(t *testing.T) {
	tests := []struct {
		name    string
		agent   *AgentSpec
		wantErr error
	}{
		{
			name:    "empty name",
			agent:   &AgentSpec{Name: ""},
			wantErr: ErrAgentNameRequired,
		},
		{
			name:    "valid name",
			agent:   &AgentSpec{Name: "my-agent"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.agent.Validate()
			if err != tt.wantErr {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestModelSpec_Validate(t *testing.T) {
	tests := []struct {
		name    string
		model   *ModelSpec
		wantErr error
	}{
		{
			name:    "nil model",
			model:   nil,
			wantErr: ErrModelRequired,
		},
		{
			name:    "empty provider",
			model:   &ModelSpec{Provider: "", Name: "claude-3"},
			wantErr: ErrModelProviderRequired,
		},
		{
			name:    "empty name",
			model:   &ModelSpec{Provider: "anthropic", Name: ""},
			wantErr: ErrModelNameRequired,
		},
		{
			name:    "valid model",
			model:   &ModelSpec{Provider: "anthropic", Name: "claude-3-opus"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ModelValidate(tt.model)
			if err != tt.wantErr {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestReactSpec_Validate(t *testing.T) {
	tests := []struct {
		name    string
		react   *ReactSpec
		wantErr error
	}{
		{
			name:    "nil react",
			react:   nil,
			wantErr: ErrReactSpecRequired,
		},
		{
			name:    "empty tools",
			react:   &ReactSpec{Tools: []string{}},
			wantErr: ErrReactToolsRequired,
		},
		{
			name:    "valid react",
			react:   &ReactSpec{Tools: []string{"tool1"}},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ReactValidate(tt.react)
			if err != tt.wantErr {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}
