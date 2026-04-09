package engine

import (
	"context"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/tools"
)

// SystemToolRouter is the interface the ReAct loop uses to route tool calls
// that belong to the ACF system tool layer (Bash, Read, Write, Edit, Grep,
// Glob, LS).  The engine package does not import the acf package directly;
// instead it depends on this interface so that:
//
//   - non-ACF jobs pass nil (no routing, all calls go to the regular Executor)
//   - ACF jobs pass an acf.EngineAdapter that wraps *acf.ACFContext
//
// IsSystemTool must be pure (no I/O); Execute will be called only when
// IsSystemTool returns true.
type SystemToolRouter interface {
	// IsSystemTool reports whether name is a built-in system tool (e.g. "Bash").
	IsSystemTool(name string) bool
	// GetSystemToolDefs returns the LLM tool definitions for all system tools.
	// The caller appends them to the workspace-defined tool definitions.
	GetSystemToolDefs() []llm.ToolDefinition
	// Execute runs the system tool identified by call.Name and returns the
	// result in the same format as tools.Executor so the loop can treat both
	// paths uniformly.
	Execute(ctx context.Context, call llm.ToolCall) tools.Result
}

// Options carries optional, per-invocation settings for the engine functions
// (RunReactLoop, RunOneShot). All fields are optional; zero values preserve
// the original behaviour so existing callers need no changes.
type Options struct {
	// ACF is non-nil for ACF agents. When set, system tool calls are routed
	// through ACF.Execute instead of the regular tools.Executor.
	ACF SystemToolRouter

	// SystemPromptOverride, when non-empty, replaces spec.SystemPrompt as the
	// system message sent to the LLM. Used by ACF to supply the translated
	// prompt instead of the raw field from the job spec.
	SystemPromptOverride string
}
