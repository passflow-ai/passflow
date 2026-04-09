package engine

import "time"

// StepResult is a single step in the execution (LLM call or tool call).
type StepResult struct {
	Index     int       `json:"index"`
	Type      string    `json:"type"`  // llm_call | tool_call
	Name      string    `json:"name"`  // tool name or "llm"
	Input     string    `json:"input"` // what was sent
	Output    string    `json:"output"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ExecutionResult is the full result of an agent run.
type ExecutionResult struct {
	ExecutionID string            `json:"execution_id"`
	AgentID     string            `json:"agent_id"`
	TaskID      string            `json:"task_id"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Status      string            `json:"status"` // completed | failed | max_iterations
	Output      string            `json:"output"` // final LLM response
	Steps       []StepResult      `json:"steps"`
	Usage       Usage             `json:"usage"`
	Error       string            `json:"error,omitempty"`
	CompletedAt time.Time         `json:"completed_at"`
	// DecisionLog contains the structured decision-making trail for observability
	DecisionLog *DecisionLog `json:"decision_log,omitempty"`
}

// Usage aggregates token consumption across all LLM calls.
type Usage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	EstimatedCost    float64 `json:"estimated_cost"`
	LLMLatencyMs     int64   `json:"llm_latency_ms,omitempty"`
	LLMCalls         int     `json:"llm_calls,omitempty"`
	ToolCalls        int     `json:"tool_calls,omitempty"`
	ToolErrors       int     `json:"tool_errors,omitempty"`
}
