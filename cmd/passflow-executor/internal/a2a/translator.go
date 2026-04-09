package a2a

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Event is a placeholder interface for the internal event type.
// This will be replaced with the actual import from passflow-api
// when integrating: "github.com/jaak-ai/passflow-api/src/packages/eventbus/types"
//
// For now, we define a minimal interface to avoid circular dependencies
// during the design phase.
type Event struct {
	ID          string
	Type        string
	SourceAgent string
	TargetAgent string
	WorkspaceID string
	Payload     map[string]interface{}
	Priority    int
	Timestamp   time.Time
	Processed   bool
}

// EventType constants (will be imported from types package)
const (
	EventTriggerAgent   = "trigger.agent"
	EventAgentCompleted = "agent.completed"
	EventAgentFailed    = "agent.failed"
)

// Translator handles conversion between A2A messages and internal events.
type Translator struct{}

// NewTranslator creates a new Translator instance.
func NewTranslator() *Translator {
	return &Translator{}
}

// TranslateA2AToEvent converts an A2A message to an internal Event.
//
// The A2A message contains:
//   - message_id: unique identifier for the A2A message
//   - from_agent: source agent ID
//   - to_agent: target agent ID (Passflow agent)
//   - capability: the capability being invoked
//   - payload: input data for the agent
//   - context: conversation/correlation IDs for tracking
//   - callback_url: optional URL for async responses
//
// The internal Event will contain:
//   - Type: EventTriggerAgent (to trigger the target agent)
//   - SourceAgent: from_agent
//   - TargetAgent: to_agent
//   - Payload: includes A2A metadata + input data
func (t *Translator) TranslateA2AToEvent(msg Message, workspaceID string) (*Event, error) {
	if err := msg.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTranslationFailed, err)
	}

	event := &Event{
		ID:          uuid.New().String(),
		Type:        EventTriggerAgent,
		SourceAgent: msg.FromAgent,
		TargetAgent: msg.ToAgent,
		WorkspaceID: workspaceID,
		Payload:     make(map[string]interface{}),
		Priority:    1, // Normal priority
		Timestamp:   time.Now().UTC(),
		Processed:   false,
	}

	// Embed A2A metadata in payload for tracking and response generation
	event.Payload["a2a_message_id"] = msg.MessageID
	event.Payload["a2a_capability"] = msg.Capability
	event.Payload["a2a_callback_url"] = msg.CallbackURL

	if msg.Context != nil {
		event.Payload["a2a_conversation_id"] = msg.Context.ConversationID
		event.Payload["a2a_correlation_id"] = msg.Context.CorrelationID
	}

	// Copy input payload to event
	event.Payload["input"] = msg.Payload

	return event, nil
}

// TranslateEventToA2A converts an internal Event to an A2A response.
//
// The internal Event should contain:
//   - Type: EventAgentCompleted or EventAgentFailed
//   - SourceAgent: the agent that processed the request
//   - TargetAgent: the agent that initiated the request (external agent)
//   - Payload: includes A2A metadata + result data
//
// The A2A Response will contain:
//   - message_id: new unique ID for the response
//   - in_reply_to: original A2A message_id
//   - from_agent: SourceAgent
//   - to_agent: TargetAgent
//   - payload: result data
//   - status: "success" or "error"
func (t *Translator) TranslateEventToA2A(event Event) (*Response, error) {
	resp := &Response{
		MessageID: uuid.New().String(),
		FromAgent: event.SourceAgent,
		ToAgent:   event.TargetAgent,
		Timestamp: time.Now().UTC(),
		Status:    "success",
		Payload:   make(map[string]interface{}),
	}

	// Extract A2A metadata from event payload
	if msgID, ok := event.Payload["a2a_message_id"].(string); ok {
		resp.InReplyTo = msgID
	}

	// Extract context
	if conversationID, ok := event.Payload["a2a_conversation_id"].(string); ok {
		if resp.Context == nil {
			resp.Context = &MessageContext{}
		}
		resp.Context.ConversationID = conversationID
	}
	if correlationID, ok := event.Payload["a2a_correlation_id"].(string); ok {
		if resp.Context == nil {
			resp.Context = &MessageContext{}
		}
		resp.Context.CorrelationID = correlationID
	}

	// Extract result payload
	if result, ok := event.Payload["result"].(map[string]interface{}); ok {
		resp.Payload = result
	}

	// Handle errors
	if event.Type == EventAgentFailed {
		resp.Status = "error"
		errorMsg := "agent execution failed"
		if errPayload, ok := event.Payload["error"].(string); ok {
			errorMsg = errPayload
		}
		resp.Error = &ErrorInfo{
			Code:    "AGENT_EXECUTION_FAILED",
			Message: errorMsg,
			Details: make(map[string]interface{}),
		}
		// Copy error details if available
		if details, ok := event.Payload["error_details"].(map[string]interface{}); ok {
			resp.Error.Details = details
		}
	}

	return resp, nil
}

// BuildAgentCard constructs an AgentCard from agent metadata.
//
// This is a helper function to generate Agent Cards from the agent
// configuration stored in MongoDB. The agent document should include
// an "a2a_capabilities" field with capability definitions.
func (t *Translator) BuildAgentCard(
	agentID string,
	name string,
	description string,
	version string,
	capabilities []Capability,
	endpoint string,
	authConfig *AuthConfig,
) *AgentCard {
	return &AgentCard{
		AgentID:      agentID,
		Name:         name,
		Description:  description,
		Version:      version,
		Capabilities: capabilities,
		Auth:         authConfig,
		Endpoint:     endpoint,
	}
}
