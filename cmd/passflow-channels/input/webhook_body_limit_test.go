package input

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/passflow-ai/passflow/cmd/passflow-channels/domain"
)

// newTestWebhookAppWithLimit creates a Fiber app with the WebhookHandler registered
// using an explicit body size limit.
func newTestWebhookAppWithLimit(ruleStore WebhookRuleStore, maxBytes int) (*fiber.App, *captureDispatcher) {
	cap := &captureDispatcher{}
	handler := NewWebhookHandler(cap, ruleStore)
	app := fiber.New(fiber.Config{
		BodyLimit: maxBytes,
	})
	handler.Register(app)
	return app, cap
}

// ruleStoreWithSecret is a convenience factory for single-rule test stores.
func ruleStoreWithSecret(workspaceID, secret string) WebhookRuleStore {
	return &mockWebhookRuleStore{
		rules: []domain.TriggerRule{
			{
				ID:            "rule-limit-1",
				WorkspaceID:   workspaceID,
				Enabled:       true,
				ChannelType:   domain.ChannelWebhook,
				WebhookSecret: secret,
			},
		},
	}
}

// TestWebhook_BodyLimitExceeded verifies that a body larger than the configured
// limit is rejected before the handler processes it.
//
// When Fiber's BodyLimit is exceeded, app.Test() returns an error from the
// underlying fasthttp transport (the connection is aborted before the response
// is written). We assert either:
//   - err != nil (body too large, transport error), or
//   - resp.StatusCode == 413 (Fiber managed to write the status before aborting).
//
// In both cases no event must have been dispatched.
func TestWebhook_BodyLimitExceeded(t *testing.T) {
	const maxBytes = 1024 // 1 KiB limit for testing
	ruleStore := ruleStoreWithSecret("ws-limit", "secret")
	app, cap := newTestWebhookAppWithLimit(ruleStore, maxBytes)

	// Build a body that exceeds the limit.
	oversized := strings.Repeat("x", maxBytes+1)
	req, _ := http.NewRequest(http.MethodPost, "/webhook/ws-limit", strings.NewReader(oversized))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", "secret")

	resp, err := app.Test(req, -1)

	// Fiber's fasthttp transport aborts the connection when BodyLimit is
	// exceeded, which surfaces as an error from app.Test rather than a 413
	// response. Either outcome proves the limit is enforced.
	if err == nil && resp.StatusCode != fiber.StatusRequestEntityTooLarge {
		t.Errorf("expected 413 or transport error for oversized body, got status %d with no error", resp.StatusCode)
	}

	// No event should have been dispatched.
	if len(cap.events) != 0 {
		t.Errorf("expected no events dispatched for oversized body, got %d", len(cap.events))
	}
}

// TestWebhook_BodyWithinLimit verifies that a body within the limit is processed normally.
func TestWebhook_BodyWithinLimit(t *testing.T) {
	const maxBytes = 1024 * 1024 // 1 MiB (production default)
	ruleStore := ruleStoreWithSecret("ws-limit2", "secret")
	app, cap := newTestWebhookAppWithLimit(ruleStore, maxBytes)

	body := `{"event":"test","key":"value"}`
	req, _ := http.NewRequest(http.MethodPost, "/webhook/ws-limit2", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", "secret")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 for body within limit, got %d", resp.StatusCode)
	}

	if len(cap.events) != 1 {
		t.Errorf("expected 1 dispatched event, got %d", len(cap.events))
	}
}

// TestWebhook_DefaultAppEnforcesLimit verifies that the production app constant
// MaxWebhookBodyBytes is defined and equals 1 MiB.
func TestWebhook_DefaultAppEnforcesLimit(t *testing.T) {
	const expectedLimit = 1 << 20 // 1 MiB
	if MaxWebhookBodyBytes != expectedLimit {
		t.Errorf("MaxWebhookBodyBytes = %d, want %d", MaxWebhookBodyBytes, expectedLimit)
	}
}
