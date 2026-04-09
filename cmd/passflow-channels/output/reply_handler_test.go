package output

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jaak-ai/passflow-channels/domain"
)

type stubDeliverySender struct {
	channel *domain.OutputChannel
	content string
	err     error
	calls   int
}

func (s *stubDeliverySender) Send(_ context.Context, ch *domain.OutputChannel, content string) error {
	s.calls++
	s.channel = ch
	s.content = content
	return s.err
}

type stubRuleLookup struct {
	rules map[string]*domain.TriggerRule
}

func (s *stubRuleLookup) GetByID(ruleID string) *domain.TriggerRule {
	return s.rules[ruleID]
}

func TestReplyHandler_DeliversSlackReplyUsingConversationID(t *testing.T) {
	sender := &stubDeliverySender{}
	store := &stubRuleLookup{
		rules: map[string]*domain.TriggerRule{
			"rule-1": {
				ID:          "rule-1",
				WorkspaceID: "ws-1",
				Action: domain.Action{
					OutputChannel: &domain.OutputChannel{
						Type: domain.ChannelSlack,
						Config: map[string]any{
							"token": "xoxb-test-token",
						},
					},
				},
			},
		},
	}

	app := fiber.New()
	NewReplyHandler(sender, store, nil).RegisterInternal(app)

	body := `{
		"workspace_id":"ws-1",
		"execution_id":"exec-1",
		"trigger_rule_id":"rule-1",
		"output":"hola mundo",
		"channel":{
			"type":"slack",
			"conversation_id":"C123"
		}
	}`
	req := httptest.NewRequest(http.MethodPost, "/execution-replies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result deliveryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.Status != "delivered" {
		t.Fatalf("status = %s, want delivered", result.Status)
	}
	if sender.calls != 1 {
		t.Fatalf("sender calls = %d, want 1", sender.calls)
	}
	if sender.content != "hola mundo" {
		t.Fatalf("content = %s, want hola mundo", sender.content)
	}
	if got, _ := sender.channel.Config["channel"].(string); got != "C123" {
		t.Fatalf("channel config = %s, want C123", got)
	}
}

func TestReplyHandler_DeliversSlackReplyUsingThreadTSWhenAvailable(t *testing.T) {
	sender := &stubDeliverySender{}
	store := &stubRuleLookup{
		rules: map[string]*domain.TriggerRule{
			"rule-1": {
				ID:          "rule-1",
				WorkspaceID: "ws-1",
				Action: domain.Action{
					OutputChannel: &domain.OutputChannel{
						Type: domain.ChannelSlack,
						Config: map[string]any{
							"token": "xoxb-test-token",
						},
					},
				},
			},
		},
	}

	app := fiber.New()
	NewReplyHandler(sender, store, nil).RegisterInternal(app)

	body := `{
		"workspace_id":"ws-1",
		"execution_id":"exec-thread-1",
		"trigger_rule_id":"rule-1",
		"output":"hola hilo",
		"channel":{
			"type":"slack",
			"conversation_id":"C123",
			"thread_ts":"1710000000.001"
		}
	}`
	req := httptest.NewRequest(http.MethodPost, "/execution-replies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got, _ := sender.channel.Config["thread_ts"].(string); got != "1710000000.001" {
		t.Fatalf("thread_ts = %s, want 1710000000.001", got)
	}
}

func TestReplyHandler_SkipsWhenNoOutputChannelConfigured(t *testing.T) {
	sender := &stubDeliverySender{}
	store := &stubRuleLookup{
		rules: map[string]*domain.TriggerRule{
			"rule-1": {
				ID:          "rule-1",
				WorkspaceID: "ws-1",
				Action:      domain.Action{},
			},
		},
	}

	app := fiber.New()
	NewReplyHandler(sender, store, nil).RegisterInternal(app)

	body := `{
		"workspace_id":"ws-1",
		"execution_id":"exec-1",
		"trigger_rule_id":"rule-1",
		"output":"hola mundo"
	}`
	req := httptest.NewRequest(http.MethodPost, "/execution-replies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if sender.calls != 0 {
		t.Fatalf("sender calls = %d, want 0", sender.calls)
	}

	var result deliveryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.Status != "skipped" || result.Reason != "no_output_channel" {
		t.Fatalf("unexpected response: %+v", result)
	}
}

func TestReplyHandler_RejectsWorkspaceMismatch(t *testing.T) {
	sender := &stubDeliverySender{}
	store := &stubRuleLookup{
		rules: map[string]*domain.TriggerRule{
			"rule-1": {
				ID:          "rule-1",
				WorkspaceID: "ws-2",
				Action: domain.Action{
					OutputChannel: &domain.OutputChannel{Type: domain.ChannelWebhook, Config: map[string]any{"url": "https://example.com"}},
				},
			},
		},
	}

	app := fiber.New()
	NewReplyHandler(sender, store, nil).RegisterInternal(app)

	body := `{
		"workspace_id":"ws-1",
		"execution_id":"exec-1",
		"trigger_rule_id":"rule-1",
		"output":"hola mundo"
	}`
	req := httptest.NewRequest(http.MethodPost, "/execution-replies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
	if sender.calls != 0 {
		t.Fatalf("sender calls = %d, want 0", sender.calls)
	}
}

func TestReplyHandler_DerivesSMSReplyAddressesFromChannelMetadata(t *testing.T) {
	sender := &stubDeliverySender{}
	store := &stubRuleLookup{
		rules: map[string]*domain.TriggerRule{
			"rule-1": {
				ID:          "rule-1",
				WorkspaceID: "ws-1",
				Action: domain.Action{
					OutputChannel: &domain.OutputChannel{
						Type: domain.ChannelSMS,
						Config: map[string]any{
							"account_sid": "TEST_ACCT_00000000000000000000",
							"auth_token":  "secret",
						},
					},
				},
			},
		},
	}

	app := fiber.New()
	NewReplyHandler(sender, store, nil).RegisterInternal(app)

	body := `{
		"workspace_id":"ws-1",
		"execution_id":"exec-2",
		"trigger_rule_id":"rule-1",
		"output":"hola sms",
		"channel":{
			"type":"sms",
			"sender_id":"+15551234567",
			"account_id":"+15557654321"
		}
	}`
	req := httptest.NewRequest(http.MethodPost, "/execution-replies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got, _ := sender.channel.Config["to"].(string); got != "+15551234567" {
		t.Fatalf("to = %s, want +15551234567", got)
	}
	if got, _ := sender.channel.Config["from"].(string); got != "+15557654321" {
		t.Fatalf("from = %s, want +15557654321", got)
	}
}

func TestReplyHandler_DerivesWhatsAppReplyTargetFromChannelMetadata(t *testing.T) {
	sender := &stubDeliverySender{}
	store := &stubRuleLookup{
		rules: map[string]*domain.TriggerRule{
			"rule-1": {
				ID:          "rule-1",
				WorkspaceID: "ws-1",
				Action: domain.Action{
					OutputChannel: &domain.OutputChannel{
						Type: domain.ChannelWhatsApp,
						Config: map[string]any{
							"token": "meta-token",
						},
					},
				},
			},
		},
	}

	app := fiber.New()
	NewReplyHandler(sender, store, nil).RegisterInternal(app)

	body := `{
		"workspace_id":"ws-1",
		"execution_id":"exec-3",
		"trigger_rule_id":"rule-1",
		"output":"hola whatsapp",
		"channel":{
			"type":"whatsapp",
			"sender_id":"5215512345678",
			"account_id":"phone-42"
		}
	}`
	req := httptest.NewRequest(http.MethodPost, "/execution-replies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got, _ := sender.channel.Config["to"].(string); got != "5215512345678" {
		t.Fatalf("to = %s, want 5215512345678", got)
	}
	if got, _ := sender.channel.Config["phone_number_id"].(string); got != "phone-42" {
		t.Fatalf("phone_number_id = %s, want phone-42", got)
	}
}
