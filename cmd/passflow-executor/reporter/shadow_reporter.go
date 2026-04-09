package reporter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// shadowResultStream is the Redis Stream key for shadow execution results.
	shadowResultStream = "passflow.shadow.results"
	// shadowSchemaVersion is the current schema version for shadow results.
	shadowSchemaVersion = "1.0"
)

// ShadowReporter publishes shadow execution results to a Redis Stream.
type ShadowReporter struct {
	rdb       *redis.Client
	streamKey string
}

// NewShadowReporter creates a new ShadowReporter that publishes to the
// passflow.shadow.results stream.
func NewShadowReporter(rdb *redis.Client) *ShadowReporter {
	return &ShadowReporter{
		rdb:       rdb,
		streamKey: shadowResultStream,
	}
}

// ShadowMetrics contains execution metrics for a shadow run.
type ShadowMetrics struct {
	TotalTokens           int     `json:"total_tokens"`
	PromptTokens          int     `json:"prompt_tokens"`
	CompletionTokens      int     `json:"completion_tokens"`
	EstimatedCostUSD      float64 `json:"estimated_cost_usd"`
	TotalDurationMs       int64   `json:"total_duration_ms"`
	Iterations            int     `json:"iterations"`
	ToolCallsCount        int     `json:"tool_calls_count"`
	ToolErrorsCount       int     `json:"tool_errors_count"`
	MockedToolsCount      int     `json:"mocked_tools_count"`
	PassthroughToolsCount int     `json:"passthrough_tools_count"`
}

// ShadowResultPayload is the payload published to the passflow.shadow.results
// stream after a shadow execution completes.
type ShadowResultPayload struct {
	SchemaVersion       string          `json:"schema_version"`
	ShadowExecutionID   string          `json:"shadow_execution_id"`
	OriginalExecutionID string          `json:"original_execution_id"`
	AnalysisBatchID     string          `json:"analysis_batch_id"`
	TenantID            string          `json:"tenant_id"`
	AgentID             string          `json:"agent_id"`
	AgentVersion        string          `json:"agent_version"`
	TraceID             string          `json:"trace_id"`
	IdempotencyKey      string          `json:"idempotency_key"`
	Status              string          `json:"status"` // completed | failed | aborted
	Output              string          `json:"output"`
	Metrics             ShadowMetrics   `json:"metrics"`
	Steps               json.RawMessage `json:"steps"`
	Error               string          `json:"error,omitempty"`
	CreatedAt           string          `json:"created_at"`
}

// PublishResult publishes a shadow execution result to the Redis Stream via XADD.
func (sr *ShadowReporter) PublishResult(ctx context.Context, result *ShadowResultPayload) error {
	if result.SchemaVersion == "" {
		result.SchemaVersion = shadowSchemaVersion
	}
	if result.CreatedAt == "" {
		result.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	payload, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("shadow_reporter: failed to marshal result: %w", err)
	}

	_, err = sr.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: sr.streamKey,
		Values: map[string]interface{}{
			"payload": string(payload),
		},
	}).Result()
	if err != nil {
		return fmt.Errorf("shadow_reporter: failed to XADD to %s: %w", sr.streamKey, err)
	}

	return nil
}
