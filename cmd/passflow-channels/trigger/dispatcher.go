package trigger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jaak-ai/passflow-channels/domain"
)

// defaultFallbackModelProvider and defaultFallbackModelID are the built-in
// fallback values used by publishToStream when both the API call fails and no
// override is configured via environment variables.
const (
	defaultFallbackModelProvider = "anthropic"
	defaultFallbackModelID       = "claude-sonnet-4-5"
)

// AgentExecutionRequest is posted to passflow-api's internal agent execution endpoint.
type AgentExecutionRequest struct {
	Input         string                  `json:"input"`
	ExecutionMode string                  `json:"execution_mode"`
	MaxIterations int                     `json:"max_iterations"`
	Source        string                  `json:"source"`
	TriggerRuleID string                  `json:"trigger_rule_id"`
	Channel       *ChannelTriggerMetadata `json:"channel,omitempty"`
}

// ChannelTriggerMetadata carries the normalized channel/session identity that
// originated an execution request.
type ChannelTriggerMetadata struct {
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

// idempotencyKeyTTL is how long we remember an event as processed.
// After this period, a duplicate event would be processed again.
const idempotencyKeyTTL = 24 * time.Hour

// idempotencyKeyPrefix is the Redis key prefix for idempotency checks.
const idempotencyKeyPrefix = "passflow:channels:processed:"

// Dispatcher evaluates events against rules and fires actions.
// It includes built-in idempotency checking via Redis SetNX to prevent
// duplicate event processing.
type Dispatcher struct {
	rdb          *redis.Client
	streamKey    string
	apiURL       string
	serviceToken string
	httpClient   *http.Client
	store        RuleStore
	// defaultModelProvider and defaultModelID are used as the fallback model
	// when publishToStream is invoked (i.e. the API is unavailable).
	// They are configurable via DEFAULT_MODEL_PROVIDER and DEFAULT_MODEL_ID
	// environment variables.
	defaultModelProvider string
	defaultModelID       string
}

// RuleStore is the interface for retrieving rules.
type RuleStore interface {
	GetRulesForWorkspace(workspaceID string) []domain.TriggerRule
	GetAllRules() []domain.TriggerRule
}

// NewDispatcher creates a new Dispatcher. The fallback model provider and ID
// used by publishToStream are read from the DEFAULT_MODEL_PROVIDER and
// DEFAULT_MODEL_ID environment variables, falling back to the built-in
// defaults when the variables are unset or empty.
func NewDispatcher(rdb *redis.Client, streamKey, apiURL, serviceToken string, store RuleStore) *Dispatcher {
	provider := os.Getenv("DEFAULT_MODEL_PROVIDER")
	if provider == "" {
		provider = defaultFallbackModelProvider
	}
	modelID := os.Getenv("DEFAULT_MODEL_ID")
	if modelID == "" {
		modelID = defaultFallbackModelID
	}
	return NewDispatcherWithConfig(rdb, streamKey, apiURL, serviceToken, store, provider, modelID)
}

// NewDispatcherWithConfig creates a Dispatcher with explicitly provided
// fallback model settings. Use this constructor in tests and when callers
// need direct control over the fallback model without environment variables.
func NewDispatcherWithConfig(
	rdb *redis.Client,
	streamKey, apiURL, serviceToken string,
	store RuleStore,
	defaultModelProvider, defaultModelID string,
) *Dispatcher {
	return &Dispatcher{
		rdb:                  rdb,
		streamKey:            streamKey,
		apiURL:               apiURL,
		serviceToken:         serviceToken,
		httpClient:           &http.Client{Timeout: 10 * time.Second},
		store:                store,
		defaultModelProvider: defaultModelProvider,
		defaultModelID:       defaultModelID,
	}
}

// Dispatch evaluates an event against all applicable rules and fires matching actions.
// It includes idempotency checking: if an event+rule combination was already processed
// within the TTL window, it skips dispatching to prevent duplicate executions.
func (d *Dispatcher) Dispatch(ctx context.Context, event domain.Event) {
	var rules []domain.TriggerRule
	if event.WorkspaceID != "" {
		rules = d.store.GetRulesForWorkspace(event.WorkspaceID)
	} else {
		rules = d.store.GetAllRules()
	}

	for _, rule := range rules {
		if rule.ChannelType != event.Channel {
			continue
		}
		if !Matches(rule, event) {
			continue
		}

		// Idempotency check: use Redis SetNX to atomically mark this event+rule as processed.
		// If the key already exists, this is a duplicate and we skip it.
		// If Redis is not available (nil client), we proceed without idempotency check.
		if d.rdb != nil {
			idempotencyKey := fmt.Sprintf("%s%s:%s", idempotencyKeyPrefix, event.ID, rule.ID)
			wasSet, err := d.rdb.SetNX(ctx, idempotencyKey, "1", idempotencyKeyTTL).Result()
			if err != nil {
				log.Printf("[trigger] idempotency check failed for event %s rule %s: %v (proceeding anyway)",
					event.ID, rule.ID, err)
			} else if !wasSet {
				log.Printf("[trigger] skipping duplicate event %s for rule %s", event.ID, rule.ID)
				continue
			}
		}

		input, err := RenderInput(rule.Action, event)
		if err != nil {
			log.Printf("[trigger] failed to render input for rule %s: %v", rule.ID, err)
			continue
		}

		targetType := rule.Action.GetTargetType()
		targetID := rule.Action.GetTargetID()

		log.Printf("[trigger] rule %q matched event %s → dispatching %s %s",
			rule.Name, event.ID, targetType, targetID)

		// Fire action directly without semaphore - rate limiting should be done at consumer level.
		// The Redis Stream provides natural backpressure.
		d.fireAction(ctx, rule, event, input)
	}
}

// fireAction creates a task via the passflow-api.
func (d *Dispatcher) fireAction(ctx context.Context, rule domain.TriggerRule, event domain.Event, input string) {
	workspaceID := event.WorkspaceID
	if workspaceID == "" {
		workspaceID = rule.WorkspaceID
	}
	if workspaceID == "" {
		log.Printf("[trigger] skipping rule %s: workspace ID is required", rule.ID)
		return
	}

	targetType := string(rule.Action.GetTargetType())
	targetID := rule.Action.GetTargetID()
	if targetType == string(domain.TargetPipeline) {
		d.fireWorkflowAction(ctx, rule, workspaceID, targetID, input)
		return
	}
	if targetID == "" {
		log.Printf("[trigger] skipping rule %s: target agent ID is required", rule.ID)
		return
	}

	req := AgentExecutionRequest{
		Input:         input,
		ExecutionMode: rule.Action.ExecutionMode,
		MaxIterations: rule.Action.MaxIterations,
		Source:        "passflow-channels",
		TriggerRuleID: rule.ID,
		Channel:       buildChannelTriggerMetadata(event),
	}
	if req.ExecutionMode == "" {
		req.ExecutionMode = "one_shot"
	}

	body, err := json.Marshal(req)
	if err != nil {
		log.Printf("[trigger] failed to marshal task request: %v", err)
		return
	}

	url := fmt.Sprintf("%s/api/v1/internal/workspaces/%s/agents/%s/execute", d.apiURL, workspaceID, targetID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		log.Printf("[trigger] failed to create HTTP request: %v", err)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Service-Token", d.serviceToken)

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("[trigger] API request failed: %v", err)
		// Fallback: publish directly to Redis Stream
		d.publishToStream(ctx, rule, event, input)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("[trigger] API returned %d for agent execution", resp.StatusCode)
		d.publishToStream(ctx, rule, event, input)
	}
}

func (d *Dispatcher) fireWorkflowAction(
	ctx context.Context,
	rule domain.TriggerRule,
	workspaceID string,
	workflowID string,
	input string,
) {
	body, err := json.Marshal(map[string]any{
		"triggerType": string(rule.ChannelType),
		"triggerData": map[string]any{
			"input":           input,
			"trigger_rule_id": rule.ID,
			"channel":         string(rule.ChannelType),
		},
	})
	if err != nil {
		log.Printf("[trigger] failed to marshal workflow execution request: %v", err)
		return
	}

	url := fmt.Sprintf("%s/api/v1/internal/workspaces/%s/workflows/%s/execute", d.apiURL, workspaceID, workflowID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		log.Printf("[trigger] failed to create workflow execution request: %v", err)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Service-Token", d.serviceToken)

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("[trigger] workflow execution request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("[trigger] workflow execution returned %d for workflow %s", resp.StatusCode, workflowID)
	}
}

// buildStreamSpec constructs the job specification map for the Redis Stream
// fallback. It uses the Dispatcher's configured fallback model values instead
// of hardcoded strings, making the model selectable via environment variables.
func (d *Dispatcher) buildStreamSpec(rule domain.TriggerRule, event domain.Event, input string) map[string]any {
	workspaceID := event.WorkspaceID
	if workspaceID == "" {
		workspaceID = rule.WorkspaceID
	}

	mode := rule.Action.ExecutionMode
	if mode == "" {
		mode = "one_shot"
	}

	return map[string]any{
		"task_id":        uuid.New().String(),
		"execution_id":   uuid.New().String(),
		"workspace_id":   workspaceID,
		"agent_id":       rule.Action.GetTargetID(), // Use helper for backwards compat
		"target_type":    string(rule.Action.GetTargetType()),
		"target_id":      rule.Action.GetTargetID(),
		"model_provider": d.defaultModelProvider,
		"model_id":       d.defaultModelID,
		"system_prompt":  "",
		"mode":           mode,
		"max_iterations": rule.Action.MaxIterations,
		"input":          input,
		"trigger_metadata": map[string]any{
			"source":          "passflow-channels",
			"trigger_rule_id": rule.ID,
			"channel":         buildChannelTriggerMetadata(event),
		},
	}
}

// publishToStream is a fallback that publishes directly to the Redis Stream
// when the API is unavailable. The model provider and ID are taken from the
// Dispatcher's configuration (set via DEFAULT_MODEL_PROVIDER and
// DEFAULT_MODEL_ID environment variables) rather than being hardcoded.
func (d *Dispatcher) publishToStream(ctx context.Context, rule domain.TriggerRule, event domain.Event, input string) {
	if d.rdb == nil {
		log.Printf("[trigger] cannot publish fallback execution for rule %s: redis is not configured", rule.ID)
		return
	}

	spec := d.buildStreamSpec(rule, event, input)
	executionID, _ := spec["execution_id"].(string)

	payload, err := json.Marshal(spec)
	if err != nil {
		log.Printf("[trigger] failed to marshal stream payload: %v", err)
		return
	}

	err = d.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: d.streamKey,
		Values: map[string]any{"payload": string(payload)},
	}).Err()

	if err != nil {
		log.Printf("[trigger] failed to publish to stream: %v", err)
	} else {
		log.Printf("[trigger] published execution %s directly to stream", executionID)
	}
}

func buildChannelTriggerMetadata(event domain.Event) *ChannelTriggerMetadata {
	metadata := &ChannelTriggerMetadata{
		Type:    string(event.Channel),
		EventID: event.ID,
	}

	switch event.Channel {
	case domain.ChannelWhatsApp:
		metadata.SenderID = firstNonEmpty(event.Fields, "from")
		metadata.AccountID = firstNonEmpty(event.Fields, "phone_number_id", "waba_id")
		metadata.MessageID = firstNonEmpty(event.Fields, "message_id")
	case domain.ChannelSMS:
		metadata.SenderID = firstNonEmpty(event.Fields, "from")
		metadata.AccountID = firstNonEmpty(event.Fields, "to")
		metadata.MessageID = firstNonEmpty(event.Fields, "message_sid")
	case domain.ChannelSlack:
		metadata.SenderID = firstNonEmpty(event.Fields, "user")
		metadata.AccountID = firstNonEmpty(event.Fields, "team_id")
		metadata.MessageID = firstNonEmpty(event.Fields, "ts")
		metadata.ConversationID = firstNonEmpty(event.Fields, "channel")
		metadata.ThreadTS = firstNonEmpty(event.Fields, "thread_ts")
	case domain.ChannelEmail:
		metadata.SenderID = firstNonEmpty(event.Fields, "from_email", "from")
		metadata.MessageID = firstNonEmpty(event.Fields, "message_id")
	case domain.ChannelWebhook:
		metadata.MessageID = event.ID
	case domain.ChannelCron:
		metadata.MessageID = event.ID
	}

	metadata.SessionScope, metadata.SessionKey = deriveSessionIdentity(event, metadata)

	return metadata
}

func deriveSessionIdentity(event domain.Event, metadata *ChannelTriggerMetadata) (string, string) {
	switch event.Channel {
	case domain.ChannelSlack:
		channelID := firstNonEmpty(event.Fields, "channel")
		if channelID != "" && metadata.SenderID != "" {
			return "channel_member", joinNonEmpty(":", string(event.Channel), metadata.AccountID, channelID, metadata.SenderID)
		}
		if metadata.SenderID != "" {
			return "member", joinNonEmpty(":", string(event.Channel), metadata.AccountID, metadata.SenderID)
		}
	case domain.ChannelWhatsApp, domain.ChannelSMS:
		if metadata.SenderID != "" {
			return "external_sender", joinNonEmpty(":", string(event.Channel), metadata.AccountID, metadata.SenderID)
		}
	case domain.ChannelEmail:
		if metadata.SenderID != "" {
			return "email_sender", joinNonEmpty(":", string(event.Channel), metadata.SenderID)
		}
	}

	return "", ""
}

func firstNonEmpty(fields map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := fields[key]; value != "" {
			return value
		}
	}
	return ""
}

func joinNonEmpty(sep string, parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return strings.Join(filtered, sep)
}
