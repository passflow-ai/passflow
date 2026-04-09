package job

import "fmt"

// Execution mode constants for the Spec.ExecutionMode field.
const (
	ExecutionModeNormal = "normal"
	ExecutionModeShadow = "shadow"
)

// ShadowConfig carries the configuration for a shadow execution.
// It is embedded in the job Spec when ExecutionMode is "shadow".
type ShadowConfig struct {
	// OriginalExecutionID is the execution being replayed.
	OriginalExecutionID string `json:"original_execution_id"`

	// AnalysisBatchID groups shadow runs for batch comparison.
	AnalysisBatchID string `json:"analysis_batch_id"`

	// InjectedSystemPrompt overrides the agent's system prompt for this shadow run.
	InjectedSystemPrompt string `json:"injected_system_prompt,omitempty"`

	// InjectedToolDescs overrides individual tool descriptions by name.
	InjectedToolDescs map[string]string `json:"injected_tool_descriptions,omitempty"`

	// MockedToolOutputs maps tool names to their canned JSON responses.
	MockedToolOutputs map[string]string `json:"mocked_tool_outputs"`

	// ReplayPolicy controls strictness: "strict" aborts on missing mocks,
	// "best_effort" allows live execution for deterministic reads.
	ReplayPolicy string `json:"replay_policy"`

	// TraceID propagates distributed tracing context.
	TraceID string `json:"trace_id"`
}

// Validate returns an error if any required field is missing or invalid.
func (sc *ShadowConfig) Validate() error {
	if sc.OriginalExecutionID == "" {
		return fmt.Errorf("shadow_config: original_execution_id is required")
	}
	if sc.AnalysisBatchID == "" {
		return fmt.Errorf("shadow_config: analysis_batch_id is required")
	}
	if sc.MockedToolOutputs == nil {
		return fmt.Errorf("shadow_config: mocked_tool_outputs must not be nil")
	}
	if sc.ReplayPolicy != "strict" && sc.ReplayPolicy != "best_effort" {
		return fmt.Errorf("shadow_config: replay_policy must be 'strict' or 'best_effort', got %q", sc.ReplayPolicy)
	}
	if sc.TraceID == "" {
		return fmt.Errorf("shadow_config: trace_id is required")
	}
	return nil
}
