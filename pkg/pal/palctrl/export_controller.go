package palctrl

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	agentdomain "github.com/jaak-ai/passflow-api/src/usecase/agent/domain"
	"github.com/jaak-ai/passflow-api/src/usecase/pal/palparser"
	"github.com/jaak-ai/passflow-api/src/usecase/pal/palres"
)

// ExportController handles agent export to PAL format.
type ExportController struct {
	agentRepo agentdomain.AgentRepository
}

// NewExportController creates a new ExportController.
func NewExportController(agentRepo agentdomain.AgentRepository) *ExportController {
	return &ExportController{
		agentRepo: agentRepo,
	}
}

// Handle exports an agent to PAL format (YAML or JSON).
// @Summary Export agent to PAL format
// @Description Export an agent as PAL specification in YAML or JSON format
// @Tags PAL
// @Accept json
// @Produce json
// @Param agentId path string true "Agent ID"
// @Param workspaceId path string true "Workspace ID"
// @Param format query string false "Export format (yaml or json, default: json)" Enums(json,yaml)
// @Success 200 {object} palres.ExportPALResponse
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Agent not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/workspace/{workspaceId}/pal/export/{agentId} [get]
func (c *ExportController) Handle(ctx *fiber.Ctx) error {
	workspaceID := ctx.Params("workspaceId")
	agentID := ctx.Params("agentId")
	format := ctx.Query("format", "json")

	if workspaceID == "" || agentID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error":   "bad_request",
			"message": "workspace ID and agent ID are required",
		})
	}

	// Validate format
	if format != "json" && format != "yaml" {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error":   "bad_request",
			"message": "format must be 'json' or 'yaml'",
		})
	}

	// Get agent from repository
	agent, err := c.agentRepo.FindByID(ctx.Context(), workspaceID, agentID)
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

	// Export agent to PAL format
	var content string
	var exportErr error

	if format == "yaml" {
		yamlBytes, err := palparser.ExportToYAML(agent)
		if err != nil {
			exportErr = err
		} else {
			content = string(yamlBytes)
		}
	} else {
		jsonBytes, err := palparser.ExportToJSON(agent)
		if err != nil {
			exportErr = err
		} else {
			content = string(jsonBytes)
		}
	}

	if exportErr != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error":   "export_error",
			"message": exportErr.Error(),
		})
	}

	response := palres.NewExportPALResponse(
		agent.ID,
		agent.Name,
		format,
		content,
		time.Now().UTC().Format(time.RFC3339),
	)

	return ctx.Status(http.StatusOK).JSON(response)
}

// Handler returns the Fiber handler function.
func (c *ExportController) Handler() fiber.Handler {
	return c.Handle
}
