# A2A Package

This package implements the Agent-to-Agent (A2A) protocol for inter-agent communication in Passflow.

## Overview

The A2A protocol allows agents from different systems to discover and communicate with each other using a standardized message format. This enables Passflow agents to interoperate with external agents (e.g., Slack bots, CRM agents, email agents) without requiring custom integrations.

## Architecture

```
External Agent → A2A Message → Translator → Internal Event → EventBus → agent-executor
                                                                              ↓
External Agent ← A2A Response ← Translator ← Internal Event ← EventBus ← Result
```

## Components

### Message Types

- **Message**: Incoming A2A message from external agent
- **Response**: Outgoing A2A response to external agent
- **AgentCard**: Metadata describing an agent's capabilities
- **Capability**: A specific operation an agent can perform

### Translator

Handles bidirectional translation between A2A messages and internal events:

- `TranslateA2AToEvent(msg Message, workspaceID string) (*Event, error)`: Converts A2A message to internal event
- `TranslateEventToA2A(event Event) (*Response, error)`: Converts internal event to A2A response
- `BuildAgentCard(...)`: Constructs Agent Card metadata

## Usage

### Receiving an A2A Message

```go
translator := a2a.NewTranslator()

// Parse A2A message from HTTP request
var msg a2a.Message
json.Unmarshal(requestBody, &msg)

// Validate message
if err := msg.Validate(); err != nil {
    // Handle validation error
}

// Translate to internal event
event, err := translator.TranslateA2AToEvent(msg, "workspace-123")
if err != nil {
    // Handle translation error
}

// Publish to EventBus
eventBus.Publish(ctx, *event)
```

### Sending an A2A Response

```go
translator := a2a.NewTranslator()

// Subscribe to agent completion events
eventBus.Subscribe(ctx, types.EventAgentCompleted, func(ctx context.Context, event types.Event) error {
    // Translate to A2A response
    resp, err := translator.TranslateEventToA2A(event)
    if err != nil {
        return err
    }

    // Send response to external agent (sync or via callback)
    if callbackURL := event.Payload["a2a_callback_url"]; callbackURL != "" {
        http.Post(callbackURL.(string), "application/json", respJSON)
    }

    return nil
})
```

### Publishing Agent Cards

```go
translator := a2a.NewTranslator()

capabilities := []a2a.Capability{
    {
        CapabilityID: "qualify_lead",
        Description:  "Analyzes lead and assigns rep",
        InputSchema:  {...},
        OutputSchema: {...},
    },
}

card := translator.BuildAgentCard(
    "passflow-agent-sales",
    "Sales Lead Qualifier",
    "Qualifies incoming sales leads",
    "1.0.0",
    capabilities,
    "https://passflow.dev.jaak.ai/api/v1/a2a/messages",
    &a2a.AuthConfig{
        Type:          "bearer",
        TokenEndpoint: "https://passflow.dev.jaak.ai/api/v1/auth/a2a-token",
    },
)

// Serve at /.well-known/agents/passflow-agent-sales.json
```

## Message Flow Example

### 1. External Agent Sends Message

```json
POST /api/v1/a2a/messages
{
  "message_id": "msg-123",
  "from_agent": "external-crm",
  "to_agent": "passflow-agent-sales",
  "capability": "qualify_lead",
  "payload": {
    "lead_email": "prospect@example.com",
    "lead_name": "Jane Doe"
  },
  "context": {
    "conversation_id": "conv-456"
  },
  "timestamp": "2026-03-09T10:30:00Z"
}
```

### 2. A2A Gateway Translates to Internal Event

```go
Event{
  ID:          "event-uuid",
  Type:        "trigger.agent",
  SourceAgent: "external-crm",
  TargetAgent: "passflow-agent-sales",
  WorkspaceID: "ws-123",
  Payload: {
    "a2a_message_id": "msg-123",
    "a2a_capability": "qualify_lead",
    "a2a_conversation_id": "conv-456",
    "input": {
      "lead_email": "prospect@example.com",
      "lead_name": "Jane Doe"
    }
  }
}
```

### 3. Agent Executor Processes and Publishes Result

```go
Event{
  Type:        "agent.completed",
  SourceAgent: "passflow-agent-sales",
  TargetAgent: "external-crm",
  Payload: {
    "a2a_message_id": "msg-123",
    "a2a_conversation_id": "conv-456",
    "result": {
      "fit_score": 85,
      "assigned_rep": "john@example.com"
    }
  }
}
```

### 4. A2A Gateway Translates to Response

```json
{
  "message_id": "msg-response-uuid",
  "in_reply_to": "msg-123",
  "from_agent": "passflow-agent-sales",
  "to_agent": "external-crm",
  "payload": {
    "fit_score": 85,
    "assigned_rep": "john@example.com"
  },
  "context": {
    "conversation_id": "conv-456"
  },
  "timestamp": "2026-03-09T10:30:15Z",
  "status": "success"
}
```

## Testing

Run unit tests:

```bash
cd agent-executor/internal/a2a
go test -v
```

## References

- [Google A2A Protocol Spec (Draft)](https://github.com/google/agentic-ai/blob/main/a2a-spec.md)
- ADR-003: A2A Gateway Design
- ADR-001: EDA & Agentic AI Alignment
