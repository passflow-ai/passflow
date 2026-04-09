package palreq

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"
)

// Apply request errors.
var (
	ErrWorkspaceIDRequired = errors.New("workspace_id is required")
)

// ApplyPALRequest represents the request body for applying PAL content.
type ApplyPALRequest struct {
	Content     string `json:"content" validate:"required"`
	WorkspaceID string `json:"workspace_id" validate:"required"`
	DryRun      bool   `json:"dry_run,omitempty"`
}

// NewApplyPALRequest parses the request body and returns an ApplyPALRequest.
func NewApplyPALRequest(ctx *fiber.Ctx) (*ApplyPALRequest, error) {
	req := new(ApplyPALRequest)
	if err := ctx.BodyParser(req); err != nil {
		return nil, ErrInvalidJSON
	}
	return req, nil
}

// ParseJSON parses a JSON byte slice into an ApplyPALRequest.
func (r *ApplyPALRequest) ParseJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// Validate validates the apply PAL request.
func (r *ApplyPALRequest) Validate() error {
	if r.Content == "" {
		return ErrContentRequired
	}

	if r.WorkspaceID == "" {
		return ErrWorkspaceIDRequired
	}

	return nil
}

// IsDryRun returns whether this is a dry run.
func (r *ApplyPALRequest) IsDryRun() bool {
	return r.DryRun
}
