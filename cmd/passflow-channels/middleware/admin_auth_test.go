package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jaak-ai/passflow-channels/middleware"
)

// newTestApp builds a minimal Fiber app with AdminAuth applied to /admin/v1.
func newTestApp(serviceToken string) *fiber.App {
	app := fiber.New()
	admin := app.Group("/admin/v1", middleware.AdminAuth(serviceToken))
	admin.Get("/rules", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}

func doAdminRequest(app *fiber.App, token string) *http.Response {
	req := httptest.NewRequest(http.MethodGet, "/admin/v1/rules", nil)
	if token != "" {
		req.Header.Set("X-Service-Token", token)
	}
	resp, _ := app.Test(req, -1)
	return resp
}

// TestAdminAuth_EmptyToken_AllRequestsRejectedWith503 verifies the fail-closed
// behaviour: when no service token is configured, ALL requests are rejected.
func TestAdminAuth_EmptyToken_AllRequestsRejectedWith503(t *testing.T) {
	app := newTestApp("") // no token configured
	resp := doAdminRequest(app, "")
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Errorf("expected 503 when service token not configured, got %d", resp.StatusCode)
	}
}

// TestAdminAuth_EmptyToken_EvenWithHeader_Rejected ensures that sending any
// token when none is configured still results in 503 (fail closed, not 401).
func TestAdminAuth_EmptyToken_EvenWithHeaderPresent_Rejected(t *testing.T) {
	app := newTestApp("")
	resp := doAdminRequest(app, "any-token")
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Errorf("expected 503 regardless of header when token not configured, got %d", resp.StatusCode)
	}
}

// TestAdminAuth_CorrectToken_Accepted verifies that a valid token grants access.
func TestAdminAuth_CorrectToken_Accepted(t *testing.T) {
	app := newTestApp("super-secret")
	resp := doAdminRequest(app, "super-secret")
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 for correct token, got %d", resp.StatusCode)
	}
}

// TestAdminAuth_WrongToken_Rejected verifies that an incorrect token returns 401.
func TestAdminAuth_WrongToken_Rejected(t *testing.T) {
	app := newTestApp("super-secret")
	resp := doAdminRequest(app, "wrong-token")
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401 for wrong token, got %d", resp.StatusCode)
	}
}

// TestAdminAuth_NoHeader_Rejected verifies that a missing header returns 401
// (not 503) when a token IS configured.
func TestAdminAuth_NoHeader_Rejected(t *testing.T) {
	app := newTestApp("super-secret")
	resp := doAdminRequest(app, "")
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401 when token missing but service is configured, got %d", resp.StatusCode)
	}
}

// TestAdminAuth_ConstantTimeCompare_CorrectSecretAfterLongerWrong verifies that
// timing-safe comparison is used (functional test — timing is not directly
// testable in a unit test, but we verify correctness for both orderings).
func TestAdminAuth_CorrectToken_Various(t *testing.T) {
	tokens := []string{"abc", "a-much-longer-service-token-value-here", "x"}
	for _, tok := range tokens {
		app := newTestApp(tok)
		resp := doAdminRequest(app, tok)
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("token %q: expected 200, got %d", tok, resp.StatusCode)
		}
	}
}
