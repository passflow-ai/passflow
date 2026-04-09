package acf

import (
	"context"
	"encoding/json"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/tools"
)

// EngineAdapter wraps *ACFContext so that it satisfies the engine.SystemToolRouter
// interface without creating an import cycle (engine → acf → engine).
//
// Callers in main.go create an EngineAdapter from the prepared ACFContext and
// pass it to the engine via engine.Options.ACF.  The engine package only sees
// the interface; it never imports the acf package.
type EngineAdapter struct {
	ctx *ACFContext
}

// NewEngineAdapter wraps acfCtx in a type that satisfies engine.SystemToolRouter.
// Returns nil when acfCtx is nil so that a simple nil-check in the caller is
// sufficient to decide whether ACF routing is active.
func NewEngineAdapter(acfCtx *ACFContext) *EngineAdapter {
	if acfCtx == nil {
		return nil
	}
	return &EngineAdapter{ctx: acfCtx}
}

// IsSystemTool delegates to ACFContext.IsSystemTool.
func (a *EngineAdapter) IsSystemTool(name string) bool {
	return a.ctx.IsSystemTool(name)
}

// GetSystemToolDefs delegates to ACFContext.GetSystemToolDefs.
func (a *EngineAdapter) GetSystemToolDefs() []llm.ToolDefinition {
	return a.ctx.GetSystemToolDefs()
}

// Execute converts an llm.ToolCall into an acf.ToolCall, dispatches it to
// ACFContext.ExecuteSystemTool, and converts the result back into a
// tools.Result so the ReAct loop can handle both paths uniformly.
func (a *EngineAdapter) Execute(ctx context.Context, call llm.ToolCall) tools.Result {
	// Parse the JSON-string arguments into a map for acf.ToolCall.Params.
	params := make(map[string]interface{})
	if call.Arguments != "" {
		// Ignore parse errors: if the LLM sent malformed JSON we pass an empty
		// params map; the SystemToolProvider will return a descriptive error.
		_ = json.Unmarshal([]byte(call.Arguments), &params)
	}

	acfCall := ToolCall{
		Tool:   call.Name,
		Params: params,
		CallID: call.ID,
	}

	result := a.ctx.ExecuteSystemTool(ctx, acfCall)

	return tools.Result{
		ToolCallID: call.ID,
		ToolName:   call.Name,
		Output:     result.Output,
		Error:      result.Error,
	}
}
