package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/config"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/engine"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/job"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/pkg/telemetry"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Stubs for processJob dependency injection
// ---------------------------------------------------------------------------

// stubReporter implements jobReporter for unit tests.
type stubReporter struct {
	startedErr     error
	failedErr      error
	completedErr   error
	completedCalls int
	failedCalls    int
	lastCompleted  *engine.ExecutionResult
	workflowSteps  []string
}

func (s *stubReporter) ReportStarted(_ context.Context, _ job.Spec) error {
	return s.startedErr
}

func (s *stubReporter) ReportFailed(_ context.Context, _ job.Spec, _ string) error {
	s.failedCalls++
	return s.failedErr
}

func (s *stubReporter) ReportCompleted(_ context.Context, _ job.Spec, result *engine.ExecutionResult) error {
	s.completedCalls++
	s.lastCompleted = result
	return s.completedErr
}

func (s *stubReporter) ReportWorkflowStarted(_ context.Context, _ string, _ string, _ map[string]any) error {
	return s.startedErr
}

func (s *stubReporter) ReportWorkflowStep(_ context.Context, _ string, _ string, stepID string, _ string, _ string, _ map[string]any, _ map[string]any, _ string) error {
	s.workflowSteps = append(s.workflowSteps, stepID)
	return nil
}

func (s *stubReporter) ReportWorkflowCompleted(_ context.Context, _ string, _ string, _ map[string]any) error {
	s.completedCalls++
	return s.completedErr
}

func (s *stubReporter) ReportWorkflowFailed(_ context.Context, _ string, _ string, _ string, _ map[string]any) error {
	s.failedCalls++
	return s.failedErr
}

// stubLLMClient is an llm.Client that always returns a minimal successful response.
type stubLLMClient struct{}

func (s *stubLLMClient) Complete(_ context.Context, _ []llm.Message, _ []llm.ToolDefinition) (*llm.Response, error) {
	return &llm.Response{
		Content: "stub response",
		Usage: llm.Usage{
			PromptTokens:     100,
			CompletionTokens: 25,
			TotalTokens:      125,
		},
	}, nil
}

// errorLLMClient is an llm.Client that always returns a network error.
type errorLLMClient struct{}

func (e *errorLLMClient) Complete(_ context.Context, _ []llm.Message, _ []llm.ToolDefinition) (*llm.Response, error) {
	return nil, errors.New("network error: connection refused")
}

type jsonLLMClient struct{}

func (s *jsonLLMClient) Complete(_ context.Context, _ []llm.Message, _ []llm.ToolDefinition) (*llm.Response, error) {
	return &llm.Response{
		Content: `{"score":80,"confidence":0.91}`,
		Usage: llm.Usage{
			PromptTokens:     120,
			CompletionTokens: 30,
			TotalTokens:      150,
		},
	}, nil
}

// makeValidSpec builds a minimal valid job.Spec for tests.
func makeValidSpec() job.Spec {
	return job.Spec{
		TaskID:        "task-1",
		ExecutionID:   "exec-1",
		WorkspaceID:   "ws-1",
		AgentID:       "agent-1",
		ModelProvider: "openai",
		ModelID:       "gpt-4o",
		Mode:          job.ModeOneShot,
		Input:         "say hello",
		MaxIterations: 1,
	}
}

// makeACFSpec builds a valid job.Spec with a CADConfig (IsACFAgent == true).
func makeACFSpec() job.Spec {
	return job.Spec{
		TaskID:        "task-acf-1",
		ExecutionID:   "exec-acf-1",
		WorkspaceID:   "ws-1",
		AgentID:       "agent-acf-1",
		ModelProvider: "anthropic",
		ModelID:       "claude-3-5-sonnet",
		Mode:          job.ModeOneShot,
		Input:         "list files in the project",
		MaxIterations: 3,
		CADConfig: &job.CADConfig{
			SystemPrompt: "You are a helpful assistant.",
			SourceType:   "claude_code",
		},
	}
}

func makeWorkflowSpec() job.Spec {
	return job.Spec{
		TaskID:        "task-wf-1",
		ExecutionID:   "exec-wf-1",
		WorkspaceID:   "ws-1",
		AgentID:       "workflow-1",
		ModelProvider: "openai",
		ModelID:       "gpt-4o",
		Mode:          job.ModeWorkflow,
		Input:         "process this workflow",
		MaxIterations: 2,
		WorkflowGraph: &job.WorkflowGraphSpec{
			StartStepID: "step-transform",
			Steps: []job.WorkflowStepSpec{
				{ID: "step-transform", Name: "Transform", Type: "transform", NextStepID: "step-agent"},
				{ID: "step-agent", Name: "Agent Task", Type: "agent_task", NextStepID: "step-end"},
				{ID: "step-end", Name: "End", Type: "end"},
			},
		},
		BuilderMetadata: map[string]string{
			"workflow_id":              "workflow-1",
			"workflow_current_step_id": "step-transform",
		},
	}
}

func makeWorkflowBranchSpec() job.Spec {
	return job.Spec{
		TaskID:        "task-wf-branch",
		ExecutionID:   "exec-wf-branch",
		WorkspaceID:   "ws-1",
		AgentID:       "workflow-branch",
		ModelProvider: "openai",
		ModelID:       "gpt-4o",
		Mode:          job.ModeWorkflow,
		Input:         "branch this workflow",
		MaxIterations: 2,
		WorkflowGraph: &job.WorkflowGraphSpec{
			StartStepID: "step-agent",
			Steps: []job.WorkflowStepSpec{
				{ID: "step-agent", Name: "Agent Task", Type: "agent_task", NextStepID: "step-branch"},
				{
					ID:   "step-branch",
					Name: "Branch",
					Type: "branch",
					Branches: []job.WorkflowBranchSpec{
						{ID: "b1", Name: "high-score", Condition: "step-agent.score >= 70", NextStepID: "step-high"},
						{ID: "b2", Name: "default", Condition: "default", NextStepID: "step-low", IsDefault: true},
					},
				},
				{ID: "step-high", Name: "High Path", Type: "transform", NextStepID: "step-end"},
				{ID: "step-low", Name: "Low Path", Type: "transform", NextStepID: "step-end"},
				{ID: "step-end", Name: "End", Type: "end"},
			},
		},
		BuilderMetadata: map[string]string{
			"workflow_id": "workflow-branch",
		},
	}
}

// makeTestDeps returns the shared config, logger, and metrics used across tests.
func makeTestDeps(t *testing.T) (*config.Config, *zap.Logger, *telemetry.Metrics) {
	t.Helper()
	cfg := &config.Config{}
	logger := zap.NewNop()
	metrics, err := telemetry.InitMetrics()
	if err != nil {
		t.Fatalf("InitMetrics: %v", err)
	}
	return cfg, logger, metrics
}

func TestDefaultLLMFactory_Gateway(t *testing.T) {
	t.Setenv("SERVICE_TOKEN", "test-service-token")
	t.Setenv("LLM_GATEWAY_URL", "http://passflow-llm-gateway")

	client, err := defaultLLMFactory(context.Background(), "gateway", "auto", "test-workspace-id", &config.Config{
		APIURL:       "http://localhost:8080",
		ServiceToken: "test-token",
	})
	if err != nil {
		t.Fatalf("defaultLLMFactory returned error: %v", err)
	}
	if client == nil {
		t.Fatal("defaultLLMFactory returned nil client")
	}

	// Keep the envs referenced so the compiler does not optimize the setup away
	// in future refactors that gate on these values.
	if os.Getenv("SERVICE_TOKEN") == "" || os.Getenv("LLM_GATEWAY_URL") == "" {
		t.Fatal("gateway env configuration missing")
	}
}

// ---------------------------------------------------------------------------
// processJob ACK/NACK policy tests
// ---------------------------------------------------------------------------

// TestProcessJob_InvalidSpec_ACKs verifies that a malformed spec is ACKed
// (returns true) to prevent an infinite retry loop on a permanent error.
func TestProcessJob_InvalidSpec_ACKs(t *testing.T) {
	cfg, logger, metrics := makeTestDeps(t)
	rep := &stubReporter{}

	badSpec := job.Spec{} // missing required fields → Validate() returns a message

	got := processJob(context.Background(), cfg, rep, badSpec, logger, metrics,
		func(_ context.Context, _, _, _ string, _ *config.Config) (llm.Client, error) {
			t.Error("factory should not be called for an invalid spec")
			return nil, nil
		}, nil)

	if !got {
		t.Error("processJob with invalid spec must return true (ACK), got false (NACK)")
	}
	if rep.failedCalls != 1 {
		t.Errorf("expected ReportFailed to be called once, got %d", rep.failedCalls)
	}
}

// TestProcessJob_LLMFactoryFailure_NACKs verifies that an LLM client creation
// failure returns false (NACK) so the job stays in the PEL for re-delivery.
// Factory failures are transient: misconfigured credentials may be fixed by an
// operator without restarting the stream.
func TestProcessJob_LLMFactoryFailure_NACKs(t *testing.T) {
	cfg, logger, metrics := makeTestDeps(t)
	rep := &stubReporter{}

	got := processJob(context.Background(), cfg, rep, makeValidSpec(), logger, metrics,
		func(_ context.Context, _, _, _ string, _ *config.Config) (llm.Client, error) {
			return nil, errors.New("credentials not found: secret missing")
		}, nil)

	if got {
		t.Error("processJob with LLM factory failure must return false (NACK), got true (ACK)")
	}
}

// TestProcessJob_ReportCompletedNetworkError_NACKs verifies that a network
// error when reporting completion returns false so the result can be retried.
func TestProcessJob_ReportCompletedNetworkError_NACKs(t *testing.T) {
	cfg, logger, metrics := makeTestDeps(t)
	rep := &stubReporter{
		completedErr: errors.New("network error: dial tcp: connect: connection refused"),
	}

	got := processJob(context.Background(), cfg, rep, makeValidSpec(), logger, metrics,
		func(_ context.Context, _, _, _ string, _ *config.Config) (llm.Client, error) {
			return &stubLLMClient{}, nil
		}, nil)

	if got {
		t.Error("processJob with ReportCompleted network error must return false (NACK), got true (ACK)")
	}
}

// TestProcessJob_Success_ACKs verifies the happy path: successful execution
// and successful report both return true (ACK).
func TestProcessJob_Success_ACKs(t *testing.T) {
	cfg, logger, metrics := makeTestDeps(t)
	rep := &stubReporter{}

	got := processJob(context.Background(), cfg, rep, makeValidSpec(), logger, metrics,
		func(_ context.Context, _, _, _ string, _ *config.Config) (llm.Client, error) {
			return &stubLLMClient{}, nil
		}, nil)

	if !got {
		t.Error("processJob with successful run must return true (ACK), got false (NACK)")
	}
	if rep.completedCalls != 1 {
		t.Errorf("expected ReportCompleted to be called once, got %d", rep.completedCalls)
	}
	if rep.lastCompleted == nil {
		t.Fatal("expected completed result to be captured")
	}
	if rep.lastCompleted.Usage.LLMCalls < 1 {
		t.Errorf("expected at least 1 llm call, got %d", rep.lastCompleted.Usage.LLMCalls)
	}
	if rep.lastCompleted.Metadata["model_provider"] != "openai" {
		t.Errorf("expected model_provider metadata to be set, got %q", rep.lastCompleted.Metadata["model_provider"])
	}
	if rep.lastCompleted.Metadata["estimated_cost_usd"] == "" {
		t.Error("expected estimated_cost_usd metadata to be set")
	}
}

// TestProcessJob_EngineError_ACKs verifies that engine-level LLM failures
// (where the engine handles the error internally and returns a "failed" result
// rather than a Go error) still produce an ACK. The engine wraps the failure
// in ExecutionResult.Status="failed" and calls ReportCompleted — not
// ReportFailed — so the result is authoritatively stored in passflow-api.
// Re-delivering the job would just run the same failing LLM call again.
func TestProcessJob_EngineError_ACKs(t *testing.T) {
	cfg, logger, metrics := makeTestDeps(t)
	rep := &stubReporter{}

	// errorLLMClient causes the engine to produce a failed ExecutionResult.
	// The engine returns (failedResult, nil) — engineErr is nil — so
	// processJob goes through the ReportCompleted path.
	got := processJob(context.Background(), cfg, rep, makeValidSpec(), logger, metrics,
		func(_ context.Context, _, _, _ string, _ *config.Config) (llm.Client, error) {
			return &errorLLMClient{}, nil
		}, nil)

	if !got {
		t.Error("processJob with engine-internal LLM error must return true (ACK), got false (NACK)")
	}
	if rep.completedCalls != 1 {
		t.Errorf("expected ReportCompleted to be called once (with failed result), got %d", rep.completedCalls)
	}
}

// ---------------------------------------------------------------------------
// TestRedactURL_WithPassword
// ---------------------------------------------------------------------------

func TestRedactURL_WithPassword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantFull bool   // true = expect the exact full URL back (no creds to redact)
		mustHave string // substring that must appear in the output
		mustNot  string // substring that must NOT appear in the output
	}{
		{
			name:     "redis URL with password",
			input:    "redis://:s3cr3t@localhost:6379/0",
			mustHave: "[REDACTED]",
			mustNot:  "s3cr3t",
		},
		{
			name:     "redis URL with user and password",
			input:    "redis://user:p@ssw0rd@redis.example.com:6379",
			mustHave: "[REDACTED]",
			mustNot:  "p@ssw0rd",
		},
		{
			name:     "redis URL without credentials",
			input:    "redis://localhost:6379",
			wantFull: true,
			mustHave: "redis://localhost:6379",
		},
		{
			name:     "empty string returns empty",
			input:    "",
			wantFull: true,
			mustHave: "",
		},
		{
			name:     "unparseable URL returns [REDACTED]",
			input:    "://this is not a url",
			mustHave: "[REDACTED]",
		},
		{
			name:     "URL with password preserves host and port",
			input:    "redis://:hunter2@myredis.prod:6380/1",
			mustHave: "myredis.prod:6380",
			mustNot:  "hunter2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactURL(tt.input)

			if tt.wantFull && got != tt.input {
				t.Errorf("redactURL(%q) = %q, want original %q", tt.input, got, tt.input)
			}
			if tt.mustHave != "" && len(got) > 0 {
				found := false
				for i := 0; i <= len(got)-len(tt.mustHave); i++ {
					if got[i:i+len(tt.mustHave)] == tt.mustHave {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("redactURL(%q) = %q, must contain %q", tt.input, got, tt.mustHave)
				}
			}
			if tt.mustNot != "" {
				for i := 0; i <= len(got)-len(tt.mustNot); i++ {
					if got[i:i+len(tt.mustNot)] == tt.mustNot {
						t.Errorf("redactURL(%q) = %q, must NOT contain %q", tt.input, got, tt.mustNot)
						break
					}
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// processJob ACF integration tests
// ---------------------------------------------------------------------------

// TestProcessJob_ACFSpec_Success_ACKs verifies that processJob correctly
// prepares the ACF context for a spec with CADConfig set and completes
// without error. The stub LLM returns an immediate text response so no tools
// are invoked.
func TestProcessJob_ACFSpec_Success_ACKs(t *testing.T) {
	cfg, logger, metrics := makeTestDeps(t)
	rep := &stubReporter{}

	got := processJob(context.Background(), cfg, rep, makeACFSpec(), logger, metrics,
		func(_ context.Context, _, _, _ string, _ *config.Config) (llm.Client, error) {
			return &stubLLMClient{}, nil
		}, nil)

	if !got {
		t.Error("processJob with ACF spec must return true (ACK), got false (NACK)")
	}
	if rep.completedCalls != 1 {
		t.Errorf("expected ReportCompleted to be called once, got %d", rep.completedCalls)
	}
}

// TestProcessJob_ACFSpec_LLMFactoryFailure_NACKs verifies that an ACF spec
// follows the same NACK policy on LLM factory failure as a non-ACF spec.
func TestProcessJob_ACFSpec_LLMFactoryFailure_NACKs(t *testing.T) {
	cfg, logger, metrics := makeTestDeps(t)
	rep := &stubReporter{}

	got := processJob(context.Background(), cfg, rep, makeACFSpec(), logger, metrics,
		func(_ context.Context, _, _, _ string, _ *config.Config) (llm.Client, error) {
			return nil, errors.New("credentials not found")
		}, nil)

	if got {
		t.Error("processJob with ACF spec + LLM factory failure must return false (NACK), got true (ACK)")
	}
}

func TestProcessJob_WorkflowGraph_ACKs(t *testing.T) {
	cfg, logger, metrics := makeTestDeps(t)
	rep := &stubReporter{}

	got := processJob(context.Background(), cfg, rep, makeWorkflowSpec(), logger, metrics,
		func(_ context.Context, _, _, _ string, _ *config.Config) (llm.Client, error) {
			return &stubLLMClient{}, nil
		}, nil)

	if !got {
		t.Error("processJob with workflow_graph spec must return true (ACK), got false (NACK)")
	}
	if rep.completedCalls == 0 {
		t.Error("expected workflow completion to be reported")
	}
}

func TestProcessJob_WorkflowGraph_BranchSelection(t *testing.T) {
	cfg, logger, metrics := makeTestDeps(t)
	rep := &stubReporter{}

	got := processJob(context.Background(), cfg, rep, makeWorkflowBranchSpec(), logger, metrics,
		func(_ context.Context, _, _, _ string, _ *config.Config) (llm.Client, error) {
			return &jsonLLMClient{}, nil
		}, nil)

	if !got {
		t.Error("processJob with workflow branch spec must return true (ACK), got false (NACK)")
	}
	joined := strings.Join(rep.workflowSteps, ",")
	if !strings.Contains(joined, "step-high") {
		t.Fatalf("expected workflow to take high-score branch, got steps %s", joined)
	}
	if strings.Contains(joined, "step-low") {
		t.Fatalf("did not expect workflow to take low-score branch, got steps %s", joined)
	}
}
