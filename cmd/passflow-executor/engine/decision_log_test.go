package engine

import (
	"testing"
	"time"

	"github.com/jaak-ai/passflow-agent-executor/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDecisionLog(t *testing.T) {
	// Given: execution parameters
	executionID := "exec-123"
	agentID := "agent-456"
	modelProvider := "anthropic"
	modelID := "claude-3-5-sonnet"

	// When: creating a new decision log
	log := NewDecisionLog(executionID, agentID, modelProvider, modelID)

	// Then: log should be initialized with correct values
	require.NotNil(t, log)
	assert.Equal(t, executionID, log.ExecutionID)
	assert.Equal(t, agentID, log.AgentID)
	assert.Equal(t, modelProvider, log.ModelProvider)
	assert.Equal(t, modelID, log.ModelID)
	assert.Empty(t, log.Entries)
	assert.Nil(t, log.CompletedAt)
	assert.WithinDuration(t, time.Now().UTC(), log.CreatedAt, time.Second)
}

func TestDecisionLog_AddDecision(t *testing.T) {
	// Given: a decision log
	log := NewDecisionLog("exec-123", "agent-456", "openai", "gpt-4")

	// When: adding a decision entry
	entry := DecisionLogEntry{
		Iteration:   0,
		Timestamp:   time.Now().UTC(),
		LLMResponse: "I will search for the file",
		TokenUsage: llm.Usage{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		},
		DurationMs: 250,
	}
	log.AddDecision(entry)

	// Then: entry should be added to the log
	assert.Len(t, log.Entries, 1)
	assert.Equal(t, entry, log.Entries[0])
}

func TestDecisionLog_AddMultipleDecisions(t *testing.T) {
	// Given: a decision log
	log := NewDecisionLog("exec-123", "agent-456", "anthropic", "claude-3-5-sonnet")

	// When: adding multiple decision entries
	for i := 0; i < 3; i++ {
		entry := DecisionLogEntry{
			Iteration:   i,
			Timestamp:   time.Now().UTC(),
			LLMResponse: "Decision " + string(rune(i)),
			DurationMs:  int64(100 * (i + 1)),
		}
		log.AddDecision(entry)
	}

	// Then: all entries should be in order
	assert.Len(t, log.Entries, 3)
	for i := 0; i < 3; i++ {
		assert.Equal(t, i, log.Entries[i].Iteration)
	}
}

func TestDecisionLog_Complete(t *testing.T) {
	// Given: a decision log
	log := NewDecisionLog("exec-123", "agent-456", "openai", "gpt-4")
	require.Nil(t, log.CompletedAt)

	// When: completing the log
	log.Complete()

	// Then: CompletedAt should be set
	require.NotNil(t, log.CompletedAt)
	assert.WithinDuration(t, time.Now().UTC(), *log.CompletedAt, time.Second)
}

func TestDecisionLog_TotalDecisions(t *testing.T) {
	// Given: a decision log with multiple entries
	log := NewDecisionLog("exec-123", "agent-456", "anthropic", "claude-3-5-sonnet")

	assert.Equal(t, 0, log.TotalDecisions())

	// When: adding entries
	for i := 0; i < 5; i++ {
		log.AddDecision(DecisionLogEntry{
			Iteration: i,
			Timestamp: time.Now().UTC(),
		})
	}

	// Then: total should match
	assert.Equal(t, 5, log.TotalDecisions())
}

func TestDecisionLog_TotalTokens(t *testing.T) {
	// Given: a decision log with entries that have token usage
	log := NewDecisionLog("exec-123", "agent-456", "openai", "gpt-4")

	log.AddDecision(DecisionLogEntry{
		Iteration: 0,
		Timestamp: time.Now().UTC(),
		TokenUsage: llm.Usage{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		},
	})

	log.AddDecision(DecisionLogEntry{
		Iteration: 1,
		Timestamp: time.Now().UTC(),
		TokenUsage: llm.Usage{
			PromptTokens:     200,
			CompletionTokens: 100,
			TotalTokens:      300,
		},
	})

	// When: calculating total tokens
	total := log.TotalTokens()

	// Then: tokens should be summed correctly
	assert.Equal(t, 300, total.PromptTokens)
	assert.Equal(t, 150, total.CompletionTokens)
	assert.Equal(t, 450, total.TotalTokens)
}

func TestDecisionLog_TotalDurationMs(t *testing.T) {
	// Given: a decision log with timed entries
	log := NewDecisionLog("exec-123", "agent-456", "anthropic", "claude-3-5-sonnet")

	log.AddDecision(DecisionLogEntry{
		Iteration:  0,
		Timestamp:  time.Now().UTC(),
		DurationMs: 250,
	})

	log.AddDecision(DecisionLogEntry{
		Iteration:  1,
		Timestamp:  time.Now().UTC(),
		DurationMs: 380,
	})

	log.AddDecision(DecisionLogEntry{
		Iteration:  2,
		Timestamp:  time.Now().UTC(),
		DurationMs: 120,
	})

	// When: calculating total duration
	totalDuration := log.TotalDurationMs()

	// Then: durations should be summed
	assert.Equal(t, int64(750), totalDuration)
}

func TestDecisionLog_EmptyLog(t *testing.T) {
	// Given: an empty decision log
	log := NewDecisionLog("exec-123", "agent-456", "openai", "gpt-4")

	// Then: totals should be zero
	assert.Equal(t, 0, log.TotalDecisions())
	assert.Equal(t, int64(0), log.TotalDurationMs())

	totalTokens := log.TotalTokens()
	assert.Equal(t, 0, totalTokens.PromptTokens)
	assert.Equal(t, 0, totalTokens.CompletionTokens)
	assert.Equal(t, 0, totalTokens.TotalTokens)
}

func TestDecisionLogEntry_WithToolCalls(t *testing.T) {
	// Given: a decision entry with tool calls
	entry := DecisionLogEntry{
		Iteration:   0,
		Timestamp:   time.Now().UTC(),
		LLMResponse: "I will read the file",
		ToolCalls: []llm.ToolCall{
			{
				ID:        "call-1",
				Name:      "Read",
				Arguments: `{"file_path": "/tmp/test.txt"}`,
			},
		},
		ToolResults: []DecisionToolResult{
			{
				ToolCallID: "call-1",
				ToolName:   "Read",
				Output:     "File contents here",
				DurationMs: 50,
			},
		},
	}

	// Then: entry should have tool calls and results
	assert.Len(t, entry.ToolCalls, 1)
	assert.Equal(t, "Read", entry.ToolCalls[0].Name)
	assert.Len(t, entry.ToolResults, 1)
	assert.Equal(t, "call-1", entry.ToolResults[0].ToolCallID)
}

func TestDecisionToolResult_WithError(t *testing.T) {
	// Given: a tool result with an error
	result := DecisionToolResult{
		ToolCallID: "call-1",
		ToolName:   "Bash",
		Output:     "",
		Error:      "command not found",
		DurationMs: 10,
	}

	// Then: error should be captured
	assert.NotEmpty(t, result.Error)
	assert.Equal(t, "command not found", result.Error)
	assert.Empty(t, result.Output)
}
