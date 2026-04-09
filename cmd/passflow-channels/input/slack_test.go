package input

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/passflow-ai/passflow/cmd/passflow-channels/domain"
	"github.com/passflow-ai/passflow/cmd/passflow-channels/trigger"
)

// noopSlackDispatcher is a safe no-op for Slack handler tests.
type noopSlackDispatcher struct{}

func (n *noopSlackDispatcher) Dispatch(_ context.Context, _ domain.Event) {}

type recordingSlackDispatcher struct {
	calls int
	last  domain.Event
}

func (r *recordingSlackDispatcher) Dispatch(_ context.Context, event domain.Event) {
	r.calls++
	r.last = event
}

// Compile-time check: *trigger.Dispatcher satisfies slackDispatcher.
var _ slackDispatcher = (*trigger.Dispatcher)(nil)

// newTestSlackApp creates a Fiber app with SlackHandler registered using a stub dispatcher.
func newTestSlackApp(signingSecret string) *fiber.App {
	return newTestSlackAppWithDispatcher(signingSecret, &noopSlackDispatcher{})
}

func newTestSlackAppWithDispatcher(signingSecret string, dispatcher slackDispatcher) *fiber.App {
	handler := &SlackHandler{
		signingSecret: signingSecret,
		dispatcher:    dispatcher,
	}
	app := fiber.New()
	handler.Register(app)
	return app
}

// buildSlackRequest crafts a signed (or unsigned) POST request to the Slack events endpoint.
func buildSlackRequest(body, signingSecret, workspaceID string, sign bool) *http.Request {
	ts := fmt.Sprintf("%d", time.Now().Unix())
	req := httptest.NewRequest(http.MethodPost,
		"/channels/slack/"+workspaceID+"/events",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	if sign {
		baseString := fmt.Sprintf("v0:%s:%s", ts, body)
		mac := hmac.New(sha256.New, []byte(signingSecret))
		mac.Write([]byte(baseString))
		sig := "v0=" + hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Slack-Request-Timestamp", ts)
		req.Header.Set("X-Slack-Signature", sig)
	}
	return req
}

// TestSlack_EmptySigningSecret_Rejected verifies that when no signing secret is
// configured the handler returns 503 (service misconfigured) for every request.
func TestSlack_EmptySigningSecret_Rejected(t *testing.T) {
	app := newTestSlackApp("") // no secret configured

	body := `{"type":"event_callback","event":{"type":"message","text":"hi","user":"U1","channel":"C1","ts":"1"}}`
	req := buildSlackRequest(body, "", "ws-1", false)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Errorf("expected 503 when signing secret is empty, got %d", resp.StatusCode)
	}
}

// TestSlack_EmptySigningSecret_AlsoRejectsSignedRequests verifies that even if
// a caller sends what looks like a valid signature, we still reject when no secret
// is configured (the service is misconfigured).
func TestSlack_EmptySigningSecret_AlsoRejectsSignedRequests(t *testing.T) {
	app := newTestSlackApp("")

	body := `{"type":"event_callback"}`
	// Caller tries to sign with some key — server has no secret configured.
	req := buildSlackRequest(body, "some-key", "ws-1", true)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Errorf("expected 503 (misconfigured), got %d", resp.StatusCode)
	}
}

// TestSlack_ValidSignature_Accepted verifies that a correctly signed request
// is accepted when the signing secret is set.
func TestSlack_ValidSignature_Accepted(t *testing.T) {
	secret := "test-signing-secret"
	app := newTestSlackApp(secret)

	body := `{"type":"event_callback","event":{"type":"message","text":"hello","user":"U1","channel":"C1","ts":"1"}}`
	req := buildSlackRequest(body, secret, "ws-1", true)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 for valid signature, got %d", resp.StatusCode)
	}
}

// TestSlack_InvalidSignature_Rejected verifies that a wrong signature returns 401.
func TestSlack_InvalidSignature_Rejected(t *testing.T) {
	app := newTestSlackApp("correct-secret")

	body := `{"type":"event_callback"}`
	req := buildSlackRequest(body, "wrong-secret", "ws-1", true)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401 for invalid signature, got %d", resp.StatusCode)
	}
}

// TestSlack_MissingSignatureHeaders_Rejected verifies that a request without
// Slack headers is rejected with 401 (not bypassed) when secret is configured.
func TestSlack_MissingSignatureHeaders_Rejected(t *testing.T) {
	app := newTestSlackApp("some-secret")

	body := `{"type":"event_callback"}`
	req := httptest.NewRequest(http.MethodPost, "/channels/slack/ws-1/events", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No X-Slack-* headers.

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401 when signature headers are absent, got %d", resp.StatusCode)
	}
}

// TestSlack_URLVerification_AllowedWithSignature verifies that the Slack URL
// verification challenge is handled when a proper signature is provided.
func TestSlack_URLVerification_AllowedWithSignature(t *testing.T) {
	secret := "challenge-secret"
	app := newTestSlackApp(secret)

	body := `{"type":"url_verification","challenge":"3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXfDM"}`
	req := buildSlackRequest(body, secret, "ws-1", true)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 for url_verification with valid signature, got %d", resp.StatusCode)
	}
}

func TestSlack_BotMessage_Ignored(t *testing.T) {
	secret := "bot-secret"
	dispatcher := &recordingSlackDispatcher{}
	app := newTestSlackAppWithDispatcher(secret, dispatcher)

	body := `{"type":"event_callback","event":{"type":"message","subtype":"bot_message","bot_id":"B1","text":"agent reply","channel":"C1","ts":"1"}}`
	req := buildSlackRequest(body, secret, "ws-1", true)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for ignored bot event, got %d", resp.StatusCode)
	}
	if dispatcher.calls != 0 {
		t.Fatalf("dispatcher calls = %d, want 0", dispatcher.calls)
	}
}

func TestSlack_PreservesThreadTSInNormalizedFields(t *testing.T) {
	secret := "thread-secret"
	dispatcher := &recordingSlackDispatcher{}
	app := newTestSlackAppWithDispatcher(secret, dispatcher)

	body := `{"type":"event_callback","team_id":"T1","event":{"type":"message","text":"hello","user":"U1","channel":"C1","ts":"1710000000.100","thread_ts":"1710000000.001"}}`
	req := buildSlackRequest(body, secret, "ws-1", true)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if dispatcher.calls != 1 {
		t.Fatalf("dispatcher calls = %d, want 1", dispatcher.calls)
	}
	if got := dispatcher.last.Fields["thread_ts"]; got != "1710000000.001" {
		t.Fatalf("thread_ts = %q, want 1710000000.001", got)
	}
}
