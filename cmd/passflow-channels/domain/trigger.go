package domain

import "time"

// ChannelType identifies the source/destination of a channel.
type ChannelType string

const (
	ChannelSlack    ChannelType = "slack"
	ChannelWebhook  ChannelType = "webhook"
	ChannelEmail    ChannelType = "email"
	ChannelCron     ChannelType = "cron"
	ChannelWhatsApp ChannelType = "whatsapp"
	ChannelSMS      ChannelType = "sms"
	ChannelWebChat  ChannelType = "webchat"
)

// AuthType defines the authentication strategy for webhooks.
type AuthType string

const (
	// AuthCustomHeader uses a custom header with a plain secret (legacy behavior).
	// Header: X-Webhook-Secret or custom header name.
	AuthCustomHeader AuthType = "custom_header"

	// AuthGitHubHMAC uses GitHub's HMAC-SHA256 signature verification.
	// Header: X-Hub-Signature-256: sha256=<hex-digest>
	AuthGitHubHMAC AuthType = "github_hmac"

	// AuthStripeSignature uses Stripe's timestamp+signature verification.
	// Header: Stripe-Signature: t=<timestamp>,v1=<signature>
	AuthStripeSignature AuthType = "stripe_signature"

	// AuthShopifyHMAC uses Shopify's HMAC-SHA256 verification.
	// Header: X-Shopify-Hmac-Sha256: <base64-digest>
	AuthShopifyHMAC AuthType = "shopify_hmac"

	// AuthNone disables authentication (use only for testing/internal).
	AuthNone AuthType = "none"
)

// AuthStrategy defines how to authenticate incoming webhooks.
type AuthStrategy struct {
	// Type is the authentication strategy to use.
	Type AuthType `json:"type"`

	// Secret is the shared secret used for verification.
	Secret string `json:"secret"`

	// Header is the custom header name (only for AuthCustomHeader).
	// Defaults to "X-Webhook-Secret" if empty.
	Header string `json:"header,omitempty"`
}

// TargetType identifies what kind of resource a trigger action targets.
type TargetType string

const (
	// TargetAgent triggers a single agent execution.
	TargetAgent TargetType = "agent"

	// TargetPipeline triggers a pipeline/workflow execution.
	TargetPipeline TargetType = "pipeline"
)

// TriggerRule defines an "if X then Y" rule.
// When an event arrives that matches Condition, Action is dispatched.
type TriggerRule struct {
	ID          string      `json:"id"`
	WorkspaceID string      `json:"workspace_id"`
	Name        string      `json:"name"`
	Enabled     bool        `json:"enabled"`
	ChannelType ChannelType `json:"channel_type"`
	Condition   Condition   `json:"condition"`
	Action      Action      `json:"action"`
	CreatedAt   time.Time   `json:"created_at"`

	// WebhookSecret is DEPRECATED. Use Auth instead.
	// Kept for backwards compatibility with existing rules.
	WebhookSecret string `json:"webhook_secret,omitempty"`

	// Auth defines the authentication strategy for webhook rules.
	// If nil and WebhookSecret is set, defaults to AuthCustomHeader behavior.
	Auth *AuthStrategy `json:"auth,omitempty"`
}

// Condition defines the "if X" part of a rule.
type Condition struct {
	// Field to inspect in the incoming event.
	// Slack:   "text" | "user" | "channel"
	// Email:   "subject" | "from" | "body"
	// Webhook: "body.<json.path>" | "header.<name>"
	// Cron:    (ignored — always matches)
	Field string `json:"field"`

	// Operator: "contains" | "equals" | "starts_with" | "matches_regex" | "always"
	Operator string `json:"operator"`

	// Value to compare against.
	Value string `json:"value"`
}

// Action defines the "then Y" part — what to do when the condition matches.
type Action struct {
	// TargetType specifies whether to trigger an agent or a pipeline.
	// Defaults to "agent" if empty for backwards compatibility.
	TargetType TargetType `json:"target_type,omitempty"`

	// TargetID is the ID of the agent or pipeline to trigger.
	// Use this field instead of AgentID for new rules.
	TargetID string `json:"target_id,omitempty"`

	// AgentID is DEPRECATED. Use TargetType + TargetID instead.
	// Kept for backwards compatibility with existing rules.
	AgentID string `json:"agent_id,omitempty"`

	// InputTemplate is a Go template rendered with the event data as context.
	// Example: "Summarize this email: {{.Subject}}\n\n{{.Body}}"
	InputTemplate string `json:"input_template"`

	// ExecutionMode: "one_shot" | "react_loop"
	ExecutionMode string `json:"execution_mode"`
	MaxIterations int    `json:"max_iterations"`

	// OutputChannel: where to send the agent's response.
	OutputChannel *OutputChannel `json:"output_channel,omitempty"`
}

// GetTargetType returns the effective target type, defaulting to agent.
func (a Action) GetTargetType() TargetType {
	if a.TargetType != "" {
		return a.TargetType
	}
	return TargetAgent
}

// GetTargetID returns the effective target ID, using AgentID as fallback.
func (a Action) GetTargetID() string {
	if a.TargetID != "" {
		return a.TargetID
	}
	return a.AgentID
}

// OutputChannel defines where to deliver the agent's response.
type OutputChannel struct {
	Type   ChannelType    `json:"type"`
	Config map[string]any `json:"config"`
}

// Event is a normalized incoming event from any channel.
type Event struct {
	ID          string            `json:"id"`
	WorkspaceID string            `json:"workspace_id"`
	Channel     ChannelType       `json:"channel"`
	Fields      map[string]string `json:"fields"` // normalized key-value pairs
	Raw         map[string]any    `json:"raw"`    // original payload
	ReceivedAt  time.Time         `json:"received_at"`
}
