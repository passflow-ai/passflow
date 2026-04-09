package output

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/jaak-ai/passflow-channels/domain"
)

const (
	replyDeliveryKeyPrefix = "passflow:channels:reply-delivered:"
	replyDeliveryTTL       = 7 * 24 * time.Hour
)

// DeliverySender is the subset of Sender used by the reply handler.
type DeliverySender interface {
	Send(ctx context.Context, ch *domain.OutputChannel, content string) error
}

// RuleLookup is the subset of the rule store used to resolve output channels.
type RuleLookup interface {
	GetByID(ruleID string) *domain.TriggerRule
}

// ChannelReplyMetadata carries the channel/session identity of the originating event.
type ChannelReplyMetadata struct {
	Type           string `json:"type"`
	EventID        string `json:"event_id,omitempty"`
	SenderID       string `json:"sender_id,omitempty"`
	AccountID      string `json:"account_id,omitempty"`
	MessageID      string `json:"message_id,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	ThreadTS       string `json:"thread_ts,omitempty"`
	SessionScope   string `json:"session_scope,omitempty"`
	SessionKey     string `json:"session_key,omitempty"`
}

type deliveryRequest struct {
	WorkspaceID   string                `json:"workspace_id"`
	ExecutionID   string                `json:"execution_id"`
	TriggerRuleID string                `json:"trigger_rule_id"`
	Output        string                `json:"output"`
	Channel       *ChannelReplyMetadata `json:"channel,omitempty"`
}

type deliveryResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

// ReplyHandler delivers completed execution output to the configured channel destination.
type ReplyHandler struct {
	sender DeliverySender
	store  RuleLookup
	rdb    *redis.Client
}

// NewReplyHandler creates a reply handler for internal delivery requests.
func NewReplyHandler(sender DeliverySender, store RuleLookup, rdb *redis.Client) *ReplyHandler {
	return &ReplyHandler{
		sender: sender,
		store:  store,
		rdb:    rdb,
	}
}

// RegisterInternal mounts the internal reply endpoint.
func (h *ReplyHandler) RegisterInternal(router fiber.Router) {
	router.Post("/execution-replies", h.Handle)
}

// Handle delivers a completed execution reply using the trigger rule's configured output channel.
func (h *ReplyHandler) Handle(c *fiber.Ctx) error {
	var req deliveryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.WorkspaceID == "" || req.ExecutionID == "" || req.TriggerRuleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "workspace_id, execution_id, and trigger_rule_id are required",
		})
	}
	if req.Output == "" {
		return c.JSON(deliveryResponse{Status: "skipped", Reason: "empty_output"})
	}

	if h.alreadyDelivered(c.Context(), req.ExecutionID) {
		return c.JSON(deliveryResponse{Status: "duplicate"})
	}

	rule := h.store.GetByID(req.TriggerRuleID)
	if rule == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "trigger rule not found"})
	}
	if rule.WorkspaceID != req.WorkspaceID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "trigger rule does not belong to workspace"})
	}

	replyChannel, status, reason := resolveReplyChannel(rule.Action.OutputChannel, req.Channel)
	if status == "skipped" {
		return c.JSON(deliveryResponse{Status: status, Reason: reason})
	}

	if err := h.sender.Send(c.Context(), replyChannel, req.Output); err != nil {
		log.Printf("[output.reply] failed to deliver execution %s via %s: %v", req.ExecutionID, replyChannel.Type, err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}

	h.markDelivered(c.Context(), req.ExecutionID)
	return c.JSON(deliveryResponse{Status: "delivered"})
}

func (h *ReplyHandler) alreadyDelivered(ctx context.Context, executionID string) bool {
	if h.rdb == nil {
		return false
	}
	exists, err := h.rdb.Exists(ctx, replyDeliveryKeyPrefix+executionID).Result()
	if err != nil {
		log.Printf("[output.reply] failed to check delivery idempotency for execution %s: %v", executionID, err)
		return false
	}
	return exists > 0
}

func (h *ReplyHandler) markDelivered(ctx context.Context, executionID string) {
	if h.rdb == nil {
		return
	}
	if err := h.rdb.Set(ctx, replyDeliveryKeyPrefix+executionID, "1", replyDeliveryTTL).Err(); err != nil {
		log.Printf("[output.reply] failed to mark delivery for execution %s: %v", executionID, err)
	}
}

func resolveReplyChannel(base *domain.OutputChannel, meta *ChannelReplyMetadata) (*domain.OutputChannel, string, string) {
	if base == nil {
		return nil, "skipped", "no_output_channel"
	}

	channel := cloneOutputChannel(base)

	switch channel.Type {
	case domain.ChannelSlack:
		derivedConversation := false
		if stringConfig(channel.Config, "channel") == "" {
			if meta == nil || meta.ConversationID == "" {
				return nil, "skipped", "missing_conversation_id"
			}
			channel.Config["channel"] = meta.ConversationID
			derivedConversation = true
		}
		if derivedConversation && stringConfig(channel.Config, "thread_ts") == "" && meta != nil && meta.ThreadTS != "" {
			channel.Config["thread_ts"] = meta.ThreadTS
		}
	case domain.ChannelEmail:
		if stringConfig(channel.Config, "to") == "" {
			if meta == nil || meta.SenderID == "" {
				return nil, "skipped", "missing_sender_id"
			}
			channel.Config["to"] = meta.SenderID
		}
	case domain.ChannelSMS:
		if stringConfig(channel.Config, "to") == "" {
			if meta == nil || meta.SenderID == "" {
				return nil, "skipped", "missing_sender_id"
			}
			channel.Config["to"] = meta.SenderID
		}
		if stringConfig(channel.Config, "from") == "" {
			if meta == nil || meta.AccountID == "" {
				return nil, "skipped", "missing_account_id"
			}
			channel.Config["from"] = meta.AccountID
		}
	case domain.ChannelWhatsApp:
		if stringConfig(channel.Config, "to") == "" {
			if meta == nil || meta.SenderID == "" {
				return nil, "skipped", "missing_sender_id"
			}
			channel.Config["to"] = meta.SenderID
		}
		if stringConfig(channel.Config, "phone_number_id") == "" {
			if meta == nil || meta.AccountID == "" {
				return nil, "skipped", "missing_account_id"
			}
			channel.Config["phone_number_id"] = meta.AccountID
		}
	case domain.ChannelWebhook:
		// Static destination only; sender validation happens at send time.
	default:
		return nil, "skipped", fmt.Sprintf("unsupported_output_type:%s", channel.Type)
	}

	return channel, "deliver", ""
}

func cloneOutputChannel(ch *domain.OutputChannel) *domain.OutputChannel {
	if ch == nil {
		return nil
	}

	configCopy := make(map[string]any, len(ch.Config))
	for key, value := range ch.Config {
		configCopy[key] = value
	}

	return &domain.OutputChannel{
		Type:   ch.Type,
		Config: configCopy,
	}
}

func stringConfig(config map[string]any, key string) string {
	if config == nil {
		return ""
	}
	value, _ := config[key].(string)
	return value
}

var _ DeliverySender = (*Sender)(nil)
