package tools

import "sync"

// ShadowStats tracks shadow-mode interception statistics.
type ShadowStats struct {
	MockedCount      int
	PassthroughCount int
	AbortedTools     []string
}

// ShadowInterceptor decides how to handle each tool call in shadow mode:
// return a mock, allow live execution, or abort. It is attached to an Executor
// via SetShadowMode and consulted during ExecuteAll.
type ShadowInterceptor struct {
	policy        *ShadowPolicy
	mockedOutputs map[string]string // tool_name -> mock response JSON
	replayPolicy  string            // "strict" | "best_effort"

	mu    sync.Mutex
	stats ShadowStats
}

// NewShadowInterceptor creates a new interceptor with the given policy and
// pre-captured tool outputs. replayPolicy controls behaviour when a mock is
// missing: "strict" aborts for non-deterministic reads, "best_effort" allows
// live execution.
func NewShadowInterceptor(policy *ShadowPolicy, mockedOutputs map[string]string, replayPolicy string) *ShadowInterceptor {
	if mockedOutputs == nil {
		mockedOutputs = make(map[string]string)
	}
	return &ShadowInterceptor{
		policy:        policy,
		mockedOutputs: mockedOutputs,
		replayPolicy:  replayPolicy,
		stats: ShadowStats{
			AbortedTools: make([]string, 0),
		},
	}
}

// Intercept decides how to handle a tool call in shadow mode.
//
// Returns:
//   - mockOutput: the canned response (non-empty when the tool should be mocked)
//   - executeReal: true if the tool should execute normally
//   - abort: true if the shadow run must stop (write tool with no mock)
//
// Exactly one of the three outcomes is active per call.
func (si *ShadowInterceptor) Intercept(toolName string, httpMethod string) (mockOutput string, executeReal bool, abort bool) {
	si.mu.Lock()
	defer si.mu.Unlock()

	captured, hasMock := si.mockedOutputs[toolName]

	// If a mock exists, always use it regardless of classification.
	if hasMock {
		si.stats.MockedCount++
		return captured, false, false
	}

	// No mock available -- decide based on tool classification.
	class := si.policy.ClassifyTool(toolName, httpMethod)

	switch class {
	case ToolClassWrite:
		// Write tools cannot execute live in shadow mode. Abort.
		si.stats.AbortedTools = append(si.stats.AbortedTools, toolName)
		return "", false, true

	case ToolClassReadDeterministic:
		// Safe to execute live.
		si.stats.PassthroughCount++
		return "", true, false

	case ToolClassReadNonDeterministic:
		// In strict mode, abort. In best_effort mode, allow live execution.
		if si.replayPolicy == "strict" {
			si.stats.AbortedTools = append(si.stats.AbortedTools, toolName)
			return "", false, true
		}
		// best_effort: allow passthrough.
		si.stats.PassthroughCount++
		return "", true, false

	default:
		// Unknown classification defaults to write (safest). Abort.
		si.stats.AbortedTools = append(si.stats.AbortedTools, toolName)
		return "", false, true
	}
}

// Stats returns a snapshot of the interception statistics.
func (si *ShadowInterceptor) Stats() ShadowStats {
	si.mu.Lock()
	defer si.mu.Unlock()
	return ShadowStats{
		MockedCount:      si.stats.MockedCount,
		PassthroughCount: si.stats.PassthroughCount,
		AbortedTools:     append([]string(nil), si.stats.AbortedTools...),
	}
}
