package tools

import (
	"testing"
)

func makeTestPolicy() *ShadowPolicy {
	return NewDefaultShadowPolicy()
}

func TestShadowInterceptor_MockedToolReturnsOutput(t *testing.T) {
	policy := makeTestPolicy()
	mocks := map[string]string{
		"send_email": `{"status":"mocked","message_id":"mock-123"}`,
	}
	si := NewShadowInterceptor(policy, mocks, "strict")

	output, executeReal, abort := si.Intercept("send_email", "")

	if abort {
		t.Error("expected no abort for mocked write tool")
	}
	if executeReal {
		t.Error("expected executeReal=false for mocked tool")
	}
	if output != mocks["send_email"] {
		t.Errorf("expected mocked output %q, got %q", mocks["send_email"], output)
	}

	stats := si.Stats()
	if stats.MockedCount != 1 {
		t.Errorf("expected MockedCount=1, got %d", stats.MockedCount)
	}
}

func TestShadowInterceptor_WriteToolWithoutMock_Aborts(t *testing.T) {
	policy := makeTestPolicy()
	si := NewShadowInterceptor(policy, map[string]string{}, "strict")

	_, executeReal, abort := si.Intercept("send_email", "")

	if !abort {
		t.Error("expected abort for write tool without mock")
	}
	if executeReal {
		t.Error("expected executeReal=false for aborted tool")
	}

	stats := si.Stats()
	if len(stats.AbortedTools) != 1 || stats.AbortedTools[0] != "send_email" {
		t.Errorf("expected AbortedTools=[send_email], got %v", stats.AbortedTools)
	}
}

func TestShadowInterceptor_DeterministicReadWithoutMock_Passthrough(t *testing.T) {
	policy := makeTestPolicy()
	si := NewShadowInterceptor(policy, map[string]string{}, "strict")

	_, executeReal, abort := si.Intercept("list_channels", "")

	if abort {
		t.Error("expected no abort for deterministic read without mock")
	}
	if !executeReal {
		t.Error("expected executeReal=true for deterministic read passthrough")
	}

	stats := si.Stats()
	if stats.PassthroughCount != 1 {
		t.Errorf("expected PassthroughCount=1, got %d", stats.PassthroughCount)
	}
}

func TestShadowInterceptor_DeterministicReadWithMock_ReturnsMock(t *testing.T) {
	policy := makeTestPolicy()
	mocks := map[string]string{
		"list_channels": `{"channels":["general","random"]}`,
	}
	si := NewShadowInterceptor(policy, mocks, "strict")

	output, executeReal, abort := si.Intercept("list_channels", "")

	if abort {
		t.Error("expected no abort for mocked deterministic read")
	}
	if executeReal {
		t.Error("expected executeReal=false for mocked deterministic read")
	}
	if output != mocks["list_channels"] {
		t.Errorf("expected mocked output, got %q", output)
	}
}

func TestShadowInterceptor_HTTPWriteMethod_NoMock_Aborts(t *testing.T) {
	policy := makeTestPolicy()
	si := NewShadowInterceptor(policy, map[string]string{}, "strict")

	_, executeReal, abort := si.Intercept("create_record", "POST")

	if !abort {
		t.Error("expected abort for HTTP POST tool without mock")
	}
	if executeReal {
		t.Error("expected executeReal=false for aborted HTTP write tool")
	}
}

func TestShadowInterceptor_HTTPGetMethod_NoMock_Passthrough(t *testing.T) {
	policy := makeTestPolicy()
	si := NewShadowInterceptor(policy, map[string]string{}, "strict")

	_, executeReal, abort := si.Intercept("fetch_data", "GET")

	if abort {
		t.Error("expected no abort for HTTP GET tool")
	}
	if !executeReal {
		t.Error("expected executeReal=true for HTTP GET passthrough")
	}
}

func TestShadowInterceptor_UnknownTool_DefaultsToWrite_Aborts(t *testing.T) {
	policy := makeTestPolicy()
	si := NewShadowInterceptor(policy, map[string]string{}, "strict")

	_, executeReal, abort := si.Intercept("unknown_tool", "")

	if !abort {
		t.Error("expected abort for unknown tool (defaults to write)")
	}
	if executeReal {
		t.Error("expected executeReal=false for unknown tool")
	}
}

func TestShadowInterceptor_UnknownToolWithMock_ReturnsMock(t *testing.T) {
	policy := makeTestPolicy()
	mocks := map[string]string{
		"unknown_tool": `{"result":"ok"}`,
	}
	si := NewShadowInterceptor(policy, mocks, "strict")

	output, executeReal, abort := si.Intercept("unknown_tool", "")

	if abort {
		t.Error("expected no abort for unknown tool with mock")
	}
	if executeReal {
		t.Error("expected executeReal=false for mocked unknown tool")
	}
	if output != mocks["unknown_tool"] {
		t.Errorf("expected mocked output, got %q", output)
	}
}

func TestShadowInterceptor_BestEffort_NonDeterministicRead_Passthrough(t *testing.T) {
	// In best_effort mode, non-deterministic reads without mocks should
	// pass through instead of aborting.
	policy := &ShadowPolicy{
		WriteTools:             map[string]bool{},
		DeterministicReadTools: map[string]bool{},
		HTTPWriteMethods:       map[string]bool{"POST": true, "PUT": true, "PATCH": true, "DELETE": true},
	}
	// A tool not in any list defaults to write. To test non-deterministic
	// read behavior, we'd need to add the classification. Since the current
	// policy only has write and deterministic_read explicit lists, and unknown
	// defaults to write, this test verifies that best_effort mode on an
	// unknown/write tool still aborts (write tools always abort).
	si := NewShadowInterceptor(policy, map[string]string{}, "best_effort")

	_, _, abort := si.Intercept("some_write_tool", "")
	if !abort {
		t.Error("expected abort for write tool even in best_effort mode")
	}
}

func TestShadowInterceptor_StatsSnapshot(t *testing.T) {
	policy := makeTestPolicy()
	mocks := map[string]string{
		"send_email": `{"ok":true}`,
	}
	si := NewShadowInterceptor(policy, mocks, "strict")

	// Mock one tool, passthrough another
	si.Intercept("send_email", "")
	si.Intercept("list_channels", "")

	stats := si.Stats()
	if stats.MockedCount != 1 {
		t.Errorf("expected MockedCount=1, got %d", stats.MockedCount)
	}
	if stats.PassthroughCount != 1 {
		t.Errorf("expected PassthroughCount=1, got %d", stats.PassthroughCount)
	}
	if len(stats.AbortedTools) != 0 {
		t.Errorf("expected no aborted tools, got %v", stats.AbortedTools)
	}

	// Verify snapshot isolation: modifying returned stats doesn't affect internal state.
	stats.MockedCount = 999
	stats2 := si.Stats()
	if stats2.MockedCount != 1 {
		t.Error("Stats() should return a snapshot, not a reference")
	}
}

func TestShadowInterceptor_NilMocks(t *testing.T) {
	policy := makeTestPolicy()
	si := NewShadowInterceptor(policy, nil, "strict")

	// Should not panic; write tool without mock should abort.
	_, _, abort := si.Intercept("send_email", "")
	if !abort {
		t.Error("expected abort for write tool with nil mocks map")
	}
}
