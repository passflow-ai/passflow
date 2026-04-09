package output

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"net/url"
	"strings"
	"testing"

	"github.com/jaak-ai/passflow-channels/config"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// --- sanitizeHeader tests ---

func TestSanitizeHeader_Clean(t *testing.T) {
	got := sanitizeHeader("user@example.com")
	if got != "user@example.com" {
		t.Errorf("expected unchanged value, got %q", got)
	}
}

func TestSanitizeHeader_StripsCR(t *testing.T) {
	got := sanitizeHeader("user@example.com\r")
	if strings.Contains(got, "\r") {
		t.Errorf("sanitizeHeader should strip \\r, got %q", got)
	}
}

func TestSanitizeHeader_StripLF(t *testing.T) {
	got := sanitizeHeader("user@example.com\n")
	if strings.Contains(got, "\n") {
		t.Errorf("sanitizeHeader should strip \\n, got %q", got)
	}
}

func TestSanitizeHeader_StripsCRLF_InjectionAttempt(t *testing.T) {
	// The injection vector uses CRLF to start a new header line.
	// sanitizeHeader must remove the CRLF so the injected text cannot start
	// a new header — the attacker's payload becomes harmless concatenated text.
	malicious := "victim@example.com\r\nBCC: attacker@evil.com"
	got := sanitizeHeader(malicious)
	if strings.Contains(got, "\r") || strings.Contains(got, "\n") {
		t.Errorf("sanitizeHeader should strip CRLF sequences, got %q", got)
	}
	// The resulting value must not contain a bare newline that could start a
	// new SMTP header line.  The text "BCC:" may remain as harmless data but
	// must not be preceded by a line break.
	if strings.ContainsAny(got, "\r\n") {
		t.Errorf("sanitizeHeader left CR/LF in output, got %q", got)
	}
}

func TestSanitizeHeader_MultipleNewlines(t *testing.T) {
	input := "foo\r\nbar\r\nbaz"
	got := sanitizeHeader(input)
	if strings.ContainsAny(got, "\r\n") {
		t.Errorf("sanitizeHeader should strip all newlines, got %q", got)
	}
}

// --- validateEmailAddress tests ---

func TestValidateEmailAddress_Valid(t *testing.T) {
	cases := []string{
		"user@example.com",
		"user+tag@sub.domain.org",
		"User Name <user@example.com>",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			if err := validateEmailAddress(c); err != nil {
				t.Errorf("expected valid address, got error: %v", err)
			}
		})
	}
}

func TestValidateEmailAddress_Invalid(t *testing.T) {
	cases := []string{
		"notanemail",
		"@nodomain",
		"",
		"user@",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			if err := validateEmailAddress(c); err == nil {
				t.Errorf("expected error for %q, got nil", c)
			}
		})
	}
}

func TestValidateEmailAddress_RejectsInjectedCRLF(t *testing.T) {
	malicious := "victim@example.com\r\nBCC: attacker@evil.com"
	if err := validateEmailAddress(malicious); err == nil {
		t.Errorf("expected error for CRLF injection attempt, got nil")
	}
}

// --- sendEmail integration: header injection rejected ---

func TestSendEmail_RejectsHeaderInjectionInTo(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "localhost",
		SMTPPort:     "587",
		SMTPUser:     "user",
		SMTPPassword: "pass",
		SMTPFrom:     "from@example.com",
	}
	s := New(cfg)
	emailCfg := map[string]interface{}{
		"to":      "victim@example.com\r\nBCC: attacker@evil.com",
		"subject": "Hello",
	}
	err := s.sendEmail(emailCfg, "body content")
	if err == nil {
		t.Error("expected error for CRLF injection in 'to' field, got nil")
	}
}

func TestSendEmail_RejectsHeaderInjectionInSubject(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "localhost",
		SMTPPort:     "587",
		SMTPUser:     "user",
		SMTPPassword: "pass",
		SMTPFrom:     "from@example.com",
	}
	s := New(cfg)
	emailCfg := map[string]interface{}{
		"to":      "user@example.com",
		"subject": "Hello\r\nBCC: attacker@evil.com",
	}
	err := s.sendEmail(emailCfg, "body content")
	if err == nil {
		t.Error("expected error for CRLF injection in 'subject' field, got nil")
	}
}

func TestSendEmail_RejectsInvalidEmailAddress(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "localhost",
		SMTPPort:     "587",
		SMTPUser:     "user",
		SMTPPassword: "pass",
		SMTPFrom:     "from@example.com",
	}
	s := New(cfg)
	emailCfg := map[string]interface{}{
		"to":      "notanemail",
		"subject": "Hello",
	}
	err := s.sendEmail(emailCfg, "body content")
	if err == nil {
		t.Error("expected error for invalid email address, got nil")
	}
}

// Verify net/mail is usable (compilation check).
var _ = mail.ParseAddress

func TestSendSMS_UsesTwilioMessagesAPI(t *testing.T) {
	var authHeader string
	var bodyValues url.Values
	var requestPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		requestPath = r.URL.Path
		data, _ := io.ReadAll(r.Body)
		bodyValues, _ = url.ParseQuery(string(data))
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	s := New(&config.Config{})
	err := s.sendSMS(context.Background(), map[string]interface{}{
		"api_url":     server.URL,
		"account_sid": "TEST_ACCT_00000000000000000000",
		"auth_token":  "test-auth-token",
		"from":        "+15557654321",
		"to":          "+15551234567",
	}, "hola sms")
	if err != nil {
		t.Fatalf("sendSMS returned error: %v", err)
	}

	expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("TEST_ACCT_00000000000000000000:test-auth-token"))
	if authHeader != expectedAuth {
		t.Fatalf("Authorization header = %q, want %q", authHeader, expectedAuth)
	}
	if requestPath != "/2010-04-01/Accounts/TEST_ACCT_00000000000000000000/Messages.json" {
		t.Fatalf("request path = %q, want Twilio Messages path", requestPath)
	}
	if bodyValues.Get("From") != "+15557654321" {
		t.Fatalf("From = %q, want +15557654321", bodyValues.Get("From"))
	}
	if bodyValues.Get("To") != "+15551234567" {
		t.Fatalf("To = %q, want +15551234567", bodyValues.Get("To"))
	}
	if bodyValues.Get("Body") != "hola sms" {
		t.Fatalf("Body = %q, want hola sms", bodyValues.Get("Body"))
	}
}

func TestSendWhatsApp_UsesMetaMessagesAPI(t *testing.T) {
	var authHeader string
	var requestPath string
	var payload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		requestPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := New(&config.Config{})
	err := s.sendWhatsApp(context.Background(), map[string]interface{}{
		"api_url":         server.URL,
		"token":           "meta-access-token",
		"phone_number_id": "phone-42",
		"to":              "5215512345678",
	}, "hola whatsapp")
	if err != nil {
		t.Fatalf("sendWhatsApp returned error: %v", err)
	}

	if authHeader != "Bearer meta-access-token" {
		t.Fatalf("Authorization header = %q, want Bearer meta-access-token", authHeader)
	}
	if requestPath != "/phone-42/messages" {
		t.Fatalf("request path = %q, want /phone-42/messages", requestPath)
	}
	if payload["messaging_product"] != "whatsapp" {
		t.Fatalf("messaging_product = %#v, want whatsapp", payload["messaging_product"])
	}
	if payload["to"] != "5215512345678" {
		t.Fatalf("to = %#v, want 5215512345678", payload["to"])
	}
	textPayload, ok := payload["text"].(map[string]any)
	if !ok {
		t.Fatalf("expected text payload, got %#v", payload["text"])
	}
	if textPayload["body"] != "hola whatsapp" {
		t.Fatalf("text.body = %#v, want hola whatsapp", textPayload["body"])
	}
}

func TestSendSlack_IncludesThreadTSWhenConfigured(t *testing.T) {
	oldClient := httpClient
	defer func() {
		httpClient = oldClient
	}()

	var payload map[string]any
	httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != "https://slack.com/api/chat.postMessage" {
				t.Fatalf("unexpected URL: %s", req.URL.String())
			}
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			}, nil
		}),
	}

	s := New(&config.Config{})
	err := s.sendSlack(context.Background(), map[string]interface{}{
		"token":     "xoxb-test-token",
		"channel":   "C123",
		"thread_ts": "1710000000.001",
	}, "hola slack")
	if err != nil {
		t.Fatalf("sendSlack returned error: %v", err)
	}
	if payload["thread_ts"] != "1710000000.001" {
		t.Fatalf("thread_ts = %#v, want 1710000000.001", payload["thread_ts"])
	}
}
