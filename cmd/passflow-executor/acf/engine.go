package acf

import (
	"context"
	"time"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/job"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
	"go.uber.org/zap"
)

// systemToolNames lists every tool name handled by SystemToolProvider.
// Keeping this as a package-level slice lets IsSystemTool do a single
// linear scan without allocating a map on every call.
var systemToolNames = []string{"Bash", "Read", "Write", "Edit", "Grep", "Glob", "LS"}

// systemToolDefs is the canonical LLM ToolDefinition for each system tool.
// These definitions are returned by GetSystemToolDefs and registered alongside
// any workspace-defined HTTP/integration tools during the ReAct loop.
var systemToolDefs = []llm.ToolDefinition{
	{
		Name:        "Bash",
		Description: "Execute a shell command inside the sandboxed work directory. Returns stdout and stderr.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The shell command to execute.",
				},
			},
			"required": []string{"command"},
		},
	},
	{
		Name:        "Read",
		Description: "Read the full contents of a file inside the work directory.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute or relative path to the file to read.",
				},
			},
			"required": []string{"file_path"},
		},
	},
	{
		Name:        "Write",
		Description: "Write content to a file, creating parent directories as needed.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Destination file path.",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Content to write.",
				},
			},
			"required": []string{"file_path", "content"},
		},
	},
	{
		Name:        "Edit",
		Description: "Perform an exact-string replacement in a file (replaces the first occurrence).",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to edit.",
				},
				"old_string": map[string]interface{}{
					"type":        "string",
					"description": "Exact string to replace.",
				},
				"new_string": map[string]interface{}{
					"type":        "string",
					"description": "Replacement string.",
				},
			},
			"required": []string{"file_path", "old_string"},
		},
	},
	{
		Name:        "Grep",
		Description: "Search for lines matching a regular expression across all files in the work directory.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "Regular expression to search for.",
				},
			},
			"required": []string{"pattern"},
		},
	},
	{
		Name:        "Glob",
		Description: "Expand a glob pattern relative to the work directory and return matching paths.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "Glob pattern (e.g. **/*.go).",
				},
			},
			"required": []string{"pattern"},
		},
	},
	{
		Name:        "LS",
		Description: "List the contents of a directory inside the work directory.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Directory to list. Defaults to the work directory root.",
				},
			},
			"required": []string{},
		},
	},
}

// ACFContext holds every ACF-specific resource needed during a single job
// execution. It is created by PrepareACFExecution and passed into the ReAct
// loop integration layer; it is NOT passed into the existing react.go
// directly (that wiring is a separate task).
type ACFContext struct {
	// SystemPrompt is the translated system prompt for the target LLM provider.
	SystemPrompt string

	// SystemTools executes the sandboxed Claude Code built-in tools.
	SystemTools *SystemToolProvider

	// Auditor records every tool call and result for the full audit trail.
	Auditor *ExecutionAuditor

	// Translator is retained so callers can re-translate if needed.
	Translator *PromptTranslator

	// Provider is the normalised LLM provider name (anthropic, openai, gemini, …).
	Provider string

	// Policy is the workspace-level policy that governs tool access.
	// May be nil when no policy was specified in the job spec.
	Policy *Policy

	// CADConfig is the raw CAD configuration from the job spec.
	CADConfig *job.CADConfig
}

// PrepareACFExecution initialises all ACF components for the given job spec.
// It returns nil when spec.IsACFAgent() is false so that the caller can use a
// simple nil-check as the ACF gate:
//
//	if acfCtx := acf.PrepareACFExecution(spec); acfCtx != nil { ... }
//
// The workDir parameter is the filesystem path in which the pod will operate;
// typically the value of the WORK_DIR environment variable.
// logger may be nil; when nil, structured audit logging is skipped.
func PrepareACFExecution(spec *job.Spec, workDir string, logger *zap.Logger) *ACFContext {
	if !spec.IsACFAgent() {
		return nil
	}

	// Convert job.Policy → acf.Policy (structurally identical but in different
	// packages to keep the job package free of ACF dependencies).
	var policy *Policy
	if spec.Policy != nil {
		policy = &Policy{
			AllowedTools:   spec.Policy.AllowedTools,
			DeniedTools:    spec.Policy.DeniedTools,
			MaxBashTimeout: spec.Policy.MaxBashTimeout,
			AllowNetwork:   spec.Policy.AllowNetwork,
			AllowWrite:     spec.Policy.AllowWrite,
		}
	}

	// Convert job.CADConfig → acf.CADConfig for the PromptTranslator.
	cadForTranslator := jobCADToACFCAD(spec.CADConfig)

	// Collect all tool definitions the agent declares so the translator can
	// embed provider-specific tool calling instructions.
	toolDefs := buildToolDefs(spec)

	translator := NewPromptTranslator()
	systemPrompt := translator.Translate(spec.ModelProvider, cadForTranslator, toolDefs)

	auditor := NewExecutionAuditor(spec.ExecutionID, spec.ModelProvider, logger)
	auditor.LogPromptTranslation(spec.CADConfig.SourceType, spec.ModelProvider, len(systemPrompt))

	return &ACFContext{
		SystemPrompt: systemPrompt,
		SystemTools:  NewSystemToolProvider(workDir, policy),
		Auditor:      auditor,
		Translator:   translator,
		Provider:     spec.ModelProvider,
		Policy:       policy,
		CADConfig:    spec.CADConfig,
	}
}

// GetSystemToolDefs returns the LLM ToolDefinition slice for every system tool.
// The caller appends these to the workspace-defined tool definitions before the
// first LLM call so the model can invoke Bash, Read, Write, etc.
//
// When an ACF policy restricts tools, definitions for denied tools are still
// returned here; enforcement happens at execution time inside SystemToolProvider
// so that the LLM receives a descriptive error rather than failing silently.
func (c *ACFContext) GetSystemToolDefs() []llm.ToolDefinition {
	// Return a copy to prevent callers from mutating the package-level slice.
	defs := make([]llm.ToolDefinition, len(systemToolDefs))
	copy(defs, systemToolDefs)
	return defs
}

// ExecuteSystemTool routes a tool call to the SystemToolProvider, times the
// execution, and records both the call and the result via the Auditor.
// It always returns a non-nil *ToolResult; errors are surfaced through
// ToolResult.Error so the ReAct loop can feed them back to the LLM.
func (c *ACFContext) ExecuteSystemTool(ctx context.Context, call ToolCall) *ToolResult {
	c.Auditor.LogToolCall(call.Tool, call.Params)

	start := time.Now()
	result, _ := c.SystemTools.Execute(ctx, &call)
	// Execute always returns a non-nil result and a nil error (errors are
	// embedded in ToolResult.Error), so the second return value is ignored.
	duration := time.Since(start)

	c.Auditor.LogToolResult(call.Tool, result.Output, result.Error, duration)

	return result
}

// IsSystemTool reports whether name matches one of the built-in system tool
// names (Bash, Read, Write, Edit, Grep, Glob, LS). The comparison is
// case-sensitive to match the exact names registered in systemToolDefs.
func (c *ACFContext) IsSystemTool(name string) bool {
	for _, n := range systemToolNames {
		if n == name {
			return true
		}
	}
	return false
}

// --- internal helpers ---

// jobCADToACFCAD converts a *job.CADConfig into a *CADConfig (acf package).
// Both types are structurally equivalent; the conversion keeps the job package
// free of a direct import of the acf package.
func jobCADToACFCAD(src *job.CADConfig) *CADConfig {
	if src == nil {
		return nil
	}

	skills := make([]SkillDef, len(src.Skills))
	for i, s := range src.Skills {
		skills[i] = SkillDef{Name: s.Name, Content: s.Content}
	}

	contextFiles := make(map[string]string, len(src.ContextFiles))
	for k, v := range src.ContextFiles {
		contextFiles[k] = v
	}

	return &CADConfig{
		SystemPrompt: src.SystemPrompt,
		Skills:       skills,
		ContextFiles: contextFiles,
		SourceType:   src.SourceType,
	}
}

// buildToolDefs builds the []ToolDef list for the PromptTranslator from the
// job spec's declared tools plus the system tools.
func buildToolDefs(spec *job.Spec) []ToolDef {
	defs := make([]ToolDef, 0, len(systemToolNames)+len(spec.Tools))

	// System tools always come first.
	for _, name := range systemToolNames {
		defs = append(defs, ToolDef{
			Name:        name,
			Description: descriptionForSystemTool(name),
			Type:        "system",
		})
	}

	// Workspace-declared tools (HTTP, integration, etc.).
	for _, t := range spec.Tools {
		defs = append(defs, ToolDef{
			Name:        t.Name,
			Description: t.Description,
			Type:        t.Type,
		})
	}

	return defs
}

// descriptionForSystemTool returns a concise one-line description for a system
// tool name. Used when building the PromptTranslator ToolDef list.
func descriptionForSystemTool(name string) string {
	for _, def := range systemToolDefs {
		if def.Name == name {
			return def.Description
		}
	}
	return name
}
