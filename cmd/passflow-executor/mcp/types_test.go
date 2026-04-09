package mcp

import (
	"encoding/json"
	"testing"
)

func TestToolJSON(t *testing.T) {
	tool := Tool{
		Name:        "slack.send_message",
		Description: "Send a message",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
	}

	data, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Tool
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Name != tool.Name {
		t.Errorf("expected %q, got %q", tool.Name, decoded.Name)
	}
}
