package output

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"github.com/jaak-ai/passflow-channels/config"
	"github.com/jaak-ai/passflow-channels/domain"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// privateIPRanges lists CIDR blocks that must not be reachable from webhook
// output destinations (SSRF prevention).
var privateIPRanges []*net.IPNet

func init() {
	blocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16", // link-local / AWS metadata
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 private
	}
	for _, block := range blocks {
		_, cidr, err := net.ParseCIDR(block)
		if err == nil {
			privateIPRanges = append(privateIPRanges, cidr)
		}
	}
}

// isPrivateIP reports whether ip falls within any of the known private ranges.
func isPrivateIP(ip net.IP) bool {
	for _, block := range privateIPRanges {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// validateWebhookURL validates that the target URL is safe to POST to:
//   - scheme must be https
//   - host must not resolve to a private or loopback IP
//   - "localhost" is explicitly blocked
func validateWebhookURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("output.webhook: URL is empty")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("output.webhook: invalid URL: %w", err)
	}

	if parsed.Scheme != "https" {
		return fmt.Errorf("output.webhook: URL scheme must be https, got %q", parsed.Scheme)
	}

	host := parsed.Hostname()
	if strings.EqualFold(host, "localhost") {
		return fmt.Errorf("output.webhook: URL host %q is not allowed (SSRF prevention)", host)
	}

	// Resolve the hostname to IPs and check each one.
	ips, err := net.LookupHost(host)
	if err != nil {
		// If the host is already an IP address, LookupHost returns it as-is.
		// Try parsing directly.
		if ip := net.ParseIP(host); ip != nil {
			if isPrivateIP(ip) {
				return fmt.Errorf("output.webhook: URL host %q resolves to a private IP (SSRF prevention)", host)
			}
			return nil
		}
		// DNS resolution failure — allow through; actual dial will fail.
		return nil
	}

	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip != nil && isPrivateIP(ip) {
			return fmt.Errorf("output.webhook: URL host %q resolves to private IP %s (SSRF prevention)", host, ipStr)
		}
	}
	return nil
}

// redactConfig returns a copy of cfg with sensitive fields (e.g. "token")
// replaced with a redacted placeholder. The original map is never mutated.
//
// Security note: Slack tokens (xoxb-*, xoxp-*) and similar credentials are
// stored as cleartext in OutputChannel.Config in Redis. This is a known
// architectural risk — tokens should be encrypted at rest in a future
// iteration. Until then, redactConfig MUST be used whenever config values are
// logged to prevent credential leakage in log aggregators.
func redactConfig(cfg map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{}, len(cfg))
	for k, v := range cfg {
		if isSensitiveConfigKey(k) {
			if s, ok := v.(string); ok && s != "" {
				// Preserve a short prefix so the log is still useful for
				// identifying which token was used (e.g. "xoxb-***").
				prefix := s
				if len(prefix) > 6 {
					prefix = prefix[:6]
				}
				copy[k] = prefix + "***"
			} else {
				copy[k] = "***"
			}
		} else {
			copy[k] = v
		}
	}
	return copy
}

func isSensitiveConfigKey(key string) bool {
	lower := strings.ToLower(key)
	return strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "password")
}

// sanitizeHeader strips CR and LF characters from a header value to prevent
// SMTP header injection attacks.
func sanitizeHeader(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}

// validateEmailAddress validates that s is a well-formed email address and
// contains no CRLF sequences that could be used for header injection.
func validateEmailAddress(s string) error {
	if strings.ContainsAny(s, "\r\n") {
		return fmt.Errorf("output.email: address contains illegal CRLF characters")
	}
	if _, err := mail.ParseAddress(s); err != nil {
		return fmt.Errorf("output.email: invalid email address %q: %w", s, err)
	}
	return nil
}

func validateE164PhoneNumber(value, channel, field string) error {
	if value == "" {
		return fmt.Errorf("output.%s: missing %s", channel, field)
	}
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("output.%s: %s contains illegal CRLF characters", channel, field)
	}
	if !strings.HasPrefix(value, "+") {
		return fmt.Errorf("output.%s: %s must be in E.164 format", channel, field)
	}
	digits := strings.TrimPrefix(value, "+")
	if len(digits) < 8 || len(digits) > 15 || !isDigitsOnly(digits) {
		return fmt.Errorf("output.%s: %s must be in E.164 format", channel, field)
	}
	return nil
}

func normalizeWhatsAppRecipient(value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("output.whatsapp: missing recipient")
	}
	if strings.ContainsAny(value, "\r\n") {
		return "", fmt.Errorf("output.whatsapp: recipient contains illegal CRLF characters")
	}

	normalized := strings.TrimPrefix(strings.TrimSpace(value), "+")
	if len(normalized) < 8 || len(normalized) > 15 || !isDigitsOnly(normalized) {
		return "", fmt.Errorf("output.whatsapp: recipient must contain only digits")
	}
	return normalized, nil
}

func isDigitsOnly(value string) bool {
	if value == "" {
		return false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// Sender delivers agent output to an output channel.
type Sender struct {
	cfg *config.Config
}

// New creates a new output Sender.
func New(cfg *config.Config) *Sender {
	return &Sender{cfg: cfg}
}

// Send delivers content to the specified output channel.
func (s *Sender) Send(ctx context.Context, ch *domain.OutputChannel, content string) error {
	if ch == nil {
		return nil
	}

	switch ch.Type {
	case domain.ChannelSlack:
		return s.sendSlack(ctx, ch.Config, content)
	case domain.ChannelEmail:
		return s.sendEmail(ch.Config, content)
	case domain.ChannelWebhook:
		return s.sendWebhook(ctx, ch.Config, content)
	case domain.ChannelSMS:
		return s.sendSMS(ctx, ch.Config, content)
	case domain.ChannelWhatsApp:
		return s.sendWhatsApp(ctx, ch.Config, content)
	default:
		return fmt.Errorf("output: unsupported channel type %q", ch.Type)
	}
}

// sendSlack posts a message to a Slack channel.
func (s *Sender) sendSlack(ctx context.Context, cfg map[string]interface{}, content string) error {
	token, _ := cfg["token"].(string)
	channel, _ := cfg["channel"].(string)
	threadTS, _ := cfg["thread_ts"].(string)
	if token == "" || channel == "" {
		return fmt.Errorf("output.slack: missing token or channel")
	}
	// SECURITY NOTE: Slack tokens are stored as cleartext in Redis.
	// This is a known architectural risk — tokens should be encrypted at rest.
	// Always use redactConfig when logging this configuration value.
	log.Printf("[output.slack] sending to channel %q (config: %v)", channel, redactConfig(cfg))

	payload := map[string]interface{}{
		"channel": channel,
		"text":    content,
	}
	if threadTS != "" {
		payload["thread_ts"] = threadTS
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://slack.com/api/chat.postMessage", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("output.slack: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("output.slack: returned status %d", resp.StatusCode)
	}
	return nil
}

// sendEmail sends the content as an email via SMTP.
func (s *Sender) sendEmail(cfg map[string]interface{}, content string) error {
	to, _ := cfg["to"].(string)
	subject, _ := cfg["subject"].(string)
	if to == "" {
		return fmt.Errorf("output.email: missing 'to' address")
	}

	// Reject CRLF injection before any processing.
	if err := validateEmailAddress(to); err != nil {
		return err
	}
	if strings.ContainsAny(subject, "\r\n") {
		return fmt.Errorf("output.email: subject contains illegal CRLF characters")
	}

	// Sanitize header values as a defence-in-depth measure.
	to = sanitizeHeader(to)
	subject = sanitizeHeader(subject)

	if subject == "" {
		subject = "Passflow Agent Output"
	}

	if s.cfg.SMTPHost == "" {
		return fmt.Errorf("output.email: SMTP not configured")
	}

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	msg := strings.Join([]string{
		"From: " + s.cfg.SMTPFrom,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		content,
	}, "\r\n")

	addr := s.cfg.SMTPHost + ":" + s.cfg.SMTPPort
	return smtp.SendMail(addr, auth, s.cfg.SMTPFrom, []string{to}, []byte(msg))
}

// sendWebhook POSTs the content to a callback URL.
func (s *Sender) sendWebhook(ctx context.Context, cfg map[string]interface{}, content string) error {
	url, _ := cfg["url"].(string)
	if url == "" {
		return fmt.Errorf("output.webhook: missing 'url'")
	}

	// Validate the URL before making the request to prevent SSRF attacks.
	if err := validateWebhookURL(url); err != nil {
		return err
	}

	payload := map[string]interface{}{
		"output":    content,
		"timestamp": time.Now().UTC(),
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Optional auth header
	if token, ok := cfg["token"].(string); ok && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("output.webhook: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("output.webhook: callback returned status %d", resp.StatusCode)
	}
	return nil
}

func (s *Sender) sendSMS(ctx context.Context, cfg map[string]interface{}, content string) error {
	accountSID, _ := cfg["account_sid"].(string)
	authToken, _ := cfg["auth_token"].(string)
	from, _ := cfg["from"].(string)
	to, _ := cfg["to"].(string)
	apiURL, _ := cfg["api_url"].(string)

	if accountSID == "" || !strings.HasPrefix(accountSID, "AC") {
		return fmt.Errorf("output.sms: missing or invalid account_sid")
	}
	if authToken == "" {
		return fmt.Errorf("output.sms: missing auth_token")
	}
	if err := validateE164PhoneNumber(from, "sms", "from"); err != nil {
		return err
	}
	if err := validateE164PhoneNumber(to, "sms", "to"); err != nil {
		return err
	}
	if apiURL == "" {
		apiURL = "https://api.twilio.com"
	}

	log.Printf("[output.sms] sending SMS from %q to %q (config: %v)", from, to, redactConfig(cfg))

	form := url.Values{}
	form.Set("From", from)
	form.Set("To", to)
	form.Set("Body", content)

	reqURL := strings.TrimRight(apiURL, "/") + "/2010-04-01/Accounts/" + accountSID + "/Messages.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(accountSID, authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("output.sms: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("output.sms: returned status %d", resp.StatusCode)
	}
	return nil
}

func (s *Sender) sendWhatsApp(ctx context.Context, cfg map[string]interface{}, content string) error {
	token, _ := cfg["token"].(string)
	phoneNumberID, _ := cfg["phone_number_id"].(string)
	to, _ := cfg["to"].(string)
	apiURL, _ := cfg["api_url"].(string)

	if token == "" {
		return fmt.Errorf("output.whatsapp: missing token")
	}
	if phoneNumberID == "" {
		return fmt.Errorf("output.whatsapp: missing phone_number_id")
	}
	normalizedTo, err := normalizeWhatsAppRecipient(to)
	if err != nil {
		return err
	}
	if apiURL == "" {
		apiURL = "https://graph.facebook.com/v23.0"
	}

	log.Printf("[output.whatsapp] sending WhatsApp message via phone_number_id %q (config: %v)", phoneNumberID, redactConfig(cfg))

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                normalizedTo,
		"type":              "text",
		"text": map[string]string{
			"body": content,
		},
	}
	body, _ := json.Marshal(payload)

	reqURL := strings.TrimRight(apiURL, "/") + "/" + phoneNumberID + "/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("output.whatsapp: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("output.whatsapp: returned status %d", resp.StatusCode)
	}
	return nil
}
