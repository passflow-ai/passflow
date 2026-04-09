package guardian

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/passflow-ai/passflow/pkg/common/setup"
)

// Session represents the authenticated user session.
type Session struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	FullName    string   `json:"full_name"`
	WorkspaceID string   `json:"workspace_id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions,omitempty"`
}

// Claims represents JWT token claims.
type Claims struct {
	jwt.RegisteredClaims
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	FullName    string   `json:"full_name"`
	WorkspaceID string   `json:"workspace_id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions,omitempty"`
}

// GenerateAccessToken creates a new JWT access token.
func GenerateAccessToken(session Session) (string, error) {
	cfg := setup.GetConfig()

	claims := Claims{
		UserID:      session.UserID,
		Email:       session.Email,
		FullName:    session.FullName,
		WorkspaceID: session.WorkspaceID,
		Role:        session.Role,
		Permissions: session.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(
				time.Duration(cfg.JWT.AccessExpiration) * time.Second,
			)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "passflow-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Signature))
}

// GenerateRefreshToken creates a new JWT refresh token.
func GenerateRefreshToken(userID string) (string, error) {
	cfg := setup.GetConfig()

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(
			time.Duration(cfg.JWT.RefreshExpiration) * time.Second,
		)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "passflow-api",
		Subject:   userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Signature))
}

// ParseToken parses and validates a JWT token string.
func ParseToken(tokenString string) (*Claims, error) {
	cfg := setup.GetConfig()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWT.Signature), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ExtractTokenFromHeader extracts the JWT token from the Authorization header.
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is empty")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// AuthMiddleware creates a Fiber middleware for JWT authentication.
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		tokenString, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		claims, err := ParseToken(tokenString)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid or expired token")
		}

		session := Session{
			UserID:      claims.UserID,
			Email:       claims.Email,
			FullName:    claims.FullName,
			WorkspaceID: claims.WorkspaceID,
			Role:        claims.Role,
			Permissions: claims.Permissions,
		}

		c.Locals("session", &session)
		c.Locals("user_id", claims.UserID)
		c.Locals("workspace_id", claims.WorkspaceID)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

// GetSession retrieves the session from the Fiber context.
func GetSession(c *fiber.Ctx) *Session {
	session, ok := c.Locals("session").(*Session)
	if !ok {
		return nil
	}
	return session
}

// RoleMiddleware creates a middleware that checks for required roles.
func RoleMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		session := GetSession(c)
		if session == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Session not found")
		}

		for _, allowedRole := range allowedRoles {
			if session.Role == allowedRole {
				return c.Next()
			}
		}

		return fiber.NewError(fiber.StatusForbidden, "Insufficient permissions")
	}
}

// PermissionMiddleware creates a middleware that checks for required permissions.
func PermissionMiddleware(requiredPermissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		session := GetSession(c)
		if session == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Session not found")
		}

		permMap := make(map[string]bool)
		for _, p := range session.Permissions {
			permMap[p] = true
		}

		for _, required := range requiredPermissions {
			if !permMap[required] {
				return fiber.NewError(fiber.StatusForbidden,
					fmt.Sprintf("Missing required permission: %s", required))
			}
		}

		return c.Next()
	}
}
