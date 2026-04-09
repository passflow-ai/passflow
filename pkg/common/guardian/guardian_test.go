package guardian

import (
	"testing"
	"time"

	"github.com/passflow-ai/passflow/pkg/common/setup"
)

func TestGenerateAccessToken(t *testing.T) {
	// Setup test configuration
	setup.UpdateConfigForTesting(setup.Configuration{
		JWT: setup.JWTConfiguration{
			Signature:         "test-secret-key",
			AccessExpiration:  3600,
			RefreshExpiration: 604800,
		},
	})

	session := Session{
		UserID:      "user-123",
		Email:       "test@example.com",
		FullName:    "Test User",
		WorkspaceID: "workspace-456",
		Role:        "admin",
		Permissions: []string{"read", "write"},
	}

	token, err := GenerateAccessToken(session)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}
}

func TestParseToken(t *testing.T) {
	setup.UpdateConfigForTesting(setup.Configuration{
		JWT: setup.JWTConfiguration{
			Signature:         "test-secret-key",
			AccessExpiration:  3600,
			RefreshExpiration: 604800,
		},
	})

	session := Session{
		UserID:      "user-123",
		Email:       "test@example.com",
		FullName:    "Test User",
		WorkspaceID: "workspace-456",
		Role:        "admin",
	}

	token, err := GenerateAccessToken(session)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if claims.UserID != session.UserID {
		t.Errorf("Expected UserID %s, got %s", session.UserID, claims.UserID)
	}

	if claims.Email != session.Email {
		t.Errorf("Expected Email %s, got %s", session.Email, claims.Email)
	}

	if claims.WorkspaceID != session.WorkspaceID {
		t.Errorf("Expected WorkspaceID %s, got %s", session.WorkspaceID, claims.WorkspaceID)
	}

	if claims.Role != session.Role {
		t.Errorf("Expected Role %s, got %s", session.Role, claims.Role)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	setup.UpdateConfigForTesting(setup.Configuration{
		JWT: setup.JWTConfiguration{
			Signature:         "test-secret-key",
			AccessExpiration:  3600,
			RefreshExpiration: 604800,
		},
	})

	token, err := GenerateRefreshToken("user-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}
}

func TestExtractTokenFromHeader(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		expectToken string
		expectError bool
	}{
		{
			name:        "valid header",
			header:      "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectError: false,
		},
		{
			name:        "empty header",
			header:      "",
			expectToken: "",
			expectError: true,
		},
		{
			name:        "missing bearer",
			header:      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectToken: "",
			expectError: true,
		},
		{
			name:        "wrong prefix",
			header:      "Basic eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectToken: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ExtractTokenFromHeader(tt.header)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if token != tt.expectToken {
				t.Errorf("Expected token %s, got %s", tt.expectToken, token)
			}
		})
	}
}

func TestParseInvalidToken(t *testing.T) {
	setup.UpdateConfigForTesting(setup.Configuration{
		JWT: setup.JWTConfiguration{
			Signature: "test-secret-key",
		},
	})

	_, err := ParseToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestTokenExpiration(t *testing.T) {
	setup.UpdateConfigForTesting(setup.Configuration{
		JWT: setup.JWTConfiguration{
			Signature:        "test-secret-key",
			AccessExpiration: 1, // 1 second
		},
	})

	session := Session{
		UserID: "user-123",
	}

	token, err := GenerateAccessToken(session)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	_, err = ParseToken(token)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}
