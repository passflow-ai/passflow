package palparser

import (
	"context"
	"errors"
	"fmt"

	"github.com/passflow-ai/passflow/pkg/pal/domain"
)

// Validation mode constants
const (
	ModeStrict = "strict"
	ModeWarn   = "warn"
)

// Validation error codes
const (
	ErrToolNotFound              = "TOOL_NOT_FOUND"
	ErrToolValidationFailedCode  = "TOOL_VALIDATION_FAILED"
	ErrSecretNotFound            = "SECRET_NOT_FOUND"
	ErrSecretValidationFailedCode = "SECRET_VALIDATION_FAILED"
	ErrModelProviderInvalid      = "MODEL_PROVIDER_INVALID"
	ErrModelNameInvalid          = "MODEL_NAME_INVALID"
	ErrInvalidValidationMode     = "INVALID_VALIDATION_MODE"
	ErrInvalidValidationInput    = "INVALID_VALIDATION_INPUT"
	ErrBlockedToolInGuardrails   = "BLOCKED_TOOL_IN_GUARDRAILS"
	ErrAllowedToolsConflict      = "ALLOWED_TOOLS_CONFLICT"
)

var (
	ToolValidationFailedErr   = errors.New("tool validation failed")
	SecretValidationFailedErr = errors.New("secret validation failed")
)

// ToolValidator defines the interface for validating tool existence
type ToolValidator interface {
	ToolExists(ctx context.Context, workspaceID, toolID string) (bool, error)
}

// SecretValidator defines the interface for validating secret existence
type SecretValidator interface {
	SecretExists(ctx context.Context, workspaceID, secretKey string) (bool, error)
}

// ModelValidator defines the interface for validating model providers and names
type ModelValidator interface {
	IsValidProvider(provider string) bool
	IsValidModel(provider, modelName string) bool
}

// ValidationIssue represents a single validation issue
type ValidationIssue struct {
	Code    string
	Message string
	Field   string
	Details map[string]interface{}
}

// ValidationResult represents the result of semantic validation
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationIssue
	Warnings []ValidationIssue
}

// Validator performs semantic validation of PAL specs
type Validator struct {
	toolValidator   ToolValidator
	secretValidator SecretValidator
	modelValidator  ModelValidator
}

// NewValidator creates a new semantic validator
func NewValidator(toolVal ToolValidator, secretVal SecretValidator, modelVal ModelValidator) *Validator {
	return &Validator{
		toolValidator:   toolVal,
		secretValidator: secretVal,
		modelValidator:  modelVal,
	}
}

// ValidateSemantics validates a PAL spec against external dependencies
// mode can be ModeStrict (fail on first error) or ModeWarn (collect all issues)
func (v *Validator) ValidateSemantics(
	ctx context.Context,
	workspaceID string,
	spec *domain.PALSpec,
	mode string,
) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationIssue{},
		Warnings: []ValidationIssue{},
	}

	// Validate input parameters
	if err := v.validateInput(spec, workspaceID); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Code:    ErrInvalidValidationInput,
			Message: err.Error(),
			Field:   "spec",
		})
		return result, err
	}

	// Validate mode
	if mode != ModeStrict && mode != ModeWarn {
		err := fmt.Errorf("invalid validation mode: %s, must be %s or %s", mode, ModeStrict, ModeWarn)
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Code:    ErrInvalidValidationMode,
			Message: err.Error(),
			Field:   "mode",
		})
		return result, err
	}

	// Validate model provider and name
	modelIssue := v.validateModel(spec)
	if modelIssue.Code != "" {
		if mode == ModeStrict {
			result.Valid = false
			result.Errors = append(result.Errors, modelIssue)
			return result, fmt.Errorf("semantic validation failed: %s", modelIssue.Message)
		}
		result.Valid = false
		result.Warnings = append(result.Warnings, modelIssue)
	}

	// Validate react spec
	if spec.React == nil {
		issue := ValidationIssue{
			Code:    ErrInvalidValidationInput,
			Message: "React spec is required",
			Field:   "react",
		}
		if mode == ModeStrict {
			result.Valid = false
			result.Errors = append(result.Errors, issue)
			return result, errors.New("semantic validation failed: React spec is required")
		}
		result.Valid = false
		result.Warnings = append(result.Warnings, issue)
	}

	// Validate tools existence
	if spec.React != nil {
		if len(spec.React.Tools) == 0 {
			issue := ValidationIssue{
				Code:    ErrInvalidValidationInput,
				Message: "At least one tool is required in react spec",
				Field:   "react.tools",
			}
			if mode == ModeStrict {
				result.Valid = false
				result.Errors = append(result.Errors, issue)
				return result, errors.New("semantic validation failed: At least one tool is required in react spec")
			}
			result.Valid = false
			result.Warnings = append(result.Warnings, issue)
		} else {
			if errs := v.validateTools(ctx, workspaceID, spec.React.Tools); len(errs) > 0 {
				if mode == ModeStrict {
					result.Valid = false
					result.Errors = append(result.Errors, errs...)
					if len(errs) > 0 {
						return result, fmt.Errorf("semantic validation failed: %s", errs[0].Message)
					}
				}
				result.Valid = false
				result.Errors = append(result.Errors, errs...)
			}
		}
	}

	// Validate guardrails if present
	if spec.Guardrails != nil {
		if errs := v.validateGuardrails(ctx, workspaceID, spec); len(errs) > 0 {
			if mode == ModeStrict {
				result.Valid = false
				result.Errors = append(result.Errors, errs...)
				if len(errs) > 0 {
					return result, fmt.Errorf("semantic validation failed: %s", errs[0].Message)
				}
			}
			result.Valid = false
			result.Warnings = append(result.Warnings, errs...)
		}
	}

	return result, nil
}

// validateInput validates the input parameters
func (v *Validator) validateInput(spec *domain.PALSpec, workspaceID string) error {
	if spec == nil {
		return errors.New("spec is required")
	}

	if workspaceID == "" {
		return errors.New("workspace ID is required")
	}

	return nil
}

// validateModel validates the model provider and name
func (v *Validator) validateModel(spec *domain.PALSpec) ValidationIssue {
	if spec.Model == nil {
		return ValidationIssue{
			Code:    ErrInvalidValidationInput,
			Message: "Model spec is required",
			Field:   "model",
		}
	}

	// Validate provider
	if !v.modelValidator.IsValidProvider(spec.Model.Provider) {
		return ValidationIssue{
			Code:    ErrModelProviderInvalid,
			Message: fmt.Sprintf("invalid model provider: %s", spec.Model.Provider),
			Field:   "model.provider",
			Details: map[string]interface{}{
				"provider": spec.Model.Provider,
			},
		}
	}

	// Validate model name
	if !v.modelValidator.IsValidModel(spec.Model.Provider, spec.Model.Name) {
		return ValidationIssue{
			Code:    ErrModelNameInvalid,
			Message: fmt.Sprintf("invalid model name: %s for provider %s", spec.Model.Name, spec.Model.Provider),
			Field:   "model.name",
			Details: map[string]interface{}{
				"provider": spec.Model.Provider,
				"name":     spec.Model.Name,
			},
		}
	}

	return ValidationIssue{}
}

// validateTools validates that all tools exist in the workspace
func (v *Validator) validateTools(ctx context.Context, workspaceID string, tools []string) []ValidationIssue {
	var issues []ValidationIssue

	for _, toolID := range tools {
		exists, err := v.toolValidator.ToolExists(ctx, workspaceID, toolID)
		if err != nil {
			issues = append(issues, ValidationIssue{
				Code:    ErrToolValidationFailedCode,
				Message: fmt.Sprintf("error validating tool %s: %v", toolID, err),
				Field:   "react.tools",
				Details: map[string]interface{}{
					"tool_id": toolID,
					"error":   err.Error(),
				},
			})
			continue
		}

		if !exists {
			issues = append(issues, ValidationIssue{
				Code:    ErrToolNotFound,
				Message: fmt.Sprintf("tool not found in workspace: %s", toolID),
				Field:   "react.tools",
				Details: map[string]interface{}{
					"tool_id": toolID,
				},
			})
		}
	}

	return issues
}

// validateGuardrails validates guardrails configuration
func (v *Validator) validateGuardrails(ctx context.Context, workspaceID string, spec *domain.PALSpec) []ValidationIssue {
	var issues []ValidationIssue

	if spec.Guardrails == nil || spec.React == nil {
		return issues
	}

	guardrails := spec.Guardrails
	reactTools := spec.React.Tools

	// Check blocked tools
	if len(guardrails.BlockedTools) > 0 {
		for _, blockedTool := range guardrails.BlockedTools {
			for _, reactTool := range reactTools {
				if blockedTool == reactTool {
					issues = append(issues, ValidationIssue{
						Code:    ErrBlockedToolInGuardrails,
						Message: fmt.Sprintf("tool %s is used in react but blocked in guardrails", blockedTool),
						Field:   "guardrails.blocked_tools",
						Details: map[string]interface{}{
							"tool_id": blockedTool,
						},
					})
				}
			}
		}
	}

	// Check allowed tools
	if len(guardrails.AllowedTools) > 0 {
		allowedMap := make(map[string]bool)
		for _, tool := range guardrails.AllowedTools {
			allowedMap[tool] = true
		}

		for _, reactTool := range reactTools {
			if !allowedMap[reactTool] {
				issues = append(issues, ValidationIssue{
					Code:    ErrAllowedToolsConflict,
					Message: fmt.Sprintf("tool %s is used in react but not in allowed_tools list", reactTool),
					Field:   "guardrails.allowed_tools",
					Details: map[string]interface{}{
						"tool_id": reactTool,
					},
				})
			}
		}
	}

	return issues
}
