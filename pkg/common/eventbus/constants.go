package eventbus

import (
	"strings"

	"github.com/passflow-ai/passflow/pkg/common/eventbus/types"
)

// Stream names for topic-based routing.
// Each domain has its own Redis Stream to reduce noise and improve throughput.
const (
	// StreamAgents handles agent lifecycle and domain events processed by agents
	// (incidents, leads, PRs, etc.)
	StreamAgents = "passflow:agents"

	// StreamTasks handles task-related events
	StreamTasks = "passflow:tasks"

	// StreamTriggers handles trigger events (cron, slack, email, webhook)
	StreamTriggers = "passflow:triggers"

	// StreamSync handles git synchronization events
	StreamSync = "passflow:sync"
)

// EventDomain represents the business domain of an event.
type EventDomain string

const (
	DomainAgents   EventDomain = "agents"
	DomainTasks    EventDomain = "tasks"
	DomainTriggers EventDomain = "triggers"
	DomainSync     EventDomain = "sync"
)

// String returns the string representation of EventDomain.
func (d EventDomain) String() string {
	return string(d)
}

// GetStreamForEventType returns the appropriate Redis Stream name for an event type.
// This is the main routing function that determines where events are published.
func GetStreamForEventType(eventType types.EventType) string {
	domain := GetDomainForEventType(eventType)
	return getStreamForDomain(domain)
}

// GetDomainForEventType returns the business domain for an event type.
// Domain events (incidents, leads, PRs) are routed to the agents domain
// since they are processed by agent executors.
func GetDomainForEventType(eventType types.EventType) EventDomain {
	eventStr := string(eventType)

	// Agent lifecycle events
	if strings.HasPrefix(eventStr, "agent.") {
		return DomainAgents
	}

	// Task events
	if strings.HasPrefix(eventStr, "task.") {
		return DomainTasks
	}

	// Trigger events
	if strings.HasPrefix(eventStr, "trigger.") {
		return DomainTriggers
	}

	// Sync events
	if strings.HasPrefix(eventStr, "sync.") {
		return DomainSync
	}

	// Domain events processed by agents: incidents, leads, PRs
	// These are business events that trigger agent workflows
	if strings.HasPrefix(eventStr, "incident.") ||
		strings.HasPrefix(eventStr, "lead.") ||
		strings.HasPrefix(eventStr, "pr.") {
		return DomainAgents
	}

	// Default to agents domain for unknown event types
	return DomainAgents
}

// getStreamForDomain maps a domain to its Redis Stream name.
func getStreamForDomain(domain EventDomain) string {
	switch domain {
	case DomainAgents:
		return StreamAgents
	case DomainTasks:
		return StreamTasks
	case DomainTriggers:
		return StreamTriggers
	case DomainSync:
		return StreamSync
	default:
		return StreamAgents
	}
}
