package palrouter

import (
	"github.com/gofiber/fiber/v2"
	agentdomain "github.com/passflow-ai/passflow/pkg/agent/domain"
	"github.com/passflow-ai/passflow/pkg/pal/palctrl"
)

// PALRouter handles routing for PAL (Passflow Agent Language) endpoints.
type PALRouter struct {
	agentRepo    agentdomain.AgentRepository
	validateCtrl *palctrl.ValidateController
	applyCtrl    *palctrl.ApplyController
	exportCtrl   *palctrl.ExportController
	diffCtrl     *palctrl.DiffController
}

// NewPALRouter creates a new PALRouter with the given agent repository.
func NewPALRouter(agentRepo agentdomain.AgentRepository) *PALRouter {
	return &PALRouter{
		agentRepo:    agentRepo,
		validateCtrl: palctrl.NewValidateController(),
		applyCtrl:    palctrl.NewApplyController(agentRepo),
		exportCtrl:   palctrl.NewExportController(agentRepo),
		diffCtrl:     palctrl.NewDiffController(agentRepo),
	}
}

// Register registers all PAL routes on the given router group.
func (r *PALRouter) Register(router fiber.Router) {
	// Global PAL endpoints (not workspace-scoped)
	pal := router.Group("/pal")
	pal.Post("/validate", r.validateCtrl.Handler())
	pal.Post("/apply", r.applyCtrl.Handler())
	pal.Post("/diff", r.diffCtrl.Handler())

	// Workspace-scoped export endpoint
	workspace := router.Group("/workspaces/:workspaceId")
	workspace.Get("/pal/export/:agentId", r.exportCtrl.Handler())
}

// RouteConfig represents a route configuration.
type RouteConfig struct {
	Path    string
	Method  string
	Handler string
}

// GetRoutes returns the route configuration for documentation.
func GetRoutes() []RouteConfig {
	return []RouteConfig{
		{Path: "/pal/validate", Method: fiber.MethodPost, Handler: "ValidatePAL"},
		{Path: "/pal/apply", Method: fiber.MethodPost, Handler: "ApplyPAL"},
		{Path: "/pal/diff", Method: fiber.MethodPost, Handler: "DiffPAL"},
		{Path: "/workspaces/:workspaceId/pal/export/:agentId", Method: fiber.MethodGet, Handler: "ExportPAL"},
	}
}
