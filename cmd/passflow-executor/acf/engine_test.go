package acf

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// minimalSpec returns a valid job.Spec with a CADConfig set (IsACFAgent == true).
func minimalSpec(t *testing.T) *job.Spec {
	t.Helper()
	return &job.Spec{
		ExecutionID:   "exec-001",
		TaskID:        "task-001",
		WorkspaceID:   "ws-001",
		AgentID:       "agent-001",
		ModelProvider: "anthropic",
		ModelID:       "claude-3-5-sonnet",
		Input:         "Hello",
		Mode:          job.ModeReactLoop,
		CADConfig: &job.CADConfig{
			SystemPrompt: "You are a helpful assistant.",
			SourceType:   "claude_code",
			Skills: []job.SkillDef{
				{Name: "go-expert", Content: "You are an expert in Go."},
			},
			ContextFiles: map[string]string{
				"CLAUDE.md": "# Project\nThis is the project root.",
			},
		},
	}
}

// tempDir creates a temporary directory for the test and registers cleanup.
func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "acf-engine-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// ---------------------------------------------------------------------------
// PrepareACFExecution
// ---------------------------------------------------------------------------

func TestPrepareACFExecution_ValidCADConfig(t *testing.T) {
	spec := minimalSpec(t)
	dir := tempDir(t)

	ctx := PrepareACFExecution(spec, dir, nil)

	require.NotNil(t, ctx, "expected non-nil ACFContext for ACF agent")
	assert.NotEmpty(t, ctx.SystemPrompt, "system prompt should be populated")
	assert.NotNil(t, ctx.SystemTools, "SystemTools should be initialised")
	assert.NotNil(t, ctx.Auditor, "Auditor should be initialised")
	assert.NotNil(t, ctx.Translator, "Translator should be retained")
	assert.Equal(t, "anthropic", ctx.Provider)
	assert.Nil(t, ctx.Policy, "Policy should be nil when not set in spec")
	assert.Equal(t, spec.CADConfig, ctx.CADConfig)
}

func TestPrepareACFExecution_NilCADConfig(t *testing.T) {
	spec := &job.Spec{
		ExecutionID:   "exec-002",
		TaskID:        "task-002",
		WorkspaceID:   "ws-001",
		AgentID:       "agent-001",
		ModelProvider: "openai",
		ModelID:       "gpt-4o",
		Input:         "Hello",
		// CADConfig intentionally nil
	}

	result := PrepareACFExecution(spec, tempDir(t), nil)
	assert.Nil(t, result, "PrepareACFExecution should return nil when CADConfig is not set")
}

func TestPrepareACFExecution_WithPolicy(t *testing.T) {
	spec := minimalSpec(t)
	spec.Policy = &job.Policy{
		AllowedTools:   []string{"Read", "Grep", "Glob"},
		DeniedTools:    []string{"Bash"},
		MaxBashTimeout: 30,
		AllowNetwork:   false,
		AllowWrite:     false,
	}

	ctx := PrepareACFExecution(spec, tempDir(t), nil)

	require.NotNil(t, ctx)
	require.NotNil(t, ctx.Policy)
	assert.Equal(t, []string{"Read", "Grep", "Glob"}, ctx.Policy.AllowedTools)
	assert.Equal(t, []string{"Bash"}, ctx.Policy.DeniedTools)
	assert.Equal(t, 30, ctx.Policy.MaxBashTimeout)
	assert.False(t, ctx.Policy.AllowNetwork)
	assert.False(t, ctx.Policy.AllowWrite)
}

func TestPrepareACFExecution_SystemPromptContainsCADContent(t *testing.T) {
	spec := minimalSpec(t)
	ctx := PrepareACFExecution(spec, tempDir(t), nil)

	require.NotNil(t, ctx)
	// The translated prompt should embed the agent's base system prompt.
	assert.Contains(t, ctx.SystemPrompt, "You are a helpful assistant.")
	// It should also embed the skill content.
	assert.Contains(t, ctx.SystemPrompt, "You are an expert in Go.")
	// And the context file content.
	assert.Contains(t, ctx.SystemPrompt, "This is the project root.")
}

func TestPrepareACFExecution_AuditorLogsPromptTranslation(t *testing.T) {
	spec := minimalSpec(t)
	ctx := PrepareACFExecution(spec, tempDir(t), nil)

	require.NotNil(t, ctx)
	entries := ctx.Auditor.Entries()
	require.Len(t, entries, 1, "one prompt_translation entry expected after preparation")
	assert.Equal(t, "prompt_translation", entries[0].Type)
	assert.Equal(t, "claude_code", entries[0].Input["source_type"])
}

// ---------------------------------------------------------------------------
// GetSystemToolDefs
// ---------------------------------------------------------------------------

func TestGetSystemToolDefs_ReturnsAllSevenTools(t *testing.T) {
	ctx := PrepareACFExecution(minimalSpec(t), tempDir(t), nil)
	require.NotNil(t, ctx)

	defs := ctx.GetSystemToolDefs()

	assert.Len(t, defs, len(systemToolNames),
		"expected one definition per system tool name")

	// Build a name set for O(1) lookup.
	nameSet := make(map[string]bool, len(defs))
	for _, d := range defs {
		nameSet[d.Name] = true
		assert.NotEmpty(t, d.Description, "tool %q must have a description", d.Name)
		assert.NotNil(t, d.Parameters, "tool %q must have a Parameters schema", d.Name)
	}

	for _, expected := range systemToolNames {
		assert.True(t, nameSet[expected], "expected tool %q in definitions", expected)
	}
}

func TestGetSystemToolDefs_ReturnsCopy(t *testing.T) {
	ctx := PrepareACFExecution(minimalSpec(t), tempDir(t), nil)
	require.NotNil(t, ctx)

	defs1 := ctx.GetSystemToolDefs()
	defs1[0].Name = "MUTATED"

	defs2 := ctx.GetSystemToolDefs()
	assert.NotEqual(t, "MUTATED", defs2[0].Name,
		"mutating the returned slice must not affect subsequent calls")
}

func TestGetSystemToolDefs_ParameterSchemasDirectly(t *testing.T) {
	ctx := PrepareACFExecution(minimalSpec(t), tempDir(t), nil)
	require.NotNil(t, ctx)

	defs := ctx.GetSystemToolDefs()

	type defByName = map[string]map[string]interface{}
	paramsByName := make(defByName, len(defs))
	for _, d := range defs {
		paramsByName[d.Name] = d.Parameters
	}

	assertRequired := func(tool, param string) {
		t.Helper()
		params, ok := paramsByName[tool]
		require.True(t, ok, "tool %q not found", tool)
		req, ok := params["required"]
		require.True(t, ok, "tool %q: missing 'required' key", tool)
		reqs, ok := req.([]string)
		require.True(t, ok, "tool %q: 'required' not []string", tool)
		for _, r := range reqs {
			if r == param {
				return
			}
		}
		t.Errorf("tool %q: %q not in required %v", tool, param, reqs)
	}

	assertRequired("Bash", "command")
	assertRequired("Read", "file_path")
	assertRequired("Write", "file_path")
	assertRequired("Write", "content")
	assertRequired("Edit", "file_path")
	assertRequired("Edit", "old_string")
	assertRequired("Grep", "pattern")
	assertRequired("Glob", "pattern")
}

// ---------------------------------------------------------------------------
// ExecuteSystemTool
// ---------------------------------------------------------------------------

func TestExecuteSystemTool_DelegatesToSystemToolProvider(t *testing.T) {
	dir := tempDir(t)

	// Write a known file so "Read" has something to return.
	testFile := filepath.Join(dir, "hello.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("hello world"), 0644))

	ctx := PrepareACFExecution(minimalSpec(t), dir, nil)
	require.NotNil(t, ctx)

	call := ToolCall{
		Tool:   "Read",
		CallID: "call-read-1",
		Params: map[string]interface{}{
			"file_path": testFile,
		},
	}

	result := ctx.ExecuteSystemTool(context.Background(), call)

	require.NotNil(t, result)
	assert.Empty(t, result.Error, "reading an existing file should not error")
	assert.Contains(t, result.Output, "hello world")
}

func TestExecuteSystemTool_PropagatesError(t *testing.T) {
	dir := tempDir(t)
	ctx := PrepareACFExecution(minimalSpec(t), dir, nil)
	require.NotNil(t, ctx)

	call := ToolCall{
		Tool:   "Read",
		CallID: "call-read-missing",
		Params: map[string]interface{}{
			"file_path": filepath.Join(dir, "nonexistent.txt"),
		},
	}

	result := ctx.ExecuteSystemTool(context.Background(), call)

	require.NotNil(t, result)
	assert.NotEmpty(t, result.Error, "reading a missing file should produce an error")
}

func TestExecuteSystemTool_LogsViaAuditor(t *testing.T) {
	dir := tempDir(t)

	testFile := filepath.Join(dir, "data.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("content"), 0644))

	ctx := PrepareACFExecution(minimalSpec(t), dir, nil)
	require.NotNil(t, ctx)

	// Entries before the call: just the prompt_translation logged during Prepare.
	beforeCount := len(ctx.Auditor.Entries())

	call := ToolCall{
		Tool:   "Read",
		CallID: "call-audit-1",
		Params: map[string]interface{}{"file_path": testFile},
	}
	ctx.ExecuteSystemTool(context.Background(), call)

	entries := ctx.Auditor.Entries()
	// Expect two new entries: tool_call + tool_result
	assert.Len(t, entries, beforeCount+2,
		"ExecuteSystemTool should add a tool_call and a tool_result entry")

	assert.Equal(t, "tool_call", entries[beforeCount].Type)
	assert.Equal(t, "Read", entries[beforeCount].ToolName)

	assert.Equal(t, "tool_result", entries[beforeCount+1].Type)
	assert.Equal(t, "Read", entries[beforeCount+1].ToolName)
}

func TestExecuteSystemTool_PolicyDenialLoggedAsError(t *testing.T) {
	spec := minimalSpec(t)
	// Deny Bash via policy.
	spec.Policy = &job.Policy{DeniedTools: []string{"Bash"}}

	dir := tempDir(t)
	ctx := PrepareACFExecution(spec, dir, nil)
	require.NotNil(t, ctx)

	call := ToolCall{
		Tool:   "Bash",
		CallID: "call-bash-denied",
		Params: map[string]interface{}{"command": "echo hi"},
	}

	result := ctx.ExecuteSystemTool(context.Background(), call)

	require.NotNil(t, result)
	assert.NotEmpty(t, result.Error, "denied Bash call should return an error")
	assert.Contains(t, result.Error, "denied")

	// The tool_result audit entry should record the error.
	entries := ctx.Auditor.Entries()
	var resultEntry *AuditEntry
	for i := range entries {
		if entries[i].Type == "tool_result" && entries[i].ToolName == "Bash" {
			resultEntry = &entries[i]
			break
		}
	}
	require.NotNil(t, resultEntry, "expected a tool_result audit entry for the denied Bash call")
	assert.NotEmpty(t, resultEntry.Error)
}

// ---------------------------------------------------------------------------
// IsSystemTool
// ---------------------------------------------------------------------------

func TestIsSystemTool_RecognisesAllSystemTools(t *testing.T) {
	ctx := PrepareACFExecution(minimalSpec(t), tempDir(t), nil)
	require.NotNil(t, ctx)

	for _, name := range systemToolNames {
		assert.True(t, ctx.IsSystemTool(name), "expected %q to be recognised as a system tool", name)
	}
}

func TestIsSystemTool_ReturnsFalseForNonSystemTool(t *testing.T) {
	ctx := PrepareACFExecution(minimalSpec(t), tempDir(t), nil)
	require.NotNil(t, ctx)

	nonSystemTools := []string{
		"jira_update",
		"send_slack_message",
		"http_request",
		"bash", // case-sensitive
		"read", // case-sensitive
		"",
	}

	for _, name := range nonSystemTools {
		assert.False(t, ctx.IsSystemTool(name), "%q should NOT be a system tool", name)
	}
}

// ---------------------------------------------------------------------------
// jobCADToACFCAD (internal helper — tested via PrepareACFExecution)
// ---------------------------------------------------------------------------

func TestJobCADToACFCAD_NilInput(t *testing.T) {
	result := jobCADToACFCAD(nil)
	assert.Nil(t, result)
}

func TestJobCADToACFCAD_FullConversion(t *testing.T) {
	src := &job.CADConfig{
		SystemPrompt: "base prompt",
		SourceType:   "openai",
		SourceRef:    "gpt-pilot/v1",
		Skills: []job.SkillDef{
			{Name: "testing", Content: "write tests"},
		},
		ContextFiles: map[string]string{
			"README.md": "# Readme",
		},
	}

	dst := jobCADToACFCAD(src)

	require.NotNil(t, dst)
	assert.Equal(t, "base prompt", dst.SystemPrompt)
	assert.Equal(t, "openai", dst.SourceType)
	require.Len(t, dst.Skills, 1)
	assert.Equal(t, "testing", dst.Skills[0].Name)
	assert.Equal(t, "write tests", dst.Skills[0].Content)
	assert.Equal(t, "# Readme", dst.ContextFiles["README.md"])
}
