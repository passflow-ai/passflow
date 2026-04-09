package cel

import (
	"errors"
	"fmt"
	"sort"

	"github.com/google/cel-go/cel"
)

// Errors returned by the evaluator.
var (
	ErrEmptyExpression   = errors.New("expression cannot be empty")
	ErrInvalidExpression = errors.New("invalid expression syntax")
	ErrEvaluationFailed  = errors.New("expression evaluation failed")
	ErrNonBooleanResult  = errors.New("expression must evaluate to boolean")
)

// Evaluator provides CEL expression evaluation capabilities.
type Evaluator struct {
	baseEnv *cel.Env
}

// NewEvaluator creates a new CEL evaluator with standard declarations.
func NewEvaluator() (*Evaluator, error) {
	env, err := cel.NewEnv(
		cel.Variable("input", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("output", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("task", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("tool", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("step", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("user", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("context", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %w", err)
	}

	return &Evaluator{baseEnv: env}, nil
}

// Evaluate evaluates a CEL expression against the provided context.
// Returns true/false for boolean expressions.
func (e *Evaluator) Evaluate(expression string, context map[string]interface{}) (bool, error) {
	if expression == "" {
		return false, ErrEmptyExpression
	}

	// Create a new environment with dynamic variables from context
	env, err := e.createEnvWithContext(context)
	if err != nil {
		return false, fmt.Errorf("failed to create environment: %w", err)
	}

	// Parse and check the expression
	ast, issues := env.Parse(expression)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("%w: %v", ErrInvalidExpression, issues.Err())
	}

	checkedAst, issues := env.Check(ast)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("%w: %v", ErrInvalidExpression, issues.Err())
	}

	// Create the program
	prg, err := env.Program(checkedAst)
	if err != nil {
		return false, fmt.Errorf("failed to create program: %w", err)
	}

	// Prepare activation (variables)
	activation := e.prepareActivation(context)

	// Evaluate
	out, _, err := prg.Eval(activation)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrEvaluationFailed, err)
	}

	// Convert result to boolean
	result, ok := out.Value().(bool)
	if !ok {
		return false, ErrNonBooleanResult
	}

	return result, nil
}

// EvaluateWithDefaults evaluates an expression using defaults for missing variables.
func (e *Evaluator) EvaluateWithDefaults(expression string, context, defaults map[string]interface{}) (bool, error) {
	// Merge defaults with context (context takes precedence)
	merged := make(map[string]interface{})
	for k, v := range defaults {
		merged[k] = v
	}
	for k, v := range context {
		merged[k] = v
	}
	return e.Evaluate(expression, merged)
}

// ValidateExpression checks if an expression is syntactically valid.
func (e *Evaluator) ValidateExpression(expression string) error {
	if expression == "" {
		return ErrEmptyExpression
	}

	// Parse the expression
	_, issues := e.baseEnv.Parse(expression)
	if issues != nil && issues.Err() != nil {
		return fmt.Errorf("%w: %v", ErrInvalidExpression, issues.Err())
	}

	return nil
}

// createEnvWithContext creates a CEL environment with variables from context.
func (e *Evaluator) createEnvWithContext(context map[string]interface{}) (*cel.Env, error) {
	var opts []cel.EnvOption

	// Standard variable names that we want to declare as map types
	standardVars := map[string]bool{
		"input":   true,
		"output":  true,
		"task":    true,
		"tool":    true,
		"step":    true,
		"user":    true,
		"context": true,
	}

	// Add declarations for all top-level keys in context
	for key := range context {
		if standardVars[key] {
			// Use map type for standard variables
			opts = append(opts, cel.Variable(key, cel.MapType(cel.StringType, cel.DynType)))
		} else {
			// Use dynamic type for other variables
			opts = append(opts, cel.Variable(key, cel.DynType))
		}
	}

	// Add standard declarations that are not in context
	for varName := range standardVars {
		if _, exists := context[varName]; !exists {
			opts = append(opts, cel.Variable(varName, cel.MapType(cel.StringType, cel.DynType)))
		}
	}

	return cel.NewEnv(opts...)
}

// prepareActivation converts the context map into CEL-compatible values.
func (e *Evaluator) prepareActivation(context map[string]interface{}) map[string]interface{} {
	activation := make(map[string]interface{})
	for key, value := range context {
		activation[key] = e.convertValue(value)
	}
	return activation
}

// convertValue converts Go values to CEL-compatible values.
func (e *Evaluator) convertValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			result[k] = e.convertValue(val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = e.convertValue(val)
		}
		return result
	case []string:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = val
		}
		return result
	default:
		return value
	}
}

// EvaluationMode defines how rules are evaluated.
type EvaluationMode string

const (
	EvaluationModeFirstMatch EvaluationMode = "first_match"
	EvaluationModeCollectAll EvaluationMode = "collect_all"
)

// RuleDefinition represents a policy rule to evaluate.
type RuleDefinition struct {
	ID        string
	Name      string
	Condition string
	Action    string
	Priority  int
	IsActive  bool
	Config    map[string]interface{}
}

// RuleResult represents the result of evaluating a rule.
type RuleResult struct {
	RuleID   string
	RuleName string
	Matched  bool
	Action   string
	Config   map[string]interface{}
	Error    error
}

// PolicyEvaluator evaluates policy rules against context.
type PolicyEvaluator struct {
	eval *Evaluator
}

// NewPolicyEvaluator creates a new policy evaluator.
func NewPolicyEvaluator() *PolicyEvaluator {
	eval, _ := NewEvaluator()
	return &PolicyEvaluator{eval: eval}
}

// EvaluateRules evaluates a set of rules against the context.
func (p *PolicyEvaluator) EvaluateRules(
	rules []RuleDefinition,
	context map[string]interface{},
	mode EvaluationMode,
) ([]RuleResult, error) {
	// Filter active rules and sort by priority (descending)
	activeRules := make([]RuleDefinition, 0)
	for _, rule := range rules {
		if rule.IsActive {
			activeRules = append(activeRules, rule)
		}
	}

	sort.Slice(activeRules, func(i, j int) bool {
		return activeRules[i].Priority > activeRules[j].Priority
	})

	var results []RuleResult

	for _, rule := range activeRules {
		matched, err := p.eval.Evaluate(rule.Condition, context)

		result := RuleResult{
			RuleID:   rule.ID,
			RuleName: rule.Name,
			Matched:  matched && err == nil,
			Action:   rule.Action,
			Config:   rule.Config,
			Error:    err,
		}

		if matched && err == nil {
			results = append(results, result)

			if mode == EvaluationModeFirstMatch {
				break
			}
		} else if err != nil {
			results = append(results, result)
		}
	}

	return results, nil
}

// GetMatchedActions returns only the matched rule actions.
func (p *PolicyEvaluator) GetMatchedActions(
	rules []RuleDefinition,
	context map[string]interface{},
	mode EvaluationMode,
) ([]string, error) {
	results, err := p.EvaluateRules(rules, context, mode)
	if err != nil {
		return nil, err
	}

	actions := make([]string, 0)
	for _, result := range results {
		if result.Matched {
			actions = append(actions, result.Action)
		}
	}

	return actions, nil
}

// ShouldRequireApproval checks if any rule requires approval.
func (p *PolicyEvaluator) ShouldRequireApproval(
	rules []RuleDefinition,
	context map[string]interface{},
) (bool, *RuleResult, error) {
	results, err := p.EvaluateRules(rules, context, EvaluationModeFirstMatch)
	if err != nil {
		return false, nil, err
	}

	for _, result := range results {
		if result.Matched && result.Action == "require_approval" {
			return true, &result, nil
		}
	}

	return false, nil, nil
}

// ShouldDeny checks if any rule denies the action.
func (p *PolicyEvaluator) ShouldDeny(
	rules []RuleDefinition,
	context map[string]interface{},
) (bool, *RuleResult, error) {
	results, err := p.EvaluateRules(rules, context, EvaluationModeFirstMatch)
	if err != nil {
		return false, nil, err
	}

	for _, result := range results {
		if result.Matched && result.Action == "deny" {
			return true, &result, nil
		}
	}

	return false, nil, nil
}
