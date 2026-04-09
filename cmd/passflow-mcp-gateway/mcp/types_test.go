package mcp

import (
	"encoding/json"
	"testing"
)

func TestToolDefinitionJSON(t *testing.T) {
	tool := Tool{
		Name:        "slack.send_message",
		Description: "Send a message to Slack",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"channel": map[string]interface{}{"type": "string"},
				"text":    map[string]interface{}{"type": "string"},
			},
			"required": []string{"channel", "text"},
		},
	}

	data, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded Tool
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Name != tool.Name {
		t.Errorf("expected name %q, got %q", tool.Name, decoded.Name)
	}
}

func TestCallRequestJSON(t *testing.T) {
	req := CallRequest{
		Name: "slack.send_message",
		Arguments: map[string]interface{}{
			"channel": "#general",
			"text":    "Hello",
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CallRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Name != req.Name {
		t.Errorf("expected name %q, got %q", req.Name, decoded.Name)
	}
}
