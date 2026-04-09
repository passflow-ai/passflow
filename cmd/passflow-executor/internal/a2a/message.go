package a2a

import "time"

// Message represents an A2A message sent from one agent to another.
// Based on Google's Agent-to-Agent protocol specification.
type Message struct {
	MessageID   string                 `json:"message_id"`
	FromAgent   string                 `json:"from_agent"`
	ToAgent     string                 `json:"to_agent"`
	Capability  string                 `json:"capability,omitempty"`
	Payload     map[string]interface{} `json:"payload"`
	Context     *MessageContext        `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Auth        *AuthInfo              `json:"auth,omitempty"`
	CallbackURL string                 `json:"callback_url,omitempty"`
}

// MessageContext holds conversation and correlation IDs for tracking.
type MessageContext struct {
	ConversationID string `json:"conversation_id,omitempty"`
	CorrelationID  string `json:"correlation_id,omitempty"`
}

// AuthInfo contains authentication information for the message.
type AuthInfo struct {
	Type  string `json:"type"`  // "bearer", "mtls"
	Token string `json:"token,omitempty"`
}

// Response represents an A2A response message.
type Response struct {
	MessageID string                 `json:"message_id"`
	InReplyTo string                 `json:"in_reply_to"`
	FromAgent string                 `json:"from_agent"`
	ToAgent   string                 `json:"to_agent"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Context   *MessageContext        `json:"context,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Status    string                 `json:"status"` // "success", "error"
	Error     *ErrorInfo             `json:"error,omitempty"`
}

// ErrorInfo contains error details in an A2A response.
type ErrorInfo struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// AgentCard describes an agent's capabilities and metadata.
// Published at /.well-known/agent.json or /.well-known/agents/{agent_id}.json
type AgentCard struct {
	AgentID      string       `json:"agent_id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Version      string       `json:"version"`
	Capabilities []Capability `json:"capabilities"`
	Auth         *AuthConfig  `json:"auth"`
	Endpoint     string       `json:"endpoint"`
}

// Capability describes a specific capability an agent can perform.
type Capability struct {
	CapabilityID string                 `json:"capability_id"`
	Description  string                 `json:"description"`
	InputSchema  map[string]interface{} `json:"input_schema"`
	OutputSchema map[string]interface{} `json:"output_schema"`
}

// AuthConfig describes the authentication mechanism for the agent.
type AuthConfig struct {
	Type          string `json:"type"` // "bearer", "mtls"
	TokenEndpoint string `json:"token_endpoint,omitempty"`
}

// AgentCardList is a collection of agent cards.
type AgentCardList struct {
	Agents []AgentCard `json:"agents"`
}

// Validate checks if the A2A message has all required fields.
func (m *Message) Validate() error {
	if m.MessageID == "" {
		return ErrEmptyMessageID
	}
	if m.FromAgent == "" {
		return ErrEmptyFromAgent
	}
	if m.ToAgent == "" {
		return ErrEmptyToAgent
	}
	if m.Timestamp.IsZero() {
		return ErrEmptyTimestamp
	}
	return nil
}

// Validate checks if the A2A response has all required fields.
func (r *Response) Validate() error {
	if r.MessageID == "" {
		return ErrEmptyMessageID
	}
	if r.FromAgent == "" {
		return ErrEmptyFromAgent
	}
	if r.ToAgent == "" {
		return ErrEmptyToAgent
	}
	if r.Status == "" {
		return ErrEmptyStatus
	}
	if r.Status != "success" && r.Status != "error" {
		return ErrInvalidStatus
	}
	if r.Status == "error" && r.Error == nil {
		return ErrMissingError
	}
	return nil
}

// Validate checks if the agent card has all required fields.
func (a *AgentCard) Validate() error {
	if a.AgentID == "" {
		return ErrEmptyAgentID
	}
	if a.Name == "" {
		return ErrEmptyName
	}
	if a.Endpoint == "" {
		return ErrEmptyEndpoint
	}
	return nil
}
