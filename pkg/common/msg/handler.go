package msg

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/passflow-ai/passflow/pkg/common/logger"
	"github.com/passflow-ai/passflow/pkg/common/security/redactor"
)

// IErrorHandler defines the interface for error handling.
type IErrorHandler interface {
	Handler(ctx *fiber.Ctx, err error) error
}

type fiberErrorHandler struct {
	log logger.Logger
}

// NewFiberErrorHandler creates a new error handler with the specified logger.
func NewFiberErrorHandler(log logger.Logger) IErrorHandler {
	return &fiberErrorHandler{
		log: log,
	}
}

// Handler processes errors and returns appropriate HTTP responses.
func (e *fiberErrorHandler) Handler(ctx *fiber.Ctx, err error) error {
	traceID := ctx.Get("X-Trace-ID")
	if traceID == "" {
		traceID = ctx.Locals("trace_id", "").(string)
	}

	if err == nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			TraceID:    traceID,
			StatusCode: fiber.StatusInternalServerError,
			ErrorCode:  "UNKNOWN_ERROR",
			Message:    "An unknown error occurred",
		})
	}

	e.logError(ctx, err, traceID)

	var errorResponse *ErrorResponse
	if errors.As(err, &errorResponse) {
		errorResponse.TraceID = traceID
		return ctx.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	var fiberError *fiber.Error
	if errors.As(err, &fiberError) {
		return ctx.Status(fiberError.Code).JSON(ErrorResponse{
			TraceID:    traceID,
			StatusCode: fiberError.Code,
			ErrorCode:  e.getErrorCode(fiberError.Code),
			Message:    fiberError.Message,
		})
	}

	return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		TraceID:    traceID,
		StatusCode: fiber.StatusInternalServerError,
		ErrorCode:  "INTERNAL_ERROR",
		Message:    redactor.RedactString(err.Error()),
	})
}

func (e *fiberErrorHandler) logError(ctx *fiber.Ctx, err error, traceID string) {
	if e.log == nil {
		return
	}

	e.log.Error(
		fmt.Sprintf("[%s %s] %s", ctx.Method(), ctx.Path(), redactor.RedactString(err.Error())),
		traceID,
	)
}

func (e *fiberErrorHandler) getErrorCode(statusCode int) string {
	switch statusCode {
	case fiber.StatusBadRequest:
		return "BAD_REQUEST"
	case fiber.StatusUnauthorized:
		return "UNAUTHORIZED"
	case fiber.StatusForbidden:
		return "FORBIDDEN"
	case fiber.StatusNotFound:
		return "NOT_FOUND"
	case fiber.StatusConflict:
		return "CONFLICT"
	case fiber.StatusUnprocessableEntity:
		return "VALIDATION_ERROR"
	case fiber.StatusTooManyRequests:
		return "RATE_LIMITED"
	default:
		return "INTERNAL_ERROR"
	}
}
