package input

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jaak-ai/passflow-channels/domain"
)

// AuthVerifier handles authentication verification for incoming webhooks.
type AuthVerifier struct{}

// NewAuthVerifier creates a new AuthVerifier.
func NewAuthVerifier() *AuthVerifier {
	return &AuthVerifier{}
}

// Verify checks the request against the configured auth strategy.
// Returns nil if authentication succeeds, an error otherwise.
func (v *AuthVerifier) Verify(auth *domain.AuthStrategy, headers map[string]string, body []byte) error {
	if auth == nil {
		return fmt.Errorf("no auth strategy configured")
	}

	switch auth.Type {
	case domain.AuthNone:
		return nil
	case domain.AuthCustomHeader:
		return v.verifyCustomHeader(auth, headers)
	case domain.AuthGitHubHMAC:
		return v.verifyGitHubHMAC(auth.Secret, headers, body)
	case domain.AuthStripeSignature:
		return v.verifyStripeSignature(auth.Secret, headers, body)
	case domain.AuthShopifyHMAC:
		return v.verifyShopifyHMAC(auth.Secret, headers, body)
	default:
		return fmt.Errorf("unknown auth type: %s", auth.Type)
	}
}

// verifyCustomHeader checks the X-Webhook-Secret header (or custom header) using constant-time comparison.
func (v *AuthVerifier) verifyCustomHeader(auth *domain.AuthStrategy, headers map[string]string) error {
	headerName := auth.Header
	if headerName == "" {
		headerName = "X-Webhook-Secret"
	}

	provided := headers[headerName]
	if provided == "" {
		// Try case-insensitive lookup
		for k, val := range headers {
			if strings.EqualFold(k, headerName) {
				provided = val
				break
			}
		}
	}

	if subtle.ConstantTimeCompare([]byte(provided), []byte(auth.Secret)) != 1 {
		return fmt.Errorf("custom header mismatch")
	}
	return nil
}

// verifyGitHubHMAC verifies GitHub webhook signatures.
// GitHub sends: X-Hub-Signature-256: sha256=<hex-digest>
// The signature is HMAC-SHA256 of the raw body using the webhook secret.
func (v *AuthVerifier) verifyGitHubHMAC(secret string, headers map[string]string, body []byte) error {
	signature := getHeaderCaseInsensitive(headers, "X-Hub-Signature-256")
	if signature == "" {
		return fmt.Errorf("missing X-Hub-Signature-256 header")
	}

	// GitHub signature format: sha256=<hex>
	if !strings.HasPrefix(signature, "sha256=") {
		return fmt.Errorf("invalid GitHub signature format")
	}

	expectedSig, err := hex.DecodeString(signature[7:])
	if err != nil {
		return fmt.Errorf("invalid hex in signature: %w", err)
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	actualSig := mac.Sum(nil)

	if !hmac.Equal(expectedSig, actualSig) {
		return fmt.Errorf("GitHub HMAC verification failed")
	}
	return nil
}

// verifyStripeSignature verifies Stripe webhook signatures.
// Stripe sends: Stripe-Signature: t=<timestamp>,v1=<signature>[,v1=<signature>...]
// The signed payload is: <timestamp>.<raw-body>
// Stripe allows up to 5 minutes clock skew.
func (v *AuthVerifier) verifyStripeSignature(secret string, headers map[string]string, body []byte) error {
	signature := getHeaderCaseInsensitive(headers, "Stripe-Signature")
	if signature == "" {
		return fmt.Errorf("missing Stripe-Signature header")
	}

	// Parse signature header
	parts := strings.Split(signature, ",")
	var timestamp string
	var signatures []string

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, value := kv[0], kv[1]
		switch key {
		case "t":
			timestamp = value
		case "v1":
			signatures = append(signatures, value)
		}
	}

	if timestamp == "" {
		return fmt.Errorf("missing timestamp in Stripe signature")
	}
	if len(signatures) == 0 {
		return fmt.Errorf("missing v1 signature in Stripe signature")
	}

	// Verify timestamp is within tolerance (5 minutes)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}

	const tolerance = 5 * time.Minute
	eventTime := time.Unix(ts, 0)
	now := time.Now()
	if now.Sub(eventTime) > tolerance || eventTime.Sub(now) > tolerance {
		return fmt.Errorf("timestamp outside tolerance window")
	}

	// Compute expected signature: HMAC-SHA256(secret, timestamp.body)
	signedPayload := fmt.Sprintf("%s.%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Check if any of the provided signatures match
	for _, sig := range signatures {
		if subtle.ConstantTimeCompare([]byte(sig), []byte(expectedSig)) == 1 {
			return nil
		}
	}

	return fmt.Errorf("Stripe signature verification failed")
}

// verifyShopifyHMAC verifies Shopify webhook signatures.
// Shopify sends: X-Shopify-Hmac-Sha256: <base64-digest>
// The signature is HMAC-SHA256 of the raw body, base64-encoded.
func (v *AuthVerifier) verifyShopifyHMAC(secret string, headers map[string]string, body []byte) error {
	signature := getHeaderCaseInsensitive(headers, "X-Shopify-Hmac-Sha256")
	if signature == "" {
		return fmt.Errorf("missing X-Shopify-Hmac-Sha256 header")
	}

	expectedSig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("invalid base64 in signature: %w", err)
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	actualSig := mac.Sum(nil)

	if !hmac.Equal(expectedSig, actualSig) {
		return fmt.Errorf("Shopify HMAC verification failed")
	}
	return nil
}

// getHeaderCaseInsensitive returns the value of a header, performing case-insensitive lookup.
func getHeaderCaseInsensitive(headers map[string]string, name string) string {
	// Try exact match first
	if val, ok := headers[name]; ok {
		return val
	}
	// Fall back to case-insensitive search
	for k, v := range headers {
		if strings.EqualFold(k, name) {
			return v
		}
	}
	return ""
}

// GetEffectiveAuth returns the effective auth strategy for a rule.
// If the rule has a new Auth configured, it returns that.
// Otherwise, it creates a legacy AuthCustomHeader strategy from WebhookSecret.
func GetEffectiveAuth(rule domain.TriggerRule) *domain.AuthStrategy {
	if rule.Auth != nil {
		return rule.Auth
	}
	if rule.WebhookSecret != "" {
		return &domain.AuthStrategy{
			Type:   domain.AuthCustomHeader,
			Secret: rule.WebhookSecret,
			Header: "X-Webhook-Secret",
		}
	}
	return nil
}
