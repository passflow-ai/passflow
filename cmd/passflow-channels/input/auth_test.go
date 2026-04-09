package input

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/jaak-ai/passflow-channels/domain"
)

func TestAuthVerifier_VerifyCustomHeader(t *testing.T) {
	v := NewAuthVerifier()

	tests := []struct {
		name    string
		auth    *domain.AuthStrategy
		headers map[string]string
		wantErr bool
	}{
		{
			name: "valid secret with default header",
			auth: &domain.AuthStrategy{
				Type:   domain.AuthCustomHeader,
				Secret: "my-secret",
			},
			headers: map[string]string{"X-Webhook-Secret": "my-secret"},
			wantErr: false,
		},
		{
			name: "valid secret with custom header",
			auth: &domain.AuthStrategy{
				Type:   domain.AuthCustomHeader,
				Secret: "my-secret",
				Header: "X-Custom-Header",
			},
			headers: map[string]string{"X-Custom-Header": "my-secret"},
			wantErr: false,
		},
		{
			name: "invalid secret",
			auth: &domain.AuthStrategy{
				Type:   domain.AuthCustomHeader,
				Secret: "my-secret",
			},
			headers: map[string]string{"X-Webhook-Secret": "wrong-secret"},
			wantErr: true,
		},
		{
			name: "missing header",
			auth: &domain.AuthStrategy{
				Type:   domain.AuthCustomHeader,
				Secret: "my-secret",
			},
			headers: map[string]string{},
			wantErr: true,
		},
		{
			name: "case insensitive header lookup",
			auth: &domain.AuthStrategy{
				Type:   domain.AuthCustomHeader,
				Secret: "my-secret",
			},
			headers: map[string]string{"x-webhook-secret": "my-secret"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Verify(tt.auth, tt.headers, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthVerifier_VerifyGitHubHMAC(t *testing.T) {
	v := NewAuthVerifier()
	secret := "github-secret"
	body := []byte(`{"action":"push","ref":"refs/heads/main"}`)

	// Compute valid signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	validSig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name    string
		headers map[string]string
		body    []byte
		wantErr bool
	}{
		{
			name:    "valid signature",
			headers: map[string]string{"X-Hub-Signature-256": validSig},
			body:    body,
			wantErr: false,
		},
		{
			name:    "invalid signature",
			headers: map[string]string{"X-Hub-Signature-256": "sha256=invalid"},
			body:    body,
			wantErr: true,
		},
		{
			name:    "missing header",
			headers: map[string]string{},
			body:    body,
			wantErr: true,
		},
		{
			name:    "wrong body",
			headers: map[string]string{"X-Hub-Signature-256": validSig},
			body:    []byte(`{"different":"body"}`),
			wantErr: true,
		},
		{
			name:    "missing sha256 prefix",
			headers: map[string]string{"X-Hub-Signature-256": hex.EncodeToString(mac.Sum(nil))},
			body:    body,
			wantErr: true,
		},
	}

	auth := &domain.AuthStrategy{
		Type:   domain.AuthGitHubHMAC,
		Secret: secret,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Verify(auth, tt.headers, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthVerifier_VerifyStripeSignature(t *testing.T) {
	v := NewAuthVerifier()
	secret := "whsec_stripe-secret"
	body := []byte(`{"id":"evt_123","type":"checkout.session.completed"}`)
	timestamp := time.Now().Unix()

	// Compute valid signature
	signedPayload := fmt.Sprintf("%d.%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	validSig := hex.EncodeToString(mac.Sum(nil))

	validHeader := fmt.Sprintf("t=%d,v1=%s", timestamp, validSig)

	tests := []struct {
		name    string
		headers map[string]string
		body    []byte
		wantErr bool
	}{
		{
			name:    "valid signature",
			headers: map[string]string{"Stripe-Signature": validHeader},
			body:    body,
			wantErr: false,
		},
		{
			name:    "invalid signature",
			headers: map[string]string{"Stripe-Signature": fmt.Sprintf("t=%d,v1=invalid", timestamp)},
			body:    body,
			wantErr: true,
		},
		{
			name:    "missing header",
			headers: map[string]string{},
			body:    body,
			wantErr: true,
		},
		{
			name: "expired timestamp",
			headers: map[string]string{
				"Stripe-Signature": fmt.Sprintf("t=%d,v1=%s", time.Now().Add(-10*time.Minute).Unix(), validSig),
			},
			body:    body,
			wantErr: true,
		},
		{
			name:    "multiple signatures (one valid)",
			headers: map[string]string{"Stripe-Signature": fmt.Sprintf("t=%d,v1=invalid,v1=%s", timestamp, validSig)},
			body:    body,
			wantErr: false,
		},
	}

	auth := &domain.AuthStrategy{
		Type:   domain.AuthStripeSignature,
		Secret: secret,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Verify(auth, tt.headers, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthVerifier_VerifyShopifyHMAC(t *testing.T) {
	v := NewAuthVerifier()
	secret := "shopify-secret"
	body := []byte(`{"id":123,"email":"customer@example.com"}`)

	// Compute valid signature (base64 encoded)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	validSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name    string
		headers map[string]string
		body    []byte
		wantErr bool
	}{
		{
			name:    "valid signature",
			headers: map[string]string{"X-Shopify-Hmac-Sha256": validSig},
			body:    body,
			wantErr: false,
		},
		{
			name:    "invalid signature",
			headers: map[string]string{"X-Shopify-Hmac-Sha256": base64.StdEncoding.EncodeToString([]byte("invalid"))},
			body:    body,
			wantErr: true,
		},
		{
			name:    "missing header",
			headers: map[string]string{},
			body:    body,
			wantErr: true,
		},
		{
			name:    "invalid base64",
			headers: map[string]string{"X-Shopify-Hmac-Sha256": "not-valid-base64!!!"},
			body:    body,
			wantErr: true,
		},
	}

	auth := &domain.AuthStrategy{
		Type:   domain.AuthShopifyHMAC,
		Secret: secret,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Verify(auth, tt.headers, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthVerifier_VerifyAuthNone(t *testing.T) {
	v := NewAuthVerifier()

	auth := &domain.AuthStrategy{
		Type: domain.AuthNone,
	}

	err := v.Verify(auth, map[string]string{}, nil)
	if err != nil {
		t.Errorf("Verify() with AuthNone should not return error, got %v", err)
	}
}

func TestAuthVerifier_NilAuth(t *testing.T) {
	v := NewAuthVerifier()

	err := v.Verify(nil, map[string]string{}, nil)
	if err == nil {
		t.Error("Verify() with nil auth should return error")
	}
}

func TestGetEffectiveAuth(t *testing.T) {
	tests := []struct {
		name string
		rule domain.TriggerRule
		want *domain.AuthStrategy
	}{
		{
			name: "new auth takes precedence",
			rule: domain.TriggerRule{
				Auth: &domain.AuthStrategy{
					Type:   domain.AuthGitHubHMAC,
					Secret: "github-secret",
				},
				WebhookSecret: "legacy-secret",
			},
			want: &domain.AuthStrategy{
				Type:   domain.AuthGitHubHMAC,
				Secret: "github-secret",
			},
		},
		{
			name: "fallback to legacy webhook secret",
			rule: domain.TriggerRule{
				WebhookSecret: "legacy-secret",
			},
			want: &domain.AuthStrategy{
				Type:   domain.AuthCustomHeader,
				Secret: "legacy-secret",
				Header: "X-Webhook-Secret",
			},
		},
		{
			name: "no auth configured",
			rule: domain.TriggerRule{},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetEffectiveAuth(tt.rule)
			if tt.want == nil {
				if got != nil {
					t.Errorf("GetEffectiveAuth() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Errorf("GetEffectiveAuth() = nil, want %v", tt.want)
				return
			}
			if got.Type != tt.want.Type || got.Secret != tt.want.Secret {
				t.Errorf("GetEffectiveAuth() = %v, want %v", got, tt.want)
			}
		})
	}
}
