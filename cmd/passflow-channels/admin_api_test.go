package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/passflow-ai/passflow/cmd/passflow-channels/config"
	"github.com/passflow-ai/passflow/cmd/passflow-channels/domain"
	"github.com/passflow-ai/passflow/cmd/passflow-channels/input"
	"github.com/passflow-ai/passflow/cmd/passflow-channels/output"
	"github.com/passflow-ai/passflow/cmd/passflow-channels/store"
	"github.com/stretchr/testify/require"
)

func setupAdminAPITest(t *testing.T) (*fiber.App, *store.RuleStore, func()) {
	t.Helper()

	mr := miniredis.NewMiniRedis()
	require.NoError(t, mr.Start())

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	ruleStore := store.New(ctx, rdb)
	eventLogger := input.NewWebhookEventLogger(rdb)

	app := fiber.New()
	registerAdminAPI(app, ruleStore, eventLogger, &config.Config{ServiceToken: "test-token"})

	cleanup := func() {
		_ = rdb.Close()
		mr.Close()
	}

	return app, ruleStore, cleanup
}

func TestRegisterAdminAPI_CreateRuleRejectsInvalidOutputChannel(t *testing.T) {
	app, _, cleanup := setupAdminAPITest(t)
	defer cleanup()

	body := bytes.NewBufferString(`{
		"name":"invalid reply rule",
		"enabled":true,
		"channel_type":"webhook",
		"action":{
			"target_id":"agent-1",
			"output_channel":{
				"type":"slack",
				"config":{"token":"xoxb-secret-token"}
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/admin/v1/workspaces/ws-1/rules", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "test-token")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRegisterAdminAPI_UpdateRulePreservesSensitiveOutputConfig(t *testing.T) {
	app, ruleStore, cleanup := setupAdminAPITest(t)
	defer cleanup()

	ctx := context.Background()
	existing := domain.TriggerRule{
		ID:          "rule-1",
		WorkspaceID: "ws-1",
		Name:        "reply rule",
		Enabled:     true,
		ChannelType: domain.ChannelSlack,
		Action: domain.Action{
			TargetID: "agent-1",
			OutputChannel: &domain.OutputChannel{
				Type: domain.ChannelSlack,
				Config: map[string]any{
					"token":   "xoxb-secret-token",
					"channel": "C123",
				},
			},
		},
		Auth: &domain.AuthStrategy{
			Type:   domain.AuthGitHubHMAC,
			Secret: "webhook-secret",
		},
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	require.NoError(t, ruleStore.Upsert(ctx, existing))

	redacted := output.RedactedOutputChannel(existing.Action.OutputChannel)
	payload := map[string]any{
		"name":         existing.Name,
		"enabled":      existing.Enabled,
		"channel_type": string(existing.ChannelType),
		"condition":    existing.Condition,
		"created_at":   existing.CreatedAt,
		"auth": map[string]any{
			"type": existing.Auth.Type,
		},
		"action": map[string]any{
			"target_id": existing.Action.TargetID,
			"output_channel": map[string]any{
				"type":   string(redacted.Type),
				"config": redacted.Config,
			},
		},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/admin/v1/workspaces/ws-1/rules/rule-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "test-token")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}

	stored := ruleStore.GetByID("rule-1")
	require.NotNil(t, stored)
	if got, _ := stored.Action.OutputChannel.Config["token"].(string); got != "xoxb-secret-token" {
		t.Fatalf("expected raw token to be preserved, got %q", got)
	}
	if stored.Auth == nil || stored.Auth.Secret != "webhook-secret" {
		t.Fatalf("expected auth secret to be preserved, got %+v", stored.Auth)
	}

	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	if bytes.Contains(responseBody, []byte("xoxb-secret-token")) {
		t.Fatal("expected response to redact output channel secret")
	}
	if bytes.Contains(responseBody, []byte("webhook-secret")) {
		t.Fatal("expected response to redact auth secret")
	}
}
