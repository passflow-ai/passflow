package acf

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseToolCall_Anthropic(t *testing.T) {
	raw := json.RawMessage(`{"name":"Read","input":{"file_path":"/tmp/test.go"}}`)
	tc, err := ParseToolCall("anthropic", "call_123", raw)
	require.NoError(t, err)
	assert.Equal(t, "Read", tc.Tool)
	assert.Equal(t, "call_123", tc.CallID)
	assert.Equal(t, "/tmp/test.go", tc.Params["file_path"])
}

func TestParseToolCall_OpenAI(t *testing.T) {
	raw := json.RawMessage(`{"name":"Read","arguments":"{\"file_path\":\"/tmp/test.go\"}"}`)
	tc, err := ParseToolCall("openai", "call_456", raw)
	require.NoError(t, err)
	assert.Equal(t, "Read", tc.Tool)
	assert.Equal(t, "/tmp/test.go", tc.Params["file_path"])
}

func TestFormatToolResult_Anthropic(t *testing.T) {
	result := &ToolResult{CallID: "c1", Output: "file contents", Error: ""}
	formatted, err := FormatToolResult("anthropic", result)
	require.NoError(t, err)
	assert.Contains(t, string(formatted), "file contents")
}

func TestFormatToolResult_OpenAI(t *testing.T) {
	result := &ToolResult{CallID: "c1", Output: "file contents", Error: ""}
	formatted, err := FormatToolResult("openai", result)
	require.NoError(t, err)
	assert.Contains(t, string(formatted), "file contents")
}

func TestFormatToolResult_WithError(t *testing.T) {
	result := &ToolResult{CallID: "c1", Error: "permission denied"}
	formatted, err := FormatToolResult("anthropic", result)
	require.NoError(t, err)
	assert.Contains(t, string(formatted), "permission denied")
}
