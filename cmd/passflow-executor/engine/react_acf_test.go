package engine

import (
	"context"
	"strings"
	"testing"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/job"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/tools"
)

// ---------------------------------------------------------------------------
// Stubs
// ---------------------------------------------------------------------------

// stubLLMClient returns a canned response sequence. Each Complete call pops
// the front of the responses slice; if the slice is exhausted it returns the
// last element.
type stubLLMClient struct {
	responses []*llm.Response
	callCount int
}

func (s *stubLLMClient) Complete(_ context.Context, _ []llm.Message, _ []llm.ToolDefinition) (*llm.Response, error) {
	idx := s.callCount
	if idx >= len(s.responses) {
		idx = len(s.responses) - 1
	}
	s.callCount++
	return s.responses[idx], nil
}

// stubSystemToolRouter is a test double for engine.SystemToolRouter.
type stubSystemToolRouter struct {
	systemTools map[string]bool
	defs        []llm.ToolDefinition
	// executed records calls in order: "toolName:args"
	executed []string
}

func (s *stubSystemToolRouter) IsSystemTool(name string) bool {
	return s.systemTools[name]
}

func (s *stubSystemToolRouter) GetSystemToolDefs() []llm.ToolDefinition {
	return s.defs
}

func (s *stubSystemToolRouter) Execute(_ context.Context, call llm.ToolCall) tools.Result {
	s.executed = append(s.executed, call.Name+":"+call.Arguments)
	return tools.Result{
		ToolCallID: call.ID,
		ToolName:   call.Name,
		Output:     `{"result":"system tool executed: ` + call.Name + `"}`,
	}
}

// minimalACFSpec returns a valid Spec with a CADConfig so IsACFAgent() == true.
func minimalACFSpec() *job.Spec {
	return &job.Spec{
		TaskID:        "task-acf",
		ExecutionID:   "exec-acf",
		WorkspaceID:   "ws-1",
		AgentID:       "agent-acf",
		ModelProvider: "anthropic",
		ModelID:       "claude-3-5-sonnet",
		Mode:          job.ModeReactLoop,
		MaxIterations: 5,
		Input:         "list the files",
		SystemPrompt:  "original system prompt",
		CADConfig: &job.CADConfig{
			SystemPrompt: "CAD system prompt",
			SourceType:   "claude_code",
		},
	}
}

// minimalSpec returns a plain (non-ACF) spec.
func minimalSpec() *job.Spec {
	return &job.Spec{
		TaskID:        "task-plain",
		ExecutionID:   "exec-plain",
		WorkspaceID:   "ws-1",
		AgentID:       "agent-plain",
		ModelProvider: "openai",
		ModelID:       "gpt-4o",
		Mode:          job.ModeReactLoop,
		MaxIterations: 5,
		Input:         "hello",
		SystemPrompt:  "be helpful",
	}
}

// ---------------------------------------------------------------------------
// Options zero value — backward compatibility
// ---------------------------------------------------------------------------

// TestRunReactLoop_NoACF_BackwardCompat verifies that existing callers that
// pass no Options see the same behaviour as before: the loop completes using
// only the regular toolExecutor and spec.SystemPrompt.
func TestRunReactLoop_NoACF_BackwardCompat(t *testing.T) {
	spec := minimalSpec()
	llmClient := &stubLLMClient{
		responses: []*llm.Response{
			{Content: "hello back"}, // no tool calls → loop exits after one iteration
		},
	}
	exec := tools.New(nil)

	result, err := RunReactLoop(context.Background(), spec, llmClient, exec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "completed" {
		t.Errorf("status = %q, want %q", result.Status, "completed")
	}
	if result.Output != "hello back" {
		t.Errorf("output = %q, want %q", result.Output, "hello back")
	}
}

// ---------------------------------------------------------------------------
// SystemPromptOverride
// ---------------------------------------------------------------------------

// TestRunReactLoop_ACF_SystemPromptOverride verifies that when
// Options.SystemPromptOverride is set, the loop uses that prompt instead of
// spec.SystemPrompt. We verify indirectly: the spec's original SystemPrompt
// is "original system prompt", the override is "ACF translated prompt", and
// the original spec must not be mutated.
func TestRunReactLoop_ACF_SystemPromptOverride(t *testing.T) {
	spec := minimalACFSpec()
	originalPrompt := spec.SystemPrompt

	llmClient := &stubLLMClient{
		responses: []*llm.Response{{Content: "done"}},
	}
	exec := tools.New(nil)

	opts := Options{SystemPromptOverride: "ACF translated prompt"}
	result, err := RunReactLoop(context.Background(), spec, llmClient, exec, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "completed" {
		t.Errorf("status = %q, want %q", result.Status, "completed")
	}
	// The original spec must not have been mutated.
	if spec.SystemPrompt != originalPrompt {
		t.Errorf("spec.SystemPrompt was mutated: got %q, want %q", spec.SystemPrompt, originalPrompt)
	}
}

// ---------------------------------------------------------------------------
// System tool routing
// ---------------------------------------------------------------------------

// TestRunReactLoop_ACF_SystemToolRoutedThroughACF verifies that when the LLM
// returns a call to a system tool (e.g. "Read") and an ACF router is set,
// the call is dispatched to the router rather than the regular executor.
func TestRunReactLoop_ACF_SystemToolRoutedThroughACF(t *testing.T) {
	spec := minimalACFSpec()
	router := &stubSystemToolRouter{
		systemTools: map[string]bool{"Read": true},
		defs: []llm.ToolDefinition{
			{Name: "Read", Description: "read a file", Parameters: map[string]interface{}{"type": "object"}},
		},
	}

	llmClient := &stubLLMClient{
		responses: []*llm.Response{
			{
				Content: "",
				ToolCalls: []llm.ToolCall{
					{ID: "c1", Name: "Read", Arguments: `{"file_path":"README.md"}`},
				},
			},
			{Content: "file contents processed"}, // final response after tool
		},
	}
	exec := tools.New(nil) // no workspace tools

	opts := Options{ACF: router}
	result, err := RunReactLoop(context.Background(), spec, llmClient, exec, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "completed" {
		t.Errorf("status = %q, want %q", result.Status, "completed")
	}
	if len(router.executed) != 1 {
		t.Fatalf("expected 1 system tool execution, got %d: %v", len(router.executed), router.executed)
	}
	if !strings.HasPrefix(router.executed[0], "Read:") {
		t.Errorf("expected execution of 'Read', got %q", router.executed[0])
	}
}

// TestRunReactLoop_ACF_RegularToolNotRoutedThroughACF verifies that a regular
// tool (one that IsSystemTool returns false for) is executed by the regular
// executor, not the ACF router.
func TestRunReactLoop_ACF_RegularToolNotRoutedThroughACF(t *testing.T) {
	spec := minimalACFSpec()
	router := &stubSystemToolRouter{
		systemTools: map[string]bool{"Read": true},
	}

	// workspace tool "http_search" is NOT a system tool
	exec := tools.New([]job.ToolConfig{
		{
			Name:        "http_search",
			Description: "search the web",
			Type:        "http",
			HTTPConfig: &job.HTTPToolConfig{
				URL:    "https://example.com/search",
				Method: "GET",
			},
		},
	})

	llmClient := &stubLLMClient{
		responses: []*llm.Response{
			{
				Content: "",
				ToolCalls: []llm.ToolCall{
					{ID: "c2", Name: "http_search", Arguments: `{"q":"golang"}`},
				},
			},
			{Content: "search done"},
		},
	}

	opts := Options{ACF: router}
	result, err := RunReactLoop(context.Background(), spec, llmClient, exec, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Regular tool executed by the executor → router.executed must be empty.
	if len(router.executed) != 0 {
		t.Errorf("ACF router should not have been called for a regular tool; got executions: %v", router.executed)
	}
	// Result may be completed or failed (http call will fail in unit test without
	// a live server), but the important thing is ACF was NOT invoked.
	_ = result
}

// TestRunReactLoop_ACF_SystemToolDefsAppended verifies that system tool
// definitions from the ACF router are appended to the tool definitions sent
// to the LLM. We capture what the LLM receives by wrapping the client.
func TestRunReactLoop_ACF_SystemToolDefsAppended(t *testing.T) {
	spec := minimalACFSpec()

	var capturedDefs []llm.ToolDefinition
	capturingClient := &capturingLLMClient{
		response: &llm.Response{Content: "ok"},
		capture:  &capturedDefs,
	}

	router := &stubSystemToolRouter{
		systemTools: map[string]bool{"Bash": true},
		defs: []llm.ToolDefinition{
			{Name: "Bash", Description: "run shell cmd", Parameters: map[string]interface{}{"type": "object"}},
		},
	}

	workspaceTool := job.ToolConfig{
		Name: "my_http_tool", Description: "http tool", Type: "http",
		HTTPConfig: &job.HTTPToolConfig{URL: "https://api.example.com", Method: "GET"},
	}
	exec := tools.New([]job.ToolConfig{workspaceTool})

	opts := Options{ACF: router}
	_, err := RunReactLoop(context.Background(), spec, capturingClient, exec, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The definitions sent to the LLM should contain both the workspace tool
	// AND the system tool.
	nameSet := make(map[string]bool, len(capturedDefs))
	for _, d := range capturedDefs {
		nameSet[d.Name] = true
	}
	if !nameSet["my_http_tool"] {
		t.Error("workspace tool 'my_http_tool' missing from LLM tool definitions")
	}
	if !nameSet["Bash"] {
		t.Error("ACF system tool 'Bash' missing from LLM tool definitions")
	}
}

// capturingLLMClient captures the tool definitions on the first call.
type capturingLLMClient struct {
	response *llm.Response
	capture  *[]llm.ToolDefinition
}

func (c *capturingLLMClient) Complete(_ context.Context, _ []llm.Message, defs []llm.ToolDefinition) (*llm.Response, error) {
	if *c.capture == nil {
		// Capture only on the first call.
		copied := make([]llm.ToolDefinition, len(defs))
		copy(copied, defs)
		*c.capture = copied
	}
	return c.response, nil
}

// ---------------------------------------------------------------------------
// Mixed calls: some system, some regular
// ---------------------------------------------------------------------------

// TestRunReactLoop_ACF_MixedToolCalls verifies that when the LLM returns both
// a system tool call and a regular tool call in one response, each is routed
// to the correct handler.
func TestRunReactLoop_ACF_MixedToolCalls(t *testing.T) {
	spec := minimalACFSpec()
	router := &stubSystemToolRouter{
		systemTools: map[string]bool{"Read": true},
	}

	// workspace tool
	exec := tools.New([]job.ToolConfig{
		{
			Name: "noop_http", Type: "http",
			HTTPConfig: &job.HTTPToolConfig{URL: "https://example.com", Method: "GET"},
		},
	})

	llmClient := &stubLLMClient{
		responses: []*llm.Response{
			{
				ToolCalls: []llm.ToolCall{
					{ID: "sys-call", Name: "Read", Arguments: `{"file_path":"main.go"}`},
					{ID: "reg-call", Name: "noop_http", Arguments: `{}`},
				},
			},
			{Content: "all done"},
		},
	}

	opts := Options{ACF: router}
	result, err := RunReactLoop(context.Background(), spec, llmClient, exec, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// System tool must have been routed through ACF.
	if len(router.executed) != 1 || !strings.HasPrefix(router.executed[0], "Read:") {
		t.Errorf("expected exactly one ACF execution of 'Read', got: %v", router.executed)
	}
	// Result should reach "completed" (the regular tool may fail in unit test,
	// but the loop records the error and continues).
	_ = result
}

// ---------------------------------------------------------------------------
// spec mutation guard
// ---------------------------------------------------------------------------

// TestRunReactLoop_ACF_SpecNotMutated ensures RunReactLoop never mutates the
// caller's Spec when Options.SystemPromptOverride is set.
func TestRunReactLoop_ACF_SpecNotMutated(t *testing.T) {
	spec := minimalACFSpec()
	original := *spec

	llmClient := &stubLLMClient{responses: []*llm.Response{{Content: "done"}}}
	exec := tools.New(nil)

	_, err := RunReactLoop(context.Background(), spec, llmClient, exec, Options{
		SystemPromptOverride: "override",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.SystemPrompt != original.SystemPrompt {
		t.Errorf("spec.SystemPrompt was mutated: got %q, want %q", spec.SystemPrompt, original.SystemPrompt)
	}
}
