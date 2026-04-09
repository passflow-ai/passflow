package input

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/passflow-ai/passflow/cmd/passflow-channels/domain"
	"github.com/passflow-ai/passflow/cmd/passflow-channels/trigger"
)

// smsDispatcher is the narrow interface SMSHandler needs from the Dispatcher.
// Defined here so tests can inject a stub without a live Redis connection.
type smsDispatcher interface {
	Dispatch(ctx context.Context, event domain.Event)
}

// SMSHandler handles inbound SMS messages via Twilio webhooks.
// Route: POST /channels/sms/webhook/:workspaceId
//
// Twilio sends form-encoded POST requests with fields such as From, To, Body,
// and MessageSid. The handler normalizes these into an internal Event and
// dispatches via the trigger system. It responds with empty TwiML to acknowledge
// receipt without sending a reply SMS.
type SMSHandler struct {
	dispatcher smsDispatcher
}

// NewSMSHandler creates a new SMSHandler.
func NewSMSHandler(dispatcher *trigger.Dispatcher) *SMSHandler {
	return &SMSHandler{dispatcher: dispatcher}
}

// Register registers SMS routes on the Fiber app.
func (h *SMSHandler) Register(app *fiber.App) {
	app.Post("/channels/sms/webhook/:workspaceId", h.handleInbound)
}

// handleInbound processes an inbound SMS from Twilio (POST, form-encoded).
//
// Twilio sends these form fields (among others):
//   - MessageSid: unique message identifier
//   - From: sender phone number (E.164 format, e.g. "+15551234567")
//   - To: recipient phone number
//   - Body: message text
//   - NumMedia: number of media attachments
//   - MediaUrl0..N: URLs for media attachments (if any)
//   - MediaContentType0..N: MIME types for media attachments
//
// The handler normalizes these into an Event and dispatches it. It always
// returns a minimal TwiML response to acknowledge receipt.
func (h *SMSHandler) handleInbound(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")

	// Parse Twilio form fields.
	from := c.FormValue("From")
	to := c.FormValue("To")
	body := c.FormValue("Body")
	messageSid := c.FormValue("MessageSid")

	if from == "" || messageSid == "" {
		log.Printf("[sms] Received request with missing required fields (From=%q, MessageSid=%q)", from, messageSid)
		// Still return TwiML to avoid Twilio retries.
		return h.respondTwiML(c)
	}

	fields := map[string]string{
		"from":        from,
		"to":          to,
		"text":        body,
		"message_sid": messageSid,
	}

	// Extract media attachments if present.
	numMedia := c.FormValue("NumMedia")
	if numMedia != "" && numMedia != "0" {
		fields["num_media"] = numMedia
		// Twilio sends MediaUrl0, MediaUrl1, etc.
		for i := 0; ; i++ {
			mediaURL := c.FormValue("MediaUrl" + itoa(i))
			if mediaURL == "" {
				break
			}
			fields["media_url_"+itoa(i)] = mediaURL
			contentType := c.FormValue("MediaContentType" + itoa(i))
			if contentType != "" {
				fields["media_content_type_"+itoa(i)] = contentType
			}
		}
	}

	raw := map[string]any{
		"from":        from,
		"to":          to,
		"body":        body,
		"message_sid": messageSid,
	}

	ev := domain.Event{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		Channel:     domain.ChannelSMS,
		Fields:      fields,
		Raw:         raw,
		ReceivedAt:  time.Now(),
	}

	h.dispatcher.Dispatch(context.Background(), ev)

	return h.respondTwiML(c)
}

// respondTwiML sends an empty TwiML response to Twilio, acknowledging receipt
// without sending a reply SMS.
func (h *SMSHandler) respondTwiML(c *fiber.Ctx) error {
	c.Set("Content-Type", "application/xml")
	return c.SendString("<Response></Response>")
}

// itoa converts a non-negative integer to a string without importing strconv.
// Only used for small media index numbers (0-9 typically).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := make([]byte, 0, 4)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	// Reverse.
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}
