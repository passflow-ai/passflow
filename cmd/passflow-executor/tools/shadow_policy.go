package tools

import "strings"

// ToolClassification describes how a tool behaves with respect to side effects.
// Shadow mode uses this classification to decide whether a tool must be mocked,
// may execute live, or requires a captured artifact.
type ToolClassification string

const (
	// ToolClassWrite identifies tools that produce side effects (send email,
	// create issue, etc.). These are ALWAYS mocked in shadow mode.
	ToolClassWrite ToolClassification = "write"

	// ToolClassReadDeterministic identifies read-only tools whose output is
	// stable across runs (list_channels, read_file, etc.). These MAY execute
	// live in shadow mode when no captured output is available.
	ToolClassReadDeterministic ToolClassification = "read_deterministic"

	// ToolClassReadNonDeterministic identifies read-only tools whose output may
	// change between runs (search, query_database). These must use captured
	// output when available, and abort if no capture exists.
	ToolClassReadNonDeterministic ToolClassification = "read_non_deterministic"
)

// ShadowPolicy encodes the hard rules for tool behaviour during shadow-mode
// execution. See docs/contracts/shadow-mode-policy.md for the full
// specification.
type ShadowPolicy struct {
	// WriteTools lists tool names that MUST be mocked in shadow mode.
	WriteTools map[string]bool

	// DeterministicReadTools lists tool names that CAN execute live in shadow
	// mode when no captured output is available.
	DeterministicReadTools map[string]bool

	// HTTPWriteMethods lists HTTP methods that classify an HTTP-type tool as a
	// write tool (e.g. POST, PUT, PATCH, DELETE).
	HTTPWriteMethods map[string]bool
}

// NewDefaultShadowPolicy returns a ShadowPolicy pre-populated with the
// canonical tool classifications defined in the shadow-mode contract.
func NewDefaultShadowPolicy() *ShadowPolicy {
	return &ShadowPolicy{
		WriteTools: map[string]bool{
			"send_email":     true,
			"send_message":   true,
			"create_issue":   true,
			"create_comment": true,
			"update_crm":     true,
			"delete_record":  true,
			"write_file":     true,
		},
		DeterministicReadTools: map[string]bool{
			"list_channels": true,
			"list_issues":   true,
			"get_repo":      true,
			"read_file":     true,
			"glob":          true,
			"grep":          true,
			"ls":            true,
		},
		HTTPWriteMethods: map[string]bool{
			"POST":   true,
			"PUT":    true,
			"PATCH":  true,
			"DELETE": true,
		},
	}
}

// ClassifyTool returns the classification for a tool. For HTTP-type tools,
// httpMethod must be provided (e.g. "GET", "POST"). For non-HTTP tools,
// httpMethod should be empty.
//
// Unknown tools default to ToolClassWrite (the safest classification).
func (p *ShadowPolicy) ClassifyTool(toolName string, httpMethod string) ToolClassification {
	method := strings.ToUpper(httpMethod)

	// If an HTTP method is provided, classify based on the method.
	if method != "" {
		if p.HTTPWriteMethods[method] {
			return ToolClassWrite
		}
		return ToolClassReadDeterministic
	}

	// Explicit write tool.
	if p.WriteTools[toolName] {
		return ToolClassWrite
	}

	// Explicit deterministic read tool.
	if p.DeterministicReadTools[toolName] {
		return ToolClassReadDeterministic
	}

	// Unknown tools are treated as non-deterministic reads if they are not
	// explicitly listed. However, per policy rule 5, unknown tools default to
	// write (safest) since classification must be explicit.
	return ToolClassWrite
}

// MustMock reports whether the given tool must return a mocked response instead
// of executing live. hasCapturedOutput indicates whether a captured artifact
// from a prior execution exists for this tool invocation.
//
// Rules:
//   - Write tools: always mocked (regardless of captured output).
//   - Deterministic read tools: mocked only if captured output exists.
//   - Non-deterministic read tools: mocked if captured output exists.
//   - Unknown tools (classified as write): always mocked.
func (p *ShadowPolicy) MustMock(toolName string, httpMethod string, hasCapturedOutput bool) bool {
	class := p.ClassifyTool(toolName, httpMethod)
	switch class {
	case ToolClassWrite:
		return true
	case ToolClassReadDeterministic:
		// Use captured output if available; otherwise allow live execution.
		return hasCapturedOutput
	case ToolClassReadNonDeterministic:
		// Must use captured output if available; if not, ShouldAbortIfNoMock
		// handles the failure case.
		return hasCapturedOutput
	default:
		// Safety fallback: unknown classification always mocks.
		return true
	}
}

// ShouldAbortIfNoMock reports whether the shadow run must abort when the given
// tool has no mock available and no captured output exists.
//
// Rules:
//   - Write tools: abort (cannot execute live, no mock available).
//   - Deterministic read tools: do NOT abort (can safely execute live).
//   - Non-deterministic read tools: abort (output not reproducible).
//   - Unknown tools (classified as write): abort.
func (p *ShadowPolicy) ShouldAbortIfNoMock(toolName string, httpMethod string) bool {
	class := p.ClassifyTool(toolName, httpMethod)
	switch class {
	case ToolClassWrite:
		return true
	case ToolClassReadDeterministic:
		return false
	case ToolClassReadNonDeterministic:
		return true
	default:
		return true
	}
}
