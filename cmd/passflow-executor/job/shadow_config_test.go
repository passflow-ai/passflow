package job

import (
	"testing"
)

func TestShadowConfig_Validate_Valid(t *testing.T) {
	sc := &ShadowConfig{
		OriginalExecutionID: "exec-orig-1",
		AnalysisBatchID:     "batch-1",
		MockedToolOutputs:   map[string]string{"send_email": `{"ok":true}`},
		ReplayPolicy:        "strict",
		TraceID:             "trace-123",
	}

	if err := sc.Validate(); err != nil {
		t.Errorf("expected valid ShadowConfig, got error: %v", err)
	}
}

func TestShadowConfig_Validate_BestEffort(t *testing.T) {
	sc := &ShadowConfig{
		OriginalExecutionID: "exec-orig-1",
		AnalysisBatchID:     "batch-1",
		MockedToolOutputs:   map[string]string{},
		ReplayPolicy:        "best_effort",
		TraceID:             "trace-123",
	}

	if err := sc.Validate(); err != nil {
		t.Errorf("expected valid ShadowConfig with best_effort policy, got error: %v", err)
	}
}

func TestShadowConfig_Validate_MissingOriginalExecutionID(t *testing.T) {
	sc := &ShadowConfig{
		AnalysisBatchID:   "batch-1",
		MockedToolOutputs: map[string]string{},
		ReplayPolicy:      "strict",
		TraceID:           "trace-123",
	}

	if err := sc.Validate(); err == nil {
		t.Error("expected error for missing original_execution_id, got nil")
	}
}

func TestShadowConfig_Validate_MissingAnalysisBatchID(t *testing.T) {
	sc := &ShadowConfig{
		OriginalExecutionID: "exec-orig-1",
		MockedToolOutputs:   map[string]string{},
		ReplayPolicy:        "strict",
		TraceID:             "trace-123",
	}

	if err := sc.Validate(); err == nil {
		t.Error("expected error for missing analysis_batch_id, got nil")
	}
}

func TestShadowConfig_Validate_NilMockedToolOutputs(t *testing.T) {
	sc := &ShadowConfig{
		OriginalExecutionID: "exec-orig-1",
		AnalysisBatchID:     "batch-1",
		MockedToolOutputs:   nil,
		ReplayPolicy:        "strict",
		TraceID:             "trace-123",
	}

	if err := sc.Validate(); err == nil {
		t.Error("expected error for nil mocked_tool_outputs, got nil")
	}
}

func TestShadowConfig_Validate_InvalidReplayPolicy(t *testing.T) {
	sc := &ShadowConfig{
		OriginalExecutionID: "exec-orig-1",
		AnalysisBatchID:     "batch-1",
		MockedToolOutputs:   map[string]string{},
		ReplayPolicy:        "invalid",
		TraceID:             "trace-123",
	}

	if err := sc.Validate(); err == nil {
		t.Error("expected error for invalid replay_policy, got nil")
	}
}

func TestShadowConfig_Validate_MissingTraceID(t *testing.T) {
	sc := &ShadowConfig{
		OriginalExecutionID: "exec-orig-1",
		AnalysisBatchID:     "batch-1",
		MockedToolOutputs:   map[string]string{},
		ReplayPolicy:        "strict",
	}

	if err := sc.Validate(); err == nil {
		t.Error("expected error for missing trace_id, got nil")
	}
}

func TestSpec_IsShadow(t *testing.T) {
	s := &Spec{ExecutionMode: ExecutionModeShadow}
	if !s.IsShadow() {
		t.Error("expected IsShadow() == true for shadow mode")
	}

	s.ExecutionMode = ExecutionModeNormal
	if s.IsShadow() {
		t.Error("expected IsShadow() == false for normal mode")
	}

	s.ExecutionMode = ""
	if s.IsShadow() {
		t.Error("expected IsShadow() == false for empty mode")
	}
}

func TestSpec_Validate_ShadowMode_RequiresShadowConfig(t *testing.T) {
	s := &Spec{
		TaskID:        "task-1",
		ExecutionID:   "exec-1",
		ModelProvider: "openai",
		ModelID:       "gpt-4o",
		Input:         "hello",
		ExecutionMode: ExecutionModeShadow,
	}

	if msg := s.Validate(); msg == "" {
		t.Error("expected validation error for shadow mode without shadow_config")
	}
}

func TestSpec_Validate_ShadowMode_WithValidConfig(t *testing.T) {
	s := &Spec{
		TaskID:        "task-1",
		ExecutionID:   "exec-1",
		ModelProvider: "openai",
		ModelID:       "gpt-4o",
		Input:         "hello",
		ExecutionMode: ExecutionModeShadow,
		ShadowConfig: &ShadowConfig{
			OriginalExecutionID: "orig-1",
			AnalysisBatchID:     "batch-1",
			MockedToolOutputs:   map[string]string{},
			ReplayPolicy:        "strict",
			TraceID:             "trace-1",
		},
	}

	if msg := s.Validate(); msg != "" {
		t.Errorf("expected valid spec, got error: %s", msg)
	}
}

func TestSpec_Validate_DefaultsExecutionModeToNormal(t *testing.T) {
	s := &Spec{
		TaskID:        "task-1",
		ExecutionID:   "exec-1",
		ModelProvider: "openai",
		ModelID:       "gpt-4o",
		Input:         "hello",
	}

	s.Validate()
	if s.ExecutionMode != ExecutionModeNormal {
		t.Errorf("expected execution_mode to default to %q, got %q", ExecutionModeNormal, s.ExecutionMode)
	}
}
