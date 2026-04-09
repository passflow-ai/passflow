package engine

import (
	"time"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
)

// DecisionLogEntry represents a single decision made by the LLM
// in the ReAct loop, including the reasoning, tool calls, and results.
type DecisionLogEntry struct {
	// Iteration is the ReAct loop iteration number (0-indexed)
	Iteration int `json:"iteration"`

	// Timestamp when this decision was made
	Timestamp time.Time `json:"timestamp"`

	// Prompt is the system + user messages sent to the LLM
	Prompt []llm.Message `json:"prompt"`

	// LLMResponse is the raw text response from the LLM
	LLMResponse string `json:"llm_response"`

	// ToolCalls are the tool invocations decided by the LLM
	ToolCalls []llm.ToolCall `json:"tool_calls,omitempty"`

	// ToolResults contains the results from executing the tool calls
	ToolResults []DecisionToolResult `json:"tool_results,omitempty"`

	// TokenUsage tracks token consumption for this decision
	TokenUsage llm.Usage `json:"token_usage"`

	// DurationMs is the time taken for the LLM call in milliseconds
	DurationMs int64 `json:"duration_ms"`
}

// DecisionToolResult represents the result of a single tool execution
// within a decision iteration.
type DecisionToolResult struct {
	// ToolCallID links this result to the corresponding ToolCall
	ToolCallID string `json:"tool_call_id"`

	// ToolName is the name of the tool that was executed
	ToolName string `json:"tool_name"`

	// Output is the tool's response (success or error message)
	Output string `json:"output"`

	// Error is set if the tool execution failed
	Error string `json:"error,omitempty"`

	// DurationMs is the time taken to execute the tool in milliseconds
	DurationMs int64 `json:"duration_ms"`
}

// DecisionLog maintains the complete decision-making trail
// for a single agent execution.
type DecisionLog struct {
	// ExecutionID uniquely identifies this agent execution
	ExecutionID string `json:"execution_id"`

	// AgentID identifies which agent made these decisions
	AgentID string `json:"agent_id"`

	// ModelProvider and ModelID identify the LLM used
	ModelProvider string `json:"model_provider"`
	ModelID       string `json:"model_id"`

	// Entries contains all decision iterations in chronological order
	Entries []DecisionLogEntry `json:"entries"`

	// CreatedAt marks when the decision log was started
	CreatedAt time.Time `json:"created_at"`

	// CompletedAt marks when the execution finished
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// NewDecisionLog creates a new decision log for an execution.
func NewDecisionLog(executionID, agentID, modelProvider, modelID string) *DecisionLog {
	return &DecisionLog{
		ExecutionID:   executionID,
		AgentID:       agentID,
		ModelProvider: modelProvider,
		ModelID:       modelID,
		Entries:       make([]DecisionLogEntry, 0),
		CreatedAt:     time.Now().UTC(),
	}
}

// AddDecision appends a new decision entry to the log.
func (dl *DecisionLog) AddDecision(entry DecisionLogEntry) {
	dl.Entries = append(dl.Entries, entry)
}

// Complete marks the decision log as finished.
func (dl *DecisionLog) Complete() {
	now := time.Now().UTC()
	dl.CompletedAt = &now
}

// TotalDecisions returns the number of decision iterations recorded.
func (dl *DecisionLog) TotalDecisions() int {
	return len(dl.Entries)
}

// TotalTokens calculates the cumulative token usage across all decisions.
func (dl *DecisionLog) TotalTokens() llm.Usage {
	var total llm.Usage
	for _, entry := range dl.Entries {
		total.PromptTokens += entry.TokenUsage.PromptTokens
		total.CompletionTokens += entry.TokenUsage.CompletionTokens
		total.TotalTokens += entry.TokenUsage.TotalTokens
	}
	return total
}

// TotalDurationMs returns the total LLM call time in milliseconds.
func (dl *DecisionLog) TotalDurationMs() int64 {
	var total int64
	for _, entry := range dl.Entries {
		total += entry.DurationMs
	}
	return total
}
