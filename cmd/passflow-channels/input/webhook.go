package input

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jaak-ai/passflow-channels/domain"
)

// MaxWebhookBodyBytes is the maximum permitted size of a webhook request body.
// Requests exceeding this limit are rejected with HTTP 413 before any parsing
// occurs, preventing memory exhaustion from oversized payloads.
// Callers that embed the WebhookHandler should configure their Fiber app with
//
//	fiber.Config{BodyLimit: MaxWebhookBodyBytes}
const MaxWebhookBodyBytes = 1 << 20 // 1 MiB

// EventDispatcher is the interface for dispatching normalized events.
// Satisfied by *trigger.Dispatcher.
type EventDispatcher interface {
	Dispatch(event domain.Event)
}

// WebhookRuleStore is the read-only subset of store.RuleStore needed by
// WebhookHandler to look up per-workspace secrets.
type WebhookRuleStore interface {
	GetRulesForWorkspace(workspaceID string) []domain.TriggerRule
}

// EventLogger is the interface for logging webhook events.
// Satisfied by *WebhookEventLogger.
type EventLogger interface {
	Log(ctx context.Context, event WebhookEvent) error
}

// WebhookHandler handles incoming generic webhooks.
// Route: POST /webhook/:workspaceId
// Route: POST /webhook/:workspaceId/:ruleHint
//
// Authentication is performed based on each rule's AuthStrategy configuration.
// Supported strategies: custom_header (legacy), github_hmac, stripe_signature, shopify_hmac, none.
// If no matching authenticated rule is found the request is rejected with 403.
type WebhookHandler struct {
	dispatcher   EventDispatcher
	ruleStore    WebhookRuleStore
	authVerifier *AuthVerifier
	eventLogger  EventLogger
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(dispatcher EventDispatcher, ruleStore WebhookRuleStore) *WebhookHandler {
	return &WebhookHandler{
		dispatcher:   dispatcher,
		ruleStore:    ruleStore,
		authVerifier: NewAuthVerifier(),
		eventLogger:  nil,
	}
}

// NewWebhookHandlerWithLogger creates a new WebhookHandler with event logging enabled.
func NewWebhookHandlerWithLogger(dispatcher EventDispatcher, ruleStore WebhookRuleStore, eventLogger EventLogger) *WebhookHandler {
	return &WebhookHandler{
		dispatcher:   dispatcher,
		ruleStore:    ruleStore,
		authVerifier: NewAuthVerifier(),
		eventLogger:  eventLogger,
	}
}

// Register registers webhook routes on the Fiber app.
func (h *WebhookHandler) Register(app *fiber.App) {
	app.Post("/webhook/:workspaceId", h.handle)
	app.Post("/webhook/:workspaceId/:ruleHint", h.handle)
}

func (h *WebhookHandler) handle(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	body := c.Body()

	// Extract headers into a map for auth verification.
	headers := make(map[string]string)
	c.Request().Header.VisitAll(func(k, v []byte) {
		headers[string(k)] = string(v)
	})

	// Authenticate the request against the workspace's configured rules.
	if err := h.authenticate(workspaceID, headers, body); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}

	// Parse body as generic JSON.
	var raw map[string]any
	if err := c.BodyParser(&raw); err != nil {
		// Fall back to raw string body.
		raw = map[string]any{"body": string(body)}
	}

	// Normalize fields.
	fields := flattenFields(raw)

	// Add headers as fields.
	for k, v := range headers {
		fields["header."+k] = v
	}

	event := domain.Event{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		Channel:     domain.ChannelWebhook,
		Fields:      fields,
		Raw:         raw,
		ReceivedAt:  time.Now(),
	}

	h.dispatcher.Dispatch(event)

	// Log event asynchronously if logger is configured.
	if h.eventLogger != nil {
		matchedRuleID := h.findMatchingRuleID(workspaceID, headers, body)
		if matchedRuleID != "" {
			// Use background context for async logging to avoid using Fiber's context
			// after the request completes, which would cause a race condition.
			go func() {
				webhookEvent := WebhookEvent{
					RuleID:         matchedRuleID,
					Timestamp:      time.Now(),
					StatusCode:     fiber.StatusOK,
					PayloadPreview: truncatePayload(string(body), 100),
					ErrorMessage:   "",
				}
				_ = h.eventLogger.Log(context.Background(), webhookEvent)
			}()
		}
	}

	return c.JSON(fiber.Map{"ok": true, "event_id": event.ID})
}

// authenticate checks that the request is authenticated against at least one
// webhook rule for the workspace. It supports multiple authentication strategies
// including custom headers, GitHub HMAC, Stripe signatures, and Shopify HMAC.
// Returns an error if authentication fails for all rules.
func (h *WebhookHandler) authenticate(workspaceID string, headers map[string]string, body []byte) error {
	rules := h.ruleStore.GetRulesForWorkspace(workspaceID)

	// Collect webhook rules with their effective auth strategies.
	var webhookRules []domain.TriggerRule
	for _, r := range rules {
		if r.ChannelType == domain.ChannelWebhook {
			auth := GetEffectiveAuth(r)
			if auth != nil {
				webhookRules = append(webhookRules, r)
			}
		}
	}

	// Fail closed: if no auth is configured, reject the request.
	if len(webhookRules) == 0 {
		return fmt.Errorf("no webhook auth configured for workspace %q", workspaceID)
	}

	// Try to authenticate against each rule's strategy.
	for _, r := range webhookRules {
		auth := GetEffectiveAuth(r)
		if err := h.authVerifier.Verify(auth, headers, body); err == nil {
			return nil
		}
	}

	return fmt.Errorf("webhook authentication failed for workspace %q", workspaceID)
}

// flattenFields extracts top-level string values from a map into a flat key-value map.
func flattenFields(m map[string]any) map[string]string {
	fields := make(map[string]string)
	for k, v := range m {
		switch val := v.(type) {
		case string:
			fields[k] = val
		case fmt.Stringer:
			fields[k] = val.String()
		default:
			if b, err := json.Marshal(v); err == nil {
				fields[k] = string(b)
			}
		}
	}
	return fields
}

// findMatchingRuleID attempts to find the rule ID that authenticated the request.
// Returns the first rule ID that matches the request, or empty string if none found.
func (h *WebhookHandler) findMatchingRuleID(workspaceID string, headers map[string]string, body []byte) string {
	rules := h.ruleStore.GetRulesForWorkspace(workspaceID)

	for _, r := range rules {
		if r.ChannelType == domain.ChannelWebhook {
			auth := GetEffectiveAuth(r)
			if auth != nil && h.authVerifier.Verify(auth, headers, body) == nil {
				return r.ID
			}
		}
	}

	return ""
}

// DispatcherFunc is a function type that satisfies EventDispatcher.
// It enables wiring *trigger.Dispatcher (which takes a context) to the
// handler without changing the dispatcher's public API.
type DispatcherFunc func(event domain.Event)

func (f DispatcherFunc) Dispatch(event domain.Event) { f(event) }
