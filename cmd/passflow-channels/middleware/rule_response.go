package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jaak-ai/passflow-channels/domain"
	"github.com/jaak-ai/passflow-channels/input"
	"github.com/jaak-ai/passflow-channels/output"
)

// SanitizeRuleForResponse redacts secrets before returning a rule through the
// admin API.
func SanitizeRuleForResponse(rule domain.TriggerRule, events []input.WebhookEvent) fiber.Map {
	auth, secretMasked := sanitizeAuth(rule.Auth)
	action := rule.Action
	action.OutputChannel = output.RedactedOutputChannel(rule.Action.OutputChannel)

	response := fiber.Map{
		"id":           rule.ID,
		"workspace_id": rule.WorkspaceID,
		"name":         rule.Name,
		"enabled":      rule.Enabled,
		"channel_type": rule.ChannelType,
		"auth":         auth,
		"condition":    rule.Condition,
		"action":       action,
		"created_at":   rule.CreatedAt,
	}

	if secretMasked != "" {
		response["secret_masked"] = secretMasked
	}
	if events != nil {
		response["recent_events"] = events
	}

	return response
}

// SanitizeRulesForResponse redacts secrets across a rule list response.
func SanitizeRulesForResponse(rules []domain.TriggerRule) []fiber.Map {
	if len(rules) == 0 {
		return []fiber.Map{}
	}

	response := make([]fiber.Map, 0, len(rules))
	for _, rule := range rules {
		response = append(response, SanitizeRuleForResponse(rule, nil))
	}
	return response
}

func sanitizeAuth(auth *domain.AuthStrategy) (*domain.AuthStrategy, string) {
	if auth == nil {
		return nil, ""
	}

	masked := maskSecret(auth.Secret)
	sanitized := *auth
	sanitized.Secret = ""
	return &sanitized, masked
}

func maskSecret(secret string) string {
	if secret == "" {
		return ""
	}
	if len(secret) <= 4 {
		return "****"
	}
	return secret[:4] + "****"
}
