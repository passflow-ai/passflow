package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// sentinel handler used to confirm the next handler was (or was not) called.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
})

// =============================================================================
// Static Token Tests (existing behavior)
// =============================================================================

func TestBearerAuth_MissingToken_ServerError(t *testing.T) {
	t.Setenv("MCP_GATEWAY_TOKEN", "")
	t.Setenv("PASSFLOW_API_URL", "")

	handler := BearerAuth(okHandler)
	req := httptest.NewRequest(http.MethodPost, "/tools/list", nil)
	req.Header.Set("Authorization", "Bearer sometoken")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when token env var is unset, got %d", rec.Code)
	}
}

func TestBearerAuth_HealthBypassesAuth(t *testing.T) {
	t.Setenv("MCP_GATEWAY_TOKEN", "")
	t.Setenv("PASSFLOW_API_URL", "")

	handler := BearerAuth(okHandler)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for /health with no token, got %d", rec.Code)
	}
}

func TestBearerAuth_NoAuthorizationHeader(t *testing.T) {
	t.Setenv("MCP_GATEWAY_TOKEN", "secret")
	t.Setenv("PASSFLOW_API_URL", "")

	handler := BearerAuth(okHandler)
	req := httptest.NewRequest(http.MethodPost, "/tools/list", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing Authorization header, got %d", rec.Code)
	}
	if rec.Header().Get("WWW-Authenticate") == "" {
		t.Error("expected WWW-Authenticate header in response")
	}
}

func TestBearerAuth_InvalidHeaderFormat(t *testing.T) {
	t.Setenv("MCP_GATEWAY_TOKEN", "secret")
	t.Setenv("PASSFLOW_API_URL", "")

	handler := BearerAuth(okHandler)
	req := httptest.NewRequest(http.MethodPost, "/tools/list", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for non-Bearer scheme, got %d", rec.Code)
	}
}

func TestBearerAuth_WrongToken(t *testing.T) {
	t.Setenv("MCP_GATEWAY_TOKEN", "correct-token")
	t.Setenv("PASSFLOW_API_URL", "")

	handler := BearerAuth(okHandler)
	req := httptest.NewRequest(http.MethodPost, "/tools/list", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong token, got %d", rec.Code)
	}
}

func TestBearerAuth_CorrectToken(t *testing.T) {
	t.Setenv("MCP_GATEWAY_TOKEN", "correct-token")
	t.Setenv("PASSFLOW_API_URL", "")

	handler := BearerAuth(okHandler)
	req := httptest.NewRequest(http.MethodPost, "/tools/list", nil)
	req.Header.Set("Authorization", "Bearer correct-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for correct token, got %d", rec.Code)
	}
}

func TestBearerAuth_BearerCaseInsensitive(t *testing.T) {
	t.Setenv("MCP_GATEWAY_TOKEN", "my-token")
	t.Setenv("PASSFLOW_API_URL", "")

	handler := BearerAuth(okHandler)
	req := httptest.NewRequest(http.MethodPost, "/tools/call", nil)
	req.Header.Set("Authorization", "BEARER my-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for case-insensitive Bearer scheme, got %d", rec.Code)
	}
}

// =============================================================================
// Passflow API Key Tests (pf_* prefix)
// =============================================================================

func TestBearerAuth_PassflowKey_MissingAPIURL(t *testing.T) {
	t.Setenv("MCP_GATEWAY_TOKEN", "")
	t.Setenv("PASSFLOW_API_URL", "")

	handler := BearerAuth(okHandler)
	req := httptest.NewRequest(http.MethodPost, "/tools/list", nil)
	req.Header.Set("Authorization", "Bearer pf_live_abc123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when PASSFLOW_API_URL is unset for pf_ key, got %d", rec.Code)
	}
}

func TestBearerAuth_PassflowKey_ValidKey(t *testing.T) {
	// Mock passflow-api server
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.URL.Path != "/api/v1/auth/me" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("X-API-Key") != "pf_live_validkey123" {
			t.Errorf("unexpected API key: %s", r.Header.Get("X-API-Key"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"user123","email":"test@example.com","name":"Test User","role":"admin","workspaceIds":["ws1"]}`))
	}))
	defer mockAPI.Close()

	t.Setenv("MCP_GATEWAY_TOKEN", "")
	t.Setenv("PASSFLOW_API_URL", mockAPI.URL)

	handler := BearerAuth(okHandler)
	req := httptest.NewRequest(http.MethodPost, "/tools/list", nil)
	req.Header.Set("Authorization", "Bearer pf_live_validkey123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for valid pf_ key, got %d", rec.Code)
	}
}

func TestBearerAuth_PassflowKey_InvalidKey(t *testing.T) {
	// Mock passflow-api server returning 401
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer mockAPI.Close()

	t.Setenv("MCP_GATEWAY_TOKEN", "")
	t.Setenv("PASSFLOW_API_URL", mockAPI.URL)

	handler := BearerAuth(okHandler)
	req := httptest.NewRequest(http.MethodPost, "/tools/list", nil)
	req.Header.Set("Authorization", "Bearer pf_live_invalidkey")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid pf_ key, got %d", rec.Code)
	}
}

func TestBearerAuth_PassflowKey_APIError(t *testing.T) {
	// Mock passflow-api server returning 500
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer mockAPI.Close()

	t.Setenv("MCP_GATEWAY_TOKEN", "")
	t.Setenv("PASSFLOW_API_URL", mockAPI.URL)

	handler := BearerAuth(okHandler)
	req := httptest.NewRequest(http.MethodPost, "/tools/list", nil)
	req.Header.Set("Authorization", "Bearer pf_live_somekey")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for API error, got %d", rec.Code)
	}
}

// =============================================================================
// X-API-Key Header Tests
// =============================================================================

func TestBearerAuth_XAPIKeyHeader_StaticToken(t *testing.T) {
	t.Setenv("MCP_GATEWAY_TOKEN", "my-static-token")
	t.Setenv("PASSFLOW_API_URL", "")

	handler := BearerAuth(okHandler)
	req := httptest.NewRequest(http.MethodPost, "/tools/list", nil)
	req.Header.Set("X-API-Key", "my-static-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for valid X-API-Key header, got %d", rec.Code)
	}
}

func TestBearerAuth_XAPIKeyHeader_PassflowKey(t *testing.T) {
	// Mock passflow-api server
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"user123","email":"test@example.com","name":"Test","role":"admin","workspaceIds":["ws1"]}`))
	}))
	defer mockAPI.Close()

	t.Setenv("MCP_GATEWAY_TOKEN", "")
	t.Setenv("PASSFLOW_API_URL", mockAPI.URL)

	handler := BearerAuth(okHandler)
	req := httptest.NewRequest(http.MethodPost, "/tools/list", nil)
	req.Header.Set("X-API-Key", "pf_live_key123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for valid pf_ key via X-API-Key header, got %d", rec.Code)
	}
}

func TestBearerAuth_XAPIKeyHeader_TakesPrecedence(t *testing.T) {
	t.Setenv("MCP_GATEWAY_TOKEN", "correct-token")
	t.Setenv("PASSFLOW_API_URL", "")

	handler := BearerAuth(okHandler)
	req := httptest.NewRequest(http.MethodPost, "/tools/list", nil)
	// Both headers set - X-API-Key should take precedence
	req.Header.Set("X-API-Key", "correct-token")
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 when X-API-Key takes precedence, got %d", rec.Code)
	}
}

// =============================================================================
// Test key types (live vs test)
// =============================================================================

func TestBearerAuth_PassflowKey_TestEnvKey(t *testing.T) {
	// Mock passflow-api server
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify test key is passed through
		if r.Header.Get("X-API-Key") != "pf_test_abc123" {
			t.Errorf("unexpected API key: %s", r.Header.Get("X-API-Key"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"user123","email":"test@example.com","name":"Test","role":"admin","workspaceIds":["ws1"]}`))
	}))
	defer mockAPI.Close()

	t.Setenv("MCP_GATEWAY_TOKEN", "")
	t.Setenv("PASSFLOW_API_URL", mockAPI.URL)

	handler := BearerAuth(okHandler)
	req := httptest.NewRequest(http.MethodPost, "/tools/list", nil)
	req.Header.Set("Authorization", "Bearer pf_test_abc123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for valid pf_test_ key, got %d", rec.Code)
	}
}
