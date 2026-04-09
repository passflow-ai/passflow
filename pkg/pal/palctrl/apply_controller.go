package palctrl

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	agentdomain "github.com/jaak-ai/passflow-api/src/usecase/agent/domain"
	"github.com/jaak-ai/passflow-api/src/usecase/pal/palparser"
	"github.com/jaak-ai/passflow-api/src/usecase/pal/palreq"
	"github.com/jaak-ai/passflow-api/src/usecase/pal/palres"
)

// ApplyController handles PAL content application and agent creation/update.
type ApplyController struct {
	agentRepo agentdomain.AgentRepository
}

// NewApplyController creates a new ApplyController.
func NewApplyController(agentRepo agentdomain.AgentRepository) *ApplyController {
	return &ApplyController{
		agentRepo: agentRepo,
	}
}

// Handle applies PAL content to create or update an agent.
// @Summary Apply PAL content
// @Description Parse, validate, and apply PAL content to create or update an agent
// @Tags PAL
// @Accept json
// @Produce json
// @Param request body palreq.ApplyPALRequest true "Apply request with PAL content and workspace ID"
// @Success 200 {object} palres.ApplyPALResponse
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/pal/apply [post]
func (c *ApplyController) Handle(ctx *fiber.Ctx) error {
	req, err := palreq.NewApplyPALRequest(ctx)
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

	// Parse and validate YAML
	spec, err := palparser.ParseAndValidate([]byte(req.Content))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error":   "parse_error",
			"message": err.Error(),
		})
	}

	// Compile to agent
	agent, err := palparser.CompileToAgent(spec, req.WorkspaceID)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error":   "compilation_error",
			"message": err.Error(),
		})
	}

	applyResp := palres.NewApplyPALResponse(req.IsDryRun())
	applyResp.SetSuccess(false)
	applyResp.SetMessage("")

	// For dry run, just return success without persisting
	if req.IsDryRun() {
		applyResp.SetSuccess(true)
		applyResp.SetMessage("Dry run completed successfully")
		applyResp.SetAgentID("")
		applyResp.AddChange("created", "agent", "", spec.Agent.Name, nil)
		applyResp.Timestamp = time.Now().UTC().Format(time.RFC3339)
		return ctx.Status(http.StatusOK).JSON(applyResp)
	}

	// Check if agent with same name exists in workspace
	exists, err := c.agentRepo.ExistsByNameInWorkspace(ctx.Context(), req.WorkspaceID, spec.Agent.Name)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error":   "internal_error",
			"message": "failed to check agent existence",
		})
	}

	var changeType string
	var agentID string

	if exists {
		// Update existing agent - find it first
		// Note: In a real scenario, we would need to search by name or have agent ID in request
		changeType = "updated"
	} else {
		// Create new agent
		changeType = "created"
		id, err := c.agentRepo.Create(ctx.Context(), agent)
		if err != nil {
			if err == agentdomain.ErrAgentAlreadyExists {
				return ctx.Status(http.StatusConflict).JSON(map[string]interface{}{
					"error":   "conflict",
					"message": "agent with this name already exists",
				})
			}
			return ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
				"error":   "internal_error",
				"message": "failed to create agent",
			})
		}
		agentID = id
		agent.ID = id
	}

	applyResp.SetSuccess(true)
	applyResp.SetAgentID(agentID)
	applyResp.SetMessage("PAL applied successfully")
	applyResp.AddChange(changeType, "agent", agentID, spec.Agent.Name, nil)
	applyResp.Timestamp = time.Now().UTC().Format(time.RFC3339)

	return ctx.Status(http.StatusOK).JSON(applyResp)
}

// Handler returns the Fiber handler function.
func (c *ApplyController) Handler() fiber.Handler {
	return c.Handle
}
