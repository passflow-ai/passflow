package output

import (
	"testing"

	"github.com/jaak-ai/passflow-channels/domain"
)

func TestValidateReplyOutputChannel_RejectsMissingStaticSlackChannel(t *testing.T) {
	err := ValidateReplyOutputChannel(domain.ChannelWebhook, &domain.OutputChannel{
		Type: domain.ChannelSlack,
		Config: map[string]any{
			"token": "xoxb-secret-token",
		},
	})
	if err == nil {
		t.Fatal("expected validation error for missing slack channel")
	}
}

func TestValidateReplyOutputChannel_AllowsDerivedWhatsAppDestination(t *testing.T) {
	err := ValidateReplyOutputChannel(domain.ChannelWhatsApp, &domain.OutputChannel{
		Type: domain.ChannelWhatsApp,
		Config: map[string]any{
			"token": "meta-secret-token",
		},
	})
	if err != nil {
		t.Fatalf("expected whatsapp reply config to be valid, got %v", err)
	}
}

func TestMergeSensitiveOutputChannel_PreservesRedactedSecrets(t *testing.T) {
	existing := &domain.OutputChannel{
		Type: domain.ChannelSMS,
		Config: map[string]any{
			"account_sid": "AC1234567890",
			"auth_token":  "twilio-secret-token",
			"from":        "+15551234567",
			"to":          "+15557654321",
		},
	}
	incoming := &domain.OutputChannel{
		Type: domain.ChannelSMS,
		Config: map[string]any{
			"account_sid": "AC1234567890",
			"auth_token":  "twilio***",
			"from":        "+15551234567",
			"to":          "+15557654321",
		},
	}

	merged := MergeSensitiveOutputChannel(existing, incoming)
	if got, _ := merged.Config["auth_token"].(string); got != "twilio-secret-token" {
		t.Fatalf("expected auth_token to be preserved, got %q", got)
	}
}

func TestRedactedOutputChannel_MasksSensitiveFields(t *testing.T) {
	channel := &domain.OutputChannel{
		Type: domain.ChannelWhatsApp,
		Config: map[string]any{
			"token":           "meta-secret-token",
			"phone_number_id": "123456789",
			"to":              "15551234567",
		},
	}

	redacted := RedactedOutputChannel(channel)
	if got, _ := redacted.Config["token"].(string); got == "meta-secret-token" {
		t.Fatal("expected token to be redacted")
	}
	if got, _ := redacted.Config["phone_number_id"].(string); got != "123456789" {
		t.Fatalf("expected non-sensitive value to stay intact, got %q", got)
	}
}
