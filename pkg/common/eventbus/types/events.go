package types

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of event in the system.
type EventType string

const (
	// Agent lifecycle events
	EventAgentStarted   EventType = "agent.started"
	EventAgentCompleted EventType = "agent.completed"
	EventAgentFailed    EventType = "agent.failed"

	// Incident events (for monitoring agents)
	EventIncidentDetected  EventType = "incident.detected"
	EventIncidentResolved  EventType = "incident.resolved"
	EventIncidentEscalated EventType = "incident.escalated"

	// Lead events (for sales agents)
	EventLeadReceived  EventType = "lead.received"
	EventLeadQualified EventType = "lead.qualified"
	EventLeadResponded EventType = "lead.responded"
	EventLeadStale     EventType = "lead.stale"

	// Task events (for productivity agents)
	EventTaskCreated   EventType = "task.created"
	EventTaskCompleted EventType = "task.completed"
	EventTaskBlocked   EventType = "task.blocked"

	// PR events (for engineering agents)
	EventPROpened          EventType = "pr.opened"
	EventPRReviewRequested EventType = "pr.review_requested"
	EventPRMerged          EventType = "pr.merged"
	EventPRFailed          EventType = "pr.failed"

	// Trigger events
	EventTriggerAgent EventType = "trigger.agent"

	// Sync events (passflow-sync → passflow-api)
	EventAgentDetected EventType = "sync.agent_detected"

	// Human-in-the-Loop events
	EventAgentAwaitingApproval EventType = "agent.awaiting_approval"
	EventApprovalGranted       EventType = "approval.granted"
	EventApprovalRejected      EventType = "approval.rejected"
)

// Priority represents the priority level of an event.
type Priority int

const (
	PriorityLow      Priority = 0
	PriorityNormal   Priority = 1
	PriorityHigh     Priority = 2
	PriorityCritical Priority = 3
)

// String returns the string representation of Priority.
func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// IsValid checks if the priority value is valid.
func (p Priority) IsValid() bool {
	return p >= PriorityLow && p <= PriorityCritical
}

// Event represents a message exchanged between agents.
type Event struct {
	ID            string                 `json:"id"`
	CorrelationID string                 `json:"correlation_id"`
	Type          EventType              `json:"type"`
	Domain        string                 `json:"domain"` // Business domain for topic routing
	SourceAgent   string                 `json:"source_agent"`
	TargetAgent   string                 `json:"target_agent,omitempty"`
	WorkspaceID   string                 `json:"workspace_id"`
	Payload       map[string]interface{} `json:"payload"`
	Priority      Priority               `json:"priority"`
	Timestamp     time.Time              `json:"timestamp"`
	Processed     bool                   `json:"processed"`
	ProcessedAt   *time.Time             `json:"processed_at,omitempty"`
	RetryCount    int                    `json:"retry_count"`
	MaxRetries    int                    `json:"max_retries"`
}

// Validation errors
var (
	ErrEmptyEventType   = errors.New("event type cannot be empty")
	ErrEmptySourceAgent = errors.New("source agent cannot be empty")
	ErrEmptyWorkspaceID = errors.New("workspace ID cannot be empty")
	ErrInvalidPriority  = errors.New("invalid priority value")
)

// NewEvent creates a new Event with default values.
// The Domain field is automatically set based on the event type.
func NewEvent(eventType EventType, sourceAgent, workspaceID string) *Event {
	return &Event{
		ID:            uuid.New().String(),
		CorrelationID: uuid.New().String(),
		Type:          eventType,
		Domain:        getDomainForEventType(eventType),
		SourceAgent:   sourceAgent,
		WorkspaceID:   workspaceID,
		Payload:       make(map[string]interface{}),
		Priority:      PriorityNormal,
		Timestamp:     time.Now().UTC(),
		Processed:     false,
		RetryCount:    0,
		MaxRetries:    3,
	}
}

// getDomainForEventType returns the business domain for an event type.
// This is duplicated from eventbus.GetDomainForEventType to avoid circular dependency.
func getDomainForEventType(eventType EventType) string {
	eventStr := string(eventType)

	// Agent lifecycle events
	if len(eventStr) >= 6 && eventStr[:6] == "agent." {
		return "agents"
	}

	// Task events
	if len(eventStr) >= 5 && eventStr[:5] == "task." {
		return "tasks"
	}

	// Trigger events
	if len(eventStr) >= 8 && eventStr[:8] == "trigger." {
		return "triggers"
	}

	// Sync events
	if len(eventStr) >= 5 && eventStr[:5] == "sync." {
		return "sync"
	}

	// Domain events processed by agents: incidents, leads, PRs
	if (len(eventStr) >= 9 && eventStr[:9] == "incident.") ||
		(len(eventStr) >= 5 && eventStr[:5] == "lead.") ||
		(len(eventStr) >= 3 && eventStr[:3] == "pr.") {
		return "agents"
	}

	// Default to agents domain for unknown event types
	return "agents"
}

// WithTargetAgent sets the target agent for the event.
func (e *Event) WithTargetAgent(targetAgent string) *Event {
	e.TargetAgent = targetAgent
	return e
}

// WithPayload sets the payload for the event.
func (e *Event) WithPayload(payload map[string]interface{}) *Event {
	e.Payload = payload
	return e
}

// WithPriority sets the priority for the event.
func (e *Event) WithPriority(priority Priority) *Event {
	e.Priority = priority
	return e
}

// WithMaxRetries sets the maximum retry count for the event.
func (e *Event) WithMaxRetries(maxRetries int) *Event {
	e.MaxRetries = maxRetries
	return e
}

// WithCorrelationID sets the correlation ID for the event.
// This enables cross-service request tracing.
func (e *Event) WithCorrelationID(correlationID string) *Event {
	e.CorrelationID = correlationID
	return e
}

// Validate checks if the event has all required fields.
func (e *Event) Validate() error {
	if e.Type == "" {
		return ErrEmptyEventType
	}
	if e.SourceAgent == "" {
		return ErrEmptySourceAgent
	}
	if e.WorkspaceID == "" {
		return ErrEmptyWorkspaceID
	}
	if !e.Priority.IsValid() {
		return ErrInvalidPriority
	}
	return nil
}

// CanRetry checks if the event can be retried.
func (e *Event) CanRetry() bool {
	return e.RetryCount < e.MaxRetries
}

// IncrementRetry increments the retry count.
func (e *Event) IncrementRetry() {
	e.RetryCount++
}

// MarkProcessed marks the event as processed.
func (e *Event) MarkProcessed() {
	e.Processed = true
	now := time.Now().UTC()
	e.ProcessedAt = &now
}

// AgentTrigger represents a trigger configuration between agents.
type AgentTrigger struct {
	ID          string    `bson:"_id" json:"id"`
	SourceAgent string    `bson:"source_agent" json:"source_agent"`
	TargetAgent string    `bson:"target_agent" json:"target_agent"`
	EventType   EventType `bson:"event_type" json:"event_type"`
	Condition   string    `bson:"condition,omitempty" json:"condition,omitempty"`
	IsActive    bool      `bson:"is_active" json:"is_active"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}

// NewAgentTrigger creates a new AgentTrigger.
func NewAgentTrigger(sourceAgent, targetAgent string, eventType EventType) *AgentTrigger {
	now := time.Now().UTC()
	return &AgentTrigger{
		ID:          uuid.New().String(),
		SourceAgent: sourceAgent,
		TargetAgent: targetAgent,
		EventType:   eventType,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// WithCondition sets the condition for the trigger.
func (t *AgentTrigger) WithCondition(condition string) *AgentTrigger {
	t.Condition = condition
	return t
}

// Deactivate deactivates the trigger.
func (t *AgentTrigger) Deactivate() {
	t.IsActive = false
	t.UpdatedAt = time.Now().UTC()
}

// Activate activates the trigger.
func (t *AgentTrigger) Activate() {
	t.IsActive = true
	t.UpdatedAt = time.Now().UTC()
}
