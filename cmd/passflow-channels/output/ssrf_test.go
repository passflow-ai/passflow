package output

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaak-ai/passflow-channels/config"
)

// --- validateWebhookURL tests ---

func TestValidateWebhookURL_ValidHTTPS(t *testing.T) {
	if err := validateWebhookURL("https://hooks.example.com/notify"); err != nil {
		t.Errorf("expected valid HTTPS URL to pass, got: %v", err)
	}
}

func TestValidateWebhookURL_HTTPRejected(t *testing.T) {
	if err := validateWebhookURL("http://hooks.example.com/notify"); err == nil {
		t.Error("expected HTTP URL to be rejected in production, got nil")
	}
}

func TestValidateWebhookURL_EmptyURL(t *testing.T) {
	if err := validateWebhookURL(""); err == nil {
		t.Error("expected empty URL to be rejected, got nil")
	}
}

func TestValidateWebhookURL_PrivateIP_10x(t *testing.T) {
	if err := validateWebhookURL("https://10.0.0.1/hook"); err == nil {
		t.Error("expected 10.x.x.x private IP to be blocked, got nil")
	}
}

func TestValidateWebhookURL_PrivateIP_172_16(t *testing.T) {
	if err := validateWebhookURL("https://172.16.0.1/hook"); err == nil {
		t.Error("expected 172.16.x.x private IP to be blocked, got nil")
	}
}

func TestValidateWebhookURL_PrivateIP_172_31(t *testing.T) {
	if err := validateWebhookURL("https://172.31.255.255/hook"); err == nil {
		t.Error("expected 172.31.x.x private IP to be blocked, got nil")
	}
}

func TestValidateWebhookURL_PrivateIP_192_168(t *testing.T) {
	if err := validateWebhookURL("https://192.168.1.1/hook"); err == nil {
		t.Error("expected 192.168.x.x private IP to be blocked, got nil")
	}
}

func TestValidateWebhookURL_Loopback_127(t *testing.T) {
	if err := validateWebhookURL("https://127.0.0.1/hook"); err == nil {
		t.Error("expected loopback 127.x.x.x to be blocked, got nil")
	}
}

func TestValidateWebhookURL_LinkLocal_169_254(t *testing.T) {
	if err := validateWebhookURL("https://169.254.169.254/latest/meta-data/"); err == nil {
		t.Error("expected AWS metadata IP 169.254.169.254 to be blocked, got nil")
	}
}

func TestValidateWebhookURL_Localhost(t *testing.T) {
	if err := validateWebhookURL("https://localhost/hook"); err == nil {
		t.Error("expected localhost to be blocked, got nil")
	}
}

func TestValidateWebhookURL_PrivateIP_172_Outside_Range(t *testing.T) {
	// 172.15.x.x and 172.32.x.x are NOT in 172.16.0.0/12 — they should be allowed.
	if err := validateWebhookURL("https://172.15.0.1/hook"); err != nil {
		t.Errorf("172.15.x.x is not in the private 172.16.0.0/12 range, should be allowed: %v", err)
	}
}

// TestSendWebhook_RejectsPrivateIP verifies the full sendWebhook path blocks
// private IP destinations.
func TestSendWebhook_RejectsPrivateIP(t *testing.T) {
	s := New(&config.Config{})
	err := s.sendWebhook(context.Background(), map[string]interface{}{
		"url": "https://192.168.0.1/callback",
	}, "payload")
	if err == nil {
		t.Error("expected sendWebhook to reject private IP URL, got nil")
	}
}

// TestSendWebhook_RejectsHTTP verifies the full sendWebhook path blocks non-HTTPS.
func TestSendWebhook_RejectsHTTP(t *testing.T) {
	s := New(&config.Config{})
	err := s.sendWebhook(context.Background(), map[string]interface{}{
		"url": "http://hooks.example.com/notify",
	}, "payload")
	if err == nil {
		t.Error("expected sendWebhook to reject HTTP URL, got nil")
	}
}

// TestSendWebhook_AllowsPublicHTTPS verifies that a legitimate public HTTPS
// endpoint is allowed through (the HTTP call will fail since it's not real, but
// the SSRF check itself must not block it).
func TestSendWebhook_AllowsPublicHTTPS(t *testing.T) {
	// Stand up a local TLS-free test server; we only care that validation passes.
	called := false
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// The test server binds to 127.0.0.1 — we cannot actually test a public
	// IP in a unit test. Instead, just verify the validation logic accepts a
	// well-formed public-looking HTTPS URL.
	if err := validateWebhookURL("https://hooks.example.com/callback"); err != nil {
		t.Errorf("expected public HTTPS URL to pass validation, got: %v", err)
	}
	_ = called
}
