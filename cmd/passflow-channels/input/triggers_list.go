package input

import (
	"github.com/gofiber/fiber/v2"
	"github.com/passflow-ai/passflow/cmd/passflow-channels/domain"
)

// TriggerOption represents a simplified trigger for selection in dropdowns.
// It contains only the fields needed for assignment, not the full rule details.
type TriggerOption struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Enabled     bool               `json:"enabled"`
	ChannelType domain.ChannelType `json:"channelType"`
}

// TriggersListResponse is the response for the triggers list endpoint.
type TriggersListResponse struct {
	Triggers []TriggerOption `json:"triggers"`
	Total    int             `json:"total"`
}

// TriggersRuleStore is the read-only subset of store.RuleStore needed by
// TriggersListHandler to list triggers for a workspace.
type TriggersRuleStore interface {
	GetRulesForWorkspace(workspaceID string) []domain.TriggerRule
}

// TriggersListHandler handles listing available triggers for a workspace.
// Route: GET /admin/v1/workspaces/:workspaceId/triggers
//
// This endpoint returns a simplified list of triggers suitable for
// dropdown selection when assigning triggers to agents or pipelines.
type TriggersListHandler struct {
	ruleStore TriggersRuleStore
}

// NewTriggersListHandler creates a new TriggersListHandler.
func NewTriggersListHandler(ruleStore TriggersRuleStore) *TriggersListHandler {
	return &TriggersListHandler{ruleStore: ruleStore}
}

// Register registers the triggers list route on the Fiber app.
// Note: This creates its own /admin/v1 group without auth middleware.
// For production use, prefer RegisterOn with an authenticated router.
func (h *TriggersListHandler) Register(app *fiber.App) {
	admin := app.Group("/admin/v1")
	admin.Get("/workspaces/:workspaceId/triggers", h.Handle)
}

// RegisterOn registers the triggers list route on an existing router group.
// Use this to mount on an authenticated admin group.
func (h *TriggersListHandler) RegisterOn(router fiber.Router) {
	router.Get("/workspaces/:workspaceId/triggers", h.Handle)
}

// Handle processes the GET /admin/v1/workspaces/:workspaceId/triggers request.
func (h *TriggersListHandler) Handle(ctx *fiber.Ctx) error {
	workspaceID := ctx.Params("workspaceId")

	rules := h.ruleStore.GetRulesForWorkspace(workspaceID)

	triggers := make([]TriggerOption, 0, len(rules))
	for _, rule := range rules {
		triggers = append(triggers, TriggerOption{
			ID:          rule.ID,
			Name:        rule.Name,
			Enabled:     rule.Enabled,
			ChannelType: rule.ChannelType,
		})
	}

	return ctx.JSON(TriggersListResponse{
		Triggers: triggers,
		Total:    len(triggers),
	})
}
