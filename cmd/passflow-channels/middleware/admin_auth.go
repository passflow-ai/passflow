// Package middleware provides Fiber middleware functions for the channels service.
package middleware

import (
	"crypto/subtle"
	"log"

	"github.com/gofiber/fiber/v2"
)

// AdminAuth returns a Fiber middleware that guards the admin API with a service
// token. It fails CLOSED: if serviceToken is empty (not configured), every
// request is rejected with 503 Service Unavailable so that a misconfigured
// deployment cannot expose an unauthenticated admin surface.
//
// When a token is configured, it uses constant-time comparison to prevent
// timing attacks. Wrong or missing tokens are rejected with 401 Unauthorized.
func AdminAuth(serviceToken string) fiber.Handler {
	if serviceToken == "" {
		log.Println("[middleware] WARNING: PASSFLOW_SERVICE_TOKEN is not set — admin API is DISABLED (fail closed)")
	}

	return func(c *fiber.Ctx) error {
		// Fail closed: token not configured → service unavailable.
		if serviceToken == "" {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "admin API is not available: service token not configured",
			})
		}

		provided := c.Get("X-Service-Token")
		// Use constant-time comparison to prevent timing-based token discovery.
		if subtle.ConstantTimeCompare([]byte(provided), []byte(serviceToken)) != 1 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		return c.Next()
	}
}
