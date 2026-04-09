package gemini

import (
	"io"
	"testing"

	"github.com/jaak-ai/passflow-agent-executor/llm"
)

func TestClient_ImplementsInterface(t *testing.T) {
	var _ llm.Client = (*Client)(nil)
}

// TestClient_ImplementsCloser verifies that *Client satisfies io.Closer so
// callers can release the underlying gRPC connection at shutdown.
func TestClient_ImplementsCloser(t *testing.T) {
	var _ io.Closer = (*Client)(nil)
}

// TestClient_CloseIsIdempotent verifies that calling Close on a nil-inner
// client (simulated via a zero-value struct) does not panic.
// A real genai.Client is never available in unit tests without a live key,
// so we only test the method signature and nil-safety here.
func TestClient_CloseIsIdempotent(t *testing.T) {
	c := &Client{} // client field is nil — Close must handle this gracefully
	// Close should not panic even with a nil inner client.
	// We accept either nil or a non-nil error; what matters is no panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Close() panicked with nil inner client: %v", r)
		}
	}()
	_ = c.Close()
}

func TestConvertMessages(t *testing.T) {
	msgs := []llm.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}

	contents := convertMessages(msgs)
	if len(contents) != 2 {
		t.Errorf("got %d contents, want 2", len(contents))
	}
	if contents[0].Role != "user" {
		t.Errorf("first role = %q, want %q", contents[0].Role, "user")
	}
	if contents[1].Role != "model" {
		t.Errorf("second role = %q, want %q", contents[1].Role, "model")
	}
}

func TestConvertMessages_SkipsSystem(t *testing.T) {
	msgs := []llm.Message{
		{Role: "system", Content: "You are helpful"},
		{Role: "user", Content: "Hello"},
	}

	contents := convertMessages(msgs)
	if len(contents) != 1 {
		t.Errorf("got %d contents, want 1 (system should be skipped)", len(contents))
	}
}
