package agentenum

// AgentStatus represents the possible states of an agent.
type AgentStatus string

const (
	StatusIdle    AgentStatus = "idle"
	StatusRunning AgentStatus = "running"
	StatusStopped AgentStatus = "stopped"
	StatusError   AgentStatus = "error"
)

// IsValid checks if the status is a valid agent status.
func (s AgentStatus) IsValid() bool {
	switch s {
	case StatusIdle, StatusRunning, StatusStopped, StatusError:
		return true
	}
	return false
}

// String returns the string representation of the status.
func (s AgentStatus) String() string {
	return string(s)
}

// CanTransitionTo checks if this status can transition to the target status.
func (s AgentStatus) CanTransitionTo(target AgentStatus) bool {
	switch s {
	case StatusIdle:
		return target == StatusRunning || target == StatusStopped
	case StatusRunning:
		return target == StatusIdle || target == StatusStopped || target == StatusError
	case StatusStopped:
		return target == StatusIdle || target == StatusRunning
	case StatusError:
		return target == StatusIdle || target == StatusStopped
	}
	return false
}

// IsActive returns true if the agent is actively running.
func (s AgentStatus) IsActive() bool {
	return s == StatusRunning
}

// AllStatuses returns all valid agent statuses.
func AllStatuses() []AgentStatus {
	return []AgentStatus{
		StatusIdle,
		StatusRunning,
		StatusStopped,
		StatusError,
	}
}
