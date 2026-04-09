package acf

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewExecutionAuditor verifies the auditor is initialised with empty entries.
func TestNewExecutionAuditor(t *testing.T) {
	a := NewExecutionAuditor("exec-1", "anthropic", nil)
	require.NotNil(t, a)
	assert.Equal(t, "exec-1", a.executionID)
	assert.Equal(t, "anthropic", a.provider)
	assert.Empty(t, a.Entries())
}

// TestLogToolCall_CreatesEntry verifies a tool_call entry is appended with correct fields.
func TestLogToolCall_CreatesEntry(t *testing.T) {
	a := NewExecutionAuditor("exec-2", "openai", nil)

	params := map[string]interface{}{"file_path": "/tmp/foo.go"}
	a.LogToolCall("Read", params)

	entries := a.Entries()
	require.Len(t, entries, 1)

	e := entries[0]
	assert.Equal(t, "tool_call", e.Type)
	assert.Equal(t, "exec-2", e.ExecutionID)
	assert.Equal(t, "openai", e.Provider)
	assert.Equal(t, "Read", e.ToolName)
	assert.Equal(t, "/tmp/foo.go", e.Input["file_path"])
	assert.WithinDuration(t, time.Now().UTC(), e.Timestamp, 2*time.Second)
}

// TestLogToolResult_IncludesDuration verifies duration and output are recorded correctly.
func TestLogToolResult_IncludesDuration(t *testing.T) {
	a := NewExecutionAuditor("exec-3", "gemini", nil)

	duration := 250 * time.Millisecond
	a.LogToolResult("Bash", "hello world", "", duration)

	entries := a.Entries()
	require.Len(t, entries, 1)

	e := entries[0]
	assert.Equal(t, "tool_result", e.Type)
	assert.Equal(t, "Bash", e.ToolName)
	assert.Equal(t, "hello world", e.Output)
	assert.Empty(t, e.Error)
	assert.Equal(t, duration, e.Duration)
}

// TestLogToolResult_WithError verifies that an error message is stored in the entry.
func TestLogToolResult_WithError(t *testing.T) {
	a := NewExecutionAuditor("exec-4", "anthropic", nil)

	a.LogToolResult("Write", "", "permission denied", 10*time.Millisecond)

	entries := a.Entries()
	require.Len(t, entries, 1)

	e := entries[0]
	assert.Equal(t, "tool_result", e.Type)
	assert.Equal(t, "permission denied", e.Error)
	assert.Equal(t, 10*time.Millisecond, e.Duration)
}

// TestLogPolicyDenied_RecordsReason verifies the policy_denied type and reason field.
func TestLogPolicyDenied_RecordsReason(t *testing.T) {
	a := NewExecutionAuditor("exec-5", "openai", nil)

	a.LogPolicyDenied("Bash", "bash execution not allowed by workspace policy")

	entries := a.Entries()
	require.Len(t, entries, 1)

	e := entries[0]
	assert.Equal(t, "policy_denied", e.Type)
	assert.Equal(t, "Bash", e.ToolName)
	assert.Equal(t, "bash execution not allowed by workspace policy", e.Error)
	assert.Equal(t, "exec-5", e.ExecutionID)
}

// TestLogPromptTranslation_RecordsEvent verifies prompt_translation is captured with metadata.
func TestLogPromptTranslation_RecordsEvent(t *testing.T) {
	a := NewExecutionAuditor("exec-6", "gemini", nil)

	a.LogPromptTranslation("claude_code", "gemini", 1024)

	entries := a.Entries()
	require.Len(t, entries, 1)

	e := entries[0]
	assert.Equal(t, "prompt_translation", e.Type)
	assert.Equal(t, "gemini", e.Provider)
	assert.Equal(t, "claude_code", e.Input["source_type"])
	assert.Equal(t, 1024, e.Input["prompt_len"])
}

// TestEntries_ReturnsAllInOrder verifies that multiple entries are returned in insertion order.
func TestEntries_ReturnsAllInOrder(t *testing.T) {
	a := NewExecutionAuditor("exec-7", "anthropic", nil)

	a.LogToolCall("Read", nil)
	a.LogToolResult("Read", "contents", "", 5*time.Millisecond)
	a.LogPolicyDenied("Bash", "denied")
	a.LogPromptTranslation("openai", "anthropic", 512)

	entries := a.Entries()
	require.Len(t, entries, 4)

	assert.Equal(t, "tool_call", entries[0].Type)
	assert.Equal(t, "tool_result", entries[1].Type)
	assert.Equal(t, "policy_denied", entries[2].Type)
	assert.Equal(t, "prompt_translation", entries[3].Type)
}

// TestEntries_ReturnsCopy verifies that mutating the returned slice does not affect the auditor.
func TestEntries_ReturnsCopy(t *testing.T) {
	a := NewExecutionAuditor("exec-8", "openai", nil)
	a.LogToolCall("Glob", nil)

	snapshot := a.Entries()
	require.Len(t, snapshot, 1)

	// Mutate the snapshot.
	snapshot[0].Type = "tampered"

	// Original must be unchanged.
	assert.Equal(t, "tool_call", a.Entries()[0].Type)
}

// TestConcurrentWrites verifies thread safety under concurrent logging.
func TestConcurrentWrites(t *testing.T) {
	a := NewExecutionAuditor("exec-concurrent", "anthropic", nil)

	const goroutines = 50
	const callsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				a.LogToolCall("Read", map[string]interface{}{"worker": id, "iter": j})
				a.LogToolResult("Read", "ok", "", time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	entries := a.Entries()
	// Each goroutine logs 2 entries per iteration.
	assert.Len(t, entries, goroutines*callsPerGoroutine*2)
}

// TestSummary_Format verifies the summary contains expected section labels and counts.
func TestSummary_Format(t *testing.T) {
	a := NewExecutionAuditor("exec-summary", "openai", nil)

	a.LogToolCall("Read", nil)
	a.LogToolResult("Read", "data", "", 100*time.Millisecond)
	a.LogToolCall("Bash", nil)
	a.LogToolResult("Bash", "", "exit status 1", 200*time.Millisecond)
	a.LogPolicyDenied("Write", "write not allowed")
	a.LogPromptTranslation("claude_code", "openai", 2048)

	summary := a.Summary()

	assert.Contains(t, summary, "exec-summary")
	assert.Contains(t, summary, "openai")
	assert.Contains(t, summary, "total entries")
	assert.Contains(t, summary, "tool_call")
	assert.Contains(t, summary, "tool_result")
	assert.Contains(t, summary, "policy_denied")
	assert.Contains(t, summary, "prompt_translation")
	// 6 total entries logged.
	assert.Contains(t, summary, "6")
	// 1 tool_result error.
	assert.Contains(t, summary, "errors: 1")
}

// TestSummary_EmptyAuditor verifies summary works with no entries.
func TestSummary_EmptyAuditor(t *testing.T) {
	a := NewExecutionAuditor("exec-empty", "anthropic", nil)
	summary := a.Summary()
	assert.Contains(t, summary, "exec-empty")
	assert.Contains(t, summary, "total entries    : 0")
}
