package msg

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// ErrorResponse represents a standardized error response.
type ErrorResponse struct {
	TraceID     string `json:"trace_id,omitempty"`
	StatusCode  int    `json:"status_code"`
	ErrorCode   string `json:"error_code"`
	Message     string `json:"message"`
	Description string `json:"description,omitempty"`
}

// Error implements the error interface.
func (e ErrorResponse) Error() string {
	return fmt.Sprintf("[%s] %s", e.ErrorCode, e.Message)
}

// NewError creates a new ErrorResponse.
func NewError(traceID string, statusCode int, errorCode, message string) ErrorResponse {
	return ErrorResponse{
		TraceID:    traceID,
		StatusCode: statusCode,
		ErrorCode:  errorCode,
		Message:    message,
	}
}

// NewBadRequestError creates a 400 Bad Request error response.
func NewBadRequestError(message string) ErrorResponse {
	return NewError("", fiber.StatusBadRequest, "BAD_REQUEST", message)
}

// NewUnauthorizedError creates a 401 Unauthorized error response.
func NewUnauthorizedError(message string) ErrorResponse {
	return NewError("", fiber.StatusUnauthorized, "UNAUTHORIZED", message)
}

// NewForbiddenError creates a 403 Forbidden error response.
func NewForbiddenError(message string) ErrorResponse {
	return NewError("", fiber.StatusForbidden, "FORBIDDEN", message)
}

// NewNotFoundError creates a 404 Not Found error response.
func NewNotFoundError(message string) ErrorResponse {
	return NewError("", fiber.StatusNotFound, "NOT_FOUND", message)
}

// NewConflictError creates a 409 Conflict error response.
func NewConflictError(message string) ErrorResponse {
	return NewError("", fiber.StatusConflict, "CONFLICT", message)
}

// NewValidationError creates a 422 Unprocessable Entity error response.
func NewValidationError(message string) ErrorResponse {
	return NewError("", fiber.StatusUnprocessableEntity, "VALIDATION_ERROR", message)
}

// NewInternalError creates a 500 Internal Server Error response.
func NewInternalError(message string) ErrorResponse {
	return NewError("", fiber.StatusInternalServerError, "INTERNAL_ERROR", message)
}
