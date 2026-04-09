package acf

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AuditEntry represents a single audited operation.
type AuditEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	ExecutionID string                 `json:"execution_id"`
	Type        string                 `json:"type"` // "tool_call", "tool_result", "policy_denied", "prompt_translation"
	Provider    string                 `json:"provider"`
	ToolName    string                 `json:"tool_name,omitempty"`
	Input       map[string]interface{} `json:"input,omitempty"`
	Output      string                 `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
}

// ExecutionAuditor collects audit entries during an execution.
// It is safe for concurrent use.
type ExecutionAuditor struct {
	executionID string
	provider    string
	entries     []AuditEntry
	mu          sync.Mutex
	logger      *zap.Logger // optional, can be nil
}

// NewExecutionAuditor creates a new auditor for the given execution.
// logger may be nil; when nil, structured logging is skipped.
func NewExecutionAuditor(executionID, provider string, logger *zap.Logger) *ExecutionAuditor {
	return &ExecutionAuditor{
		executionID: executionID,
		provider:    provider,
		entries:     make([]AuditEntry, 0),
		logger:      logger,
	}
}

// LogToolCall records a tool invocation attempt.
func (a *ExecutionAuditor) LogToolCall(toolName string, params map[string]interface{}) {
	entry := AuditEntry{
		Timestamp:   time.Now().UTC(),
		ExecutionID: a.executionID,
		Type:        "tool_call",
		Provider:    a.provider,
		ToolName:    toolName,
		Input:       params,
	}

	a.mu.Lock()
	a.entries = append(a.entries, entry)
	a.mu.Unlock()

	if a.logger != nil {
		a.logger.Info("tool_call",
			zap.String("execution_id", a.executionID),
			zap.String("provider", a.provider),
			zap.String("tool", toolName),
		)
	}
}

// LogToolResult records the result of a tool execution.
func (a *ExecutionAuditor) LogToolResult(toolName, output, errMsg string, duration time.Duration) {
	entry := AuditEntry{
		Timestamp:   time.Now().UTC(),
		ExecutionID: a.executionID,
		Type:        "tool_result",
		Provider:    a.provider,
		ToolName:    toolName,
		Output:      output,
		Error:       errMsg,
		Duration:    duration,
	}

	a.mu.Lock()
	a.entries = append(a.entries, entry)
	a.mu.Unlock()

	if a.logger != nil {
		fields := []zap.Field{
			zap.String("execution_id", a.executionID),
			zap.String("provider", a.provider),
			zap.String("tool", toolName),
			zap.Duration("duration", duration),
		}
		if errMsg != "" {
			fields = append(fields, zap.String("error", errMsg))
			a.logger.Warn("tool_result", fields...)
		} else {
			a.logger.Info("tool_result", fields...)
		}
	}
}

// LogPolicyDenied records a policy denial.
func (a *ExecutionAuditor) LogPolicyDenied(toolName string, reason string) {
	entry := AuditEntry{
		Timestamp:   time.Now().UTC(),
		ExecutionID: a.executionID,
		Type:        "policy_denied",
		Provider:    a.provider,
		ToolName:    toolName,
		Error:       reason,
	}

	a.mu.Lock()
	a.entries = append(a.entries, entry)
	a.mu.Unlock()

	if a.logger != nil {
		a.logger.Warn("policy_denied",
			zap.String("execution_id", a.executionID),
			zap.String("provider", a.provider),
			zap.String("tool", toolName),
			zap.String("reason", reason),
		)
	}
}

// LogPromptTranslation records a prompt translation event.
func (a *ExecutionAuditor) LogPromptTranslation(sourceType, provider string, promptLen int) {
	entry := AuditEntry{
		Timestamp:   time.Now().UTC(),
		ExecutionID: a.executionID,
		Type:        "prompt_translation",
		Provider:    provider,
		Input: map[string]interface{}{
			"source_type": sourceType,
			"prompt_len":  promptLen,
		},
	}

	a.mu.Lock()
	a.entries = append(a.entries, entry)
	a.mu.Unlock()

	if a.logger != nil {
		a.logger.Info("prompt_translation",
			zap.String("execution_id", a.executionID),
			zap.String("source_type", sourceType),
			zap.String("provider", provider),
			zap.Int("prompt_len", promptLen),
		)
	}
}

// Entries returns a snapshot of all collected audit entries in insertion order.
func (a *ExecutionAuditor) Entries() []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	snapshot := make([]AuditEntry, len(a.entries))
	copy(snapshot, a.entries)
	return snapshot
}

// Summary returns a human-readable summary of the audit trail.
func (a *ExecutionAuditor) Summary() string {
	a.mu.Lock()
	entries := make([]AuditEntry, len(a.entries))
	copy(entries, a.entries)
	a.mu.Unlock()

	var (
		toolCalls    int
		toolResults  int
		policyDenied int
		translations int
		errors       int
		totalDur     time.Duration
	)

	for _, e := range entries {
		switch e.Type {
		case "tool_call":
			toolCalls++
		case "tool_result":
			toolResults++
			totalDur += e.Duration
			if e.Error != "" {
				errors++
			}
		case "policy_denied":
			policyDenied++
		case "prompt_translation":
			translations++
		}
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "ExecutionAuditor summary [execution_id=%s provider=%s]\n", a.executionID, a.provider)
	fmt.Fprintf(&sb, "  total entries    : %d\n", len(entries))
	fmt.Fprintf(&sb, "  tool_call        : %d\n", toolCalls)
	fmt.Fprintf(&sb, "  tool_result      : %d (errors: %d, total_duration: %s)\n", toolResults, errors, totalDur)
	fmt.Fprintf(&sb, "  policy_denied    : %d\n", policyDenied)
	fmt.Fprintf(&sb, "  prompt_translation: %d\n", translations)
	return sb.String()
}
