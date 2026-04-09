package msg

import (
	"github.com/gofiber/fiber/v2"
)

// APIResponse represents a standardized API response.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success sends a successful response with data.
func Success(c *fiber.Ctx, data interface{}) error {
	return c.JSON(APIResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMessage sends a successful response with a message and data.
func SuccessWithMessage(c *fiber.Ctx, message string, data interface{}) error {
	return c.JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SendError sends an error response with a status code and message.
func SendError(c *fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(APIResponse{
		Success: false,
		Error:   message,
	})
}

// SendBadRequest sends a 400 Bad Request error response.
func SendBadRequest(c *fiber.Ctx, message string) error {
	return SendError(c, fiber.StatusBadRequest, message)
}

// SendNotFound sends a 404 Not Found error response.
func SendNotFound(c *fiber.Ctx, message string) error {
	return SendError(c, fiber.StatusNotFound, message)
}

// SendUnauthorized sends a 401 Unauthorized error response.
func SendUnauthorized(c *fiber.Ctx, message string) error {
	return SendError(c, fiber.StatusUnauthorized, message)
}

// SendForbidden sends a 403 Forbidden error response.
func SendForbidden(c *fiber.Ctx, message string) error {
	return SendError(c, fiber.StatusForbidden, message)
}

// SendInternalError sends a 500 Internal Server Error response.
func SendInternalError(c *fiber.Ctx, message string) error {
	return SendError(c, fiber.StatusInternalServerError, message)
}

// SendConflict sends a 409 Conflict error response.
func SendConflict(c *fiber.Ctx, message string) error {
	return SendError(c, fiber.StatusConflict, message)
}

// SendValidationError sends a 422 Unprocessable Entity error response.
func SendValidationError(c *fiber.Ctx, message string) error {
	return SendError(c, fiber.StatusUnprocessableEntity, message)
}
