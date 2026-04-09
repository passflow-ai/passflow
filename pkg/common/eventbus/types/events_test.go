package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPriority_String(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		expected string
	}{
		{"Low priority", PriorityLow, "low"},
		{"Normal priority", PriorityNormal, "normal"},
		{"High priority", PriorityHigh, "high"},
		{"Critical priority", PriorityCritical, "critical"},
		{"Unknown priority", Priority(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.priority.String())
		})
	}
}

func TestPriority_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		expected bool
	}{
		{"Low is valid", PriorityLow, true},
		{"Normal is valid", PriorityNormal, true},
		{"High is valid", PriorityHigh, true},
		{"Critical is valid", PriorityCritical, true},
		{"Negative is invalid", Priority(-1), false},
		{"Above critical is invalid", Priority(4), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.priority.IsValid())
		})
	}
}

func TestNewEvent(t *testing.T) {
	event := NewEvent(EventIncidentDetected, "log-monitor", "workspace-123")

	require.NotNil(t, event)
	assert.NotEmpty(t, event.ID)
	assert.NotEmpty(t, event.CorrelationID, "CorrelationID should be auto-generated")
	assert.Equal(t, EventIncidentDetected, event.Type)
	assert.Equal(t, "agents", event.Domain) // incident events route to agents domain
	assert.Equal(t, "log-monitor", event.SourceAgent)
	assert.Equal(t, "workspace-123", event.WorkspaceID)
	assert.Empty(t, event.TargetAgent)
	assert.NotNil(t, event.Payload)
	assert.Equal(t, PriorityNormal, event.Priority)
	assert.False(t, event.Processed)
	assert.Nil(t, event.ProcessedAt)
	assert.Equal(t, 0, event.RetryCount)
	assert.Equal(t, 3, event.MaxRetries)
	assert.WithinDuration(t, time.Now().UTC(), event.Timestamp, time.Second)
}

func TestNewEvent_DomainRouting(t *testing.T) {
	tests := []struct {
		name           string
		eventType      EventType
		expectedDomain string
	}{
		{
			name:           "agent event routes to agents domain",
			eventType:      EventAgentStarted,
			expectedDomain: "agents",
		},
		{
			name:           "task event routes to tasks domain",
			eventType:      EventTaskCreated,
			expectedDomain: "tasks",
		},
		{
			name:           "trigger event routes to triggers domain",
			eventType:      EventTriggerAgent,
			expectedDomain: "triggers",
		},
		{
			name:           "sync event routes to sync domain",
			eventType:      EventAgentDetected,
			expectedDomain: "sync",
		},
		{
			name:           "incident event routes to agents domain",
			eventType:      EventIncidentDetected,
			expectedDomain: "agents",
		},
		{
			name:           "lead event routes to agents domain",
			eventType:      EventLeadReceived,
			expectedDomain: "agents",
		},
		{
			name:           "PR event routes to agents domain",
			eventType:      EventPROpened,
			expectedDomain: "agents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewEvent(tt.eventType, "test-agent", "workspace-123")
			assert.Equal(t, tt.expectedDomain, event.Domain)
		})
	}
}

func TestEvent_WithTargetAgent(t *testing.T) {
	event := NewEvent(EventTriggerAgent, "source", "workspace").
		WithTargetAgent("target-agent")

	assert.Equal(t, "target-agent", event.TargetAgent)
}

func TestEvent_WithPayload(t *testing.T) {
	payload := map[string]interface{}{
		"error":    "connection timeout",
		"severity": "critical",
	}

	event := NewEvent(EventIncidentDetected, "monitor", "workspace").
		WithPayload(payload)

	assert.Equal(t, payload, event.Payload)
	assert.Equal(t, "connection timeout", event.Payload["error"])
}

func TestEvent_WithPriority(t *testing.T) {
	event := NewEvent(EventIncidentEscalated, "responder", "workspace").
		WithPriority(PriorityCritical)

	assert.Equal(t, PriorityCritical, event.Priority)
}

func TestEvent_WithMaxRetries(t *testing.T) {
	event := NewEvent(EventTaskCreated, "agent", "workspace").
		WithMaxRetries(5)

	assert.Equal(t, 5, event.MaxRetries)
}

func TestEvent_ChainedBuilders(t *testing.T) {
	payload := map[string]interface{}{"key": "value"}

	event := NewEvent(EventLeadReceived, "sales-agent", "workspace").
		WithTargetAgent("sdr-agent").
		WithPayload(payload).
		WithPriority(PriorityHigh).
		WithMaxRetries(5)

	assert.Equal(t, "sdr-agent", event.TargetAgent)
	assert.Equal(t, payload, event.Payload)
	assert.Equal(t, PriorityHigh, event.Priority)
	assert.Equal(t, 5, event.MaxRetries)
}

func TestEvent_Validate(t *testing.T) {
	tests := []struct {
		name        string
		event       *Event
		expectedErr error
	}{
		{
			name:        "Valid event",
			event:       NewEvent(EventAgentStarted, "agent", "workspace"),
			expectedErr: nil,
		},
		{
			name: "Empty event type",
			event: &Event{
				SourceAgent: "agent",
				WorkspaceID: "workspace",
				Priority:    PriorityNormal,
			},
			expectedErr: ErrEmptyEventType,
		},
		{
			name: "Empty source agent",
			event: &Event{
				Type:        EventAgentStarted,
				WorkspaceID: "workspace",
				Priority:    PriorityNormal,
			},
			expectedErr: ErrEmptySourceAgent,
		},
		{
			name: "Empty workspace ID",
			event: &Event{
				Type:        EventAgentStarted,
				SourceAgent: "agent",
				Priority:    PriorityNormal,
			},
			expectedErr: ErrEmptyWorkspaceID,
		},
		{
			name: "Invalid priority",
			event: &Event{
				Type:        EventAgentStarted,
				SourceAgent: "agent",
				WorkspaceID: "workspace",
				Priority:    Priority(99),
			},
			expectedErr: ErrInvalidPriority,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestEvent_CanRetry(t *testing.T) {
	event := NewEvent(EventAgentFailed, "agent", "workspace")

	assert.True(t, event.CanRetry())

	event.RetryCount = 2
	assert.True(t, event.CanRetry())

	event.RetryCount = 3
	assert.False(t, event.CanRetry())

	event.RetryCount = 4
	assert.False(t, event.CanRetry())
}

func TestEvent_IncrementRetry(t *testing.T) {
	event := NewEvent(EventAgentFailed, "agent", "workspace")

	assert.Equal(t, 0, event.RetryCount)

	event.IncrementRetry()
	assert.Equal(t, 1, event.RetryCount)

	event.IncrementRetry()
	assert.Equal(t, 2, event.RetryCount)
}

func TestEvent_MarkProcessed(t *testing.T) {
	event := NewEvent(EventAgentCompleted, "agent", "workspace")

	assert.False(t, event.Processed)
	assert.Nil(t, event.ProcessedAt)

	event.MarkProcessed()

	assert.True(t, event.Processed)
	require.NotNil(t, event.ProcessedAt)
	assert.WithinDuration(t, time.Now().UTC(), *event.ProcessedAt, time.Second)
}

func TestNewAgentTrigger(t *testing.T) {
	trigger := NewAgentTrigger("log-monitor", "incident-responder", EventIncidentDetected)

	require.NotNil(t, trigger)
	assert.NotEmpty(t, trigger.ID)
	assert.Equal(t, "log-monitor", trigger.SourceAgent)
	assert.Equal(t, "incident-responder", trigger.TargetAgent)
	assert.Equal(t, EventIncidentDetected, trigger.EventType)
	assert.Empty(t, trigger.Condition)
	assert.True(t, trigger.IsActive)
	assert.WithinDuration(t, time.Now().UTC(), trigger.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now().UTC(), trigger.UpdatedAt, time.Second)
}

func TestAgentTrigger_WithCondition(t *testing.T) {
	trigger := NewAgentTrigger("source", "target", EventLeadStale).
		WithCondition("hours_since_contact > 2")

	assert.Equal(t, "hours_since_contact > 2", trigger.Condition)
}

func TestAgentTrigger_Deactivate(t *testing.T) {
	trigger := NewAgentTrigger("source", "target", EventTaskBlocked)
	originalUpdatedAt := trigger.UpdatedAt

	time.Sleep(time.Millisecond * 10)
	trigger.Deactivate()

	assert.False(t, trigger.IsActive)
	assert.True(t, trigger.UpdatedAt.After(originalUpdatedAt))
}

func TestAgentTrigger_Activate(t *testing.T) {
	trigger := NewAgentTrigger("source", "target", EventPROpened)
	trigger.IsActive = false
	originalUpdatedAt := trigger.UpdatedAt

	time.Sleep(time.Millisecond * 10)
	trigger.Activate()

	assert.True(t, trigger.IsActive)
	assert.True(t, trigger.UpdatedAt.After(originalUpdatedAt))
}

func TestEventType_Constants(t *testing.T) {
	expectedEvents := map[EventType]string{
		EventAgentStarted:          "agent.started",
		EventAgentCompleted:        "agent.completed",
		EventAgentFailed:           "agent.failed",
		EventIncidentDetected:      "incident.detected",
		EventIncidentResolved:      "incident.resolved",
		EventIncidentEscalated:     "incident.escalated",
		EventLeadReceived:          "lead.received",
		EventLeadQualified:         "lead.qualified",
		EventLeadResponded:         "lead.responded",
		EventLeadStale:             "lead.stale",
		EventTaskCreated:           "task.created",
		EventTaskCompleted:         "task.completed",
		EventTaskBlocked:           "task.blocked",
		EventPROpened:              "pr.opened",
		EventPRReviewRequested:     "pr.review_requested",
		EventPRMerged:              "pr.merged",
		EventPRFailed:              "pr.failed",
		EventTriggerAgent:          "trigger.agent",
		EventAgentAwaitingApproval: "agent.awaiting_approval",
		EventApprovalGranted:       "approval.granted",
		EventApprovalRejected:      "approval.rejected",
	}

	for eventType, expected := range expectedEvents {
		assert.Equal(t, expected, string(eventType))
	}
}

func TestHITLEvents_Creation(t *testing.T) {
	t.Run("EventAgentAwaitingApproval", func(t *testing.T) {
		payload := map[string]interface{}{
			"approval_id": "approval-123",
			"task_id":     "task-456",
			"action":      "delete_production_database",
			"context":     "User requested deletion of production DB",
		}

		event := NewEvent(EventAgentAwaitingApproval, "cleanup-agent", "workspace-789").
			WithPayload(payload).
			WithPriority(PriorityHigh)

		assert.Equal(t, EventAgentAwaitingApproval, event.Type)
		assert.Equal(t, "cleanup-agent", event.SourceAgent)
		assert.Equal(t, "workspace-789", event.WorkspaceID)
		assert.Equal(t, PriorityHigh, event.Priority)
		assert.Equal(t, "approval-123", event.Payload["approval_id"])
		assert.NoError(t, event.Validate())
	})

	t.Run("EventApprovalGranted", func(t *testing.T) {
		payload := map[string]interface{}{
			"approval_id": "approval-123",
			"task_id":     "task-456",
			"approved_by": "user-789",
		}

		event := NewEvent(EventApprovalGranted, "system", "workspace-789").
			WithTargetAgent("cleanup-agent").
			WithPayload(payload)

		assert.Equal(t, EventApprovalGranted, event.Type)
		assert.Equal(t, "cleanup-agent", event.TargetAgent)
		assert.Equal(t, "approval-123", event.Payload["approval_id"])
		assert.NoError(t, event.Validate())
	})

	t.Run("EventApprovalRejected", func(t *testing.T) {
		payload := map[string]interface{}{
			"approval_id": "approval-123",
			"task_id":     "task-456",
			"rejected_by": "user-789",
			"reason":      "Too risky for production",
		}

		event := NewEvent(EventApprovalRejected, "system", "workspace-789").
			WithTargetAgent("cleanup-agent").
			WithPayload(payload)

		assert.Equal(t, EventApprovalRejected, event.Type)
		assert.Equal(t, "cleanup-agent", event.TargetAgent)
		assert.Equal(t, "Too risky for production", event.Payload["reason"])
		assert.NoError(t, event.Validate())
	})
}
