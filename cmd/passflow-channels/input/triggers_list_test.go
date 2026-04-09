package input

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jaak-ai/passflow-channels/domain"
)

// mockTriggersRuleStore implements TriggersRuleStore for testing.
type mockTriggersRuleStore struct {
	rules map[string][]domain.TriggerRule
}

func newMockTriggersRuleStore() *mockTriggersRuleStore {
	return &mockTriggersRuleStore{
		rules: make(map[string][]domain.TriggerRule),
	}
}

func (m *mockTriggersRuleStore) GetRulesForWorkspace(workspaceID string) []domain.TriggerRule {
	return m.rules[workspaceID]
}

func (m *mockTriggersRuleStore) AddRule(rule domain.TriggerRule) {
	m.rules[rule.WorkspaceID] = append(m.rules[rule.WorkspaceID], rule)
}

func TestTriggersListHandler_Handle(t *testing.T) {
	t.Run("returns triggers for workspace", func(t *testing.T) {
		ruleStore := newMockTriggersRuleStore()
		rule := domain.TriggerRule{
			ID:          "rule-1",
			Name:        "GitHub PR Webhook",
			WorkspaceID: "ws-123",
			ChannelType: domain.ChannelWebhook,
			Enabled:     true,
		}
		ruleStore.AddRule(rule)

		handler := NewTriggersListHandler(ruleStore)
		app := fiber.New()
		handler.Register(app)

		req := httptest.NewRequest(http.MethodGet, "/admin/v1/workspaces/ws-123/triggers", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var result TriggersListResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(result.Triggers) != 1 {
			t.Errorf("expected 1 trigger, got %d", len(result.Triggers))
		}
		if result.Triggers[0].ID != "rule-1" {
			t.Errorf("expected trigger ID 'rule-1', got %q", result.Triggers[0].ID)
		}
		if result.Triggers[0].Name != "GitHub PR Webhook" {
			t.Errorf("expected trigger name 'GitHub PR Webhook', got %q", result.Triggers[0].Name)
		}
		if result.Triggers[0].ChannelType != domain.ChannelWebhook {
			t.Errorf("expected channel type 'webhook', got %q", result.Triggers[0].ChannelType)
		}
		if !result.Triggers[0].Enabled {
			t.Error("expected trigger to be enabled")
		}
	})

	t.Run("returns empty list when no triggers", func(t *testing.T) {
		ruleStore := newMockTriggersRuleStore()
		handler := NewTriggersListHandler(ruleStore)
		app := fiber.New()
		handler.Register(app)

		req := httptest.NewRequest(http.MethodGet, "/admin/v1/workspaces/ws-empty/triggers", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var result TriggersListResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(result.Triggers) != 0 {
			t.Errorf("expected 0 triggers, got %d", len(result.Triggers))
		}
		if result.Total != 0 {
			t.Errorf("expected total 0, got %d", result.Total)
		}
	})

	t.Run("returns multiple triggers for workspace", func(t *testing.T) {
		ruleStore := newMockTriggersRuleStore()
		ruleStore.AddRule(domain.TriggerRule{
			ID:          "rule-1",
			Name:        "GitHub Webhook",
			WorkspaceID: "ws-multi",
			ChannelType: domain.ChannelWebhook,
			Enabled:     true,
		})
		ruleStore.AddRule(domain.TriggerRule{
			ID:          "rule-2",
			Name:        "Slack Trigger",
			WorkspaceID: "ws-multi",
			ChannelType: domain.ChannelSlack,
			Enabled:     false,
		})
		ruleStore.AddRule(domain.TriggerRule{
			ID:          "rule-3",
			Name:        "Cron Job",
			WorkspaceID: "ws-multi",
			ChannelType: domain.ChannelCron,
			Enabled:     true,
		})

		handler := NewTriggersListHandler(ruleStore)
		app := fiber.New()
		handler.Register(app)

		req := httptest.NewRequest(http.MethodGet, "/admin/v1/workspaces/ws-multi/triggers", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var result TriggersListResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(result.Triggers) != 3 {
			t.Errorf("expected 3 triggers, got %d", len(result.Triggers))
		}
		if result.Total != 3 {
			t.Errorf("expected total 3, got %d", result.Total)
		}
	})

	t.Run("filters triggers by workspace", func(t *testing.T) {
		ruleStore := newMockTriggersRuleStore()
		ruleStore.AddRule(domain.TriggerRule{
			ID:          "rule-1",
			Name:        "WS1 Rule",
			WorkspaceID: "ws-1",
			ChannelType: domain.ChannelWebhook,
			Enabled:     true,
		})
		ruleStore.AddRule(domain.TriggerRule{
			ID:          "rule-2",
			Name:        "WS2 Rule",
			WorkspaceID: "ws-2",
			ChannelType: domain.ChannelSlack,
			Enabled:     true,
		})

		handler := NewTriggersListHandler(ruleStore)
		app := fiber.New()
		handler.Register(app)

		// Request triggers for ws-1
		req := httptest.NewRequest(http.MethodGet, "/admin/v1/workspaces/ws-1/triggers", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var result TriggersListResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(result.Triggers) != 1 {
			t.Errorf("expected 1 trigger, got %d", len(result.Triggers))
		}
		if result.Triggers[0].ID != "rule-1" {
			t.Errorf("expected trigger ID 'rule-1', got %q", result.Triggers[0].ID)
		}
		if result.Triggers[0].Name != "WS1 Rule" {
			t.Errorf("expected trigger name 'WS1 Rule', got %q", result.Triggers[0].Name)
		}
	})
}

// Ensure mockTriggersRuleStore satisfies the interface at compile time.
var _ TriggersRuleStore = (*mockTriggersRuleStore)(nil)
