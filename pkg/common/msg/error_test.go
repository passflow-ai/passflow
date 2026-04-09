package msg

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestErrorResponse_Error(t *testing.T) {
	err := ErrorResponse{
		StatusCode: fiber.StatusBadRequest,
		ErrorCode:  "TEST_ERROR",
		Message:    "Test error message",
	}

	expected := "[TEST_ERROR] Test error message"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestNewError(t *testing.T) {
	traceID := "trace-123"
	statusCode := fiber.StatusNotFound
	errorCode := "NOT_FOUND"
	message := "Resource not found"

	err := NewError(traceID, statusCode, errorCode, message)

	if err.TraceID != traceID {
		t.Errorf("Expected TraceID %s, got %s", traceID, err.TraceID)
	}

	if err.StatusCode != statusCode {
		t.Errorf("Expected StatusCode %d, got %d", statusCode, err.StatusCode)
	}

	if err.ErrorCode != errorCode {
		t.Errorf("Expected ErrorCode %s, got %s", errorCode, err.ErrorCode)
	}

	if err.Message != message {
		t.Errorf("Expected Message %s, got %s", message, err.Message)
	}
}

func TestNewBadRequestError(t *testing.T) {
	message := "Invalid input"
	err := NewBadRequestError(message)

	if err.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Expected StatusCode %d, got %d", fiber.StatusBadRequest, err.StatusCode)
	}

	if err.ErrorCode != "BAD_REQUEST" {
		t.Errorf("Expected ErrorCode BAD_REQUEST, got %s", err.ErrorCode)
	}

	if err.Message != message {
		t.Errorf("Expected Message %s, got %s", message, err.Message)
	}
}

func TestNewUnauthorizedError(t *testing.T) {
	message := "Access denied"
	err := NewUnauthorizedError(message)

	if err.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected StatusCode %d, got %d", fiber.StatusUnauthorized, err.StatusCode)
	}

	if err.ErrorCode != "UNAUTHORIZED" {
		t.Errorf("Expected ErrorCode UNAUTHORIZED, got %s", err.ErrorCode)
	}
}

func TestNewForbiddenError(t *testing.T) {
	message := "Not allowed"
	err := NewForbiddenError(message)

	if err.StatusCode != fiber.StatusForbidden {
		t.Errorf("Expected StatusCode %d, got %d", fiber.StatusForbidden, err.StatusCode)
	}

	if err.ErrorCode != "FORBIDDEN" {
		t.Errorf("Expected ErrorCode FORBIDDEN, got %s", err.ErrorCode)
	}
}

func TestNewNotFoundError(t *testing.T) {
	message := "Resource not found"
	err := NewNotFoundError(message)

	if err.StatusCode != fiber.StatusNotFound {
		t.Errorf("Expected StatusCode %d, got %d", fiber.StatusNotFound, err.StatusCode)
	}

	if err.ErrorCode != "NOT_FOUND" {
		t.Errorf("Expected ErrorCode NOT_FOUND, got %s", err.ErrorCode)
	}
}

func TestNewConflictError(t *testing.T) {
	message := "Resource already exists"
	err := NewConflictError(message)

	if err.StatusCode != fiber.StatusConflict {
		t.Errorf("Expected StatusCode %d, got %d", fiber.StatusConflict, err.StatusCode)
	}

	if err.ErrorCode != "CONFLICT" {
		t.Errorf("Expected ErrorCode CONFLICT, got %s", err.ErrorCode)
	}
}

func TestNewValidationError(t *testing.T) {
	message := "Validation failed"
	err := NewValidationError(message)

	if err.StatusCode != fiber.StatusUnprocessableEntity {
		t.Errorf("Expected StatusCode %d, got %d", fiber.StatusUnprocessableEntity, err.StatusCode)
	}

	if err.ErrorCode != "VALIDATION_ERROR" {
		t.Errorf("Expected ErrorCode VALIDATION_ERROR, got %s", err.ErrorCode)
	}
}

func TestNewInternalError(t *testing.T) {
	message := "Something went wrong"
	err := NewInternalError(message)

	if err.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("Expected StatusCode %d, got %d", fiber.StatusInternalServerError, err.StatusCode)
	}

	if err.ErrorCode != "INTERNAL_ERROR" {
		t.Errorf("Expected ErrorCode INTERNAL_ERROR, got %s", err.ErrorCode)
	}
}
