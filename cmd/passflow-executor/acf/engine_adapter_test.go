package acf

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// NewEngineAdapter
// ---------------------------------------------------------------------------

func TestNewEngineAdapter_NilACFContext(t *testing.T) {
	adapter := NewEngineAdapter(nil)
	assert.Nil(t, adapter, "NewEngineAdapter(nil) must return nil")
}

func TestNewEngineAdapter_NonNilACFContext(t *testing.T) {
	ctx := PrepareACFExecution(minimalSpec(t), tempDir(t), nil)
	require.NotNil(t, ctx)

	adapter := NewEngineAdapter(ctx)
	assert.NotNil(t, adapter, "NewEngineAdapter with a valid ACFContext must return non-nil")
}

// ---------------------------------------------------------------------------
// IsSystemTool
// ---------------------------------------------------------------------------

func TestEngineAdapter_IsSystemTool_RecognisedNames(t *testing.T) {
	adapter := NewEngineAdapter(PrepareACFExecution(minimalSpec(t), tempDir(t), nil))
	require.NotNil(t, adapter)

	for _, name := range systemToolNames {
		assert.True(t, adapter.IsSystemTool(name),
			"adapter.IsSystemTool(%q) should return true", name)
	}
}

func TestEngineAdapter_IsSystemTool_UnknownName(t *testing.T) {
	adapter := NewEngineAdapter(PrepareACFExecution(minimalSpec(t), tempDir(t), nil))
	require.NotNil(t, adapter)

	assert.False(t, adapter.IsSystemTool("http_request"))
	assert.False(t, adapter.IsSystemTool("bash")) // case-sensitive
	assert.False(t, adapter.IsSystemTool(""))
}

// ---------------------------------------------------------------------------
// GetSystemToolDefs
// ---------------------------------------------------------------------------

func TestEngineAdapter_GetSystemToolDefs_MatchesACFContext(t *testing.T) {
	acfCtx := PrepareACFExecution(minimalSpec(t), tempDir(t), nil)
	require.NotNil(t, acfCtx)

	adapter := NewEngineAdapter(acfCtx)
	adapterDefs := adapter.GetSystemToolDefs()
	directDefs := acfCtx.GetSystemToolDefs()

	require.Len(t, adapterDefs, len(directDefs))
	for i := range adapterDefs {
		assert.Equal(t, directDefs[i].Name, adapterDefs[i].Name)
		assert.Equal(t, directDefs[i].Description, adapterDefs[i].Description)
	}
}

// ---------------------------------------------------------------------------
// Execute — type conversion and delegation
// ---------------------------------------------------------------------------

func TestEngineAdapter_Execute_RoutesToSystemTool(t *testing.T) {
	dir := tempDir(t)
	testFile := filepath.Join(dir, "hello.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("adapter test content"), 0644))

	acfCtx := PrepareACFExecution(minimalSpec(t), dir, nil)
	require.NotNil(t, acfCtx)

	adapter := NewEngineAdapter(acfCtx)

	call := llm.ToolCall{
		ID:        "call-read-adapter",
		Name:      "Read",
		Arguments: `{"file_path":"` + testFile + `"}`,
	}

	result := adapter.Execute(context.Background(), call)

	assert.Equal(t, "call-read-adapter", result.ToolCallID)
	assert.Equal(t, "Read", result.ToolName)
	assert.Empty(t, result.Error, "reading an existing file must not produce an error")
	assert.Contains(t, result.Output, "adapter test content")
}

func TestEngineAdapter_Execute_PropagatesErrorInOutput(t *testing.T) {
	dir := tempDir(t)
	acfCtx := PrepareACFExecution(minimalSpec(t), dir, nil)
	require.NotNil(t, acfCtx)

	adapter := NewEngineAdapter(acfCtx)

	call := llm.ToolCall{
		ID:        "call-read-missing",
		Name:      "Read",
		Arguments: `{"file_path":"/nonexistent/path/file.txt"}`,
	}

	result := adapter.Execute(context.Background(), call)

	assert.Equal(t, "call-read-missing", result.ToolCallID)
	assert.Equal(t, "Read", result.ToolName)
	assert.NotEmpty(t, result.Error, "reading a missing file must surface an error")
}

func TestEngineAdapter_Execute_MalformedArgumentsYieldsError(t *testing.T) {
	dir := tempDir(t)
	acfCtx := PrepareACFExecution(minimalSpec(t), dir, nil)
	require.NotNil(t, acfCtx)

	adapter := NewEngineAdapter(acfCtx)

	// Malformed JSON — the adapter must not panic; the tool returns an error.
	call := llm.ToolCall{
		ID:        "call-bad-args",
		Name:      "Read",
		Arguments: `{not valid json`,
	}

	result := adapter.Execute(context.Background(), call)

	// The call should complete without panic; the SystemToolProvider will
	// return an error because file_path is missing from the empty params map.
	assert.Equal(t, "call-bad-args", result.ToolCallID)
	assert.NotEmpty(t, result.Error)
}

func TestEngineAdapter_Execute_EmptyArgumentsHandled(t *testing.T) {
	dir := tempDir(t)
	acfCtx := PrepareACFExecution(minimalSpec(t), dir, nil)
	require.NotNil(t, acfCtx)

	adapter := NewEngineAdapter(acfCtx)

	// Empty arguments string — adapter must not panic.
	call := llm.ToolCall{
		ID:        "call-empty-args",
		Name:      "Read",
		Arguments: "",
	}

	result := adapter.Execute(context.Background(), call)
	// No file_path → SystemToolProvider returns an error, but no crash.
	assert.Equal(t, "call-empty-args", result.ToolCallID)
	assert.NotEmpty(t, result.Error)
}

func TestEngineAdapter_Execute_AuditorRecordsCall(t *testing.T) {
	dir := tempDir(t)
	testFile := filepath.Join(dir, "audit.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("audit content"), 0644))

	acfCtx := PrepareACFExecution(minimalSpec(t), dir, nil)
	require.NotNil(t, acfCtx)

	adapter := NewEngineAdapter(acfCtx)
	beforeCount := len(acfCtx.Auditor.Entries())

	call := llm.ToolCall{
		ID:        "call-audit-adapter",
		Name:      "Read",
		Arguments: `{"file_path":"` + testFile + `"}`,
	}
	adapter.Execute(context.Background(), call)

	entries := acfCtx.Auditor.Entries()
	// Expect tool_call + tool_result = 2 new entries.
	assert.Len(t, entries, beforeCount+2)
}
