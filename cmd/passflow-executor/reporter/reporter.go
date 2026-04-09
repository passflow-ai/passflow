package reporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jaak-ai/passflow-agent-executor/engine"
	"github.com/jaak-ai/passflow-agent-executor/job"
)

// Reporter sends execution updates back to passflow-api.
type Reporter struct {
	apiURL       string
	serviceToken string
	httpClient   *http.Client
}

// New creates a new Reporter.
func New(apiURL, serviceToken string) *Reporter {
	return &Reporter{
		apiURL:       apiURL,
		serviceToken: serviceToken,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
	}
}

// ReportStarted notifies the API that an execution has begun.
func (r *Reporter) ReportStarted(ctx context.Context, spec job.Spec) error {
	payload := map[string]interface{}{
		"status":       "running",
		"started_at":   time.Now().UTC(),
		"agent_id":     spec.AgentID,
		"workspace_id": spec.WorkspaceID,
	}
	applyTriggerMetadata(payload, spec.TriggerMetadata)
	return r.patch(ctx, fmt.Sprintf("/api/v1/internal/executions/%s/status", spec.ExecutionID), payload)
}

// ReportStep sends a single step result to the API.
func (r *Reporter) ReportStep(ctx context.Context, executionID string, step engine.StepResult) error {
	return r.post(ctx, fmt.Sprintf("/api/v1/internal/executions/%s/steps", executionID), step)
}

// ReportCompleted sends the final execution result.
func (r *Reporter) ReportCompleted(ctx context.Context, spec job.Spec, result *engine.ExecutionResult) error {
	payload := map[string]interface{}{
		"status":       result.Status,
		"output":       result.Output,
		"steps":        result.Steps,
		"usage":        result.Usage,
		"metadata":     result.Metadata,
		"error":        result.Error,
		"completed_at": result.CompletedAt,
		"agent_id":     result.AgentID,
		"workspace_id": spec.WorkspaceID,
	}
	applyTriggerMetadata(payload, spec.TriggerMetadata)
	return r.patch(ctx, fmt.Sprintf("/api/v1/internal/executions/%s/status", result.ExecutionID), payload)
}

// ReportFailed notifies the API of a failed execution.
func (r *Reporter) ReportFailed(ctx context.Context, spec job.Spec, errMsg string) error {
	payload := map[string]interface{}{
		"status":       "failed",
		"error":        errMsg,
		"agent_id":     spec.AgentID,
		"workspace_id": spec.WorkspaceID,
		"completed_at": time.Now().UTC(),
	}
	applyTriggerMetadata(payload, spec.TriggerMetadata)
	return r.patch(ctx, fmt.Sprintf("/api/v1/internal/executions/%s/status", spec.ExecutionID), payload)
}

func (r *Reporter) ReportWorkflowStarted(ctx context.Context, workspaceID, executionID string, triggerData map[string]any) error {
	payload := map[string]interface{}{
		"workspace_id": workspaceID,
		"status":       "running",
		"trigger_data": triggerData,
	}
	return r.patch(ctx, fmt.Sprintf("/api/v1/internal/workflow-executions/%s/status", executionID), payload)
}

func (r *Reporter) ReportWorkflowStep(
	ctx context.Context,
	workspaceID, executionID, stepID, stepName, status string,
	input map[string]any,
	output map[string]any,
	errMsg string,
) error {
	now := time.Now().UTC()
	payload := map[string]interface{}{
		"workspace_id": workspaceID,
		"step_id":      stepID,
		"step_name":    stepName,
		"status":       status,
		"input":        input,
		"output":       output,
		"error":        errMsg,
		"started_at":   now,
	}
	if status == "completed" || status == "failed" || status == "cancelled" {
		payload["completed_at"] = now
	}
	return r.post(ctx, fmt.Sprintf("/api/v1/internal/workflow-executions/%s/steps", executionID), payload)
}

func (r *Reporter) ReportWorkflowCompleted(
	ctx context.Context,
	workspaceID, executionID string,
	triggerData map[string]any,
) error {
	payload := map[string]interface{}{
		"workspace_id": workspaceID,
		"status":       "completed",
		"trigger_data": triggerData,
		"completed_at": time.Now().UTC(),
	}
	return r.patch(ctx, fmt.Sprintf("/api/v1/internal/workflow-executions/%s/status", executionID), payload)
}

func (r *Reporter) ReportWorkflowFailed(
	ctx context.Context,
	workspaceID, executionID, errMsg string,
	triggerData map[string]any,
) error {
	payload := map[string]interface{}{
		"workspace_id": workspaceID,
		"status":       "failed",
		"error":        errMsg,
		"trigger_data": triggerData,
		"completed_at": time.Now().UTC(),
	}
	return r.patch(ctx, fmt.Sprintf("/api/v1/internal/workflow-executions/%s/status", executionID), payload)
}

func (r *Reporter) post(ctx context.Context, path string, payload interface{}) error {
	return r.send(ctx, http.MethodPost, path, payload)
}

func (r *Reporter) patch(ctx context.Context, path string, payload interface{}) error {
	return r.send(ctx, http.MethodPatch, path, payload)
}

func (r *Reporter) send(ctx context.Context, method, path string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("reporter: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, r.apiURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("reporter: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", r.serviceToken)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("reporter: request to %s failed: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("reporter: API returned %d for %s", resp.StatusCode, path)
	}

	return nil
}

func applyTriggerMetadata(payload map[string]interface{}, metadata *job.TriggerMetadata) {
	if metadata == nil {
		return
	}

	trigger := map[string]interface{}{}
	if metadata.Source != "" {
		trigger["source"] = metadata.Source
	}
	if metadata.TriggerRuleID != "" {
		trigger["trigger_rule_id"] = metadata.TriggerRuleID
	}
	if metadata.Channel != nil {
		trigger["channel"] = map[string]interface{}{
			"type":            metadata.Channel.Type,
			"event_id":        metadata.Channel.EventID,
			"sender_id":       metadata.Channel.SenderID,
			"account_id":      metadata.Channel.AccountID,
			"message_id":      metadata.Channel.MessageID,
			"conversation_id": metadata.Channel.ConversationID,
			"thread_ts":       metadata.Channel.ThreadTS,
			"session_scope":   metadata.Channel.SessionScope,
			"session_key":     metadata.Channel.SessionKey,
		}
	}
	if len(trigger) == 0 {
		return
	}
	payload["trigger_metadata"] = trigger
}
