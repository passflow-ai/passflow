package output

import (
	"fmt"
	"strings"

	"github.com/jaak-ai/passflow-channels/domain"
)

// RedactedOutputChannel returns a copy of the output channel with sensitive
// config values masked for admin API responses.
func RedactedOutputChannel(ch *domain.OutputChannel) *domain.OutputChannel {
	if ch == nil {
		return nil
	}

	return &domain.OutputChannel{
		Type:   ch.Type,
		Config: redactConfig(ch.Config),
	}
}

// MergeSensitiveOutputChannel preserves existing secrets when an admin update
// sends redacted placeholders or omits secret fields.
func MergeSensitiveOutputChannel(existing, incoming *domain.OutputChannel) *domain.OutputChannel {
	if incoming == nil {
		return nil
	}
	if existing == nil || existing.Type != incoming.Type {
		return incoming
	}

	if incoming.Config == nil {
		incoming.Config = map[string]any{}
	}

	redactedExisting := redactConfig(existing.Config)
	for key, existingValue := range existing.Config {
		if !isSensitiveConfigKey(key) {
			continue
		}

		incomingValue, ok := incoming.Config[key]
		if !ok {
			incoming.Config[key] = existingValue
			continue
		}

		incomingString, ok := incomingValue.(string)
		if !ok || incomingString == "" {
			incoming.Config[key] = existingValue
			continue
		}

		if redactedValue, ok := redactedExisting[key].(string); ok && incomingString == redactedValue {
			incoming.Config[key] = existingValue
		}
	}

	return incoming
}

// ValidateReplyOutputChannel validates reply delivery configuration against the
// origin channel. Missing destination fields are only allowed when they can be
// derived safely from the originating session metadata.
func ValidateReplyOutputChannel(source domain.ChannelType, ch *domain.OutputChannel) error {
	if ch == nil {
		return nil
	}

	cfg := ch.Config
	if cfg == nil {
		cfg = map[string]any{}
	}

	switch ch.Type {
	case domain.ChannelSlack:
		token, _ := cfg["token"].(string)
		channel, _ := cfg["channel"].(string)
		if token == "" {
			return fmt.Errorf("output_channel.config.token is required for slack replies")
		}
		if channel == "" && source != domain.ChannelSlack {
			return fmt.Errorf("output_channel.config.channel is required unless the source channel is slack")
		}
	case domain.ChannelEmail:
		to, _ := cfg["to"].(string)
		subject, _ := cfg["subject"].(string)
		if to == "" && source != domain.ChannelEmail {
			return fmt.Errorf("output_channel.config.to is required unless the source channel is email")
		}
		if to != "" {
			if err := validateEmailAddress(to); err != nil {
				return err
			}
		}
		if strings.ContainsAny(subject, "\r\n") {
			return fmt.Errorf("output.email: subject contains illegal CRLF characters")
		}
	case domain.ChannelWebhook:
		url, _ := cfg["url"].(string)
		if url == "" {
			return fmt.Errorf("output_channel.config.url is required for webhook replies")
		}
		if err := validateWebhookURL(url); err != nil {
			return err
		}
	case domain.ChannelSMS:
		accountSID, _ := cfg["account_sid"].(string)
		authToken, _ := cfg["auth_token"].(string)
		from, _ := cfg["from"].(string)
		to, _ := cfg["to"].(string)
		if accountSID == "" || !strings.HasPrefix(accountSID, "AC") {
			return fmt.Errorf("output.sms: missing or invalid account_sid")
		}
		if authToken == "" {
			return fmt.Errorf("output.sms: missing auth_token")
		}
		if from == "" && source != domain.ChannelSMS {
			return fmt.Errorf("output_channel.config.from is required unless the source channel is sms")
		}
		if to == "" && source != domain.ChannelSMS {
			return fmt.Errorf("output_channel.config.to is required unless the source channel is sms")
		}
		if from != "" {
			if err := validateE164PhoneNumber(from, "sms", "from"); err != nil {
				return err
			}
		}
		if to != "" {
			if err := validateE164PhoneNumber(to, "sms", "to"); err != nil {
				return err
			}
		}
	case domain.ChannelWhatsApp:
		token, _ := cfg["token"].(string)
		phoneNumberID, _ := cfg["phone_number_id"].(string)
		to, _ := cfg["to"].(string)
		if token == "" {
			return fmt.Errorf("output.whatsapp: missing token")
		}
		if phoneNumberID == "" && source != domain.ChannelWhatsApp {
			return fmt.Errorf("output_channel.config.phone_number_id is required unless the source channel is whatsapp")
		}
		if to == "" && source != domain.ChannelWhatsApp {
			return fmt.Errorf("output_channel.config.to is required unless the source channel is whatsapp")
		}
		if to != "" {
			if _, err := normalizeWhatsAppRecipient(to); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("output channel type %q is not supported", ch.Type)
	}

	return nil
}
