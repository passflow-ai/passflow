package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShadowConfig_Validate_Valid(t *testing.T) {
	tests := []struct {
		name   string
		config ShadowConfig
	}{
		{
			name: "minimal valid config with strict policy",
			config: ShadowConfig{
				OriginalExecutionID: "exec-001",
				AnalysisBatchID:     "batch-001",
				ReplayPolicy:        ReplayPolicyStrict,
				TraceID:             "trace-001",
			},
		},
		{
			name: "full config with all optional fields",
			config: ShadowConfig{
				OriginalExecutionID:      "exec-002",
				AnalysisBatchID:          "batch-002",
				InjectedSystemPrompt:     "You are a helpful assistant",
				InjectedToolDescriptions: map[string]string{"search": "Search the web"},
				MockedToolOutputs:        map[string]string{"search": `{"results": []}`},
				ReplayPolicy:             ReplayPolicyBestEffort,
				TraceID:                  "trace-002",
			},
		},
		{
			name: "config with best_effort replay policy",
			config: ShadowConfig{
				OriginalExecutionID: "exec-003",
				AnalysisBatchID:     "batch-003",
				ReplayPolicy:        ReplayPolicyBestEffort,
				TraceID:             "trace-003",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestShadowConfig_Validate_MissingRequired(t *testing.T) {
	tests := []struct {
		name        string
		config      ShadowConfig
		expectedErr error
	}{
		{
			name: "missing original_execution_id",
			config: ShadowConfig{
				AnalysisBatchID: "batch-001",
				ReplayPolicy:    ReplayPolicyStrict,
				TraceID:         "trace-001",
			},
			expectedErr: ErrEmptyOriginalExecutionID,
		},
		{
			name: "missing analysis_batch_id",
			config: ShadowConfig{
				OriginalExecutionID: "exec-001",
				ReplayPolicy:        ReplayPolicyStrict,
				TraceID:             "trace-001",
			},
			expectedErr: ErrEmptyAnalysisBatchID,
		},
		{
			name: "missing replay_policy",
			config: ShadowConfig{
				OriginalExecutionID: "exec-001",
				AnalysisBatchID:     "batch-001",
				TraceID:             "trace-001",
			},
			expectedErr: ErrInvalidReplayPolicy,
		},
		{
			name: "invalid replay_policy",
			config: ShadowConfig{
				OriginalExecutionID: "exec-001",
				AnalysisBatchID:     "batch-001",
				ReplayPolicy:        "invalid_policy",
				TraceID:             "trace-001",
			},
			expectedErr: ErrInvalidReplayPolicy,
		},
		{
			name: "missing trace_id",
			config: ShadowConfig{
				OriginalExecutionID: "exec-001",
				AnalysisBatchID:     "batch-001",
				ReplayPolicy:        ReplayPolicyStrict,
			},
			expectedErr: ErrEmptyShadowTraceID,
		},
		{
			name:        "all fields empty",
			config:      ShadowConfig{},
			expectedErr: ErrEmptyOriginalExecutionID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestShadowResult_Validate_Valid(t *testing.T) {
	tests := []struct {
		name   string
		result ShadowResult
	}{
		{
			name: "completed result",
			result: ShadowResult{
				SchemaVersion:       "1.0",
				ShadowExecutionID:   "shadow-exec-001",
				OriginalExecutionID: "exec-001",
				AnalysisBatchID:     "batch-001",
				TenantID:            "tenant-001",
				AgentID:             "agent-001",
				AgentVersion:        "1.0.0",
				TraceID:             "trace-001",
				IdempotencyKey:      "idem-001",
				Status:              ShadowStatusCompleted,
				Output:              "Task completed successfully",
				Metrics: ShadowResultMetrics{
					TotalTokens:           1500,
					PromptTokens:          1000,
					CompletionTokens:      500,
					EstimatedCostUSD:      0.015,
					TotalDurationMs:       3200,
					Iterations:            3,
					ToolCallsCount:        5,
					ToolErrorsCount:       0,
					MockedToolsCount:      2,
					PassthroughToolsCount: 3,
				},
				Steps:     []json.RawMessage{},
				CreatedAt: time.Now().UTC(),
			},
		},
		{
			name: "failed result with error",
			result: ShadowResult{
				SchemaVersion:       "1.0",
				ShadowExecutionID:   "shadow-exec-002",
				OriginalExecutionID: "exec-002",
				AnalysisBatchID:     "batch-002",
				TenantID:            "tenant-002",
				AgentID:             "agent-002",
				AgentVersion:        "2.1.0",
				TraceID:             "trace-002",
				IdempotencyKey:      "idem-002",
				Status:              ShadowStatusFailed,
				Output:              "",
				Error:               "LLM provider timeout",
				Metrics:             ShadowResultMetrics{},
				Steps:               []json.RawMessage{},
				CreatedAt:           time.Now().UTC(),
			},
		},
		{
			name: "aborted result",
			result: ShadowResult{
				SchemaVersion:       "1.0",
				ShadowExecutionID:   "shadow-exec-003",
				OriginalExecutionID: "exec-003",
				AnalysisBatchID:     "batch-003",
				TenantID:            "tenant-003",
				AgentID:             "agent-003",
				AgentVersion:        "1.0.0",
				TraceID:             "trace-003",
				IdempotencyKey:      "idem-003",
				Status:              ShadowStatusAborted,
				Error:               "Replay policy strict: tool mismatch at step 3",
				Metrics:             ShadowResultMetrics{},
				Steps:               []json.RawMessage{},
				CreatedAt:           time.Now().UTC(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.result.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestShadowResult_Validate_InvalidStatus(t *testing.T) {
	base := ShadowResult{
		SchemaVersion:       "1.0",
		ShadowExecutionID:   "shadow-exec-001",
		OriginalExecutionID: "exec-001",
		AnalysisBatchID:     "batch-001",
		TenantID:            "tenant-001",
		AgentID:             "agent-001",
		AgentVersion:        "1.0.0",
		TraceID:             "trace-001",
		IdempotencyKey:      "idem-001",
		Metrics:             ShadowResultMetrics{},
		Steps:               []json.RawMessage{},
		CreatedAt:           time.Now().UTC(),
	}

	tests := []struct {
		name   string
		status string
	}{
		{"empty status", ""},
		{"unknown status", "unknown"},
		{"running is not valid", "running"},
		{"pending is not valid", "pending"},
		{"cancelled typo", "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := base
			r.Status = tt.status
			err := r.Validate()
			assert.Equal(t, ErrInvalidShadowResultStatus, err)
		})
	}
}

func TestShadowResult_Validate_MissingRequiredFields(t *testing.T) {
	validResult := func() ShadowResult {
		return ShadowResult{
			SchemaVersion:       "1.0",
			ShadowExecutionID:   "shadow-exec-001",
			OriginalExecutionID: "exec-001",
			AnalysisBatchID:     "batch-001",
			TenantID:            "tenant-001",
			AgentID:             "agent-001",
			AgentVersion:        "1.0.0",
			TraceID:             "trace-001",
			IdempotencyKey:      "idem-001",
			Status:              ShadowStatusCompleted,
			Metrics:             ShadowResultMetrics{},
			Steps:               []json.RawMessage{},
			CreatedAt:           time.Now().UTC(),
		}
	}

	tests := []struct {
		name        string
		modify      func(*ShadowResult)
		expectedErr error
	}{
		{
			name:        "missing schema_version",
			modify:      func(r *ShadowResult) { r.SchemaVersion = "" },
			expectedErr: ErrEmptyResultSchemaVersion,
		},
		{
			name:        "missing shadow_execution_id",
			modify:      func(r *ShadowResult) { r.ShadowExecutionID = "" },
			expectedErr: ErrEmptyShadowExecutionID,
		},
		{
			name:        "missing original_execution_id",
			modify:      func(r *ShadowResult) { r.OriginalExecutionID = "" },
			expectedErr: ErrEmptyResultOriginalExecID,
		},
		{
			name:        "missing analysis_batch_id",
			modify:      func(r *ShadowResult) { r.AnalysisBatchID = "" },
			expectedErr: ErrEmptyResultAnalysisBatchID,
		},
		{
			name:        "missing tenant_id",
			modify:      func(r *ShadowResult) { r.TenantID = "" },
			expectedErr: ErrEmptyResultTenantID,
		},
		{
			name:        "missing agent_id",
			modify:      func(r *ShadowResult) { r.AgentID = "" },
			expectedErr: ErrEmptyResultAgentID,
		},
		{
			name:        "missing agent_version",
			modify:      func(r *ShadowResult) { r.AgentVersion = "" },
			expectedErr: ErrEmptyResultAgentVersion,
		},
		{
			name:        "missing trace_id",
			modify:      func(r *ShadowResult) { r.TraceID = "" },
			expectedErr: ErrEmptyResultTraceID,
		},
		{
			name:        "missing idempotency_key",
			modify:      func(r *ShadowResult) { r.IdempotencyKey = "" },
			expectedErr: ErrEmptyResultIdempotencyKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := validResult()
			tt.modify(&r)
			err := r.Validate()
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestShadowConfig_JSONRoundTrip(t *testing.T) {
	config := ShadowConfig{
		OriginalExecutionID:      "exec-001",
		AnalysisBatchID:          "batch-001",
		InjectedSystemPrompt:     "You are a test assistant",
		InjectedToolDescriptions: map[string]string{"search": "Search tool"},
		MockedToolOutputs:        map[string]string{"search": `{"results":[]}`},
		ReplayPolicy:             ReplayPolicyStrict,
		TraceID:                  "trace-001",
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)

	var decoded ShadowConfig
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, config.OriginalExecutionID, decoded.OriginalExecutionID)
	assert.Equal(t, config.AnalysisBatchID, decoded.AnalysisBatchID)
	assert.Equal(t, config.InjectedSystemPrompt, decoded.InjectedSystemPrompt)
	assert.Equal(t, config.InjectedToolDescriptions, decoded.InjectedToolDescriptions)
	assert.Equal(t, config.MockedToolOutputs, decoded.MockedToolOutputs)
	assert.Equal(t, config.ReplayPolicy, decoded.ReplayPolicy)
	assert.Equal(t, config.TraceID, decoded.TraceID)
}

func TestShadowResult_JSONRoundTrip(t *testing.T) {
	result := ShadowResult{
		SchemaVersion:       "1.0",
		ShadowExecutionID:   "shadow-exec-001",
		OriginalExecutionID: "exec-001",
		AnalysisBatchID:     "batch-001",
		TenantID:            "tenant-001",
		AgentID:             "agent-001",
		AgentVersion:        "1.0.0",
		TraceID:             "trace-001",
		IdempotencyKey:      "idem-001",
		Status:              ShadowStatusCompleted,
		Output:              "Done",
		Metrics: ShadowResultMetrics{
			TotalTokens:           1500,
			PromptTokens:          1000,
			CompletionTokens:      500,
			EstimatedCostUSD:      0.015,
			TotalDurationMs:       3200,
			Iterations:            3,
			ToolCallsCount:        5,
			ToolErrorsCount:       1,
			MockedToolsCount:      2,
			PassthroughToolsCount: 3,
		},
		Steps:     []json.RawMessage{json.RawMessage(`{"step":1}`)},
		CreatedAt: time.Date(2026, 3, 9, 12, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded ShadowResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, result.SchemaVersion, decoded.SchemaVersion)
	assert.Equal(t, result.ShadowExecutionID, decoded.ShadowExecutionID)
	assert.Equal(t, result.Status, decoded.Status)
	assert.Equal(t, result.Metrics.TotalTokens, decoded.Metrics.TotalTokens)
	assert.Equal(t, result.Metrics.EstimatedCostUSD, decoded.Metrics.EstimatedCostUSD)
	assert.Len(t, decoded.Steps, 1)
}

func TestIsValidExecutionMode(t *testing.T) {
	assert.True(t, IsValidExecutionMode(ExecutionModeNormal))
	assert.True(t, IsValidExecutionMode(ExecutionModeShadow))
	assert.False(t, IsValidExecutionMode(""))
	assert.False(t, IsValidExecutionMode("debug"))
	assert.False(t, IsValidExecutionMode("replay"))
}

func TestIsValidShadowStatus(t *testing.T) {
	assert.True(t, IsValidShadowStatus(ShadowStatusCompleted))
	assert.True(t, IsValidShadowStatus(ShadowStatusFailed))
	assert.True(t, IsValidShadowStatus(ShadowStatusAborted))
	assert.False(t, IsValidShadowStatus(""))
	assert.False(t, IsValidShadowStatus("running"))
	assert.False(t, IsValidShadowStatus("pending"))
}

func TestShadowConstants(t *testing.T) {
	// Stream names
	assert.Equal(t, "passflow.tasks.queue", StreamTasksQueue)
	assert.Equal(t, "passflow.shadow.results", StreamShadowResults)

	// Consumer groups
	assert.Equal(t, "agent-executor-group", ConsumerGroupAgentExecutor)
	assert.Equal(t, "shadow-analyzer-group", ConsumerGroupShadowAnalyzer)

	// Execution modes
	assert.Equal(t, "normal", ExecutionModeNormal)
	assert.Equal(t, "shadow", ExecutionModeShadow)

	// Replay policies
	assert.Equal(t, "strict", ReplayPolicyStrict)
	assert.Equal(t, "best_effort", ReplayPolicyBestEffort)

	// Shadow statuses
	assert.Equal(t, "completed", ShadowStatusCompleted)
	assert.Equal(t, "failed", ShadowStatusFailed)
	assert.Equal(t, "aborted", ShadowStatusAborted)
}
