package palreq

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"
)

// Diff request errors.
var (
	ErrAgentIDRequired = errors.New("agent_id is required")
)

// DiffPALRequest represents the request body for diffing PAL content against an existing agent.
type DiffPALRequest struct {
	Content string `json:"content" validate:"required"`
	AgentID string `json:"agent_id" validate:"required"`
}

// NewDiffPALRequest parses the request body and returns a DiffPALRequest.
func NewDiffPALRequest(ctx *fiber.Ctx) (*DiffPALRequest, error) {
	req := new(DiffPALRequest)
	if err := ctx.BodyParser(req); err != nil {
		return nil, ErrInvalidJSON
	}
	return req, nil
}

// ParseJSON parses a JSON byte slice into a DiffPALRequest.
func (r *DiffPALRequest) ParseJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// Validate validates the diff PAL request.
func (r *DiffPALRequest) Validate() error {
	if r.Content == "" {
		return ErrContentRequired
	}

	if r.AgentID == "" {
		return ErrAgentIDRequired
	}

	return nil
}
