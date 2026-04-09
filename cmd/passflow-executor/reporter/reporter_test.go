package reporter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaak-ai/passflow-agent-executor/job"
)

func TestReportStarted_IncludesTriggerMetadata(t *testing.T) {
	var payload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/internal/executions/exec-1/status" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rep := New(server.URL, "svc-token")
	err := rep.ReportStarted(context.Background(), job.Spec{
		ExecutionID: "exec-1",
		AgentID:     "agent-1",
		WorkspaceID: "ws-1",
		TriggerMetadata: &job.TriggerMetadata{
			Source:        "channels-service",
			TriggerRuleID: "rule-1",
			Channel: &job.ChannelTriggerMetadata{
				Type:           "slack",
				ConversationID: "C123",
				ThreadTS:       "1710000000.123",
			},
		},
	})
	if err != nil {
		t.Fatalf("ReportStarted: %v", err)
	}

	trigger, ok := payload["trigger_metadata"].(map[string]any)
	if !ok {
		t.Fatalf("missing trigger_metadata payload: %#v", payload["trigger_metadata"])
	}
	if trigger["source"] != "channels-service" {
		t.Fatalf("source = %#v, want channels-service", trigger["source"])
	}
	if trigger["trigger_rule_id"] != "rule-1" {
		t.Fatalf("trigger_rule_id = %#v, want rule-1", trigger["trigger_rule_id"])
	}
	channel, ok := trigger["channel"].(map[string]any)
	if !ok {
		t.Fatalf("missing channel metadata: %#v", trigger["channel"])
	}
	if channel["thread_ts"] != "1710000000.123" {
		t.Fatalf("thread_ts = %#v, want 1710000000.123", channel["thread_ts"])
	}
}
