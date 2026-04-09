package domain

import (
	"testing"
	"time"

	"github.com/passflow-ai/passflow/pkg/agent/agentenum"
)

func TestNewAgent(t *testing.T) {
	workspaceID := "ws-123"
	name := "Test Agent"
	persona := "You are a helpful assistant"

	agent := NewAgent(workspaceID, name, persona)

	if agent.WorkspaceID != workspaceID {
		t.Errorf("NewAgent().WorkspaceID = %v, want %v", agent.WorkspaceID, workspaceID)
	}
	if agent.Name != name {
		t.Errorf("NewAgent().Name = %v, want %v", agent.Name, name)
	}
	if agent.Persona != persona {
		t.Errorf("NewAgent().Persona = %v, want %v", agent.Persona, persona)
	}
	if agent.Status != agentenum.StatusIdle.String() {
		t.Errorf("NewAgent().Status = %v, want %v", agent.Status, agentenum.StatusIdle.String())
	}
	if agent.Integrations == nil {
		t.Error("NewAgent().Integrations should not be nil")
	}
	if agent.CreatedAt.IsZero() {
		t.Error("NewAgent().CreatedAt should not be zero")
	}
	if agent.UpdatedAt.IsZero() {
		t.Error("NewAgent().UpdatedAt should not be zero")
	}
}

func TestAgent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		agent   *Agent
		wantErr error
	}{
		{
			name: "valid agent",
			agent: &Agent{
				WorkspaceID: "ws-123",
				Name:        "Test Agent",
				Persona:     "You are helpful",
				Model: ModelConfig{
					Provider: "anthropic",
					ModelID:  "claude-opus-4-5-20251101",
				},
				Heartbeat: Heartbeat{
					Enabled:  true,
					Interval: "5m",
				},
			},
			wantErr: nil,
		},
		{
			name: "missing workspace ID",
			agent: &Agent{
				Name:    "Test Agent",
				Persona: "You are helpful",
			},
			wantErr: ErrWorkspaceIDRequired,
		},
		{
			name: "missing name",
			agent: &Agent{
				WorkspaceID: "ws-123",
				Persona:     "You are helpful",
			},
			wantErr: ErrAgentNameRequired,
		},
		{
			name: "name too long",
			agent: &Agent{
				WorkspaceID: "ws-123",
				Name:        string(make([]byte, 101)),
				Persona:     "You are helpful",
			},
			wantErr: ErrAgentNameTooLong,
		},
		{
			name: "missing persona",
			agent: &Agent{
				WorkspaceID: "ws-123",
				Name:        "Test Agent",
			},
			wantErr: ErrPersonaRequired,
		},
		{
			name: "invalid provider",
			agent: &Agent{
				WorkspaceID: "ws-123",
				Name:        "Test Agent",
				Persona:     "You are helpful",
				Model: ModelConfig{
					Provider: "invalid",
					ModelID:  "model-123",
				},
			},
			wantErr: ErrInvalidProvider,
		},
		{
			name: "provider without model ID",
			agent: &Agent{
				WorkspaceID: "ws-123",
				Name:        "Test Agent",
				Persona:     "You are helpful",
				Model: ModelConfig{
					Provider: "anthropic",
				},
			},
			wantErr: ErrModelIDRequired,
		},
		{
			name: "invalid heartbeat interval",
			agent: &Agent{
				WorkspaceID: "ws-123",
				Name:        "Test Agent",
				Persona:     "You are helpful",
				Heartbeat: Heartbeat{
					Enabled:  true,
					Interval: "3m",
				},
			},
			wantErr: ErrInvalidHeartbeatInterval,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.agent.Validate()
			if err != tt.wantErr {
				t.Errorf("Agent.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestModelConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ModelConfig
		wantErr error
	}{
		{
			name:    "empty config is valid",
			config:  ModelConfig{},
			wantErr: nil,
		},
		{
			name: "valid config",
			config: ModelConfig{
				Provider: "anthropic",
				ModelID:  "claude-opus-4-5-20251101",
			},
			wantErr: nil,
		},
		{
			name: "invalid provider",
			config: ModelConfig{
				Provider: "invalid",
				ModelID:  "model-123",
			},
			wantErr: ErrInvalidProvider,
		},
		{
			name: "missing model ID",
			config: ModelConfig{
				Provider: "openai",
			},
			wantErr: ErrModelIDRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err != tt.wantErr {
				t.Errorf("ModelConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHeartbeat_Validate(t *testing.T) {
	tests := []struct {
		name      string
		heartbeat Heartbeat
		wantErr   error
	}{
		{
			name:      "disabled heartbeat is valid",
			heartbeat: Heartbeat{Enabled: false},
			wantErr:   nil,
		},
		{
			name: "valid 5m interval",
			heartbeat: Heartbeat{
				Enabled:  true,
				Interval: "5m",
			},
			wantErr: nil,
		},
		{
			name: "valid 1h interval",
			heartbeat: Heartbeat{
				Enabled:  true,
				Interval: "1h",
			},
			wantErr: nil,
		},
		{
			name: "invalid interval",
			heartbeat: Heartbeat{
				Enabled:  true,
				Interval: "3m",
			},
			wantErr: ErrInvalidHeartbeatInterval,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.heartbeat.Validate()
			if err != tt.wantErr {
				t.Errorf("Heartbeat.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgent_GetStatus(t *testing.T) {
	agent := &Agent{Status: "running"}
	if got := agent.GetStatus(); got != agentenum.StatusRunning {
		t.Errorf("Agent.GetStatus() = %v, want %v", got, agentenum.StatusRunning)
	}
}

func TestAgent_SetStatus(t *testing.T) {
	agent := &Agent{}
	agent.SetStatus(agentenum.StatusRunning)

	if agent.Status != "running" {
		t.Errorf("Agent.Status = %v, want %v", agent.Status, "running")
	}
	if agent.UpdatedAt.IsZero() {
		t.Error("Agent.UpdatedAt should be set")
	}
}

func TestAgent_CanStart(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"can start from idle", "idle", true},
		{"can start from stopped", "stopped", true},
		{"cannot start from running", "running", false},
		{"cannot start from error", "error", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{Status: tt.status}
			if got := agent.CanStart(); got != tt.want {
				t.Errorf("Agent.CanStart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_CanStop(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"can stop from running", "running", true},
		{"can stop from idle", "idle", true},
		{"can stop from error", "error", true},
		{"cannot stop from stopped", "stopped", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{Status: tt.status}
			if got := agent.CanStop(); got != tt.want {
				t.Errorf("Agent.CanStop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_Start(t *testing.T) {
	agent := &Agent{Status: "idle"}
	err := agent.Start()

	if err != nil {
		t.Errorf("Agent.Start() error = %v, want nil", err)
	}
	if agent.Status != "running" {
		t.Errorf("Agent.Status = %v, want running", agent.Status)
	}
	if agent.LastRunAt == nil {
		t.Error("Agent.LastRunAt should be set")
	}
}

func TestAgent_Start_FromRunning(t *testing.T) {
	agent := &Agent{Status: "running"}
	err := agent.Start()

	if err == nil {
		t.Error("Agent.Start() should return error when already running")
	}
}

func TestAgent_Stop(t *testing.T) {
	agent := &Agent{Status: "running"}
	err := agent.Stop()

	if err != nil {
		t.Errorf("Agent.Stop() error = %v, want nil", err)
	}
	if agent.Status != "stopped" {
		t.Errorf("Agent.Status = %v, want stopped", agent.Status)
	}
}

func TestAgent_Stop_FromStopped(t *testing.T) {
	agent := &Agent{Status: "stopped"}
	err := agent.Stop()

	if err == nil {
		t.Error("Agent.Stop() should return error when already stopped")
	}
}

func TestAgent_IncrementSession(t *testing.T) {
	agent := &Agent{
		Stats: AgentStats{
			TotalSessions: 5,
			SessionsToday: 2,
		},
	}

	agent.IncrementSession()

	if agent.Stats.TotalSessions != 6 {
		t.Errorf("Agent.Stats.TotalSessions = %v, want 6", agent.Stats.TotalSessions)
	}
	if agent.Stats.SessionsToday != 3 {
		t.Errorf("Agent.Stats.SessionsToday = %v, want 3", agent.Stats.SessionsToday)
	}
}

func TestAgent_AddCost(t *testing.T) {
	agent := &Agent{
		Stats: AgentStats{TotalCost: 10.5},
	}

	agent.AddCost(5.25)

	if agent.Stats.TotalCost != 15.75 {
		t.Errorf("Agent.Stats.TotalCost = %v, want 15.75", agent.Stats.TotalCost)
	}
}

func TestAgent_UpdateAvgResponseTime(t *testing.T) {
	tests := []struct {
		name          string
		totalSessions int
		currentAvg    float64
		newTime       float64
		expectedAvg   float64
	}{
		{
			name:          "first session",
			totalSessions: 0,
			currentAvg:    0,
			newTime:       100,
			expectedAvg:   100,
		},
		{
			name:          "second session",
			totalSessions: 2,
			currentAvg:    100,
			newTime:       200,
			expectedAvg:   150,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				Stats: AgentStats{
					TotalSessions:   tt.totalSessions,
					AvgResponseTime: tt.currentAvg,
				},
			}
			agent.UpdateAvgResponseTime(tt.newTime)

			if agent.Stats.AvgResponseTime != tt.expectedAvg {
				t.Errorf("Agent.Stats.AvgResponseTime = %v, want %v", agent.Stats.AvgResponseTime, tt.expectedAvg)
			}
		})
	}
}

func TestValidHeartbeatIntervals(t *testing.T) {
	validIntervals := []string{"5m", "15m", "30m", "1h", "2h", "4h", "8h", "12h", "24h"}

	for _, interval := range validIntervals {
		if !ValidHeartbeatIntervals[interval] {
			t.Errorf("Interval %s should be valid", interval)
		}
	}

	invalidIntervals := []string{"1m", "3m", "10m", "3h", "6h", "1d"}
	for _, interval := range invalidIntervals {
		if ValidHeartbeatIntervals[interval] {
			t.Errorf("Interval %s should not be valid", interval)
		}
	}
}

func TestAgent_Times(t *testing.T) {
	now := time.Now().UTC()
	agent := &Agent{
		LastRunAt: &now,
		NextRunAt: &now,
	}

	if agent.LastRunAt.IsZero() {
		t.Error("LastRunAt should not be zero")
	}
	if agent.NextRunAt.IsZero() {
		t.Error("NextRunAt should not be zero")
	}
}

func TestAgentWithNewWizardFields(t *testing.T) {
	t.Run("agent with all new fields", func(t *testing.T) {
		agent := Agent{
			ID:           "test-id",
			WorkspaceID:  "ws-123",
			Name:         "Test Agent",
			Description:  "A test agent description",
			Category:     "support",
			Persona:      "You are helpful",
			Instructions: "Help users with their questions",
			Tools: []AgentTool{
				{ID: "web_search", Name: "Web Search", Enabled: true},
				{ID: "code_execution", Name: "Code Execution", Enabled: false, Config: map[string]interface{}{"timeout": 30}},
			},
			Guardrails: &AgentGuardrails{
				ProhibitedTopics:     []string{"violence", "illegal"},
				MaxResponseLength:    2000,
				RequiredLanguage:     "English",
				EscalationConditions: []string{"angry customer"},
			},
			Visibility:  "private",
			Channels:    []string{"chat", "api"},
			Icon:        "robot",
			Tags:        []string{"support", "customer"},
			Temperature: 0.7,
			MaxTokens:   4096,
			Status:      "idle",
		}

		if agent.Description != "A test agent description" {
			t.Errorf("Description = %v, want 'A test agent description'", agent.Description)
		}
		if agent.Category != "support" {
			t.Errorf("Category = %v, want 'support'", agent.Category)
		}
		if agent.Instructions != "Help users with their questions" {
			t.Errorf("Instructions = %v, want 'Help users with their questions'", agent.Instructions)
		}
		if len(agent.Tools) != 2 {
			t.Errorf("Tools length = %v, want 2", len(agent.Tools))
		}
		if agent.Tools[0].ID != "web_search" {
			t.Errorf("Tools[0].ID = %v, want 'web_search'", agent.Tools[0].ID)
		}
		if agent.Tools[1].Config["timeout"] != 30 {
			t.Errorf("Tools[1].Config[timeout] = %v, want 30", agent.Tools[1].Config["timeout"])
		}
		if agent.Guardrails == nil {
			t.Fatal("Guardrails should not be nil")
		}
		if len(agent.Guardrails.ProhibitedTopics) != 2 {
			t.Errorf("Guardrails.ProhibitedTopics length = %v, want 2", len(agent.Guardrails.ProhibitedTopics))
		}
		if agent.Guardrails.MaxResponseLength != 2000 {
			t.Errorf("Guardrails.MaxResponseLength = %v, want 2000", agent.Guardrails.MaxResponseLength)
		}
		if agent.Guardrails.RequiredLanguage != "English" {
			t.Errorf("Guardrails.RequiredLanguage = %v, want 'English'", agent.Guardrails.RequiredLanguage)
		}
		if agent.Visibility != "private" {
			t.Errorf("Visibility = %v, want 'private'", agent.Visibility)
		}
		if len(agent.Channels) != 2 {
			t.Errorf("Channels length = %v, want 2", len(agent.Channels))
		}
		if agent.Icon != "robot" {
			t.Errorf("Icon = %v, want 'robot'", agent.Icon)
		}
		if len(agent.Tags) != 2 {
			t.Errorf("Tags length = %v, want 2", len(agent.Tags))
		}
		if agent.Temperature != 0.7 {
			t.Errorf("Temperature = %v, want 0.7", agent.Temperature)
		}
		if agent.MaxTokens != 4096 {
			t.Errorf("MaxTokens = %v, want 4096", agent.MaxTokens)
		}
	})

	t.Run("agent with optional fields empty", func(t *testing.T) {
		agent := Agent{
			ID:          "test-id",
			WorkspaceID: "ws-123",
			Name:        "Minimal Agent",
			Persona:     "You are helpful",
			Status:      "idle",
		}

		if agent.Description != "" {
			t.Errorf("Description should be empty, got %v", agent.Description)
		}
		if agent.Tools != nil {
			t.Errorf("Tools should be nil, got %v", agent.Tools)
		}
		if agent.Guardrails != nil {
			t.Errorf("Guardrails should be nil, got %v", agent.Guardrails)
		}
	})
}

func TestAgentTool(t *testing.T) {
	tool := AgentTool{
		ID:      "web_search",
		Name:    "Web Search",
		Enabled: true,
		Config: map[string]interface{}{
			"max_results": 10,
			"safe_search": true,
		},
	}

	if tool.ID != "web_search" {
		t.Errorf("ID = %v, want 'web_search'", tool.ID)
	}
	if tool.Name != "Web Search" {
		t.Errorf("Name = %v, want 'Web Search'", tool.Name)
	}
	if !tool.Enabled {
		t.Error("Enabled = false, want true")
	}
	if tool.Config["max_results"] != 10 {
		t.Errorf("Config[max_results] = %v, want 10", tool.Config["max_results"])
	}
}

func TestAgentGuardrails(t *testing.T) {
	guardrails := AgentGuardrails{
		ProhibitedTopics:     []string{"violence", "illegal activities"},
		MaxResponseLength:    5000,
		RequiredLanguage:     "Spanish",
		EscalationConditions: []string{"angry", "refund request"},
	}

	if len(guardrails.ProhibitedTopics) != 2 {
		t.Errorf("ProhibitedTopics length = %v, want 2", len(guardrails.ProhibitedTopics))
	}
	if guardrails.ProhibitedTopics[0] != "violence" {
		t.Errorf("ProhibitedTopics[0] = %v, want 'violence'", guardrails.ProhibitedTopics[0])
	}
	if guardrails.MaxResponseLength != 5000 {
		t.Errorf("MaxResponseLength = %v, want 5000", guardrails.MaxResponseLength)
	}
	if guardrails.RequiredLanguage != "Spanish" {
		t.Errorf("RequiredLanguage = %v, want 'Spanish'", guardrails.RequiredLanguage)
	}
	if len(guardrails.EscalationConditions) != 2 {
		t.Errorf("EscalationConditions length = %v, want 2", len(guardrails.EscalationConditions))
	}
}
