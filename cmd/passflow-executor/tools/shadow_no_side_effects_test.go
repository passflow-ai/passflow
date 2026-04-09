package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/job"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
)

// ---------------------------------------------------------------------------
// Shadow Mode: No Side Effects Test Suite (QA-2001)
//
// These tests prove that shadow mode NEVER executes write tools for real.
// This is a CRITICAL safety property of the shadow execution system.
// ---------------------------------------------------------------------------

// newWriteToolConfigs returns tool configs for common write tools backed by
// HTTP endpoints pointing at the given server URL. Each tool uses a POST
// method so it is classified as a write tool by ShadowPolicy.
func newWriteToolConfigs(serverURL string) []job.ToolConfig {
	names := []string{"send_email", "send_message", "create_issue", "update_crm", "delete_record"}
	configs := make([]job.ToolConfig, 0, len(names))
	for _, name := range names {
		configs = append(configs, job.ToolConfig{
			Name: name,
			Type: "http",
			HTTPConfig: &job.HTTPToolConfig{
				URL:    serverURL + "/" + name,
				Method: "POST",
			},
		})
	}
	return configs
}

// makeToolCalls builds llm.ToolCall slices from tool names.
func makeToolCalls(names ...string) []llm.ToolCall {
	calls := make([]llm.ToolCall, len(names))
	for i, name := range names {
		calls[i] = llm.ToolCall{
			ID:        "call-" + name,
			Name:      name,
			Arguments: `{}`,
		}
	}
	return calls
}

// TestShadowMode_WriteToolsNeverExecute creates an executor backed by a real
// HTTP server that counts requests. In shadow mode with mocks, write tools
// must NEVER hit the real server -- all calls must return mocked outputs.
func TestShadowMode_WriteToolsNeverExecute(t *testing.T) {
	var hitCount atomic.Int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitCount.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"real":"response"}`))
	}))
	defer srv.Close()

	writeTools := newWriteToolConfigs(srv.URL)
	exec := New(writeTools)
	policy := NewDefaultShadowPolicy()

	mocks := map[string]string{
		"send_email":    `{"mocked":"send_email"}`,
		"send_message":  `{"mocked":"send_message"}`,
		"create_issue":  `{"mocked":"create_issue"}`,
		"update_crm":    `{"mocked":"update_crm"}`,
		"delete_record": `{"mocked":"delete_record"}`,
	}

	interceptor := NewShadowInterceptor(policy, mocks, "strict")
	exec.SetShadowMode(interceptor)

	calls := makeToolCalls("send_email", "send_message", "create_issue", "update_crm", "delete_record")
	results := exec.ExecuteAll(context.Background(), calls)

	// Verify: zero real HTTP calls were made.
	if got := hitCount.Load(); got != 0 {
		t.Fatalf("CRITICAL: shadow mode allowed %d real HTTP calls to write tools; expected 0", got)
	}

	// Verify: all results contain mocked output.
	for _, r := range results {
		if r.Error != "" {
			t.Errorf("unexpected error for tool %q: %s", r.ToolName, r.Error)
		}
		if !strings.Contains(r.Output, "mocked") {
			t.Errorf("tool %q output should contain mocked data, got: %s", r.ToolName, r.Output)
		}
		if !r.Mocked {
			t.Errorf("tool %q result should have Mocked=true", r.ToolName)
		}
	}

	// Verify stats.
	stats := exec.ShadowStats()
	if stats.MockedCount != 5 {
		t.Errorf("expected MockedCount=5, got %d", stats.MockedCount)
	}
	if stats.PassthroughCount != 0 {
		t.Errorf("expected PassthroughCount=0, got %d", stats.PassthroughCount)
	}
}

// TestShadowMode_ReadToolsAllowlisted verifies that deterministic read tools
// classified in the policy can pass through to real execution when no mock is
// provided. Since the test server is on loopback (blocked by SSRF), we verify
// the passthrough counter is incremented and the interceptor attempts real
// execution (which will fail due to SSRF, proving the call was NOT mocked).
func TestShadowMode_ReadToolsAllowlisted(t *testing.T) {
	// Use a non-routable IP so the request will fail at the network level but
	// the key assertion is that the interceptor delegates to the real executor
	// rather than returning a mock or aborting.
	readTools := []job.ToolConfig{
		{
			Name: "list_channels",
			Type: "http",
			HTTPConfig: &job.HTTPToolConfig{
				URL:    "http://192.0.2.1/list_channels", // TEST-NET-1, non-routable
				Method: "GET",
			},
		},
	}

	exec := New(readTools)
	policy := NewDefaultShadowPolicy()

	// No mocks provided for deterministic read tools -- they should pass through.
	interceptor := NewShadowInterceptor(policy, nil, "strict")
	exec.SetShadowMode(interceptor)

	calls := makeToolCalls("list_channels")
	_ = exec.ExecuteAll(context.Background(), calls)

	// The tool should have been passed through, not mocked.
	stats := exec.ShadowStats()
	if stats.PassthroughCount != 1 {
		t.Errorf("expected PassthroughCount=1 for allowlisted read tool, got %d", stats.PassthroughCount)
	}
	if stats.MockedCount != 0 {
		t.Errorf("expected MockedCount=0 for allowlisted read tool, got %d", stats.MockedCount)
	}
}

// TestShadowMode_ReadToolsWithMock verifies that when a mock IS provided for
// a deterministic read tool, the mock is used instead of live execution.
func TestShadowMode_ReadToolsWithMock(t *testing.T) {
	var hitCount atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	readTools := []job.ToolConfig{
		{
			Name: "list_channels",
			Type: "http",
			HTTPConfig: &job.HTTPToolConfig{
				URL:    srv.URL + "/list_channels",
				Method: "GET",
			},
		},
	}

	exec := New(readTools)
	policy := NewDefaultShadowPolicy()
	mocks := map[string]string{
		"list_channels": `{"channels":["general","random"]}`,
	}

	interceptor := NewShadowInterceptor(policy, mocks, "strict")
	exec.SetShadowMode(interceptor)

	calls := makeToolCalls("list_channels")
	results := exec.ExecuteAll(context.Background(), calls)

	// With a mock provided, the interceptor should use it.
	stats := exec.ShadowStats()
	if stats.MockedCount != 1 {
		t.Errorf("expected MockedCount=1 when mock is provided for read tool, got %d", stats.MockedCount)
	}
	if hitCount.Load() != 0 {
		t.Errorf("expected 0 real HTTP calls when mock is available, got %d", hitCount.Load())
	}
	if !strings.Contains(results[0].Output, "general") {
		t.Errorf("expected mocked channels output, got: %s", results[0].Output)
	}
}

// TestShadowMode_MissingMockForWriteTool_Aborts verifies that if a write tool
// has no mock configured, the shadow run aborts immediately.
func TestShadowMode_MissingMockForWriteTool_Aborts(t *testing.T) {
	writeTools := []job.ToolConfig{
		{Name: "send_email", Type: "http", HTTPConfig: &job.HTTPToolConfig{URL: "http://192.0.2.1/send", Method: "POST"}},
		{Name: "send_message", Type: "http", HTTPConfig: &job.HTTPToolConfig{URL: "http://192.0.2.1/msg", Method: "POST"}},
	}

	exec := New(writeTools)
	policy := NewDefaultShadowPolicy()

	// No mocks at all -- write tools should cause an abort.
	interceptor := NewShadowInterceptor(policy, nil, "strict")
	exec.SetShadowMode(interceptor)

	calls := makeToolCalls("send_email", "send_message")
	results := exec.ExecuteAll(context.Background(), calls)

	// First tool should cause abort.
	stats := exec.ShadowStats()
	if len(stats.AbortedTools) == 0 {
		t.Fatal("expected shadow run to abort when write tool has no mock, but AbortedTools is empty")
	}
	if stats.AbortedTools[0] != "send_email" {
		t.Errorf("expected first aborted tool to be %q, got %q", "send_email", stats.AbortedTools[0])
	}

	// First result should have a shadow mode error.
	if !strings.Contains(results[0].Error, "shadow mode") {
		t.Errorf("expected shadow mode error in first result, got: %q", results[0].Error)
	}

	// Second tool should also have abort error (cascaded).
	if !strings.Contains(results[1].Error, "shadow mode") && !strings.Contains(results[1].Error, "aborted") {
		t.Errorf("expected cascaded abort error in second result, got: %q", results[1].Error)
	}
}

// TestShadowMode_MissingMockForNonDeterministicRead_Aborts verifies that an
// unknown tool (classified as write per the default policy's safe default)
// without a mock causes the shadow run to abort in strict mode.
func TestShadowMode_MissingMockForNonDeterministicRead_Aborts(t *testing.T) {
	// "search_database" is not in any allowlist. As a non-HTTP tool, the
	// policy classifies it as write (safest default), so it must abort
	// without a mock.
	unknownTools := []job.ToolConfig{
		{Name: "search_database", Type: "integration", IntegrationConfig: &job.IntegrationToolConfig{Provider: "custom", Action: "search"}},
	}

	exec := New(unknownTools)
	policy := NewDefaultShadowPolicy()
	interceptor := NewShadowInterceptor(policy, nil, "strict")
	exec.SetShadowMode(interceptor)

	calls := makeToolCalls("search_database")
	results := exec.ExecuteAll(context.Background(), calls)

	stats := exec.ShadowStats()
	if len(stats.AbortedTools) == 0 {
		t.Fatal("expected shadow run to abort for unknown tool without mock")
	}
	if stats.AbortedTools[0] != "search_database" {
		t.Errorf("expected aborted tool %q, got %q", "search_database", stats.AbortedTools[0])
	}
	if !strings.Contains(results[0].Error, "shadow mode") {
		t.Errorf("expected shadow mode error, got: %q", results[0].Error)
	}
}

// TestShadowMode_StatsTracking verifies that MockedCount and PassthroughCount
// are correctly tracked across a mixed set of tool calls.
func TestShadowMode_StatsTracking(t *testing.T) {
	tools := []job.ToolConfig{
		// Write tools (will be mocked)
		{Name: "send_email", Type: "http", HTTPConfig: &job.HTTPToolConfig{URL: "http://192.0.2.1/send", Method: "POST"}},
		{Name: "create_issue", Type: "http", HTTPConfig: &job.HTTPToolConfig{URL: "http://192.0.2.1/issue", Method: "POST"}},
		// Deterministic read tool (will pass through -- no mock provided)
		{Name: "list_channels", Type: "http", HTTPConfig: &job.HTTPToolConfig{URL: "http://192.0.2.1/channels", Method: "GET"}},
	}

	exec := New(tools)
	policy := NewDefaultShadowPolicy()
	mocks := map[string]string{
		"send_email":   `{"mocked":"email"}`,
		"create_issue": `{"mocked":"issue"}`,
		// No mock for list_channels -- it should pass through as a deterministic read.
	}

	interceptor := NewShadowInterceptor(policy, mocks, "strict")
	exec.SetShadowMode(interceptor)

	calls := makeToolCalls("send_email", "list_channels", "create_issue")
	_ = exec.ExecuteAll(context.Background(), calls)

	stats := exec.ShadowStats()
	if stats.MockedCount != 2 {
		t.Errorf("expected MockedCount=2, got %d", stats.MockedCount)
	}
	if stats.PassthroughCount != 1 {
		t.Errorf("expected PassthroughCount=1, got %d", stats.PassthroughCount)
	}
	if len(stats.AbortedTools) != 0 {
		t.Errorf("expected no aborted tools, got %v", stats.AbortedTools)
	}
}

// TestShadowMode_NormalModeUnaffected verifies that without a shadow
// interceptor, ExecuteAll behaves exactly as before -- tools execute for real
// and there is no shadow interception.
func TestShadowMode_NormalModeUnaffected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"real":"data"}`))
	}))
	defer srv.Close()

	// Build tools pointing at the test server. The loopback address IS blocked
	// by SSRF, so the executor will return an SSRF error -- proving it attempted
	// real execution (no shadow interception).
	writeTools := []job.ToolConfig{
		{Name: "send_email", Type: "http", HTTPConfig: &job.HTTPToolConfig{URL: srv.URL + "/send", Method: "POST"}},
	}

	exec := New(writeTools)
	// Deliberately NOT calling exec.SetShadowMode().

	calls := makeToolCalls("send_email")
	results := exec.ExecuteAll(context.Background(), calls)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	// The result should have an error from the SSRF block (real execution
	// was attempted, not mocked).
	if results[0].Error == "" {
		// If somehow the server was reached, verify we got real data, not mocked.
		if strings.Contains(results[0].Output, "mocked") {
			t.Error("normal mode returned mocked output -- shadow interception is leaking")
		}
	} else {
		// SSRF block error is expected -- this proves real execution was attempted.
		if strings.Contains(results[0].Error, "shadow") {
			t.Error("normal mode returned shadow-related error -- interception is active without SetShadowMode")
		}
	}

	// ShadowStats should return zero values when no interceptor is set.
	stats := exec.ShadowStats()
	if stats.MockedCount != 0 || stats.PassthroughCount != 0 || len(stats.AbortedTools) != 0 {
		t.Errorf("expected zero-value ShadowStats without interceptor, got %+v", stats)
	}
}

// TestShadowMode_WriteToolsTableDriven uses table-driven tests to verify each
// individual write tool is properly mocked in shadow mode.
func TestShadowMode_WriteToolsTableDriven(t *testing.T) {
	var hitCount atomic.Int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitCount.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"real":"executed"}`))
	}))
	defer srv.Close()

	tests := []struct {
		name       string
		toolName   string
		httpMethod string
		toolType   string
		mockOutput string
	}{
		{name: "send_email is mocked", toolName: "send_email", httpMethod: "POST", toolType: "http", mockOutput: `{"status":"mocked_email_sent"}`},
		{name: "send_message is mocked", toolName: "send_message", httpMethod: "POST", toolType: "http", mockOutput: `{"status":"mocked_message_sent"}`},
		{name: "create_issue is mocked", toolName: "create_issue", httpMethod: "POST", toolType: "http", mockOutput: `{"id":"MOCK-123"}`},
		{name: "update_crm is mocked", toolName: "update_crm", httpMethod: "PUT", toolType: "http", mockOutput: `{"updated":true}`},
		{name: "delete_record is mocked", toolName: "delete_record", httpMethod: "DELETE", toolType: "http", mockOutput: `{"deleted":true}`},
		{name: "write_file is mocked", toolName: "write_file", httpMethod: "POST", toolType: "http", mockOutput: `{"written":true}`},
		{name: "unknown_tool defaults to write", toolName: "unknown_dangerous_tool", httpMethod: "", toolType: "integration", mockOutput: `{"safe":"mocked"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hitCount.Store(0)

			var toolCfg job.ToolConfig
			if tc.toolType == "http" {
				toolCfg = job.ToolConfig{
					Name: tc.toolName,
					Type: "http",
					HTTPConfig: &job.HTTPToolConfig{
						URL:    srv.URL + "/" + tc.toolName,
						Method: tc.httpMethod,
					},
				}
			} else {
				toolCfg = job.ToolConfig{
					Name: tc.toolName,
					Type: "integration",
					IntegrationConfig: &job.IntegrationToolConfig{
						Provider: "custom",
						Action:   "do",
					},
				}
			}

			exec := New([]job.ToolConfig{toolCfg})
			policy := NewDefaultShadowPolicy()
			mocks := map[string]string{tc.toolName: tc.mockOutput}

			interceptor := NewShadowInterceptor(policy, mocks, "strict")
			exec.SetShadowMode(interceptor)

			calls := makeToolCalls(tc.toolName)
			results := exec.ExecuteAll(context.Background(), calls)

			if hitCount.Load() != 0 {
				t.Errorf("CRITICAL: tool %q made %d real HTTP calls in shadow mode", tc.toolName, hitCount.Load())
			}
			if results[0].Error != "" {
				t.Errorf("unexpected error for %q: %s", tc.toolName, results[0].Error)
			}
			if results[0].Output != tc.mockOutput {
				t.Errorf("expected mock output %q, got %q", tc.mockOutput, results[0].Output)
			}

			stats := exec.ShadowStats()
			if stats.MockedCount != 1 {
				t.Errorf("expected MockedCount=1, got %d", stats.MockedCount)
			}
		})
	}
}

// TestShadowMode_HTTPMethodClassification verifies that HTTP write methods
// (POST, PUT, PATCH, DELETE) are always mocked even if the tool name is not
// in the explicit write list.
func TestShadowMode_HTTPMethodClassification(t *testing.T) {
	var hitCount atomic.Int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	writeMethods := []string{"POST", "PUT", "PATCH", "DELETE"}

	for _, method := range writeMethods {
		t.Run(method, func(t *testing.T) {
			hitCount.Store(0)

			// Use a tool name NOT in the explicit write list to prove
			// method-based classification works.
			toolCfg := job.ToolConfig{
				Name: "custom_api_call",
				Type: "http",
				HTTPConfig: &job.HTTPToolConfig{
					URL:    srv.URL + "/api",
					Method: method,
				},
			}

			exec := New([]job.ToolConfig{toolCfg})
			policy := NewDefaultShadowPolicy()
			mocks := map[string]string{"custom_api_call": `{"mocked":true}`}

			interceptor := NewShadowInterceptor(policy, mocks, "strict")
			exec.SetShadowMode(interceptor)

			calls := makeToolCalls("custom_api_call")
			_ = exec.ExecuteAll(context.Background(), calls)

			if hitCount.Load() != 0 {
				t.Errorf("CRITICAL: HTTP %s tool made real call in shadow mode", method)
			}
			stats := exec.ShadowStats()
			if stats.MockedCount != 1 {
				t.Errorf("expected MockedCount=1 for HTTP %s, got %d", method, stats.MockedCount)
			}
		})
	}
}

// TestShadowMode_MockedFlagSet verifies that the Mocked field on Result is
// true for mocked tool calls and false for passthrough calls.
func TestShadowMode_MockedFlagSet(t *testing.T) {
	tools := []job.ToolConfig{
		{Name: "send_email", Type: "http", HTTPConfig: &job.HTTPToolConfig{URL: "http://192.0.2.1/send", Method: "POST"}},
		{Name: "read_file", Type: "http", HTTPConfig: &job.HTTPToolConfig{URL: "http://192.0.2.1/read", Method: "GET"}},
	}

	exec := New(tools)
	policy := NewDefaultShadowPolicy()
	mocks := map[string]string{
		"send_email": `{"mocked":"email"}`,
		// No mock for read_file -- deterministic read, should passthrough.
	}

	interceptor := NewShadowInterceptor(policy, mocks, "strict")
	exec.SetShadowMode(interceptor)

	calls := makeToolCalls("send_email", "read_file")
	results := exec.ExecuteAll(context.Background(), calls)

	if !results[0].Mocked {
		t.Error("expected Mocked=true for send_email (write tool with mock)")
	}
	if results[1].Mocked {
		t.Error("expected Mocked=false for read_file (deterministic read, passthrough)")
	}
}
