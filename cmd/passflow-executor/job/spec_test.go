package job

import (
	"encoding/json"
	"strings"
	"testing"
)

// --- HIGH 4: Credentials exposed in job spec ---

func TestRedactCredentialsReplacesValues(t *testing.T) {
	spec := &Spec{
		TaskID:      "task-1",
		ExecutionID: "exec-1",
		Tools: []ToolConfig{
			{
				Name: "slack",
				Type: "integration",
				IntegrationConfig: &IntegrationToolConfig{
					Provider: "slack",
					Action:   "send_message",
					Credentials: map[string]string{
						"token":          "xoxb-real-secret-token",
						"signing_secret": "abc123secret",
					},
				},
			},
			{
				Name:       "fetch",
				Type:       "http",
				HTTPConfig: &HTTPToolConfig{URL: "https://example.com"},
			},
		},
	}

	redacted := spec.RedactCredentials()

	// Original must be untouched.
	for _, tool := range spec.Tools {
		if tool.IntegrationConfig == nil {
			continue
		}
		for k, v := range tool.IntegrationConfig.Credentials {
			if v == "[REDACTED]" {
				t.Errorf("original spec was mutated: tool %q credential %q is now [REDACTED]", tool.Name, k)
			}
		}
	}

	// Redacted copy must have no plaintext credentials.
	for _, tool := range redacted.Tools {
		if tool.IntegrationConfig == nil {
			continue
		}
		for k, v := range tool.IntegrationConfig.Credentials {
			if v != "[REDACTED]" {
				t.Errorf("credential %q in redacted spec is not redacted: got %q", k, v)
			}
		}
	}
}

func TestRedactCredentialsHandlesNilIntegrationConfig(t *testing.T) {
	spec := &Spec{
		TaskID:      "task-1",
		ExecutionID: "exec-1",
		Tools: []ToolConfig{
			{
				Name:       "fetch",
				Type:       "http",
				HTTPConfig: &HTTPToolConfig{URL: "https://example.com"},
			},
		},
	}

	// Must not panic.
	redacted := spec.RedactCredentials()
	if len(redacted.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(redacted.Tools))
	}
}

func TestRedactCredentialsEmptyTools(t *testing.T) {
	spec := &Spec{TaskID: "t", ExecutionID: "e"}
	redacted := spec.RedactCredentials()
	if len(redacted.Tools) != 0 {
		t.Errorf("expected 0 tools in redacted spec, got %d", len(redacted.Tools))
	}
}

func TestSensitiveFieldsReturnsCredentialPaths(t *testing.T) {
	spec := &Spec{
		Tools: []ToolConfig{
			{
				Name: "slack",
				Type: "integration",
				IntegrationConfig: &IntegrationToolConfig{
					Provider:    "slack",
					Credentials: map[string]string{"token": "secret"},
				},
			},
		},
	}

	fields := spec.SensitiveFields()
	if len(fields) == 0 {
		t.Fatal("SensitiveFields should return at least one path for an integration tool with credentials")
	}

	found := false
	for _, f := range fields {
		if strings.Contains(f, "credentials") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("SensitiveFields %v should contain a path with 'credentials'", fields)
	}
}

func TestSensitiveFieldsEmptyWhenNoIntegrations(t *testing.T) {
	spec := &Spec{
		Tools: []ToolConfig{
			{
				Name:       "fetch",
				Type:       "http",
				HTTPConfig: &HTTPToolConfig{URL: "https://example.com"},
			},
		},
	}

	fields := spec.SensitiveFields()
	if len(fields) != 0 {
		t.Errorf("expected no sensitive fields for HTTP-only tools, got %v", fields)
	}
}

func TestMultipleIntegrationToolsAllRedacted(t *testing.T) {
	spec := &Spec{
		Tools: []ToolConfig{
			{
				Name: "slack",
				Type: "integration",
				IntegrationConfig: &IntegrationToolConfig{
					Provider:    "slack",
					Credentials: map[string]string{"token": "slack-secret"},
				},
			},
			{
				Name: "github",
				Type: "integration",
				IntegrationConfig: &IntegrationToolConfig{
					Provider:    "github",
					Credentials: map[string]string{"access_token": "ghp_secret"},
				},
			},
		},
	}

	redacted := spec.RedactCredentials()
	for _, tool := range redacted.Tools {
		if tool.IntegrationConfig == nil {
			continue
		}
		for k, v := range tool.IntegrationConfig.Credentials {
			if v != "[REDACTED]" {
				t.Errorf("tool %q credential %q not redacted", tool.Name, k)
			}
		}
	}
}

// --- ACF: IsACFAgent ---

func TestIsACFAgentReturnsFalseWithoutCADConfig(t *testing.T) {
	spec := &Spec{
		TaskID:        "task-1",
		ExecutionID:   "exec-1",
		ModelProvider: "openai",
		ModelID:       "gpt-4o",
		Input:         "do something",
	}
	if spec.IsACFAgent() {
		t.Error("IsACFAgent() should return false when CADConfig is nil")
	}
}

func TestIsACFAgentReturnsTrueWithCADConfig(t *testing.T) {
	spec := &Spec{
		TaskID:        "task-2",
		ExecutionID:   "exec-2",
		ModelProvider: "anthropic",
		ModelID:       "claude-3-5-sonnet",
		Input:         "do something",
		CADConfig: &CADConfig{
			SystemPrompt: "You are a helpful agent.",
			SourceType:   "git",
			SourceRef:    "main",
		},
	}
	if !spec.IsACFAgent() {
		t.Error("IsACFAgent() should return true when CADConfig is set")
	}
}

// --- ACF: JSON serialization round-trip ---

func TestSpecWithCADConfigJSONRoundTrip(t *testing.T) {
	original := &Spec{
		TaskID:        "task-3",
		ExecutionID:   "exec-3",
		WorkspaceID:   "ws-1",
		AgentID:       "agent-1",
		ModelProvider: "anthropic",
		ModelID:       "claude-3-5-sonnet",
		Input:         "analyze this repo",
		Mode:          ModeReactLoop,
		MaxIterations: 5,
		CADConfig: &CADConfig{
			SystemPrompt: "You are a code analyst.",
			SourceType:   "git",
			SourceRef:    "refs/heads/main",
			Skills: []SkillDef{
				{Name: "summarize", Content: "Summarize the given code."},
			},
			ContextFiles: map[string]string{
				"README.md": "# My Project",
			},
			MCPServers: []MCPServerRef{
				{Name: "mcp-primary", URL: "http://mcp-gateway:8080"},
			},
		},
		Policy: &Policy{
			AllowedTools:   []string{"bash", "read_file"},
			DeniedTools:    []string{"delete_file"},
			MaxBashTimeout: 30,
			AllowNetwork:   true,
			AllowWrite:     false,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded Spec
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.CADConfig == nil {
		t.Fatal("decoded CADConfig is nil")
	}
	if decoded.CADConfig.SystemPrompt != original.CADConfig.SystemPrompt {
		t.Errorf("SystemPrompt mismatch: got %q, want %q", decoded.CADConfig.SystemPrompt, original.CADConfig.SystemPrompt)
	}
	if decoded.CADConfig.SourceType != original.CADConfig.SourceType {
		t.Errorf("SourceType mismatch: got %q, want %q", decoded.CADConfig.SourceType, original.CADConfig.SourceType)
	}
	if len(decoded.CADConfig.Skills) != 1 || decoded.CADConfig.Skills[0].Name != "summarize" {
		t.Errorf("Skills mismatch: got %+v", decoded.CADConfig.Skills)
	}
	if decoded.CADConfig.ContextFiles["README.md"] != "# My Project" {
		t.Errorf("ContextFiles mismatch: got %+v", decoded.CADConfig.ContextFiles)
	}
	if len(decoded.CADConfig.MCPServers) != 1 || decoded.CADConfig.MCPServers[0].URL != "http://mcp-gateway:8080" {
		t.Errorf("MCPServers mismatch: got %+v", decoded.CADConfig.MCPServers)
	}

	if decoded.Policy == nil {
		t.Fatal("decoded Policy is nil")
	}
	if len(decoded.Policy.AllowedTools) != 2 {
		t.Errorf("AllowedTools mismatch: got %v", decoded.Policy.AllowedTools)
	}
	if len(decoded.Policy.DeniedTools) != 1 || decoded.Policy.DeniedTools[0] != "delete_file" {
		t.Errorf("DeniedTools mismatch: got %v", decoded.Policy.DeniedTools)
	}
	if decoded.Policy.MaxBashTimeout != 30 {
		t.Errorf("MaxBashTimeout mismatch: got %d", decoded.Policy.MaxBashTimeout)
	}
	if !decoded.Policy.AllowNetwork {
		t.Error("AllowNetwork should be true")
	}
	if decoded.Policy.AllowWrite {
		t.Error("AllowWrite should be false")
	}
}

func TestSpecWithoutCADConfigOmitsACFFieldsInJSON(t *testing.T) {
	spec := &Spec{
		TaskID:        "task-4",
		ExecutionID:   "exec-4",
		ModelProvider: "openai",
		ModelID:       "gpt-4o",
		Input:         "hello",
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	raw := string(data)
	if strings.Contains(raw, "cad_config") {
		t.Errorf("cad_config should be omitted from JSON for non-ACF specs, got: %s", raw)
	}
	if strings.Contains(raw, "policy") {
		t.Errorf("policy should be omitted from JSON for non-ACF specs, got: %s", raw)
	}
}

// --- ACF: Validate regression ---

func TestValidateStillWorksForACFSpec(t *testing.T) {
	spec := &Spec{
		TaskID:        "task-5",
		ExecutionID:   "exec-5",
		ModelProvider: "anthropic",
		ModelID:       "claude-3-5-sonnet",
		Input:         "run analysis",
		CADConfig: &CADConfig{
			SystemPrompt: "You are a coding agent.",
			SourceType:   "inline",
		},
		Policy: &Policy{
			AllowNetwork: false,
			AllowWrite:   false,
		},
	}

	if msg := spec.Validate(); msg != "" {
		t.Errorf("Validate() returned error for valid ACF spec: %q", msg)
	}
	// Validate() should set defaults.
	if spec.Mode != ModeOneShot {
		t.Errorf("expected Mode %q, got %q", ModeOneShot, spec.Mode)
	}
	if spec.MaxIterations != DefaultMaxIterations {
		t.Errorf("expected MaxIterations %d, got %d", DefaultMaxIterations, spec.MaxIterations)
	}
}

func TestValidateMissingFieldsStillFailsForACFSpec(t *testing.T) {
	cases := []struct {
		name string
		spec *Spec
		want string
	}{
		{
			name: "missing task_id",
			spec: &Spec{
				ExecutionID: "exec-1", ModelProvider: "openai", ModelID: "gpt-4o", Input: "x",
				CADConfig: &CADConfig{SourceType: "git"},
			},
			want: "missing task_id",
		},
		{
			name: "missing model_provider",
			spec: &Spec{
				TaskID: "t", ExecutionID: "e", ModelID: "gpt-4o", Input: "x",
				CADConfig: &CADConfig{SourceType: "git"},
			},
			want: "missing model_provider",
		},
		{
			name: "missing input",
			spec: &Spec{
				TaskID: "t", ExecutionID: "e", ModelProvider: "openai", ModelID: "gpt-4o",
				CADConfig: &CADConfig{SourceType: "git"},
			},
			want: "missing input",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.spec.Validate()
			if got != tc.want {
				t.Errorf("Validate() = %q, want %q", got, tc.want)
			}
		})
	}
}
