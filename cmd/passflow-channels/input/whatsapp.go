package input

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jaak-ai/passflow-channels/domain"
	"github.com/jaak-ai/passflow-channels/trigger"
)

// whatsappDispatcher is the narrow interface WhatsAppHandler needs from the Dispatcher.
// Defined here so tests can inject a stub without a live Redis connection.
type whatsappDispatcher interface {
	Dispatch(ctx context.Context, event domain.Event)
}

// WhatsAppHandler handles Meta Cloud API webhooks for WhatsApp Business.
// Routes:
//
//	GET  /channels/whatsapp/webhook              — Meta verification handshake
//	POST /channels/whatsapp/webhook/:workspaceId — incoming messages
type WhatsAppHandler struct {
	verifyToken string
	dispatcher  whatsappDispatcher
}

// NewWhatsAppHandler creates a new WhatsAppHandler.
// verifyToken is the token configured in the Meta App Dashboard for webhook verification.
// If verifyToken is empty, verification requests are rejected with 503 and incoming
// messages are rejected with 503.
func NewWhatsAppHandler(verifyToken string, dispatcher *trigger.Dispatcher) *WhatsAppHandler {
	if verifyToken == "" {
		log.Println("[whatsapp] WARNING: WHATSAPP_VERIFY_TOKEN is not set — all WhatsApp requests will be rejected (503)")
	}
	return &WhatsAppHandler{verifyToken: verifyToken, dispatcher: dispatcher}
}

// Register registers WhatsApp routes on the Fiber app.
func (h *WhatsAppHandler) Register(app *fiber.App) {
	app.Get("/channels/whatsapp/webhook", h.verify)
	app.Post("/channels/whatsapp/webhook/:workspaceId", h.handleMessage)
}

// verify handles the Meta webhook verification handshake (GET).
// Meta sends hub.mode, hub.verify_token, and hub.challenge as query parameters.
// If mode == "subscribe" and the token matches, we return the challenge as plain text.
func (h *WhatsAppHandler) verify(c *fiber.Ctx) error {
	if h.verifyToken == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "whatsapp integration not configured",
		})
	}

	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode == "subscribe" && token == h.verifyToken {
		log.Println("[whatsapp] Webhook verification succeeded")
		return c.SendString(challenge)
	}

	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "verification failed"})
}

// handleMessage processes incoming WhatsApp messages from Meta Cloud API (POST).
// Meta's payload structure:
//
//	{
//	  "object": "whatsapp_business_account",
//	  "entry": [{
//	    "id": "<WABA_ID>",
//	    "changes": [{
//	      "value": {
//	        "messaging_product": "whatsapp",
//	        "metadata": { "display_phone_number": "...", "phone_number_id": "..." },
//	        "messages": [{
//	          "from": "15551234567",
//	          "id": "wamid.xxx",
//	          "timestamp": "1677000000",
//	          "type": "text",
//	          "text": { "body": "Hello!" }
//	        }]
//	      },
//	      "field": "messages"
//	    }]
//	  }]
//	}
func (h *WhatsAppHandler) handleMessage(c *fiber.Ctx) error {
	if h.verifyToken == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "whatsapp integration not configured",
		})
	}

	workspaceID := c.Params("workspaceId")
	body := c.Body()

	var payload whatsappPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON"})
	}

	// Extract and dispatch each message from the payload.
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			if change.Field != "messages" {
				continue
			}

			// Extract phone number ID from metadata for reply routing.
			phoneNumberID := change.Value.Metadata.PhoneNumberID
			displayPhone := change.Value.Metadata.DisplayPhoneNumber

			for _, msg := range change.Value.Messages {
				fields := map[string]string{
					"from":            msg.From,
					"message_id":      msg.ID,
					"timestamp":       msg.Timestamp,
					"type":            msg.Type,
					"phone_number_id": phoneNumberID,
					"display_phone":   displayPhone,
					"waba_id":         entry.ID,
				}

				// Extract text body if present.
				if msg.Text != nil {
					fields["text"] = msg.Text.Body
				}

				// Extract media info if present.
				if msg.Image != nil {
					fields["media_type"] = "image"
					fields["media_id"] = msg.Image.ID
					if msg.Image.Caption != "" {
						fields["caption"] = msg.Image.Caption
					}
				}
				if msg.Document != nil {
					fields["media_type"] = "document"
					fields["media_id"] = msg.Document.ID
					if msg.Document.Filename != "" {
						fields["filename"] = msg.Document.Filename
					}
				}

				// Build raw map for the individual message.
				raw := map[string]any{
					"from":      msg.From,
					"id":        msg.ID,
					"timestamp": msg.Timestamp,
					"type":      msg.Type,
				}
				if msg.Text != nil {
					raw["text"] = msg.Text.Body
				}

				ev := domain.Event{
					ID:          uuid.New().String(),
					WorkspaceID: workspaceID,
					Channel:     domain.ChannelWhatsApp,
					Fields:      fields,
					Raw:         raw,
					ReceivedAt:  time.Now(),
				}

				h.dispatcher.Dispatch(context.Background(), ev)
			}
		}
	}

	// Meta requires a fast 200 response; processing happens asynchronously via dispatch.
	return c.JSON(fiber.Map{"ok": true})
}

// --- Meta Cloud API payload types ---

type whatsappPayload struct {
	Object string          `json:"object"`
	Entry  []whatsappEntry `json:"entry"`
}

type whatsappEntry struct {
	ID      string           `json:"id"`
	Changes []whatsappChange `json:"changes"`
}

type whatsappChange struct {
	Value whatsappValue `json:"value"`
	Field string        `json:"field"`
}

type whatsappValue struct {
	MessagingProduct string            `json:"messaging_product"`
	Metadata         whatsappMetadata  `json:"metadata"`
	Messages         []whatsappMessage `json:"messages"`
}

type whatsappMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

type whatsappMessage struct {
	From      string            `json:"from"`
	ID        string            `json:"id"`
	Timestamp string            `json:"timestamp"`
	Type      string            `json:"type"`
	Text      *whatsappText     `json:"text,omitempty"`
	Image     *whatsappMedia    `json:"image,omitempty"`
	Document  *whatsappDocument `json:"document,omitempty"`
}

type whatsappText struct {
	Body string `json:"body"`
}

type whatsappMedia struct {
	ID      string `json:"id"`
	Caption string `json:"caption,omitempty"`
}

type whatsappDocument struct {
	ID       string `json:"id"`
	Filename string `json:"filename,omitempty"`
}
