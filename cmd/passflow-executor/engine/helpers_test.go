package engine

import (
	"testing"

	"github.com/jaak-ai/passflow-agent-executor/llm"
)

func TestFindToolArgs_MatchesById(t *testing.T) {
	calls := []llm.ToolCall{
		{ID: "call_abc", Name: "tool_one", Arguments: `{"key":"value_one"}`},
		{ID: "call_xyz", Name: "tool_two", Arguments: `{"key":"value_two"}`},
	}

	got := findToolArgs(calls, "call_abc")
	if got != `{"key":"value_one"}` {
		t.Errorf("findToolArgs by ID = %q, want %q", got, `{"key":"value_one"}`)
	}
}

func TestFindToolArgs_SecondEntry(t *testing.T) {
	calls := []llm.ToolCall{
		{ID: "call_abc", Name: "tool_one", Arguments: `{"a":1}`},
		{ID: "call_xyz", Name: "tool_two", Arguments: `{"b":2}`},
	}

	got := findToolArgs(calls, "call_xyz")
	if got != `{"b":2}` {
		t.Errorf("findToolArgs by ID = %q, want %q", got, `{"b":2}`)
	}
}

func TestFindToolArgs_UnknownIdReturnsEmpty(t *testing.T) {
	calls := []llm.ToolCall{
		{ID: "call_abc", Name: "tool_one", Arguments: `{"key":"val"}`},
	}

	got := findToolArgs(calls, "call_does_not_exist")
	if got != "" {
		t.Errorf("findToolArgs with unknown ID = %q, want empty string", got)
	}
}

func TestFindToolArgs_EmptyCallsReturnsEmpty(t *testing.T) {
	got := findToolArgs(nil, "call_abc")
	if got != "" {
		t.Errorf("findToolArgs on nil calls = %q, want empty string", got)
	}
}

// TestFindToolArgs_DoesNotMatchByName verifies that passing a tool name
// (rather than an ID) does not accidentally match, confirming the function
// correctly keys on ID and the callers must supply ToolCallID, not ToolName.
func TestFindToolArgs_DoesNotMatchByName(t *testing.T) {
	calls := []llm.ToolCall{
		{ID: "call_abc", Name: "my_tool", Arguments: `{"x":1}`},
	}

	// Passing the Name instead of the ID must return empty — the function
	// only matches on ID.
	got := findToolArgs(calls, "my_tool")
	if got != "" {
		t.Errorf("findToolArgs matched by name instead of ID; got %q, want empty", got)
	}
}
