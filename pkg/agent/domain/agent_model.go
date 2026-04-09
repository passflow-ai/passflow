package domain

import (
	"errors"
	"time"

	"github.com/passflow-ai/passflow/pkg/agent/agentenum"
)

// Validation errors.
var (
	ErrAgentNameRequired         = errors.New("agent name is required")
	ErrAgentNameTooLong          = errors.New("agent name must be 100 characters or less")
	ErrWorkspaceIDRequired       = errors.New("workspace ID is required")
	ErrPersonaRequired           = errors.New("persona is required")
	ErrInvalidProvider           = errors.New("invalid model provider")
	ErrModelIDRequired           = errors.New("model ID is required")
	ErrInvalidHeartbeatInterval  = errors.New("invalid heartbeat interval")
	ErrInvalidCPUFormat          = errors.New("cpu value must be in millicores format (e.g. '500m') or cores (e.g. '1')")
	ErrInvalidMemoryFormat       = errors.New("memory value must include unit (e.g. '512Mi', '2Gi')")
	ErrInvalidTimeout            = errors.New("max execution time must be between 0 and 86400 seconds (24h)")
	ErrInvalidConcurrentSessions = errors.New("max concurrent sessions must be between 0 and 100")
	ErrProviderNotEnabled        = errors.New("provider is not enabled for this workspace")
	ErrModelNotAllowed           = errors.New("model is not allowed for this workspace provider")
)

// Operation errors.
var (
	ErrAgentNotFound      = errors.New("agent not found")
	ErrAgentAlreadyExists = errors.New("agent with this name already exists in workspace")
)

// ResourcePreset defines named resource tiers for agent pods.
type ResourcePreset string

const (
	ResourcePresetSmall  ResourcePreset = "small"
	ResourcePresetMedium ResourcePreset = "medium"
	ResourcePresetLarge  ResourcePreset = "large"
	ResourcePresetCustom ResourcePreset = "custom"
)

// ResourceConfig defines the compute resources and execution limits for an agent pod.
type ResourceConfig struct {
	// Kubernetes resource format: "200m", "1000m", "2"
	CPURequest string `json:"cpuRequest,omitempty" bson:"cpu_request,omitempty"`
	CPULimit   string `json:"cpuLimit,omitempty" bson:"cpu_limit,omitempty"`
	// Kubernetes memory format: "256Mi", "1Gi", "2Gi"
	MemoryRequest string `json:"memoryRequest,omitempty" bson:"memory_request,omitempty"`
	MemoryLimit   string `json:"memoryLimit,omitempty" bson:"memory_limit,omitempty"`
	// Ephemeral storage for agents that need to write files (e.g. "1Gi")
	StorageRequest string `json:"storageRequest,omitempty" bson:"storage_request,omitempty"`
	// Maximum wall-clock time in seconds before the pod is killed (0 = platform default 3600)
	MaxExecutionTime int `json:"maxExecutionTime,omitempty" bson:"max_execution_time,omitempty"`
	// Maximum number of concurrent executions for this agent (0 = unlimited)
	MaxConcurrentSessions int `json:"maxConcurrentSessions,omitempty" bson:"max_concurrent_sessions,omitempty"`
	// Allowlist of tool IDs the agent may invoke (empty = all tools allowed)
	AllowedTools []string `json:"allowedTools,omitempty" bson:"allowed_tools,omitempty"`
	// Named preset used in the UI (informational only, not enforced at runtime)
	Preset ResourcePreset `json:"preset,omitempty" bson:"preset,omitempty"`
}

// DefaultResourceConfig returns sensible defaults matching the current platform limits.
func DefaultResourceConfig() ResourceConfig {
	return ResourceConfig{
		CPURequest:            "200m",
		CPULimit:              "1000m",
		MemoryRequest:         "256Mi",
		MemoryLimit:           "1Gi",
		MaxExecutionTime:      3600,
		MaxConcurrentSessions: 0,
		AllowedTools:          []string{},
		Preset:                ResourcePresetSmall,
	}
}

// PresetConfigs returns predefined resource configurations.
var PresetConfigs = map[ResourcePreset]ResourceConfig{
	ResourcePresetSmall: {
		CPURequest: "200m", CPULimit: "1000m",
		MemoryRequest: "256Mi", MemoryLimit: "512Mi",
		MaxExecutionTime: 1800, Preset: ResourcePresetSmall,
	},
	ResourcePresetMedium: {
		CPURequest: "500m", CPULimit: "2000m",
		MemoryRequest: "512Mi", MemoryLimit: "1Gi",
		MaxExecutionTime: 3600, Preset: ResourcePresetMedium,
	},
	ResourcePresetLarge: {
		CPURequest: "1000m", CPULimit: "4000m",
		MemoryRequest: "1Gi", MemoryLimit: "2Gi",
		MaxExecutionTime: 14400, Preset: ResourcePresetLarge,
	},
}

// Validate checks that resource config values are within acceptable bounds.
func (r *ResourceConfig) Validate() error {
	if r.MaxExecutionTime < 0 || r.MaxExecutionTime > 86400 {
		return ErrInvalidTimeout
	}
	if r.MaxConcurrentSessions < 0 || r.MaxConcurrentSessions > 100 {
		return ErrInvalidConcurrentSessions
	}
	return nil
}

// AgentTool represents a tool that an agent can use.
type AgentTool struct {
	ID      string                 `json:"id" bson:"id"`
	Name    string                 `json:"name" bson:"name"`
	Enabled bool                   `json:"enabled" bson:"enabled"`
	Config  map[string]interface{} `json:"config,omitempty" bson:"config,omitempty"`
}

// AgentGuardrails represents restrictions and safety configurations for an agent.
type AgentGuardrails struct {
	ProhibitedTopics     []string `json:"prohibitedTopics" bson:"prohibited_topics"`
	MaxResponseLength    int      `json:"maxResponseLength,omitempty" bson:"max_response_length,omitempty"`
	RequiredLanguage     string   `json:"requiredLanguage,omitempty" bson:"required_language,omitempty"`
	EscalationConditions []string `json:"escalationConditions,omitempty" bson:"escalation_conditions,omitempty"`
}

// Agent represents an AI agent configuration.
type Agent struct {
	ID             string                   `json:"id" bson:"_id"`
	WorkspaceID    string                   `json:"workspaceId" bson:"workspace_id"`
	Name           string                   `json:"name" bson:"name"`
	Description    string                   `json:"description,omitempty" bson:"description,omitempty"`
	Category       string                   `json:"category,omitempty" bson:"category,omitempty"`
	TemplateID     string                   `json:"templateId,omitempty" bson:"template_id,omitempty"`
	Persona        string                   `json:"persona" bson:"persona"`
	Instructions   string                   `json:"instructions,omitempty" bson:"instructions,omitempty"`
	Model          ModelConfig              `json:"model" bson:"model"`
	Heartbeat      Heartbeat                `json:"heartbeat" bson:"heartbeat"`
	Integrations   []string                 `json:"integrations" bson:"integrations"`
	Tools          []AgentTool              `json:"tools,omitempty" bson:"tools,omitempty"`
	Guardrails     *AgentGuardrails         `json:"guardrails,omitempty" bson:"guardrails,omitempty"`
	Visibility     string                   `json:"visibility,omitempty" bson:"visibility,omitempty"`
	Channels       []string                 `json:"channels,omitempty" bson:"channels,omitempty"`
	Icon           string                   `json:"icon,omitempty" bson:"icon,omitempty"`
	Tags           []string                 `json:"tags,omitempty" bson:"tags,omitempty"`
	Temperature    float64                  `json:"temperature,omitempty" bson:"temperature,omitempty"`
	MaxTokens      int                      `json:"maxTokens,omitempty" bson:"max_tokens,omitempty"`
	Resources      ResourceConfig           `json:"resources" bson:"resources"`
	TriggerIds     []string                 `json:"triggerIds,omitempty" bson:"trigger_ids,omitempty"`
	Status         string                   `json:"status" bson:"status"`
	LifecycleState agentenum.LifecycleState `json:"lifecycleState" bson:"lifecycle_state"`
	LastRunAt      *time.Time               `json:"lastRunAt,omitempty" bson:"last_run_at,omitempty"`
	NextRunAt      *time.Time               `json:"nextRunAt,omitempty" bson:"next_run_at,omitempty"`
	Stats          AgentStats               `json:"stats" bson:"stats"`
	CreatedAt      time.Time                `json:"createdAt" bson:"created_at"`
	UpdatedAt      time.Time                `json:"updatedAt" bson:"updated_at"`
}

// ModelConfig represents the AI model configuration.
type ModelConfig struct {
	Provider string `json:"provider" bson:"provider"`
	ModelID  string `json:"modelId" bson:"model_id"`
	APIKeyID string `json:"apiKeyId" bson:"api_key_id"`
}

// Heartbeat represents the heartbeat configuration.
type Heartbeat struct {
	Enabled   bool   `json:"enabled" bson:"enabled"`
	Interval  string `json:"interval" bson:"interval"`
	Checklist string `json:"checklist" bson:"checklist"`
}

// AgentStats represents the agent statistics.
type AgentStats struct {
	TotalSessions   int     `json:"totalSessions" bson:"total_sessions"`
	SessionsToday   int     `json:"sessionsToday" bson:"sessions_today"`
	TotalCost       float64 `json:"totalCost" bson:"total_cost"`
	AvgResponseTime float64 `json:"avgResponseTime" bson:"avg_response_time"`
}

// ValidHeartbeatIntervals contains all valid heartbeat interval values.
var ValidHeartbeatIntervals = map[string]bool{
	"5m":  true,
	"15m": true,
	"30m": true,
	"1h":  true,
	"2h":  true,
	"4h":  true,
	"8h":  true,
	"12h": true,
	"24h": true,
}

// NewAgent creates a new Agent with default values.
func NewAgent(workspaceID, name, persona string) *Agent {
	now := time.Now().UTC()
	return &Agent{
		WorkspaceID:  workspaceID,
		Name:         name,
		Persona:      persona,
		Integrations: []string{},
		Resources:    DefaultResourceConfig(),
		Status:       agentenum.StatusIdle.String(),
		Stats:        AgentStats{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// Validate checks if the agent has all required fields with valid values.
func (a *Agent) Validate() error {
	if a.WorkspaceID == "" {
		return ErrWorkspaceIDRequired
	}
	if a.Name == "" {
		return ErrAgentNameRequired
	}
	if len(a.Name) > 100 {
		return ErrAgentNameTooLong
	}
	if a.Persona == "" {
		return ErrPersonaRequired
	}
	if err := a.Model.Validate(); err != nil {
		return err
	}
	if err := a.Heartbeat.Validate(); err != nil {
		return err
	}
	if err := a.Resources.Validate(); err != nil {
		return err
	}
	return nil
}

// Validate checks if the model config is valid.
func (m *ModelConfig) Validate() error {
	if m.Provider != "" {
		provider := agentenum.ModelProvider(m.Provider)
		if !provider.IsValid() {
			return ErrInvalidProvider
		}
		if m.ModelID == "" {
			return ErrModelIDRequired
		}
	}
	return nil
}

// Validate checks if the heartbeat config is valid.
func (h *Heartbeat) Validate() error {
	if h.Enabled && h.Interval != "" {
		if !ValidHeartbeatIntervals[h.Interval] {
			return ErrInvalidHeartbeatInterval
		}
	}
	return nil
}

// GetStatus returns the agent status as an enum.
func (a *Agent) GetStatus() agentenum.AgentStatus {
	return agentenum.AgentStatus(a.Status)
}

// SetStatus sets the agent status.
func (a *Agent) SetStatus(status agentenum.AgentStatus) {
	a.Status = status.String()
	a.UpdatedAt = time.Now().UTC()
}

// CanStart checks if the agent can be started.
func (a *Agent) CanStart() bool {
	status := a.GetStatus()
	return status.CanTransitionTo(agentenum.StatusRunning)
}

// CanStop checks if the agent can be stopped.
func (a *Agent) CanStop() bool {
	status := a.GetStatus()
	return status.CanTransitionTo(agentenum.StatusStopped)
}

// Start transitions the agent to running status.
func (a *Agent) Start() error {
	if !a.CanStart() {
		return errors.New("agent cannot be started from current status")
	}
	now := time.Now().UTC()
	a.Status = agentenum.StatusRunning.String()
	a.LastRunAt = &now
	a.UpdatedAt = now
	return nil
}

// Stop transitions the agent to stopped status.
func (a *Agent) Stop() error {
	if !a.CanStop() {
		return errors.New("agent cannot be stopped from current status")
	}
	a.Status = agentenum.StatusStopped.String()
	a.UpdatedAt = time.Now().UTC()
	return nil
}

// IncrementSession increments the session counters.
func (a *Agent) IncrementSession() {
	a.Stats.TotalSessions++
	a.Stats.SessionsToday++
	a.UpdatedAt = time.Now().UTC()
}

// AddCost adds cost to the total cost.
func (a *Agent) AddCost(cost float64) {
	a.Stats.TotalCost += cost
	a.UpdatedAt = time.Now().UTC()
}

// UpdateAvgResponseTime updates the average response time.
func (a *Agent) UpdateAvgResponseTime(responseTime float64) {
	if a.Stats.TotalSessions == 0 {
		a.Stats.AvgResponseTime = responseTime
	} else {
		total := a.Stats.AvgResponseTime * float64(a.Stats.TotalSessions-1)
		a.Stats.AvgResponseTime = (total + responseTime) / float64(a.Stats.TotalSessions)
	}
	a.UpdatedAt = time.Now().UTC()
}
