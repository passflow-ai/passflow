package cel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEvaluator(t *testing.T) {
	eval, err := NewEvaluator()
	require.NoError(t, err)
	assert.NotNil(t, eval)
}

func TestEvaluator_Evaluate_SimpleBooleanExpressions(t *testing.T) {
	eval, err := NewEvaluator()
	require.NoError(t, err)

	tests := []struct {
		name       string
		expression string
		context    map[string]interface{}
		expected   bool
		wantErr    bool
	}{
		{
			name:       "simple true",
			expression: "true",
			context:    map[string]interface{}{},
			expected:   true,
		},
		{
			name:       "simple false",
			expression: "false",
			context:    map[string]interface{}{},
			expected:   false,
		},
		{
			name:       "numeric comparison greater than",
			expression: "amount > 1000",
			context:    map[string]interface{}{"amount": 1500},
			expected:   true,
		},
		{
			name:       "numeric comparison less than",
			expression: "amount < 1000",
			context:    map[string]interface{}{"amount": 1500},
			expected:   false,
		},
		{
			name:       "string equality",
			expression: "status == 'active'",
			context:    map[string]interface{}{"status": "active"},
			expected:   true,
		},
		{
			name:       "string inequality",
			expression: "status != 'active'",
			context:    map[string]interface{}{"status": "inactive"},
			expected:   true,
		},
		{
			name:       "logical AND",
			expression: "amount > 100 && status == 'approved'",
			context:    map[string]interface{}{"amount": 500, "status": "approved"},
			expected:   true,
		},
		{
			name:       "logical OR",
			expression: "priority == 'critical' || amount > 10000",
			context:    map[string]interface{}{"priority": "low", "amount": 15000},
			expected:   true,
		},
		{
			name:       "nested property access",
			expression: "input.amount > 1000",
			context:    map[string]interface{}{"input": map[string]interface{}{"amount": 2000}},
			expected:   true,
		},
		{
			name:       "list contains",
			expression: "'admin' in roles",
			context:    map[string]interface{}{"roles": []string{"user", "admin", "editor"}},
			expected:   true,
		},
		{
			name:       "list not contains",
			expression: "'superadmin' in roles",
			context:    map[string]interface{}{"roles": []string{"user", "admin"}},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.Evaluate(tt.expression, tt.context)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEvaluator_Evaluate_PolicyConditions(t *testing.T) {
	eval, err := NewEvaluator()
	require.NoError(t, err)

	tests := []struct {
		name       string
		expression string
		context    map[string]interface{}
		expected   bool
	}{
		{
			name:       "high risk tool requires approval",
			expression: "tool.side_effect_level == 'high_risk'",
			context: map[string]interface{}{
				"tool": map[string]interface{}{
					"side_effect_level": "high_risk",
					"name":              "delete_user",
				},
			},
			expected: true,
		},
		{
			name:       "large amount needs manager approval",
			expression: "input.amount > 10000",
			context: map[string]interface{}{
				"input": map[string]interface{}{
					"amount":   25000,
					"currency": "USD",
				},
			},
			expected: true,
		},
		{
			name:       "low confidence needs human verification",
			expression: "output.confidence < 0.8",
			context: map[string]interface{}{
				"output": map[string]interface{}{
					"confidence": 0.65,
					"result":     "uncertain",
				},
			},
			expected: true,
		},
		{
			name:       "pii detected requires compliant processing",
			expression: "input.contains_pii == true",
			context: map[string]interface{}{
				"input": map[string]interface{}{
					"contains_pii": true,
					"data_type":    "personal",
				},
			},
			expected: true,
		},
		{
			name:       "department specific rule",
			expression: "task.department == 'finance' && input.amount > 5000",
			context: map[string]interface{}{
				"task": map[string]interface{}{
					"department": "finance",
					"type":       "expense_approval",
				},
				"input": map[string]interface{}{
					"amount": 7500,
				},
			},
			expected: true,
		},
		{
			name:       "time based rule - business hours",
			expression: "hour >= 9 && hour <= 17",
			context: map[string]interface{}{
				"hour": 14,
			},
			expected: true,
		},
		{
			name:       "complex multi-condition rule",
			expression: "(task.type == 'lead_qualification' && output.score < 50) || (task.type == 'ticket_triage' && output.priority == 'critical')",
			context: map[string]interface{}{
				"task": map[string]interface{}{
					"type": "lead_qualification",
				},
				"output": map[string]interface{}{
					"score":    35,
					"priority": "medium",
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.Evaluate(tt.expression, tt.context)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEvaluator_Evaluate_ErrorCases(t *testing.T) {
	eval, err := NewEvaluator()
	require.NoError(t, err)

	tests := []struct {
		name       string
		expression string
		context    map[string]interface{}
	}{
		{
			name:       "invalid expression syntax",
			expression: "amount >",
			context:    map[string]interface{}{},
		},
		{
			name:       "undefined variable",
			expression: "undefined_var > 100",
			context:    map[string]interface{}{},
		},
		{
			name:       "type mismatch",
			expression: "amount > 'text'",
			context:    map[string]interface{}{"amount": 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := eval.Evaluate(tt.expression, tt.context)
			assert.Error(t, err)
		})
	}
}

func TestEvaluator_ValidateExpression(t *testing.T) {
	eval, err := NewEvaluator()
	require.NoError(t, err)

	tests := []struct {
		name       string
		expression string
		valid      bool
	}{
		{
			name:       "valid simple expression",
			expression: "true",
			valid:      true,
		},
		{
			name:       "valid comparison",
			expression: "amount > 100",
			valid:      true,
		},
		{
			name:       "valid complex expression",
			expression: "input.amount > 1000 && status == 'active'",
			valid:      true,
		},
		{
			name:       "invalid syntax - missing operand",
			expression: "amount >",
			valid:      false,
		},
		{
			name:       "invalid syntax - unbalanced parentheses",
			expression: "(amount > 100",
			valid:      false,
		},
		{
			name:       "empty expression",
			expression: "",
			valid:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := eval.ValidateExpression(tt.expression)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestEvaluator_EvaluateWithDefaults(t *testing.T) {
	eval, err := NewEvaluator()
	require.NoError(t, err)

	// Expression using a variable that doesn't exist in context
	// Should use the default value
	expression := "amount > 100"
	context := map[string]interface{}{} // No amount
	defaults := map[string]interface{}{"amount": 50}

	result, err := eval.EvaluateWithDefaults(expression, context, defaults)
	require.NoError(t, err)
	assert.Equal(t, false, result) // 50 > 100 is false
}

func TestPolicyEvaluator_EvaluateRules(t *testing.T) {
	policyEval := NewPolicyEvaluator()

	rules := []RuleDefinition{
		{
			ID:        "rule-1",
			Name:      "High Value Transaction",
			Condition: "input.amount > 10000",
			Action:    "require_approval",
			Priority:  100,
			IsActive:  true,
		},
		{
			ID:        "rule-2",
			Name:      "Critical Priority",
			Condition: "task.priority == 'critical'",
			Action:    "escalate",
			Priority:  200,
			IsActive:  true,
		},
		{
			ID:        "rule-3",
			Name:      "Disabled Rule",
			Condition: "true",
			Action:    "deny",
			Priority:  300,
			IsActive:  false, // Disabled
		},
	}

	context := map[string]interface{}{
		"input": map[string]interface{}{
			"amount": 25000,
		},
		"task": map[string]interface{}{
			"priority": "high",
			"type":     "payment",
		},
	}

	results, err := policyEval.EvaluateRules(rules, context, EvaluationModeFirstMatch)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "rule-1", results[0].RuleID)
	assert.Equal(t, "require_approval", results[0].Action)
	assert.True(t, results[0].Matched)
}

func TestPolicyEvaluator_EvaluateRules_CollectAll(t *testing.T) {
	policyEval := NewPolicyEvaluator()

	rules := []RuleDefinition{
		{
			ID:        "rule-1",
			Name:      "High Amount",
			Condition: "amount > 1000",
			Action:    "log",
			Priority:  100,
			IsActive:  true,
		},
		{
			ID:        "rule-2",
			Name:      "Premium User",
			Condition: "user_type == 'premium'",
			Action:    "notify",
			Priority:  200,
			IsActive:  true,
		},
	}

	context := map[string]interface{}{
		"amount":    5000,
		"user_type": "premium",
	}

	results, err := policyEval.EvaluateRules(rules, context, EvaluationModeCollectAll)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.True(t, results[0].Matched)
	assert.True(t, results[1].Matched)
}
