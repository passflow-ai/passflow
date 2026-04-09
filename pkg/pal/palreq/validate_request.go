package palreq

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"
)

// Validation errors.
var (
	ErrInvalidJSON    = errors.New("invalid JSON body")
	ErrContentRequired = errors.New("content is required")
	ErrInvalidMode    = errors.New("mode must be 'strict' or 'warn'")
)

// ValidatePALRequest represents the request body for validating PAL content.
type ValidatePALRequest struct {
	Content string `json:"content" validate:"required"`
	Mode    string `json:"mode,omitempty"` // strict | warn
}

// NewValidatePALRequest parses the request body and returns a ValidatePALRequest.
func NewValidatePALRequest(ctx *fiber.Ctx) (*ValidatePALRequest, error) {
	req := new(ValidatePALRequest)
	if err := ctx.BodyParser(req); err != nil {
		return nil, ErrInvalidJSON
	}
	return req, nil
}

// ParseJSON parses a JSON byte slice into a ValidatePALRequest.
func (r *ValidatePALRequest) ParseJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// Validate validates the validate PAL request.
func (r *ValidatePALRequest) Validate() error {
	if r.Content == "" {
		return ErrContentRequired
	}

	if r.Mode != "" && r.Mode != "strict" && r.Mode != "warn" {
		return ErrInvalidMode
	}

	return nil
}

// GetMode returns the validation mode, defaulting to "strict" if not specified.
func (r *ValidatePALRequest) GetMode() string {
	if r.Mode == "" {
		return "strict"
	}
	return r.Mode
}
