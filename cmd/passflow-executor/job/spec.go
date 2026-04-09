package job

// ResourceConfig carries the pod-level compute limits requested for this job.
// Zero-value fields fall back to the platform defaults in the passflow-agent-executor.
type ResourceConfig struct {
	CPURequest            string   `json:"cpu_request,omitempty"`
	CPULimit              string   `json:"cpu_limit,omitempty"`
	MemoryRequest         string   `json:"memory_request,omitempty"`
	MemoryLimit           string   `json:"memory_limit,omitempty"`
	StorageRequest        string   `json:"storage_request,omitempty"`
	MaxExecutionTime      int      `json:"max_execution_time,omitempty"`      // seconds; 0 = default (3600)
	MaxConcurrentSessions int      `json:"max_concurrent_sessions,omitempty"` // 0 = unlimited
	AllowedTools          []string `json:"allowed_tools,omitempty"`
}

// Spec is the payload published to the Redis Stream when a task is queued.
// The passflow-api publishes this; the passflow-agent-executor consumes it.
type Spec struct {
	TaskID          string            `json:"task_id"`
	ExecutionID     string            `json:"execution_id"`
	WorkspaceID     string            `json:"workspace_id"`
	AgentID         string            `json:"agent_id"`
	BuilderMetadata map[string]string `json:"builder_metadata,omitempty"`

	// Trace context propagation
	TraceID string `json:"trace_id,omitempty"`
	SpanID  string `json:"span_id,omitempty"`

	// LLM configuration
	ModelProvider string `json:"model_provider"` // openai | anthropic | ollama
	ModelID       string `json:"model_id"`       // gpt-4o | claude-3-5-sonnet | llama3 | …
	SystemPrompt  string `json:"system_prompt"`

	// Execution mode
	Mode          string `json:"mode"`           // one_shot | react_loop
	MaxIterations int    `json:"max_iterations"` // react_loop only; 0 → default (10)
	Input         string `json:"input"`          // user message / task description

	// Tools available to the agent
	Tools []ToolConfig `json:"tools,omitempty"`

	// Channels to notify when done
	OutputChannels []ChannelConfig `json:"output_channels,omitempty"`

	// Optional structured trigger context for executions that originated in a
	// channel or another external event source.
	TriggerMetadata *TriggerMetadata `json:"trigger_metadata,omitempty"`

	// Resource limits for the pod running this job
	Resources ResourceConfig `json:"resources,omitempty"`

	// ACF Configuration (optional — only set for imported agents)
	CADConfig *CADConfig `json:"cad_config,omitempty"`
	Policy    *Policy    `json:"policy,omitempty"`

	// Shadow execution mode: "normal" (default) or "shadow".
	ExecutionMode string        `json:"execution_mode,omitempty"`
	ShadowConfig  *ShadowConfig `json:"shadow_config,omitempty"`

	// WorkflowGraph carries the graph definition when this job executes a workflow
	// rather than a plain single-agent task.
	WorkflowGraph *WorkflowGraphSpec `json:"workflow_graph,omitempty"`
}

// ToolConfig describes a tool the agent can invoke.
type ToolConfig struct {
	// Name used by the LLM when calling the tool (must be unique within the job).
	Name        string `json:"name"`
	Description string `json:"description"`

	// Type determines the execution strategy.
	// "http"  → generic HTTP call (uses HTTPConfig)
	// others  → named integrations (slack, github, linear, etc.) using Credentials
	Type string `json:"type"`

	// HTTPConfig is set when Type == "http".
	HTTPConfig *HTTPToolConfig `json:"http_config,omitempty"`

	// IntegrationConfig is set for integration-based tools.
	IntegrationConfig *IntegrationToolConfig `json:"integration_config,omitempty"`

	// Parameters is the JSON Schema for arguments the LLM must provide.
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// HTTPToolConfig configures a generic HTTP tool.
type HTTPToolConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"` // GET | POST | PUT | DELETE | PATCH
	Headers map[string]string `json:"headers,omitempty"`
	// Body template: use {{.param_name}} placeholders for LLM-supplied args.
	BodyTemplate string `json:"body_template,omitempty"`
}

// IntegrationToolConfig configures an integration-based tool.
type IntegrationToolConfig struct {
	Provider    string            `json:"provider"`    // slack | github | linear | …
	Action      string            `json:"action"`      // send_message | create_issue | …
	Credentials map[string]string `json:"credentials"` // decrypted at API level before publishing
}

// ChannelConfig describes where to deliver the agent's output.
type ChannelConfig struct {
	Type   string                 `json:"type"`   // slack | email | webhook
	Config map[string]interface{} `json:"config"` // provider-specific params
}

// ChannelTriggerMetadata carries the external channel identity that originated
// an execution. All fields are optional except Type so older producers remain
// compatible while newer producers can attach isolation metadata.
type ChannelTriggerMetadata struct {
	Type           string `json:"type"`
	EventID        string `json:"event_id,omitempty"`
	SenderID       string `json:"sender_id,omitempty"`
	AccountID      string `json:"account_id,omitempty"`
	MessageID      string `json:"message_id,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	ThreadTS       string `json:"thread_ts,omitempty"`
	SessionScope   string `json:"session_scope,omitempty"`
	SessionKey     string `json:"session_key,omitempty"`
}

// TriggerMetadata captures structured execution-trigger context.
type TriggerMetadata struct {
	Source        string                  `json:"source,omitempty"`
	TriggerRuleID string                  `json:"trigger_rule_id,omitempty"`
	Channel       *ChannelTriggerMetadata `json:"channel,omitempty"`
}

// CADConfig contains the agent definition for ACF-compatible agents.
type CADConfig struct {
	SystemPrompt string            `json:"system_prompt"`
	Skills       []SkillDef        `json:"skills,omitempty"`
	ContextFiles map[string]string `json:"context_files,omitempty"`
	SourceType   string            `json:"source_type"`
	SourceRef    string            `json:"source_ref,omitempty"`
	MCPServers   []MCPServerRef    `json:"mcp_servers,omitempty"`
}

// SkillDef represents a named skill bundled with an ACF agent.
type SkillDef struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// MCPServerRef is a reference to an MCP server an ACF agent may use.
type MCPServerRef struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Policy defines workspace-level tool restrictions for ACF agents.
type Policy struct {
	AllowedTools   []string `json:"allowed_tools,omitempty"`
	DeniedTools    []string `json:"denied_tools,omitempty"`
	MaxBashTimeout int      `json:"max_bash_timeout,omitempty"`
	AllowNetwork   bool     `json:"allow_network"`
	AllowWrite     bool     `json:"allow_write"`
}

// IsACFAgent returns true if this job uses the Agent Compatibility Framework.
func (s *Spec) IsACFAgent() bool {
	return s.CADConfig != nil
}

// ExecutionMode constants.
const (
	ModeOneShot   = "one_shot"
	ModeReactLoop = "react_loop"
	ModeWorkflow  = "workflow_graph"
)

// DefaultMaxIterations is the fallback when MaxIterations is 0.
const DefaultMaxIterations = 10

// IsShadow reports whether this job is a shadow execution.
func (s *Spec) IsShadow() bool {
	return s.ExecutionMode == ExecutionModeShadow
}

// Validate returns an error string if the spec is invalid, empty string if OK.
func (s *Spec) Validate() string {
	if s.TaskID == "" {
		return "missing task_id"
	}
	if s.ExecutionID == "" {
		return "missing execution_id"
	}
	if s.ModelProvider == "" {
		return "missing model_provider"
	}
	if s.ModelID == "" {
		return "missing model_id"
	}
	if s.Input == "" {
		return "missing input"
	}
	if s.Mode == "" {
		s.Mode = ModeOneShot
	}
	if s.MaxIterations <= 0 {
		s.MaxIterations = DefaultMaxIterations
	}

	// Default execution mode to normal.
	if s.ExecutionMode == "" {
		s.ExecutionMode = ExecutionModeNormal
	}

	// Shadow mode requires a valid ShadowConfig.
	if s.ExecutionMode == ExecutionModeShadow {
		if s.ShadowConfig == nil {
			return "shadow mode requires shadow_config"
		}
		if err := s.ShadowConfig.Validate(); err != nil {
			return err.Error()
		}
	}

	return ""
}

type WorkflowGraphSpec struct {
	StartStepID string             `json:"start_step_id,omitempty"`
	Steps       []WorkflowStepSpec `json:"steps,omitempty"`
}

type WorkflowStepSpec struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Config     map[string]interface{} `json:"config,omitempty"`
	NextStepID string                 `json:"next_step_id,omitempty"`
	Branches   []WorkflowBranchSpec   `json:"branches,omitempty"`
}

type WorkflowBranchSpec struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Condition  string `json:"condition"`
	NextStepID string `json:"next_step_id"`
	IsDefault  bool   `json:"is_default,omitempty"`
}
