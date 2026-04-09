package palparser

import (
	"testing"

	"github.com/passflow-ai/passflow/pkg/pal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseYAML_ValidMinimalSpec(t *testing.T) {
	yamlContent := []byte(`
agent:
  name: test-agent
model:
  provider: anthropic
  name: claude-3-opus
react:
  tools:
    - tool1
`)

	spec, err := ParseYAML(yamlContent)
	require.NoError(t, err)
	assert.NotNil(t, spec)
	assert.Equal(t, "test-agent", spec.Agent.Name)
	assert.Equal(t, "anthropic", spec.Model.Provider)
	assert.Equal(t, "claude-3-opus", spec.Model.Name)
	assert.Equal(t, []string{"tool1"}, spec.React.Tools)
	assert.Nil(t, spec.Triggers)
	assert.Nil(t, spec.Guardrails)
	assert.Nil(t, spec.Approvals)
	assert.Nil(t, spec.Output)
	assert.Nil(t, spec.Pod)
}

func TestParseYAML_ValidFullSpec(t *testing.T) {
	yamlContent := []byte(`
agent:
  name: full-agent
  description: A full featured agent
model:
  provider: openai
  name: gpt-4
  config:
    temperature: 0.7
    max_tokens: 2000
triggers:
  - type: cron
    schedule: "0 0 * * *"
  - type: webhook
    endpoint: /webhook/agent
react:
  tools:
    - tool1
    - tool2
  max_iterations: 10
  stop_sequences:
    - "END"
guardrails:
  allowed_tools:
    - tool1
  rate_limits:
    requests_per_minute: 10
  timeout_seconds: 300
approvals:
  required: true
  approvers:
    - approver1@example.com
  timeout_hours: 24
output:
  format: json
  destination: webhook
  transform:
    field1: value1
pod:
  image: my-image:latest
  resources:
    cpu: "1"
    memory: "512Mi"
  environment:
    DEBUG: "true"
`)

	spec, err := ParseYAML(yamlContent)
	require.NoError(t, err)
	assert.NotNil(t, spec)

	// Agent
	assert.Equal(t, "full-agent", spec.Agent.Name)
	assert.Equal(t, "A full featured agent", spec.Agent.Description)

	// Model
	assert.Equal(t, "openai", spec.Model.Provider)
	assert.Equal(t, "gpt-4", spec.Model.Name)
	assert.Equal(t, 0.7, spec.Model.Config["temperature"])
	assert.Equal(t, 2000, spec.Model.Config["max_tokens"])

	// Triggers
	assert.Len(t, spec.Triggers, 2)
	assert.Equal(t, "cron", spec.Triggers[0].Type)
	assert.Equal(t, "0 0 * * *", spec.Triggers[0].Schedule)
	assert.Equal(t, "webhook", spec.Triggers[1].Type)
	assert.Equal(t, "/webhook/agent", spec.Triggers[1].Endpoint)

	// React
	assert.Equal(t, []string{"tool1", "tool2"}, spec.React.Tools)
	assert.Equal(t, 10, spec.React.MaxIterations)
	assert.Equal(t, []string{"END"}, spec.React.StopSequences)

	// Guardrails
	assert.Equal(t, []string{"tool1"}, spec.Guardrails.AllowedTools)
	assert.Equal(t, 10, spec.Guardrails.RateLimits["requests_per_minute"])
	assert.Equal(t, 300, spec.Guardrails.TimeoutSeconds)

	// Approvals
	assert.True(t, spec.Approvals.Required)
	assert.Equal(t, []string{"approver1@example.com"}, spec.Approvals.Approvers)
	assert.Equal(t, 24, spec.Approvals.TimeoutHours)

	// Output
	assert.Equal(t, "json", spec.Output.Format)
	assert.Equal(t, "webhook", spec.Output.Destination)
	assert.Equal(t, "value1", spec.Output.Transform["field1"])

	// Pod
	assert.Equal(t, "my-image:latest", spec.Pod.Image)
	assert.Equal(t, "1", spec.Pod.Resources["cpu"])
	assert.Equal(t, "512Mi", spec.Pod.Resources["memory"])
	assert.Equal(t, "true", spec.Pod.Environment["DEBUG"])
}

func TestParseYAML_InvalidYAML(t *testing.T) {
	yamlContent := []byte(`
agent:
  name: test-agent
  invalid: [unclosed array
`)

	spec, err := ParseYAML(yamlContent)
	assert.Error(t, err)
	assert.Nil(t, spec)
	assert.Equal(t, ErrInvalidYAML, err)
}

func TestParseYAML_MissingAgentRoot(t *testing.T) {
	yamlContent := []byte(`
model:
  provider: anthropic
  name: claude-3-opus
react:
  tools:
    - tool1
`)

	spec, err := ParseYAML(yamlContent)
	assert.Error(t, err)
	assert.Nil(t, spec)
	assert.Equal(t, ErrMissingAgentRoot, err)
}

func TestParseAndValidate_ValidSpec(t *testing.T) {
	yamlContent := []byte(`
agent:
  name: test-agent
model:
  provider: anthropic
  name: claude-3-opus
react:
  tools:
    - tool1
`)

	spec, err := ParseAndValidate(yamlContent)
	require.NoError(t, err)
	assert.NotNil(t, spec)
}

func TestParseAndValidate_InvalidYAML(t *testing.T) {
	yamlContent := []byte(`invalid: [unclosed`)

	spec, err := ParseAndValidate(yamlContent)
	assert.Error(t, err)
	assert.Nil(t, spec)
}

func TestParseAndValidate_MissingModel(t *testing.T) {
	yamlContent := []byte(`
agent:
  name: test-agent
react:
  tools:
    - tool1
`)

	spec, err := ParseAndValidate(yamlContent)
	assert.Error(t, err)
	assert.Nil(t, spec)
	assert.Equal(t, domain.ErrModelRequired, err)
}

func TestParseAndValidate_MissingReact(t *testing.T) {
	yamlContent := []byte(`
agent:
  name: test-agent
model:
  provider: anthropic
  name: claude-3-opus
`)

	spec, err := ParseAndValidate(yamlContent)
	assert.Error(t, err)
	assert.Nil(t, spec)
	assert.Equal(t, domain.ErrReactSpecRequired, err)
}

func TestParseAndValidate_MissingAgentName(t *testing.T) {
	yamlContent := []byte(`
agent:
  description: Agent without name
model:
  provider: anthropic
  name: claude-3-opus
react:
  tools:
    - tool1
`)

	spec, err := ParseAndValidate(yamlContent)
	assert.Error(t, err)
	assert.Nil(t, spec)
	assert.Equal(t, domain.ErrAgentNameRequired, err)
}

func TestParseAndValidate_EmptyReactTools(t *testing.T) {
	yamlContent := []byte(`
agent:
  name: test-agent
model:
  provider: anthropic
  name: claude-3-opus
react:
  tools: []
`)

	spec, err := ParseAndValidate(yamlContent)
	assert.Error(t, err)
	assert.Nil(t, spec)
	assert.Equal(t, domain.ErrReactToolsRequired, err)
}
