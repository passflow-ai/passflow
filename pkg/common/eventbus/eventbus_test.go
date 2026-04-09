package eventbus

import (
	"testing"

	"github.com/passflow-ai/passflow/pkg/common/eventbus/types"
	"github.com/stretchr/testify/assert"
)

func TestEventFilter_Matches_AllCriteria(t *testing.T) {
	filter := EventFilter{
		EventTypes:  []types.EventType{types.EventIncidentDetected, types.EventIncidentEscalated},
		SourceAgent: "log-monitor",
		TargetAgent: "incident-responder",
		WorkspaceID: "workspace-123",
		MinPriority: types.PriorityHigh,
	}

	matchingEvent := types.Event{
		Type:        types.EventIncidentDetected,
		SourceAgent: "log-monitor",
		TargetAgent: "incident-responder",
		WorkspaceID: "workspace-123",
		Priority:    types.PriorityCritical,
	}

	assert.True(t, filter.Matches(matchingEvent))
}

func TestEventFilter_Matches_EventTypes(t *testing.T) {
	filter := EventFilter{
		EventTypes: []types.EventType{types.EventLeadReceived, types.EventLeadQualified},
	}

	tests := []struct {
		name     string
		event    types.Event
		expected bool
	}{
		{
			name:     "Matching event type - LeadReceived",
			event:    types.Event{Type: types.EventLeadReceived},
			expected: true,
		},
		{
			name:     "Matching event type - LeadQualified",
			event:    types.Event{Type: types.EventLeadQualified},
			expected: true,
		},
		{
			name:     "Non-matching event type",
			event:    types.Event{Type: types.EventLeadStale},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, filter.Matches(tt.event))
		})
	}
}

func TestEventFilter_Matches_EmptyEventTypes(t *testing.T) {
	filter := EventFilter{
		EventTypes: []types.EventType{},
	}

	event := types.Event{Type: types.EventAgentStarted}
	assert.True(t, filter.Matches(event))
}

func TestEventFilter_Matches_SourceAgent(t *testing.T) {
	filter := EventFilter{
		SourceAgent: "specific-agent",
	}

	tests := []struct {
		name     string
		event    types.Event
		expected bool
	}{
		{
			name:     "Matching source agent",
			event:    types.Event{SourceAgent: "specific-agent"},
			expected: true,
		},
		{
			name:     "Non-matching source agent",
			event:    types.Event{SourceAgent: "other-agent"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, filter.Matches(tt.event))
		})
	}
}

func TestEventFilter_Matches_TargetAgent(t *testing.T) {
	filter := EventFilter{
		TargetAgent: "target-agent",
	}

	tests := []struct {
		name     string
		event    types.Event
		expected bool
	}{
		{
			name:     "Matching target agent",
			event:    types.Event{TargetAgent: "target-agent"},
			expected: true,
		},
		{
			name:     "Non-matching target agent",
			event:    types.Event{TargetAgent: "other-target"},
			expected: false,
		},
		{
			name:     "Empty target agent in event",
			event:    types.Event{TargetAgent: ""},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, filter.Matches(tt.event))
		})
	}
}

func TestEventFilter_Matches_WorkspaceID(t *testing.T) {
	filter := EventFilter{
		WorkspaceID: "workspace-abc",
	}

	tests := []struct {
		name     string
		event    types.Event
		expected bool
	}{
		{
			name:     "Matching workspace",
			event:    types.Event{WorkspaceID: "workspace-abc"},
			expected: true,
		},
		{
			name:     "Non-matching workspace",
			event:    types.Event{WorkspaceID: "workspace-xyz"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, filter.Matches(tt.event))
		})
	}
}

func TestEventFilter_Matches_MinPriority(t *testing.T) {
	filter := EventFilter{
		MinPriority: types.PriorityHigh,
	}

	tests := []struct {
		name     string
		event    types.Event
		expected bool
	}{
		{
			name:     "Critical priority meets minimum",
			event:    types.Event{Priority: types.PriorityCritical},
			expected: true,
		},
		{
			name:     "High priority meets minimum",
			event:    types.Event{Priority: types.PriorityHigh},
			expected: true,
		},
		{
			name:     "Normal priority below minimum",
			event:    types.Event{Priority: types.PriorityNormal},
			expected: false,
		},
		{
			name:     "Low priority below minimum",
			event:    types.Event{Priority: types.PriorityLow},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, filter.Matches(tt.event))
		})
	}
}

func TestEventFilter_Matches_EmptyFilter(t *testing.T) {
	filter := EventFilter{}

	event := types.Event{
		Type:        types.EventPRMerged,
		SourceAgent: "any-agent",
		TargetAgent: "any-target",
		WorkspaceID: "any-workspace",
		Priority:    types.PriorityLow,
	}

	assert.True(t, filter.Matches(event))
}

func TestEventFilter_Matches_CombinedFilters(t *testing.T) {
	filter := EventFilter{
		EventTypes:  []types.EventType{types.EventTaskCreated},
		WorkspaceID: "workspace-123",
		MinPriority: types.PriorityNormal,
	}

	tests := []struct {
		name     string
		event    types.Event
		expected bool
	}{
		{
			name: "All criteria match",
			event: types.Event{
				Type:        types.EventTaskCreated,
				WorkspaceID: "workspace-123",
				Priority:    types.PriorityHigh,
			},
			expected: true,
		},
		{
			name: "Event type doesn't match",
			event: types.Event{
				Type:        types.EventTaskCompleted,
				WorkspaceID: "workspace-123",
				Priority:    types.PriorityHigh,
			},
			expected: false,
		},
		{
			name: "Workspace doesn't match",
			event: types.Event{
				Type:        types.EventTaskCreated,
				WorkspaceID: "workspace-456",
				Priority:    types.PriorityHigh,
			},
			expected: false,
		},
		{
			name: "Priority doesn't match",
			event: types.Event{
				Type:        types.EventTaskCreated,
				WorkspaceID: "workspace-123",
				Priority:    types.PriorityLow,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, filter.Matches(tt.event))
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "passflow:events", config.StreamName)
	assert.Equal(t, "passflow-agents", config.ConsumerGroup)
	assert.Equal(t, "agent-consumer", config.ConsumerName)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 5, config.RetryDelaySeconds)
	assert.Equal(t, 10, config.BatchSize)
	assert.Equal(t, 5000, config.BlockTimeoutMillis)
}
