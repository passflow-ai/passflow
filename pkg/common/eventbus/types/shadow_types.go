package types

import (
	"encoding/json"
	"errors"
	"time"
)

// Shadow execution mode constants.
const (
	ExecutionModeNormal = "normal"
	ExecutionModeShadow = "shadow"
)

// Shadow result status constants.
const (
	ShadowStatusCompleted = "completed"
	ShadowStatusFailed    = "failed"
	ShadowStatusAborted   = "aborted"
)

// Replay policy constants.
const (
	ReplayPolicyStrict     = "strict"
	ReplayPolicyBestEffort = "best_effort"
)

// Stream name constants.
const (
	StreamTasksQueue    = "passflow.tasks.queue"
	StreamShadowResults = "passflow.shadow.results"
)

// Consumer group constants.
const (
	ConsumerGroupAgentExecutor  = "agent-executor-group"
	ConsumerGroupShadowAnalyzer = "shadow-analyzer-group"
)

// Validation errors for shadow types.
var (
	ErrEmptyOriginalExecutionID = errors.New("shadow_config: original_execution_id is required")
	ErrEmptyAnalysisBatchID     = errors.New("shadow_config: analysis_batch_id is required")
	ErrInvalidReplayPolicy      = errors.New("shadow_config: replay_policy must be 'strict' or 'best_effort'")
	ErrEmptyShadowTraceID       = errors.New("shadow_config: trace_id is required")

	ErrEmptyShadowExecutionID     = errors.New("shadow_result: shadow_execution_id is required")
	ErrEmptyResultOriginalExecID  = errors.New("shadow_result: original_execution_id is required")
	ErrEmptyResultAnalysisBatchID = errors.New("shadow_result: analysis_batch_id is required")
	ErrEmptyResultTenantID        = errors.New("shadow_result: tenant_id is required")
	ErrEmptyResultAgentID         = errors.New("shadow_result: agent_id is required")
	ErrEmptyResultAgentVersion    = errors.New("shadow_result: agent_version is required")
	ErrEmptyResultTraceID         = errors.New("shadow_result: trace_id is required")
	ErrEmptyResultIdempotencyKey  = errors.New("shadow_result: idempotency_key is required")
	ErrInvalidShadowResultStatus  = errors.New("shadow_result: status must be 'completed', 'failed', or 'aborted'")
	ErrEmptyResultSchemaVersion   = errors.New("shadow_result: schema_version is required")
)

// ShadowConfig carries the configuration for a shadow execution.
// It is embedded in the task queue message when execution_mode is "shadow".
type ShadowConfig struct {
	OriginalExecutionID      string            `json:"original_execution_id"`
	AnalysisBatchID          string            `json:"analysis_batch_id"`
	InjectedSystemPrompt     string            `json:"injected_system_prompt,omitempty"`
	InjectedToolDescriptions map[string]string `json:"injected_tool_descriptions,omitempty"`
	MockedToolOutputs        map[string]string `json:"mocked_tool_outputs,omitempty"`
	ReplayPolicy             string            `json:"replay_policy"`
	TraceID                  string            `json:"trace_id"`
}

// Validate checks that ShadowConfig has all required fields and valid values.
func (sc *ShadowConfig) Validate() error {
	if sc.OriginalExecutionID == "" {
		return ErrEmptyOriginalExecutionID
	}
	if sc.AnalysisBatchID == "" {
		return ErrEmptyAnalysisBatchID
	}
	if sc.ReplayPolicy != ReplayPolicyStrict && sc.ReplayPolicy != ReplayPolicyBestEffort {
		return ErrInvalidReplayPolicy
	}
	if sc.TraceID == "" {
		return ErrEmptyShadowTraceID
	}
	return nil
}

// ShadowResultMetrics contains execution metrics for a shadow run.
type ShadowResultMetrics struct {
	TotalTokens           int     `json:"total_tokens"`
	PromptTokens          int     `json:"prompt_tokens"`
	CompletionTokens      int     `json:"completion_tokens"`
	EstimatedCostUSD      float64 `json:"estimated_cost_usd"`
	TotalDurationMs       int     `json:"total_duration_ms"`
	Iterations            int     `json:"iterations"`
	ToolCallsCount        int     `json:"tool_calls_count"`
	ToolErrorsCount       int     `json:"tool_errors_count"`
	MockedToolsCount      int     `json:"mocked_tools_count"`
	PassthroughToolsCount int     `json:"passthrough_tools_count"`
}

// ShadowResult is the payload published to the passflow.shadow.results stream
// by the agent-executor after completing a shadow execution.
type ShadowResult struct {
	SchemaVersion       string              `json:"schema_version"`
	ShadowExecutionID   string              `json:"shadow_execution_id"`
	OriginalExecutionID string              `json:"original_execution_id"`
	AnalysisBatchID     string              `json:"analysis_batch_id"`
	TenantID            string              `json:"tenant_id"`
	AgentID             string              `json:"agent_id"`
	AgentVersion        string              `json:"agent_version"`
	TraceID             string              `json:"trace_id"`
	IdempotencyKey      string              `json:"idempotency_key"`
	Status              string              `json:"status"`
	Output              string              `json:"output"`
	Metrics             ShadowResultMetrics `json:"metrics"`
	Steps               []json.RawMessage   `json:"steps"`
	Error               string              `json:"error,omitempty"`
	CreatedAt           time.Time           `json:"created_at"`
}

// Validate checks that ShadowResult has all required fields and valid values.
func (sr *ShadowResult) Validate() error {
	if sr.SchemaVersion == "" {
		return ErrEmptyResultSchemaVersion
	}
	if sr.ShadowExecutionID == "" {
		return ErrEmptyShadowExecutionID
	}
	if sr.OriginalExecutionID == "" {
		return ErrEmptyResultOriginalExecID
	}
	if sr.AnalysisBatchID == "" {
		return ErrEmptyResultAnalysisBatchID
	}
	if sr.TenantID == "" {
		return ErrEmptyResultTenantID
	}
	if sr.AgentID == "" {
		return ErrEmptyResultAgentID
	}
	if sr.AgentVersion == "" {
		return ErrEmptyResultAgentVersion
	}
	if sr.TraceID == "" {
		return ErrEmptyResultTraceID
	}
	if sr.IdempotencyKey == "" {
		return ErrEmptyResultIdempotencyKey
	}
	if sr.Status != ShadowStatusCompleted && sr.Status != ShadowStatusFailed && sr.Status != ShadowStatusAborted {
		return ErrInvalidShadowResultStatus
	}
	return nil
}

// IsValidExecutionMode checks if the given mode is a recognized execution mode.
func IsValidExecutionMode(mode string) bool {
	return mode == ExecutionModeNormal || mode == ExecutionModeShadow
}

// IsValidShadowStatus checks if the given status is a recognized shadow result status.
func IsValidShadowStatus(status string) bool {
	return status == ShadowStatusCompleted || status == ShadowStatusFailed || status == ShadowStatusAborted
}
