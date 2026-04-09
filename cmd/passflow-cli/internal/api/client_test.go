package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_HandleResponse_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":    map[string]string{"id": "test-123"},
			"success": true,
		})
	}))
	defer server.Close()

	client := NewClientWithConfig(server.URL, "test-token")

	var result APIResponse[map[string]string]
	err := client.get("/test", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Data["id"] != "test-123" {
		t.Errorf("expected id=test-123, got %s", result.Data["id"])
	}
}

func TestClient_HandleResponse_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "invalid request",
		})
	}))
	defer server.Close()

	client := NewClientWithConfig(server.URL, "test-token")

	var result interface{}
	err := client.get("/test", &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "API error (400): invalid request" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestClient_Authorization(t *testing.T) {
	var capturedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": nil, "success": true})
	}))
	defer server.Close()

	client := NewClientWithConfig(server.URL, "my-jwt-token")
	client.get("/test", nil)

	expected := "Bearer my-jwt-token"
	if capturedAuth != expected {
		t.Errorf("expected Authorization=%q, got %q", expected, capturedAuth)
	}
}
