package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jaak-ai/passflow-channels/input"
	"github.com/jaak-ai/passflow-channels/store"
)

// AdminHandlers provides admin API handlers for the channels service.
type AdminHandlers struct {
	ruleStore   *store.RuleStore
	eventLogger *input.WebhookEventLogger
}

// NewAdminHandlers creates a new AdminHandlers instance.
func NewAdminHandlers(ruleStore *store.RuleStore, eventLogger *input.WebhookEventLogger) *AdminHandlers {
	return &AdminHandlers{
		ruleStore:   ruleStore,
		eventLogger: eventLogger,
	}
}

// GetRule retrieves a rule by ID with recent webhook events.
// GET /admin/v1/rules/:ruleId
func (h *AdminHandlers) GetRule(c *fiber.Ctx) error {
	ruleID := c.Params("ruleId")
	if ruleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ruleId parameter is required",
		})
	}

	rule := h.ruleStore.GetByID(ruleID)
	if rule == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "rule not found",
		})
	}

	// Fetch recent webhook events (limit 20)
	events, err := h.eventLogger.GetRecent(c.Context(), ruleID, 20)
	if err != nil {
		// Log error but don't fail the request - events are optional
		// Return rule with empty events array
		events = []input.WebhookEvent{}
	}

	return c.JSON(SanitizeRuleForResponse(*rule, events))
}
