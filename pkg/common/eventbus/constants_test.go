package eventbus

import (
	"testing"

	"github.com/passflow-ai/passflow/pkg/common/eventbus/types"
	"github.com/stretchr/testify/assert"
)

func TestStreamConstants(t *testing.T) {
	t.Run("stream names are unique", func(t *testing.T) {
		streams := []string{
			StreamAgents,
			StreamTasks,
			StreamTriggers,
			StreamSync,
		}

		uniqueStreams := make(map[string]bool)
		for _, stream := range streams {
			assert.False(t, uniqueStreams[stream], "stream %s is duplicated", stream)
			uniqueStreams[stream] = true
		}

		assert.Equal(t, 4, len(uniqueStreams))
	})

	t.Run("stream names follow naming convention", func(t *testing.T) {
		streams := []string{
			StreamAgents,
			StreamTasks,
			StreamTriggers,
			StreamSync,
		}

		for _, stream := range streams {
			assert.Contains(t, stream, "passflow:", "stream %s should start with passflow:", stream)
			assert.NotEqual(t, "passflow:", stream, "stream name cannot be just passflow:")
		}
	})
}

func TestGetStreamForEventType(t *testing.T) {
	tests := []struct {
		name       string
		eventType  types.EventType
		wantStream string
	}{
		// Agent lifecycle events -> passflow:agents
		{
			name:       "agent started event routes to agents stream",
			eventType:  types.EventAgentStarted,
			wantStream: StreamAgents,
		},
		{
			name:       "agent completed event routes to agents stream",
			eventType:  types.EventAgentCompleted,
			wantStream: StreamAgents,
		},
		{
			name:       "agent failed event routes to agents stream",
			eventType:  types.EventAgentFailed,
			wantStream: StreamAgents,
		},

		// Task events -> passflow:tasks
		{
			name:       "task created event routes to tasks stream",
			eventType:  types.EventTaskCreated,
			wantStream: StreamTasks,
		},
		{
			name:       "task completed event routes to tasks stream",
			eventType:  types.EventTaskCompleted,
			wantStream: StreamTasks,
		},
		{
			name:       "task blocked event routes to tasks stream",
			eventType:  types.EventTaskBlocked,
			wantStream: StreamTasks,
		},

		// Trigger events -> passflow:triggers
		{
			name:       "trigger agent event routes to triggers stream",
			eventType:  types.EventTriggerAgent,
			wantStream: StreamTriggers,
		},

		// Sync events -> passflow:sync
		{
			name:       "agent detected event routes to sync stream",
			eventType:  types.EventAgentDetected,
			wantStream: StreamSync,
		},

		// Domain events (incidents, leads, PRs) -> passflow:agents (processed by agents)
		{
			name:       "incident detected routes to agents stream",
			eventType:  types.EventIncidentDetected,
			wantStream: StreamAgents,
		},
		{
			name:       "incident resolved routes to agents stream",
			eventType:  types.EventIncidentResolved,
			wantStream: StreamAgents,
		},
		{
			name:       "incident escalated routes to agents stream",
			eventType:  types.EventIncidentEscalated,
			wantStream: StreamAgents,
		},
		{
			name:       "lead received routes to agents stream",
			eventType:  types.EventLeadReceived,
			wantStream: StreamAgents,
		},
		{
			name:       "lead qualified routes to agents stream",
			eventType:  types.EventLeadQualified,
			wantStream: StreamAgents,
		},
		{
			name:       "lead responded routes to agents stream",
			eventType:  types.EventLeadResponded,
			wantStream: StreamAgents,
		},
		{
			name:       "lead stale routes to agents stream",
			eventType:  types.EventLeadStale,
			wantStream: StreamAgents,
		},
		{
			name:       "PR opened routes to agents stream",
			eventType:  types.EventPROpened,
			wantStream: StreamAgents,
		},
		{
			name:       "PR review requested routes to agents stream",
			eventType:  types.EventPRReviewRequested,
			wantStream: StreamAgents,
		},
		{
			name:       "PR merged routes to agents stream",
			eventType:  types.EventPRMerged,
			wantStream: StreamAgents,
		},
		{
			name:       "PR failed routes to agents stream",
			eventType:  types.EventPRFailed,
			wantStream: StreamAgents,
		},

		// Unknown event types default to agents stream
		{
			name:       "unknown event type defaults to agents stream",
			eventType:  types.EventType("unknown.event"),
			wantStream: StreamAgents,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStreamForEventType(tt.eventType)
			assert.Equal(t, tt.wantStream, got)
		})
	}
}

func TestGetDomainForEventType(t *testing.T) {
	tests := []struct {
		name       string
		eventType  types.EventType
		wantDomain EventDomain
	}{
		{
			name:       "agent started belongs to agents domain",
			eventType:  types.EventAgentStarted,
			wantDomain: DomainAgents,
		},
		{
			name:       "task created belongs to tasks domain",
			eventType:  types.EventTaskCreated,
			wantDomain: DomainTasks,
		},
		{
			name:       "trigger agent belongs to triggers domain",
			eventType:  types.EventTriggerAgent,
			wantDomain: DomainTriggers,
		},
		{
			name:       "agent detected belongs to sync domain",
			eventType:  types.EventAgentDetected,
			wantDomain: DomainSync,
		},
		{
			name:       "incident detected belongs to agents domain",
			eventType:  types.EventIncidentDetected,
			wantDomain: DomainAgents,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDomainForEventType(tt.eventType)
			assert.Equal(t, tt.wantDomain, got)
		})
	}
}

func TestEventDomain_String(t *testing.T) {
	tests := []struct {
		domain EventDomain
		want   string
	}{
		{DomainAgents, "agents"},
		{DomainTasks, "tasks"},
		{DomainTriggers, "triggers"},
		{DomainSync, "sync"},
		{EventDomain("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.domain.String())
		})
	}
}
