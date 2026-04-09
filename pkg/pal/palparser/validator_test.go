package palparser

import (
	"context"
	"testing"

	"github.com/passflow-ai/passflow/pkg/pal/domain"
)

// mockToolValidator implements ToolValidator for testing
type mockToolValidator struct {
	tools   map[string]bool
	errOnID string
}

func newMockToolValidator() *mockToolValidator {
	return &mockToolValidator{
		tools: make(map[string]bool),
	}
}

func (m *mockToolValidator) ToolExists(ctx context.Context, workspaceID, toolID string) (bool, error) {
	if toolID == m.errOnID {
		return false, ToolValidationFailedErr
	}
	return m.tools[toolID], nil
}

func (m *mockToolValidator) addTool(toolID string) *mockToolValidator {
	m.tools[toolID] = true
	return m
}

// mockSecretValidator implements SecretValidator for testing
type mockSecretValidator struct {
	secrets map[string]bool
	errOnID string
}

func newMockSecretValidator() *mockSecretValidator {
	return &mockSecretValidator{
		secrets: make(map[string]bool),
	}
}

func (m *mockSecretValidator) SecretExists(ctx context.Context, workspaceID, secretKey string) (bool, error) {
	if secretKey == m.errOnID {
		return false, SecretValidationFailedErr
	}
	return m.secrets[secretKey], nil
}

func (m *mockSecretValidator) addSecret(secretKey string) *mockSecretValidator {
	m.secrets[secretKey] = true
	return m
}

// mockModelValidator implements ModelValidator for testing
type mockModelValidator struct {
	providers map[string]bool
	models    map[string]map[string]bool
	errOnName string
}

func newMockModelValidator() *mockModelValidator {
	return &mockModelValidator{
		providers: make(map[string]bool),
		models:    make(map[string]map[string]bool),
	}
}

func (m *mockModelValidator) IsValidProvider(provider string) bool {
	return m.providers[provider]
}

func (m *mockModelValidator) IsValidModel(provider, modelName string) bool {
	if modelName == m.errOnName {
		return false
	}
	providerModels, ok := m.models[provider]
	if !ok {
		return false
	}
	return providerModels[modelName]
}

func (m *mockModelValidator) addProvider(provider string) *mockModelValidator {
	m.providers[provider] = true
	return m
}

func (m *mockModelValidator) addModel(provider, modelName string) *mockModelValidator {
	if m.models[provider] == nil {
		m.models[provider] = make(map[string]bool)
	}
	m.models[provider][modelName] = true
	return m
}

func newValidPALSpec() *domain.PALSpec {
	return &domain.PALSpec{
		Agent: &domain.AgentSpec{
			Name:        "test-agent",
			Description: "Test agent",
		},
		Model: &domain.ModelSpec{
			Provider: "anthropic",
			Name:     "claude-3-opus",
			Config: map[string]interface{}{
				"temperature": 0.7,
				"max_tokens":  2048,
			},
		},
		React: &domain.ReactSpec{
			Tools:         []string{"tool1", "tool2"},
			MaxIterations: 10,
		},
	}
}

func TestNewValidator(t *testing.T) {
	toolVal := newMockToolValidator()
	secretVal := newMockSecretValidator()
	modelVal := newMockModelValidator()

	v := NewValidator(toolVal, secretVal, modelVal)

	if v == nil {
		t.Fatal("expected validator to be created")
	}
}

func TestValidateSemantics_StrictMode_AllValid(t *testing.T) {
	ctx := context.Background()
	spec := newValidPALSpec()

	toolVal := newMockToolValidator().addTool("tool1").addTool("tool2")
	secretVal := newMockSecretValidator()
	modelVal := newMockModelValidator().addProvider("anthropic").addModel("anthropic", "claude-3-opus")

	v := NewValidator(toolVal, secretVal, modelVal)

	result, err := v.ValidateSemantics(ctx, "workspace-1", spec, ModeStrict)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Valid {
		t.Fatal("expected validation to pass")
	}

	if len(result.Errors) > 0 {
		t.Fatalf("expected no errors, got %d", len(result.Errors))
	}

	if len(result.Warnings) > 0 {
		t.Fatalf("expected no warnings, got %d", len(result.Warnings))
	}
}

func TestValidateSemantics_StrictMode_MissingTool(t *testing.T) {
	ctx := context.Background()
	spec := newValidPALSpec()

	toolVal := newMockToolValidator().addTool("tool1")
	secretVal := newMockSecretValidator()
	modelVal := newMockModelValidator().addProvider("anthropic").addModel("anthropic", "claude-3-opus")

	v := NewValidator(toolVal, secretVal, modelVal)

	result, err := v.ValidateSemantics(ctx, "workspace-1", spec, ModeStrict)

	if err == nil {
		t.Fatal("expected error in strict mode")
	}

	if result == nil {
		t.Fatal("expected result even with error")
	}

	if result.Valid {
		t.Fatal("expected validation to fail")
	}

	if len(result.Errors) == 0 {
		t.Fatal("expected at least one error")
	}

	found := false
	for _, e := range result.Errors {
		if e.Code == ErrToolNotFound {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected tool not found error, got %v", result.Errors)
	}
}

func TestValidateSemantics_StrictMode_InvalidProvider(t *testing.T) {
	ctx := context.Background()
	spec := newValidPALSpec()
	spec.Model.Provider = "invalid-provider"

	toolVal := newMockToolValidator().addTool("tool1").addTool("tool2")
	secretVal := newMockSecretValidator()
	modelVal := newMockModelValidator()

	v := NewValidator(toolVal, secretVal, modelVal)

	result, err := v.ValidateSemantics(ctx, "workspace-1", spec, ModeStrict)

	if err == nil {
		t.Fatal("expected error for invalid provider")
	}

	if result.Valid {
		t.Fatal("expected validation to fail")
	}

	found := false
	for _, e := range result.Errors {
		if e.Code == ErrModelProviderInvalid {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected invalid provider error, got %v", result.Errors)
	}
}

func TestValidateSemantics_StrictMode_InvalidModel(t *testing.T) {
	ctx := context.Background()
	spec := newValidPALSpec()
	spec.Model.Name = "invalid-model"

	toolVal := newMockToolValidator().addTool("tool1").addTool("tool2")
	secretVal := newMockSecretValidator()
	modelVal := newMockModelValidator().addProvider("anthropic")

	v := NewValidator(toolVal, secretVal, modelVal)

	result, err := v.ValidateSemantics(ctx, "workspace-1", spec, ModeStrict)

	if err == nil {
		t.Fatal("expected error for invalid model")
	}

	if result.Valid {
		t.Fatal("expected validation to fail")
	}

	found := false
	for _, e := range result.Errors {
		if e.Code == ErrModelNameInvalid {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected invalid model error, got %v", result.Errors)
	}
}

func TestValidateSemantics_WarnMode_MultipleIssues(t *testing.T) {
	ctx := context.Background()
	spec := newValidPALSpec()
	spec.Model.Provider = "invalid-provider"

	toolVal := newMockToolValidator()
	secretVal := newMockSecretValidator()
	modelVal := newMockModelValidator()

	v := NewValidator(toolVal, secretVal, modelVal)

	result, err := v.ValidateSemantics(ctx, "workspace-1", spec, ModeWarn)

	if err != nil {
		t.Fatalf("expected no error in warn mode, got %v", err)
	}

	if result.Valid {
		t.Fatal("expected validation to fail")
	}

	totalIssues := len(result.Errors) + len(result.Warnings)
	if totalIssues == 0 {
		t.Fatal("expected at least one issue")
	}
}

func TestValidateSemantics_WarnMode_CollectsAllIssues(t *testing.T) {
	ctx := context.Background()
	spec := newValidPALSpec()
	spec.Model.Provider = "invalid-provider"
	spec.Model.Name = "invalid-model"

	toolVal := newMockToolValidator()
	secretVal := newMockSecretValidator()
	modelVal := newMockModelValidator()

	v := NewValidator(toolVal, secretVal, modelVal)

	result, err := v.ValidateSemantics(ctx, "workspace-1", spec, ModeWarn)

	if err != nil {
		t.Fatalf("expected no error in warn mode, got %v", err)
	}

	if result.Valid {
		t.Fatal("expected validation to fail")
	}

	totalIssues := len(result.Errors) + len(result.Warnings)
	if totalIssues < 2 {
		t.Fatalf("expected at least 2 issues, got %d", totalIssues)
	}
}

func TestValidateSemantics_NilSpec(t *testing.T) {
	ctx := context.Background()

	toolVal := newMockToolValidator()
	secretVal := newMockSecretValidator()
	modelVal := newMockModelValidator()

	v := NewValidator(toolVal, secretVal, modelVal)

	result, err := v.ValidateSemantics(ctx, "workspace-1", nil, ModeStrict)

	if err == nil {
		t.Fatal("expected error for nil spec")
	}

	if result == nil {
		t.Fatal("expected result even for nil spec")
	}

	if result.Valid {
		t.Fatal("expected validation to fail")
	}
}

func TestValidateSemantics_EmptyWorkspaceID(t *testing.T) {
	ctx := context.Background()
	spec := newValidPALSpec()

	toolVal := newMockToolValidator()
	secretVal := newMockSecretValidator()
	modelVal := newMockModelValidator()

	v := NewValidator(toolVal, secretVal, modelVal)

	result, err := v.ValidateSemantics(ctx, "", spec, ModeStrict)

	if err == nil {
		t.Fatal("expected error for empty workspace ID")
	}

	if result.Valid {
		t.Fatal("expected validation to fail")
	}
}

func TestValidateSemantics_WithGuardrails_AllowedTools(t *testing.T) {
	ctx := context.Background()
	spec := newValidPALSpec()
	spec.Guardrails = &domain.GuardrailsSpec{
		AllowedTools: []string{"tool1", "tool2"},
	}

	toolVal := newMockToolValidator().addTool("tool1").addTool("tool2").addTool("tool3")
	secretVal := newMockSecretValidator()
	modelVal := newMockModelValidator().addProvider("anthropic").addModel("anthropic", "claude-3-opus")

	v := NewValidator(toolVal, secretVal, modelVal)

	result, err := v.ValidateSemantics(ctx, "workspace-1", spec, ModeStrict)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Valid {
		t.Fatalf("expected validation to pass, got errors: %v", result.Errors)
	}
}

func TestValidateSemantics_WithGuardrails_BlockedTools(t *testing.T) {
	ctx := context.Background()
	spec := newValidPALSpec()
	spec.Guardrails = &domain.GuardrailsSpec{
		BlockedTools: []string{"tool1"},
	}

	toolVal := newMockToolValidator().addTool("tool1").addTool("tool2")
	secretVal := newMockSecretValidator()
	modelVal := newMockModelValidator().addProvider("anthropic").addModel("anthropic", "claude-3-opus")

	v := NewValidator(toolVal, secretVal, modelVal)

	result, err := v.ValidateSemantics(ctx, "workspace-1", spec, ModeWarn)

	if err != nil {
		t.Fatalf("expected no error in warn mode, got %v", err)
	}

	if result.Valid {
		t.Fatal("expected validation to have issues, got valid=true")
	}

	if len(result.Warnings) == 0 {
		t.Fatalf("expected warning for blocked tool in guardrails, got errors: %v, warnings: %v", result.Errors, result.Warnings)
	}
}

func TestValidateSemantics_NoReactSpec(t *testing.T) {
	ctx := context.Background()
	spec := newValidPALSpec()
	spec.React = nil

	toolVal := newMockToolValidator()
	secretVal := newMockSecretValidator()
	modelVal := newMockModelValidator().addProvider("anthropic").addModel("anthropic", "claude-3-opus")

	v := NewValidator(toolVal, secretVal, modelVal)

	result, err := v.ValidateSemantics(ctx, "workspace-1", spec, ModeStrict)

	if err == nil {
		t.Fatal("expected error for missing react spec")
	}

	if result.Valid {
		t.Fatal("expected validation to fail")
	}
}

func TestValidateSemantics_ToolValidatorError(t *testing.T) {
	ctx := context.Background()
	spec := newValidPALSpec()

	toolVal := newMockToolValidator().addTool("tool1").addTool("tool2")
	toolVal.errOnID = "tool2"

	secretVal := newMockSecretValidator()
	modelVal := newMockModelValidator().addProvider("anthropic").addModel("anthropic", "claude-3-opus")

	v := NewValidator(toolVal, secretVal, modelVal)

	result, err := v.ValidateSemantics(ctx, "workspace-1", spec, ModeStrict)

	if err == nil {
		t.Fatal("expected error from tool validator")
	}

	if result == nil {
		t.Fatal("expected result even with validator error")
	}
}

func TestValidateSemantics_EmptyToolsList(t *testing.T) {
	ctx := context.Background()
	spec := newValidPALSpec()
	spec.React.Tools = []string{}

	toolVal := newMockToolValidator()
	secretVal := newMockSecretValidator()
	modelVal := newMockModelValidator().addProvider("anthropic").addModel("anthropic", "claude-3-opus")

	v := NewValidator(toolVal, secretVal, modelVal)

	result, err := v.ValidateSemantics(ctx, "workspace-1", spec, ModeStrict)

	if err == nil {
		t.Fatal("expected error for empty tools list")
	}

	if result.Valid {
		t.Fatal("expected validation to fail")
	}
}

func TestValidateSemantics_ValidationIssueFields(t *testing.T) {
	ctx := context.Background()
	spec := newValidPALSpec()
	spec.React.Tools = []string{"missing-tool"}

	toolVal := newMockToolValidator()
	secretVal := newMockSecretValidator()
	modelVal := newMockModelValidator().addProvider("anthropic").addModel("anthropic", "claude-3-opus")

	v := NewValidator(toolVal, secretVal, modelVal)

	result, err := v.ValidateSemantics(ctx, "workspace-1", spec, ModeStrict)

	if err == nil {
		t.Fatal("expected error")
	}

	if len(result.Errors) == 0 {
		t.Fatal("expected at least one error")
	}

	issue := result.Errors[0]
	if issue.Code == "" {
		t.Fatal("expected issue code to be set")
	}

	if issue.Message == "" {
		t.Fatal("expected issue message to be set")
	}

	if issue.Field == "" {
		t.Fatal("expected issue field to be set")
	}
}
