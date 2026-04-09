package input

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jaak-ai/passflow-channels/domain"
	"github.com/jaak-ai/passflow-channels/trigger"
)

// slackDispatcher is the narrow interface SlackHandler needs from the Dispatcher.
// Defined here so tests can inject a stub without a live Redis connection.
type slackDispatcher interface {
	Dispatch(ctx context.Context, event domain.Event)
}

// SlackHandler handles Slack Events API payloads.
// Route: POST /channels/slack/:workspaceId/events
type SlackHandler struct {
	signingSecret string
	dispatcher    slackDispatcher
}

// NewSlackHandler creates a new SlackHandler.
// It logs a warning at construction time when signingSecret is empty so the
// misconfiguration is visible in startup logs.
// If signingSecret is empty every incoming request will be rejected with 503.
func NewSlackHandler(signingSecret string, dispatcher *trigger.Dispatcher) *SlackHandler {
	if signingSecret == "" {
		log.Println("[slack] WARNING: SLACK_SIGNING_SECRET is not set — all Slack requests will be rejected (503)")
	}
	return &SlackHandler{signingSecret: signingSecret, dispatcher: dispatcher}
}

// Register registers Slack routes on the Fiber app.
func (h *SlackHandler) Register(app *fiber.App) {
	app.Post("/channels/slack/:workspaceId/events", h.handle)
}

func (h *SlackHandler) handle(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	body := c.Body()

	// Reject all requests when no signing secret is configured.
	// An empty secret means the service is misconfigured; accepting requests
	// without verification would allow any caller to trigger agent actions.
	if h.signingSecret == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "slack integration not configured",
		})
	}

	// Verify Slack signature — mandatory, never skipped.
	if err := h.verifySignature(c, body); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid signature"})
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON"})
	}

	// URL verification challenge (Slack requirement when setting up Events API)
	if payloadType, _ := payload["type"].(string); payloadType == "url_verification" {
		challenge, _ := payload["challenge"].(string)
		return c.JSON(fiber.Map{"challenge": challenge})
	}

	// Normalize event fields
	fields := make(map[string]string)
	if event, ok := payload["event"].(map[string]interface{}); ok {
		if shouldIgnoreSlackEvent(event) {
			return c.JSON(fiber.Map{"ok": true})
		}
		if text, ok := event["text"].(string); ok {
			fields["text"] = text
		}
		if user, ok := event["user"].(string); ok {
			fields["user"] = user
		}
		if channel, ok := event["channel"].(string); ok {
			fields["channel"] = channel
		}
		if ts, ok := event["ts"].(string); ok {
			fields["ts"] = ts
		}
		if threadTS, ok := event["thread_ts"].(string); ok {
			fields["thread_ts"] = threadTS
		}
		if eventType, ok := event["type"].(string); ok {
			fields["event_type"] = eventType
		}
	}
	if teamID, ok := payload["team_id"].(string); ok {
		fields["team_id"] = teamID
	}

	ev := domain.Event{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		Channel:     domain.ChannelSlack,
		Fields:      fields,
		Raw:         payload,
		ReceivedAt:  time.Now(),
	}

	h.dispatcher.Dispatch(context.Background(), ev)

	return c.JSON(fiber.Map{"ok": true})
}

func shouldIgnoreSlackEvent(event map[string]interface{}) bool {
	if botID, _ := event["bot_id"].(string); botID != "" {
		return true
	}

	switch subtype, _ := event["subtype"].(string); subtype {
	case "bot_message", "message_changed", "message_deleted":
		return true
	default:
		return false
	}
}

// verifySignature validates the Slack signing secret.
func (h *SlackHandler) verifySignature(c *fiber.Ctx, body []byte) error {
	timestamp := c.Get("X-Slack-Request-Timestamp")
	signature := c.Get("X-Slack-Signature")

	if timestamp == "" || signature == "" {
		return fmt.Errorf("missing slack signature headers")
	}

	// Prevent replay attacks (>5 min old)
	ts, err := parseTimestamp(timestamp)
	if err != nil || time.Since(ts) > 5*time.Minute {
		return fmt.Errorf("request too old")
	}

	baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(h.signingSecret))
	mac.Write([]byte(baseString))
	expected := "v0=" + hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}

// year2000Unix is the Unix timestamp for 2000-01-01 00:00:00 UTC.
const year2000Unix = 946684800

// year2100Unix is the Unix timestamp for 2100-01-01 00:00:00 UTC.
const year2100Unix = 4102444800

// parseTimestamp parses a Unix-second timestamp string from a Slack header.
// It rejects:
//   - strings containing non-digit characters
//   - strings longer than 19 digits (would overflow int64)
//   - timestamps outside the range [year 2000, year 2100]
func parseTimestamp(ts string) (time.Time, error) {
	if ts == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}
	// Guard against int64 overflow: max int64 has 19 digits.
	if len(ts) > 19 {
		return time.Time{}, fmt.Errorf("timestamp too long")
	}

	sec, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp: %w", err)
	}

	if sec < year2000Unix || sec > year2100Unix {
		return time.Time{}, fmt.Errorf("timestamp %d is outside the acceptable range [2000, 2100]", sec)
	}

	return time.Unix(sec, 0), nil
}
