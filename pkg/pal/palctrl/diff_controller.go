package palctrl

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gofiber/fiber/v2"
	agentdomain "github.com/jaak-ai/passflow-api/src/usecase/agent/domain"
	"github.com/jaak-ai/passflow-api/src/usecase/pal/palparser"
	"github.com/jaak-ai/passflow-api/src/usecase/pal/palreq"
	"github.com/jaak-ai/passflow-api/src/usecase/pal/palres"
)

// DiffController handles diffing of PAL content against existing agents.
type DiffController struct {
	agentRepo agentdomain.AgentRepository
}

// NewDiffController creates a new DiffController.
func NewDiffController(agentRepo agentdomain.AgentRepository) *DiffController {
	return &DiffController{
		agentRepo: agentRepo,
	}
}

// Handle computes the diff between an existing agent and proposed PAL content.
// @Summary Compute PAL diff
// @Description Compare proposed PAL content against an existing agent and return differences
// @Tags PAL
// @Accept json
// @Produce json
// @Param request body palreq.DiffPALRequest true "Diff request with new PAL content and agent ID"
// @Success 200 {object} palres.DiffPALResponse
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Agent not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/pal/diff [post]
func (c *DiffController) Handle(ctx *fiber.Ctx) error {
	req, err := palreq.NewDiffPALRequest(ctx)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error":   "invalid_request",
			"message": err.Error(),
		})
	}

	if err := req.Validate(); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error":   "validation_error",
			"message": err.Error(),
		})
	}

	// Get current agent - need to infer workspace from context or use default
	// For now, we'll fetch without workspace check (repository should support this)
	// In production, you'd need to add workspace context to the request
	agent, err := c.agentRepo.FindByID(ctx.Context(), "", req.AgentID)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error":   "internal_error",
			"message": "failed to fetch agent",
		})
	}

	if agent == nil {
		return ctx.Status(http.StatusNotFound).JSON(map[string]interface{}{
			"error":   "not_found",
			"message": "agent not found",
		})
	}

	// Parse new content
	newSpec, err := palparser.ParseAndValidate([]byte(req.Content))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error":   "parse_error",
			"message": err.Error(),
		})
	}

	// Export current agent to spec for comparison
	currentSpec, err := palparser.ExportToSpec(agent)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error":   "export_error",
			"message": err.Error(),
		})
	}

	// Create diff response
	diffResp := palres.NewDiffPALResponse(req.AgentID)

	// Compare specs
	comparePALSpecsImpl(currentSpec, newSpec, "agent", diffResp)

	return ctx.Status(http.StatusOK).JSON(diffResp)
}

// Handler returns the Fiber handler function.
func (c *DiffController) Handler() fiber.Handler {
	return c.Handle
}

// comparePALSpecsImpl performs the actual comparison of PAL specs
func comparePALSpecsImpl(current, proposed interface{}, path string, diffResp *palres.DiffPALResponse) {
	if current == nil && proposed != nil {
		diffResp.AddChange(path, "value", "added", nil, proposed)
		return
	}

	if current != nil && proposed == nil {
		diffResp.AddChange(path, "value", "removed", current, nil)
		return
	}

	if current == nil && proposed == nil {
		return
	}

	// Compare values
	if !reflect.DeepEqual(current, proposed) {
		// For scalar types
		if isScalar(current) && isScalar(proposed) {
			if fmt.Sprintf("%v", current) != fmt.Sprintf("%v", proposed) {
				diffResp.AddChange(path, "value", "modified", current, proposed)
			}
		}
	}
}

// isScalar checks if a value is a scalar type (string, number, bool)
func isScalar(v interface{}) bool {
	switch v.(type) {
	case string, int, int64, float64, bool:
		return true
	default:
		return false
	}
}
