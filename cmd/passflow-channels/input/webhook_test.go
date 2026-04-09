package input

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/passflow-ai/passflow/cmd/passflow-channels/domain"
)

// captureDispatcher is a test double for EventDispatcher.
type captureDispatcher struct {
	events []domain.Event
}

func (d *captureDispatcher) Dispatch(event domain.Event) {
	d.events = append(d.events, event)
}

// mockWebhookRuleStore implements WebhookRuleStore for tests.
type mockWebhookRuleStore struct {
	rules []domain.TriggerRule
}

func (s *mockWebhookRuleStore) GetRulesForWorkspace(workspaceID string) []domain.TriggerRule {
	var result []domain.TriggerRule
	for _, r := range s.rules {
		if r.WorkspaceID == workspaceID {
			result = append(result, r)
		}
	}
	return result
}

// newTestWebhookApp creates a Fiber app with the WebhookHandler registered.
func newTestWebhookApp(ruleStore WebhookRuleStore) (*fiber.App, *captureDispatcher) {
	cap := &captureDispatcher{}
	handler := NewWebhookHandler(cap, ruleStore)
	app := fiber.New()
	handler.Register(app)
	return app, cap
}

// doRequest fires a test HTTP request against the Fiber app.
func doRequest(app *fiber.App, method, url, body, secret string) *http.Response {
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if secret != "" {
		req.Header.Set("X-Webhook-Secret", secret)
	}
	resp, _ := app.Test(req, -1)
	return resp
}

// TestWebhook_NoConfiguredSecret_Rejected verifies that when no webhook secret
// is configured for the workspace, the endpoint returns 403 (fail closed).
func TestWebhook_NoConfiguredSecret_Rejected(t *testing.T) {
	// Workspace has a rule but no WebhookSecret configured.
	ruleStore := &mockWebhookRuleStore{
		rules: []domain.TriggerRule{
			{
				ID:            "rule-1",
				WorkspaceID:   "ws-1",
				Enabled:       true,
				ChannelType:   domain.ChannelWebhook,
				WebhookSecret: "", // no secret — must reject
			},
		},
	}
	app, _ := newTestWebhookApp(ruleStore)
	resp := doRequest(app, http.MethodPost, "/webhook/ws-1", `{"event":"test"}`, "")
	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("expected 403 when no webhook secret configured, got %d", resp.StatusCode)
	}
}

// TestWebhook_WrongSecret_Rejected verifies that a wrong secret returns 403.
func TestWebhook_WrongSecret_Rejected(t *testing.T) {
	ruleStore := &mockWebhookRuleStore{
		rules: []domain.TriggerRule{
			{
				ID:            "rule-1",
				WorkspaceID:   "ws-2",
				Enabled:       true,
				ChannelType:   domain.ChannelWebhook,
				WebhookSecret: "correct-secret",
			},
		},
	}
	app, _ := newTestWebhookApp(ruleStore)
	resp := doRequest(app, http.MethodPost, "/webhook/ws-2", `{"event":"test"}`, "wrong-secret")
	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("expected 403 for wrong secret, got %d", resp.StatusCode)
	}
}

// TestWebhook_CorrectSecret_Accepted verifies that the correct secret allows the request.
func TestWebhook_CorrectSecret_Accepted(t *testing.T) {
	ruleStore := &mockWebhookRuleStore{
		rules: []domain.TriggerRule{
			{
				ID:            "rule-1",
				WorkspaceID:   "ws-3",
				Enabled:       true,
				ChannelType:   domain.ChannelWebhook,
				WebhookSecret: "my-secret",
			},
		},
	}
	app, _ := newTestWebhookApp(ruleStore)
	resp := doRequest(app, http.MethodPost, "/webhook/ws-3", `{"event":"test"}`, "my-secret")
	if resp.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 200 for correct secret, got %d: %s", resp.StatusCode, body)
	}
}

// TestWebhook_NoRulesForWorkspace_Rejected verifies that a workspace with no
// rules configured rejects webhook requests (no secret can be verified).
func TestWebhook_NoRulesForWorkspace_Rejected(t *testing.T) {
	ruleStore := &mockWebhookRuleStore{rules: nil}
	app, _ := newTestWebhookApp(ruleStore)
	resp := doRequest(app, http.MethodPost, "/webhook/unknown-ws", `{}`, "any-secret")
	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("expected 403 for workspace with no rules, got %d", resp.StatusCode)
	}
}

// TestWebhook_EmptySecretHeader_Rejected verifies that omitting the secret
// header is rejected even when a secret IS configured.
func TestWebhook_EmptySecretHeader_Rejected(t *testing.T) {
	ruleStore := &mockWebhookRuleStore{
		rules: []domain.TriggerRule{
			{
				ID:            "rule-1",
				WorkspaceID:   "ws-4",
				Enabled:       true,
				ChannelType:   domain.ChannelWebhook,
				WebhookSecret: "required-secret",
			},
		},
	}
	app, _ := newTestWebhookApp(ruleStore)
	// No secret header sent at all.
	resp := doRequest(app, http.MethodPost, "/webhook/ws-4", `{}`, "")
	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("expected 403 when secret header is missing, got %d", resp.StatusCode)
	}
}

// TestWebhook_MultipleRules_SecretFromAnyRule verifies that a correct secret
// shared by multiple rules for the same workspace grants access.
func TestWebhook_MultipleRules_SecretFromAnyRule(t *testing.T) {
	ruleStore := &mockWebhookRuleStore{
		rules: []domain.TriggerRule{
			{ID: "r1", WorkspaceID: "ws-5", Enabled: true, ChannelType: domain.ChannelWebhook, WebhookSecret: "secret-A"},
			{ID: "r2", WorkspaceID: "ws-5", Enabled: true, ChannelType: domain.ChannelWebhook, WebhookSecret: "secret-A"},
		},
	}
	app, _ := newTestWebhookApp(ruleStore)
	resp := doRequest(app, http.MethodPost, "/webhook/ws-5", `{}`, "secret-A")
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// TestWebhook_EventDispatchedOnSuccess verifies that a successful auth leads
// to the event being dispatched.
func TestWebhook_EventDispatchedOnSuccess(t *testing.T) {
	ruleStore := &mockWebhookRuleStore{
		rules: []domain.TriggerRule{
			{ID: "r1", WorkspaceID: "ws-6", Enabled: true, ChannelType: domain.ChannelWebhook, WebhookSecret: "tok"},
		},
	}
	app, cap := newTestWebhookApp(ruleStore)
	resp := doRequest(app, http.MethodPost, "/webhook/ws-6", `{"key":"val"}`, "tok")
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if len(cap.events) != 1 {
		t.Errorf("expected 1 dispatched event, got %d", len(cap.events))
	}
	if cap.events[0].WorkspaceID != "ws-6" {
		t.Errorf("expected workspace ws-6, got %q", cap.events[0].WorkspaceID)
	}
}

// TestWebhookHandler_LogsEvent verifies that webhook events are logged to Redis
// after successful dispatch.
func TestWebhookHandler_LogsEvent(t *testing.T) {
	ruleStore := &mockWebhookRuleStore{
		rules: []domain.TriggerRule{
			{ID: "rule-webhook-1", WorkspaceID: "ws-log", Enabled: true, ChannelType: domain.ChannelWebhook, WebhookSecret: "secret123"},
		},
	}

	// Create a mock event logger that captures log calls with sync channel.
	mockLogger := &mockEventLogger{
		logged: make([]WebhookEvent, 0),
		done:   make(chan struct{}, 1),
	}

	cap := &captureDispatcher{}
	handler := NewWebhookHandlerWithLogger(cap, ruleStore, mockLogger)
	app := fiber.New()
	handler.Register(app)

	// Send a webhook request.
	payload := `{"action":"order.created","order_id":"12345"}`
	req := httptest.NewRequest(http.MethodPost, "/webhook/ws-log", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", "secret123")

	resp, _ := app.Test(req, -1)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Wait for async logging to complete.
	<-mockLogger.done

	if len(mockLogger.logged) == 0 {
		t.Fatal("expected event to be logged, but no events were logged")
	}

	event := mockLogger.logged[0]
	if event.RuleID != "rule-webhook-1" {
		t.Errorf("expected RuleID 'rule-webhook-1', got %q", event.RuleID)
	}
	if event.StatusCode != 200 {
		t.Errorf("expected StatusCode 200, got %d", event.StatusCode)
	}
	if event.PayloadPreview == "" {
		t.Errorf("expected non-empty PayloadPreview")
	}
	if len(event.PayloadPreview) > 103 {
		t.Errorf("expected PayloadPreview to be truncated to ~100 chars, got %d", len(event.PayloadPreview))
	}
	if event.ErrorMessage != "" {
		t.Errorf("expected empty ErrorMessage, got %q", event.ErrorMessage)
	}
}

// mockEventLogger is a test double for WebhookEventLogger.
type mockEventLogger struct {
	logged []WebhookEvent
	done   chan struct{}
}

func (m *mockEventLogger) Log(ctx context.Context, event WebhookEvent) error {
	m.logged = append(m.logged, event)
	if m.done != nil {
		m.done <- struct{}{}
	}
	return nil
}
