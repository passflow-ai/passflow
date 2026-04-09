package trigger

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/passflow-ai/passflow/cmd/passflow-channels/domain"
)

// stubRuleStore satisfies RuleStore for dispatcher tests.
type stubRuleStore struct {
	rules []domain.TriggerRule
}

func (s *stubRuleStore) GetRulesForWorkspace(_ string) []domain.TriggerRule {
	return s.rules
}
func (s *stubRuleStore) GetAllRules() []domain.TriggerRule {
	return s.rules
}

// buildAlwaysRule returns a rule that always matches for the given channel type.
func buildAlwaysRule(id, agentID string, channelType domain.ChannelType) domain.TriggerRule {
	return domain.TriggerRule{
		ID:          id,
		Name:        id,
		Enabled:     true,
		ChannelType: channelType,
		Condition:   domain.Condition{Operator: "always"},
		Action: domain.Action{
			AgentID:       agentID,
			ExecutionMode: "one_shot",
		},
	}
}

// TestDispatcher_ExecutesSynchronously verifies that actions are executed
// synchronously (no goroutines) to let the consumer control rate limiting.
func TestDispatcher_ExecutesSynchronously(t *testing.T) {
	const numRules = 5

	var callOrder []int
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callOrder = append(callOrder, len(callOrder))
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	rules := make([]domain.TriggerRule, numRules)
	for i := range rules {
		rules[i] = buildAlwaysRule("rule-"+string(rune('A'+i)), "agent-1", domain.ChannelSlack)
	}

	store := &stubRuleStore{rules: rules}
	d := NewDispatcher(nil, "stream", server.URL, "test-token", store)

	event := domain.Event{
		ID:          "evt-1",
		WorkspaceID: "ws-1",
		Channel:     domain.ChannelSlack,
		Fields:      map[string]string{"text": "hello"},
		ReceivedAt:  time.Now(),
	}

	// Dispatch should complete synchronously
	d.Dispatch(context.Background(), event)

	// All actions should have fired by the time Dispatch returns
	mu.Lock()
	got := len(callOrder)
	mu.Unlock()

	if got != numRules {
		t.Errorf("expected %d actions fired synchronously, got %d", numRules, got)
	}
}

// TestDispatcher_AllMatchingRulesFired verifies that all matching rules fire
// their actions, none are dropped.
func TestDispatcher_AllMatchingRulesFired(t *testing.T) {
	const numRules = 15

	var fired int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&fired, 1)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	rules := make([]domain.TriggerRule, numRules)
	for i := range rules {
		rules[i] = buildAlwaysRule("rule-all-"+string(rune('A'+i)), "agent-2", domain.ChannelSlack)
	}

	store := &stubRuleStore{rules: rules}
	d := NewDispatcher(nil, "stream", server.URL, "test-token", store)

	event := domain.Event{
		ID:          "evt-2",
		WorkspaceID: "ws-1",
		Channel:     domain.ChannelSlack,
		Fields:      map[string]string{"text": "hello"},
		ReceivedAt:  time.Now(),
	}

	d.Dispatch(context.Background(), event)

	// Actions fire synchronously, so no need to wait
	if got := atomic.LoadInt64(&fired); got != numRules {
		t.Errorf("expected %d actions fired, got %d", numRules, got)
	}
}

func TestDispatcher_PipelineTargetCallsInternalWorkflowExecute(t *testing.T) {
	var path string
	var token string
	var payload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		token = r.Header.Get("X-Service-Token")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &payload)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	store := &stubRuleStore{rules: []domain.TriggerRule{
		{
			ID:          "rule-pipeline",
			Name:        "pipeline",
			Enabled:     true,
			ChannelType: domain.ChannelWebhook,
			Condition:   domain.Condition{Operator: "always"},
			Action: domain.Action{
				TargetType:    domain.TargetPipeline,
				TargetID:      "wf-123",
				InputTemplate: "webhook_input",
			},
		},
	}}
	d := NewDispatcher(nil, "stream", server.URL, "svc-token", store)

	d.Dispatch(context.Background(), domain.Event{
		ID:          "evt-pipeline",
		WorkspaceID: "ws-123",
		Channel:     domain.ChannelWebhook,
		Fields:      map[string]string{"text": "hello"},
		ReceivedAt:  time.Now(),
	})

	if path != "/api/v1/internal/workspaces/ws-123/workflows/wf-123/execute" {
		t.Fatalf("unexpected path: %s", path)
	}
	if token != "svc-token" {
		t.Fatalf("unexpected service token: %s", token)
	}
	if payload["triggerType"] != "webhook" {
		t.Fatalf("unexpected triggerType: %#v", payload["triggerType"])
	}
}

func TestDispatcher_AgentTargetCallsInternalAgentExecute(t *testing.T) {
	var path string
	var token string
	var payload AgentExecutionRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		token = r.Header.Get("X-Service-Token")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &payload)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	store := &stubRuleStore{rules: []domain.TriggerRule{
		{
			ID:          "rule-agent",
			Name:        "agent",
			Enabled:     true,
			ChannelType: domain.ChannelWhatsApp,
			Condition:   domain.Condition{Operator: "always"},
			Action: domain.Action{
				TargetType:    domain.TargetAgent,
				TargetID:      "agent-123",
				ExecutionMode: "react_loop",
				MaxIterations: 3,
			},
		},
	}}
	d := NewDispatcher(nil, "stream", server.URL, "svc-token", store)

	d.Dispatch(context.Background(), domain.Event{
		ID:          "evt-agent",
		WorkspaceID: "ws-123",
		Channel:     domain.ChannelWhatsApp,
		Fields: map[string]string{
			"text":            "hola",
			"from":            "5215512345678",
			"phone_number_id": "phone-42",
			"message_id":      "wamid.123",
		},
		ReceivedAt: time.Now(),
	})

	if path != "/api/v1/internal/workspaces/ws-123/agents/agent-123/execute" {
		t.Fatalf("unexpected path: %s", path)
	}
	if token != "svc-token" {
		t.Fatalf("unexpected service token: %s", token)
	}
	if payload.Input != "hola" {
		t.Fatalf("unexpected input: %s", payload.Input)
	}
	if payload.ExecutionMode != "react_loop" {
		t.Fatalf("unexpected execution mode: %s", payload.ExecutionMode)
	}
	if payload.MaxIterations != 3 {
		t.Fatalf("unexpected max iterations: %d", payload.MaxIterations)
	}
	if payload.Source != "passflow-channels" {
		t.Fatalf("unexpected source: %s", payload.Source)
	}
	if payload.TriggerRuleID != "rule-agent" {
		t.Fatalf("unexpected trigger rule ID: %s", payload.TriggerRuleID)
	}
	if payload.Channel == nil {
		t.Fatal("expected channel metadata")
	}
	if payload.Channel.Type != "whatsapp" {
		t.Fatalf("unexpected channel type: %s", payload.Channel.Type)
	}
	if payload.Channel.EventID != "evt-agent" {
		t.Fatalf("unexpected event ID: %s", payload.Channel.EventID)
	}
	if payload.Channel.SenderID != "5215512345678" {
		t.Fatalf("unexpected sender ID: %s", payload.Channel.SenderID)
	}
	if payload.Channel.AccountID != "phone-42" {
		t.Fatalf("unexpected account ID: %s", payload.Channel.AccountID)
	}
	if payload.Channel.MessageID != "wamid.123" {
		t.Fatalf("unexpected message ID: %s", payload.Channel.MessageID)
	}
	if payload.Channel.SessionScope != "external_sender" {
		t.Fatalf("unexpected session scope: %s", payload.Channel.SessionScope)
	}
	if payload.Channel.SessionKey != "whatsapp:phone-42:5215512345678" {
		t.Fatalf("unexpected session key: %s", payload.Channel.SessionKey)
	}
}
