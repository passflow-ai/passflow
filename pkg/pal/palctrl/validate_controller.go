package palctrl

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/passflow-ai/passflow/pkg/pal/palparser"
	"github.com/passflow-ai/passflow/pkg/pal/palreq"
	"github.com/passflow-ai/passflow/pkg/pal/palres"
)

// ValidateController handles PAL content validation.
type ValidateController struct{}

// NewValidateController creates a new ValidateController.
func NewValidateController() *ValidateController {
	return &ValidateController{}
}

// Handle validates PAL content and returns validation results.
// @Summary Validate PAL content
// @Description Validate PAL YAML/JSON content for syntax and semantic errors
// @Tags PAL
// @Accept json
// @Produce json
// @Param request body palreq.ValidatePALRequest true "Validation request with PAL content"
// @Success 200 {object} palres.ValidatePALResponse
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/pal/validate [post]
func (c *ValidateController) Handle(ctx *fiber.Ctx) error {
	req, err := palreq.NewValidatePALRequest(ctx)
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

	// Parse YAML content
	spec, err := palparser.ParseYAML([]byte(req.Content))
	if err != nil {
		validationResp := palres.NewValidatePALResponse(false)
		validationResp.AddError("root", err.Error(), "parse_error")
		return ctx.Status(http.StatusOK).JSON(validationResp)
	}

	// Validate the spec
	validationResp := palres.NewValidatePALResponse(true)
	if err := spec.Validate(); err != nil {
		validationResp.AddError("spec", err.Error(), "validation_error")
		return ctx.Status(http.StatusOK).JSON(validationResp)
	}

	return ctx.Status(http.StatusOK).JSON(validationResp)
}

// Handler returns the Fiber handler function.
func (c *ValidateController) Handler() fiber.Handler {
	return c.Handle
}
