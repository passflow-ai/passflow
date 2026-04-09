package tools

import "testing"

func TestClassifyTool_WriteTool(t *testing.T) {
	p := NewDefaultShadowPolicy()
	writeTools := []string{"send_email", "send_message", "create_issue", "create_comment", "update_crm", "delete_record", "write_file"}
	for _, tool := range writeTools {
		got := p.ClassifyTool(tool, "")
		if got != ToolClassWrite {
			t.Errorf("ClassifyTool(%q, \"\") = %q, want %q", tool, got, ToolClassWrite)
		}
	}
}

func TestClassifyTool_DeterministicRead(t *testing.T) {
	p := NewDefaultShadowPolicy()
	readTools := []string{"list_channels", "list_issues", "get_repo", "read_file", "glob", "grep", "ls"}
	for _, tool := range readTools {
		got := p.ClassifyTool(tool, "")
		if got != ToolClassReadDeterministic {
			t.Errorf("ClassifyTool(%q, \"\") = %q, want %q", tool, got, ToolClassReadDeterministic)
		}
	}
}

func TestClassifyTool_NonDeterministicRead(t *testing.T) {
	// Unknown tools that are not in any list default to write (safest).
	// Non-deterministic reads would need to be explicitly added to a
	// non-deterministic list. For now, unknown = write per policy rule 5.
	p := NewDefaultShadowPolicy()
	got := p.ClassifyTool("search", "")
	if got != ToolClassWrite {
		t.Errorf("ClassifyTool(\"search\", \"\") = %q, want %q (unknown defaults to write)", got, ToolClassWrite)
	}
}

func TestClassifyTool_HTTPPostIsWrite(t *testing.T) {
	p := NewDefaultShadowPolicy()
	methods := []string{"POST", "PUT", "PATCH", "DELETE"}
	for _, m := range methods {
		got := p.ClassifyTool("any_http_tool", m)
		if got != ToolClassWrite {
			t.Errorf("ClassifyTool(\"any_http_tool\", %q) = %q, want %q", m, got, ToolClassWrite)
		}
	}
}

func TestClassifyTool_HTTPGetIsRead(t *testing.T) {
	p := NewDefaultShadowPolicy()
	got := p.ClassifyTool("any_http_tool", "GET")
	if got != ToolClassReadDeterministic {
		t.Errorf("ClassifyTool(\"any_http_tool\", \"GET\") = %q, want %q", got, ToolClassReadDeterministic)
	}
}

func TestClassifyTool_HTTPMethodCaseInsensitive(t *testing.T) {
	p := NewDefaultShadowPolicy()
	got := p.ClassifyTool("any_http_tool", "post")
	if got != ToolClassWrite {
		t.Errorf("ClassifyTool(\"any_http_tool\", \"post\") = %q, want %q", got, ToolClassWrite)
	}
}

func TestMustMock_WriteToolAlwaysTrue(t *testing.T) {
	p := NewDefaultShadowPolicy()
	// Write tools must be mocked regardless of captured output.
	if !p.MustMock("send_email", "", false) {
		t.Error("MustMock(\"send_email\", \"\", false) = false, want true")
	}
	if !p.MustMock("send_email", "", true) {
		t.Error("MustMock(\"send_email\", \"\", true) = false, want true")
	}
}

func TestMustMock_DeterministicReadWithCaptureFalse(t *testing.T) {
	p := NewDefaultShadowPolicy()
	// Deterministic read without capture: can execute live, so MustMock = false.
	if p.MustMock("list_channels", "", false) {
		t.Error("MustMock(\"list_channels\", \"\", false) = true, want false")
	}
}

func TestMustMock_DeterministicReadWithCaptureTrue(t *testing.T) {
	p := NewDefaultShadowPolicy()
	// Deterministic read with capture: prefer captured output, MustMock = true.
	if !p.MustMock("list_channels", "", true) {
		t.Error("MustMock(\"list_channels\", \"\", true) = false, want true")
	}
}

func TestMustMock_NonDeterministicReadWithCaptureTrue(t *testing.T) {
	// Unknown tool "search" defaults to write, so MustMock is always true.
	p := NewDefaultShadowPolicy()
	if !p.MustMock("search", "", true) {
		t.Error("MustMock(\"search\", \"\", true) = false, want true")
	}
}

func TestMustMock_UnknownToolDefaultsToWrite(t *testing.T) {
	p := NewDefaultShadowPolicy()
	// Per policy rule 5, unknown tools default to write classification.
	if !p.MustMock("totally_unknown_tool", "", false) {
		t.Error("MustMock(\"totally_unknown_tool\", \"\", false) = false, want true")
	}
}

func TestShouldAbortIfNoMock_WriteToolTrue(t *testing.T) {
	p := NewDefaultShadowPolicy()
	if !p.ShouldAbortIfNoMock("send_email", "") {
		t.Error("ShouldAbortIfNoMock(\"send_email\", \"\") = false, want true")
	}
}

func TestShouldAbortIfNoMock_DeterministicReadFalse(t *testing.T) {
	p := NewDefaultShadowPolicy()
	if p.ShouldAbortIfNoMock("list_channels", "") {
		t.Error("ShouldAbortIfNoMock(\"list_channels\", \"\") = true, want false")
	}
}

func TestShouldAbortIfNoMock_UnknownToolTrue(t *testing.T) {
	p := NewDefaultShadowPolicy()
	// Unknown tools default to write, so abort is required.
	if !p.ShouldAbortIfNoMock("unknown_tool", "") {
		t.Error("ShouldAbortIfNoMock(\"unknown_tool\", \"\") = false, want true")
	}
}

func TestShouldAbortIfNoMock_HTTPDeleteTrue(t *testing.T) {
	p := NewDefaultShadowPolicy()
	if !p.ShouldAbortIfNoMock("http_tool", "DELETE") {
		t.Error("ShouldAbortIfNoMock(\"http_tool\", \"DELETE\") = false, want true")
	}
}

func TestShouldAbortIfNoMock_HTTPGetFalse(t *testing.T) {
	p := NewDefaultShadowPolicy()
	if p.ShouldAbortIfNoMock("http_tool", "GET") {
		t.Error("ShouldAbortIfNoMock(\"http_tool\", \"GET\") = true, want false")
	}
}
