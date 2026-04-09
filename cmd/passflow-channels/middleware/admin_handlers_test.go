package middleware

import (
	"context"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/jaak-ai/passflow-channels/domain"
	"github.com/jaak-ai/passflow-channels/input"
	"github.com/jaak-ai/passflow-channels/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHandlers(t *testing.T) (*AdminHandlers, *miniredis.Miniredis, *store.RuleStore, *input.WebhookEventLogger) {
	t.Helper()

	mr := miniredis.NewMiniRedis()
	require.NoError(t, mr.Start())

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()

	ruleStore := store.New(ctx, rdb)
	eventLogger := input.NewWebhookEventLogger(rdb)
	handlers := NewAdminHandlers(ruleStore, eventLogger)

	return handlers, mr, ruleStore, eventLogger
}

func TestGetRule_Success(t *testing.T) {
	handlers, mr, ruleStore, eventLogger := setupTestHandlers(t)
	defer mr.Close()

	ctx := context.Background()

	// Create a test rule
	rule := domain.TriggerRule{
		ID:          "rule-123",
		WorkspaceID: "ws-456",
		Name:        "Test Webhook Rule",
		Enabled:     true,
		ChannelType: domain.ChannelWebhook,
		Condition: domain.Condition{
			Field:    "body.action",
			Operator: "equals",
			Value:    "deploy",
		},
		Action: domain.Action{
			TargetType:    domain.TargetAgent,
			TargetID:      "agent-789",
			InputTemplate: "Deploy triggered: {{.body.ref}}",
			ExecutionMode: "one_shot",
			OutputChannel: &domain.OutputChannel{
				Type: domain.ChannelSlack,
				Config: map[string]any{
					"token":   "xoxb-super-secret-token",
					"channel": "C123",
				},
			},
		},
		Auth: &domain.AuthStrategy{
			Type:   domain.AuthGitHubHMAC,
			Secret: "webhook-secret-value",
		},
		CreatedAt: time.Now(),
	}
	require.NoError(t, ruleStore.Upsert(ctx, rule))

	// Log some webhook events
	event1 := input.WebhookEvent{
		RuleID:         "rule-123",
		Timestamp:      time.Now(),
		StatusCode:     200,
		PayloadPreview: `{"action":"deploy","ref":"main"}`,
	}
	require.NoError(t, eventLogger.Log(ctx, event1))

	event2 := input.WebhookEvent{
		RuleID:         "rule-123",
		Timestamp:      time.Now().Add(-1 * time.Minute),
		StatusCode:     401,
		PayloadPreview: `{"action":"test"}`,
		ErrorMessage:   "authentication failed",
	}
	require.NoError(t, eventLogger.Log(ctx, event2))

	// Create Fiber app and register route
	app := fiber.New()
	app.Get("/admin/v1/rules/:ruleId", handlers.GetRule)

	// Make request
	req := httptest.NewRequest("GET", "/admin/v1/rules/rule-123", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Check that response contains expected fields
	bodyStr := string(body)
	assert.Contains(t, bodyStr, `"id":"rule-123"`)
	assert.Contains(t, bodyStr, `"workspace_id":"ws-456"`)
	assert.Contains(t, bodyStr, `"name":"Test Webhook Rule"`)
	assert.Contains(t, bodyStr, `"enabled":true`)
	assert.Contains(t, bodyStr, `"channel_type":"webhook"`)
	assert.Contains(t, bodyStr, `"recent_events"`)
	assert.Contains(t, bodyStr, `"deploy"`)
	assert.Contains(t, bodyStr, `"authentication failed"`)
	assert.Contains(t, bodyStr, `"secret_masked":"webh****"`)
	assert.NotContains(t, bodyStr, "webhook-secret-value")
	assert.NotContains(t, bodyStr, "xoxb-super-secret-token")
}

func TestGetRule_NotFound(t *testing.T) {
	handlers, mr, _, _ := setupTestHandlers(t)
	defer mr.Close()

	app := fiber.New()
	app.Get("/admin/v1/rules/:ruleId", handlers.GetRule)

	req := httptest.NewRequest("GET", "/admin/v1/rules/nonexistent", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "rule not found")
}

func TestGetRule_MissingRuleId(t *testing.T) {
	handlers, mr, _, _ := setupTestHandlers(t)
	defer mr.Close()

	app := fiber.New()
	app.Get("/admin/v1/rules/:ruleId", handlers.GetRule)

	req := httptest.NewRequest("GET", "/admin/v1/rules/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Fiber will return 404 for missing route parameter
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestGetRule_NoEvents(t *testing.T) {
	handlers, mr, ruleStore, _ := setupTestHandlers(t)
	defer mr.Close()

	ctx := context.Background()

	// Create a rule without any logged events
	rule := domain.TriggerRule{
		ID:          "rule-no-events",
		WorkspaceID: "ws-test",
		Name:        "No Events Rule",
		Enabled:     true,
		ChannelType: domain.ChannelWebhook,
		Condition: domain.Condition{
			Field:    "body.type",
			Operator: "equals",
			Value:    "push",
		},
		Action: domain.Action{
			TargetType:    domain.TargetAgent,
			TargetID:      "agent-123",
			InputTemplate: "Process push",
			ExecutionMode: "one_shot",
		},
		CreatedAt: time.Now(),
	}
	require.NoError(t, ruleStore.Upsert(ctx, rule))

	app := fiber.New()
	app.Get("/admin/v1/rules/:ruleId", handlers.GetRule)

	req := httptest.NewRequest("GET", "/admin/v1/rules/rule-no-events", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyStr := string(body)
	assert.Contains(t, bodyStr, `"id":"rule-no-events"`)
	assert.Contains(t, bodyStr, `"recent_events":[]`)
}
