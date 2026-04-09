package agentenum

import "testing"

func TestAgentStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status AgentStatus
		want   bool
	}{
		{"idle is valid", StatusIdle, true},
		{"running is valid", StatusRunning, true},
		{"stopped is valid", StatusStopped, true},
		{"error is valid", StatusError, true},
		{"invalid status", AgentStatus("invalid"), false},
		{"empty status", AgentStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("AgentStatus.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgentStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status AgentStatus
		want   string
	}{
		{"idle", StatusIdle, "idle"},
		{"running", StatusRunning, "running"},
		{"stopped", StatusStopped, "stopped"},
		{"error", StatusError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("AgentStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgentStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name   string
		from   AgentStatus
		to     AgentStatus
		want   bool
	}{
		{"idle to running", StatusIdle, StatusRunning, true},
		{"idle to stopped", StatusIdle, StatusStopped, true},
		{"idle to error", StatusIdle, StatusError, false},
		{"running to idle", StatusRunning, StatusIdle, true},
		{"running to stopped", StatusRunning, StatusStopped, true},
		{"running to error", StatusRunning, StatusError, true},
		{"stopped to idle", StatusStopped, StatusIdle, true},
		{"stopped to running", StatusStopped, StatusRunning, true},
		{"stopped to error", StatusStopped, StatusError, false},
		{"error to idle", StatusError, StatusIdle, true},
		{"error to stopped", StatusError, StatusStopped, true},
		{"error to running", StatusError, StatusRunning, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.want {
				t.Errorf("AgentStatus.CanTransitionTo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgentStatus_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status AgentStatus
		want   bool
	}{
		{"idle is not active", StatusIdle, false},
		{"running is active", StatusRunning, true},
		{"stopped is not active", StatusStopped, false},
		{"error is not active", StatusError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsActive(); got != tt.want {
				t.Errorf("AgentStatus.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllStatuses(t *testing.T) {
	statuses := AllStatuses()
	if len(statuses) != 4 {
		t.Errorf("AllStatuses() returned %d statuses, want 4", len(statuses))
	}

	expected := map[AgentStatus]bool{
		StatusIdle:    true,
		StatusRunning: true,
		StatusStopped: true,
		StatusError:   true,
	}

	for _, s := range statuses {
		if !expected[s] {
			t.Errorf("Unexpected status in AllStatuses(): %v", s)
		}
	}
}
