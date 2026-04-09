package tools

import (
	"testing"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/job"
)

// TestCanHandle_KnownTool verifies that CanHandle returns true for a tool
// that was registered with the executor.
func TestCanHandle_KnownTool(t *testing.T) {
	exec := New([]job.ToolConfig{
		{Name: "search", Description: "search tool", Type: "http",
			HTTPConfig: &job.HTTPToolConfig{URL: "https://example.com", Method: "GET"}},
		{Name: "notify", Description: "notify tool", Type: "http",
			HTTPConfig: &job.HTTPToolConfig{URL: "https://notify.example.com", Method: "POST"}},
	})

	if !exec.CanHandle("search") {
		t.Error("CanHandle(\"search\") = false, want true")
	}
	if !exec.CanHandle("notify") {
		t.Error("CanHandle(\"notify\") = false, want true")
	}
}

// TestCanHandle_UnknownTool verifies that CanHandle returns false for a tool
// that was NOT registered (e.g. an ACF system tool like "Bash").
func TestCanHandle_UnknownTool(t *testing.T) {
	exec := New([]job.ToolConfig{
		{Name: "search", Type: "http",
			HTTPConfig: &job.HTTPToolConfig{URL: "https://example.com", Method: "GET"}},
	})

	systemTools := []string{"Bash", "Read", "Write", "Edit", "Grep", "Glob", "LS"}
	for _, name := range systemTools {
		if exec.CanHandle(name) {
			t.Errorf("CanHandle(%q) = true for system tool, want false", name)
		}
	}
}

// TestCanHandle_EmptyExecutor verifies that an executor with no tools returns
// false for any name.
func TestCanHandle_EmptyExecutor(t *testing.T) {
	exec := New(nil)
	if exec.CanHandle("anything") {
		t.Error("CanHandle on empty executor should return false")
	}
}

// TestCanHandle_CaseSensitive verifies that CanHandle is case-sensitive,
// matching the map key exactly.
func TestCanHandle_CaseSensitive(t *testing.T) {
	exec := New([]job.ToolConfig{
		{Name: "MyTool", Type: "http",
			HTTPConfig: &job.HTTPToolConfig{URL: "https://example.com", Method: "GET"}},
	})

	if !exec.CanHandle("MyTool") {
		t.Error("CanHandle(\"MyTool\") = false, want true")
	}
	if exec.CanHandle("mytool") {
		t.Error("CanHandle(\"mytool\") = true, want false (case-sensitive)")
	}
	if exec.CanHandle("MYTOOL") {
		t.Error("CanHandle(\"MYTOOL\") = true, want false (case-sensitive)")
	}
}

// TestCanHandle_EmptyString verifies that an empty tool name returns false.
func TestCanHandle_EmptyString(t *testing.T) {
	exec := New([]job.ToolConfig{
		{Name: "some_tool", Type: "http",
			HTTPConfig: &job.HTTPToolConfig{URL: "https://example.com", Method: "GET"}},
	})
	if exec.CanHandle("") {
		t.Error("CanHandle(\"\") = true, want false")
	}
}
