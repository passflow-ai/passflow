package reporter

import (
	"encoding/json"
	"testing"
)

func TestShadowResultPayload_DefaultSchemaVersion(t *testing.T) {
	payload := &ShadowResultPayload{
		ShadowExecutionID:   "shadow-exec-1",
		OriginalExecutionID: "orig-exec-1",
		AnalysisBatchID:     "batch-1",
		TenantID:            "ws-1",
		AgentID:             "agent-1",
		AgentVersion:        "gpt-4o",
		TraceID:             "trace-1",
		IdempotencyKey:      "batch-1:orig-exec-1:shadow-exec-1",
		Status:              "completed",
		Output:              "shadow output",
		Steps:               json.RawMessage("[]"),
	}

	// Verify the payload can be marshaled to JSON.
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	// Unmarshal and verify fields.
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	if decoded["shadow_execution_id"] != "shadow-exec-1" {
		t.Errorf("expected shadow_execution_id=shadow-exec-1, got %v", decoded["shadow_execution_id"])
	}
	if decoded["status"] != "completed" {
		t.Errorf("expected status=completed, got %v", decoded["status"])
	}
	if decoded["tenant_id"] != "ws-1" {
		t.Errorf("expected tenant_id=ws-1, got %v", decoded["tenant_id"])
	}
}

func TestShadowMetrics_JSONRoundTrip(t *testing.T) {
	m := ShadowMetrics{
		TotalTokens:           1500,
		PromptTokens:          1000,
		CompletionTokens:      500,
		EstimatedCostUSD:      0.015,
		TotalDurationMs:       2500,
		Iterations:            3,
		ToolCallsCount:        5,
		ToolErrorsCount:       1,
		MockedToolsCount:      3,
		PassthroughToolsCount: 2,
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal metrics: %v", err)
	}

	var decoded ShadowMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal metrics: %v", err)
	}

	if decoded.TotalTokens != 1500 {
		t.Errorf("expected TotalTokens=1500, got %d", decoded.TotalTokens)
	}
	if decoded.MockedToolsCount != 3 {
		t.Errorf("expected MockedToolsCount=3, got %d", decoded.MockedToolsCount)
	}
	if decoded.PassthroughToolsCount != 2 {
		t.Errorf("expected PassthroughToolsCount=2, got %d", decoded.PassthroughToolsCount)
	}
	if decoded.EstimatedCostUSD != 0.015 {
		t.Errorf("expected EstimatedCostUSD=0.015, got %f", decoded.EstimatedCostUSD)
	}
}

func TestShadowResultPayload_StatusValues(t *testing.T) {
	statuses := []string{"completed", "failed", "aborted"}
	for _, status := range statuses {
		payload := &ShadowResultPayload{
			ShadowExecutionID:   "exec-1",
			OriginalExecutionID: "orig-1",
			AnalysisBatchID:     "batch-1",
			TenantID:            "ws-1",
			AgentID:             "agent-1",
			AgentVersion:        "v1",
			TraceID:             "trace-1",
			IdempotencyKey:      "key-1",
			Status:              status,
			Steps:               json.RawMessage("[]"),
		}

		data, err := json.Marshal(payload)
		if err != nil {
			t.Errorf("failed to marshal payload with status %q: %v", status, err)
			continue
		}

		var decoded map[string]interface{}
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Errorf("failed to unmarshal payload with status %q: %v", status, err)
			continue
		}

		if decoded["status"] != status {
			t.Errorf("expected status=%q, got %v", status, decoded["status"])
		}
	}
}

func TestShadowResultPayload_ErrorField_OmittedWhenEmpty(t *testing.T) {
	payload := &ShadowResultPayload{
		ShadowExecutionID:   "exec-1",
		OriginalExecutionID: "orig-1",
		AnalysisBatchID:     "batch-1",
		TenantID:            "ws-1",
		AgentID:             "agent-1",
		AgentVersion:        "v1",
		TraceID:             "trace-1",
		IdempotencyKey:      "key-1",
		Status:              "completed",
		Steps:               json.RawMessage("[]"),
		Error:               "",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// The "error" field has omitempty, so it should not appear when empty.
	if _, exists := decoded["error"]; exists {
		t.Error("expected 'error' field to be omitted when empty")
	}
}
