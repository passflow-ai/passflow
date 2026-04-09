package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/engine"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/job"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/pkg/telemetry"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/tools"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

func processWorkflowGraphJob(
	ctx context.Context,
	rep jobReporter,
	spec job.Spec,
	llmClient llm.Client,
	toolExec *tools.Executor,
	logger *zap.Logger,
	metrics *telemetry.Metrics,
	startTime time.Time,
) bool {
	if spec.WorkflowGraph == nil {
		msg := "workflow_graph mode requires workflow_graph payload"
		logger.Error(msg)
		_ = rep.ReportFailed(ctx, spec, msg)
		return true
	}

	runtimeContext := map[string]any{
		"input": spec.Input,
	}
	triggerData := map[string]any{
		"workflow_current_step_id":   spec.BuilderMetadata["workflow_current_step_id"],
		"workflow_current_step_name": spec.BuilderMetadata["workflow_current_step_name"],
		"workflow_context":           runtimeContext,
	}
	if err := rep.ReportWorkflowStarted(ctx, spec.WorkspaceID, spec.ExecutionID, triggerData); err != nil {
		logger.Warn("Failed to report workflow started", zap.Error(err))
	}

	result := &engine.ExecutionResult{
		ExecutionID: spec.ExecutionID,
		AgentID:     spec.AgentID,
		TaskID:      spec.TaskID,
		Metadata:    spec.BuilderMetadata,
		Status:      "completed",
	}

	currentStepID := spec.BuilderMetadata["workflow_current_step_id"]
	if currentStepID == "" {
		currentStepID = spec.WorkflowGraph.StartStepID
	}
	if currentStepID == "" && len(spec.WorkflowGraph.Steps) > 0 {
		currentStepID = spec.WorkflowGraph.Steps[0].ID
	}

	currentInput := spec.Input
	stepIndex := 0
	for currentStepID != "" {
		step := workflowStepByID(spec.WorkflowGraph, currentStepID)
		if step == nil {
			break
		}

		triggerData["workflow_current_step_id"] = step.ID
		triggerData["workflow_current_step_name"] = step.Name
		triggerData["workflow_context"] = runtimeContext

		stepInput := map[string]any{
			"input":   currentInput,
			"context": runtimeContext,
		}
		_ = rep.ReportWorkflowStep(ctx, spec.WorkspaceID, spec.ExecutionID, step.ID, step.Name, "running", stepInput, nil, "")

		switch step.Type {
		case "input", "transform", "wait", "webhook", "webhook_wait":
			output := map[string]any{
				"auto_advanced": true,
				"step_type":     step.Type,
			}
			runtimeContext[step.ID] = output
			runtimeContext["last_step_output"] = output
			_ = rep.ReportWorkflowStep(ctx, spec.WorkspaceID, spec.ExecutionID, step.ID, step.Name, "completed", stepInput, output, "")
			result.Steps = append(result.Steps, engine.StepResult{
				Index:     stepIndex,
				Type:      "workflow_step",
				Name:      step.Name,
				Input:     currentInput,
				Output:    fmt.Sprintf("auto advanced: %s", step.Type),
				Timestamp: time.Now().UTC(),
			})
			stepIndex++
			currentStepID = nextWorkflowGraphStepID(step, runtimeContext)
		case "rule_check", "branch":
			nextStepID := nextWorkflowGraphStepID(step, runtimeContext)
			output := map[string]any{
				"auto_advanced": true,
				"step_type":     step.Type,
				"selected_next": nextStepID,
			}
			runtimeContext[step.ID] = output
			runtimeContext["last_branch_decision"] = output
			_ = rep.ReportWorkflowStep(ctx, spec.WorkspaceID, spec.ExecutionID, step.ID, step.Name, "completed", stepInput, output, "")
			result.Steps = append(result.Steps, engine.StepResult{
				Index:     stepIndex,
				Type:      "workflow_step",
				Name:      step.Name,
				Input:     currentInput,
				Output:    fmt.Sprintf("branch selected: %s", nextStepID),
				Timestamp: time.Now().UTC(),
			})
			stepIndex++
			currentStepID = nextStepID
		case "approval", "human_task":
			msg := "workflow reached unexpected approval step inside executor"
			_ = rep.ReportWorkflowStep(ctx, spec.WorkspaceID, spec.ExecutionID, step.ID, step.Name, "failed", stepInput, nil, msg)
			_ = rep.ReportWorkflowFailed(ctx, spec.WorkspaceID, spec.ExecutionID, msg, map[string]any{
				"workflow_current_step_id":   step.ID,
				"workflow_current_step_name": step.Name,
				"workflow_context":           runtimeContext,
			})
			metrics.ExecutionErrorsTotal.Add(ctx, 1, metric.WithAttributes(builderMetricAttrs(spec)...))
			return true
		case "tool_call":
			toolName := step.Name
			if toolID, ok := step.Config["toolId"].(string); ok && toolID != "" {
				toolName = toolID
			}
			argsJSON, _ := json.Marshal(map[string]any{"input": currentInput})
			toolResult := toolExec.Execute(ctx, llm.ToolCall{
				ID:        step.ID,
				Name:      toolName,
				Arguments: string(argsJSON),
			})
			decodedOutput := decodeJSONIfPossible(toolResult.Output)
			output := map[string]any{"output": decodedOutput}
			status := "completed"
			if toolResult.Error != "" {
				status = "failed"
				result.Status = "failed"
				result.Error = toolResult.Error
			}
			_ = rep.ReportWorkflowStep(ctx, spec.WorkspaceID, spec.ExecutionID, step.ID, step.Name, status, stepInput, output, toolResult.Error)
			result.Steps = append(result.Steps, engine.StepResult{
				Index:     stepIndex,
				Type:      "tool_call",
				Name:      step.Name,
				Input:     currentInput,
				Output:    toolResult.Output,
				Error:     toolResult.Error,
				Timestamp: time.Now().UTC(),
			})
			stepIndex++
			result.Usage.ToolCalls++
			if toolResult.Error != "" {
				result.Usage.ToolErrors++
			}
			if toolResult.Error != "" {
				_ = rep.ReportWorkflowFailed(ctx, spec.WorkspaceID, spec.ExecutionID, toolResult.Error, map[string]any{
					"workflow_current_step_id":   step.ID,
					"workflow_current_step_name": step.Name,
					"workflow_context":           runtimeContext,
				})
				metrics.ExecutionErrorsTotal.Add(ctx, 1, metric.WithAttributes(builderMetricAttrs(spec)...))
				return true
			}
			runtimeContext[step.ID] = decodedOutput
			runtimeContext["last_tool_output"] = decodedOutput
			currentInput = toolResult.Output
			currentStepID = nextWorkflowGraphStepID(step, runtimeContext)
		case "agent_task":
			subSpec := spec
			subSpec.Mode = job.ModeOneShot
			subSpec.Input = currentInput
			subResult, err := engine.RunOneShot(ctx, &subSpec, llmClient, toolExec)
			if err != nil {
				msg := fmt.Sprintf("workflow agent task failed: %v", err)
				_ = rep.ReportWorkflowStep(ctx, spec.WorkspaceID, spec.ExecutionID, step.ID, step.Name, "failed", stepInput, nil, msg)
				_ = rep.ReportWorkflowFailed(ctx, spec.WorkspaceID, spec.ExecutionID, msg, map[string]any{
					"workflow_current_step_id":   step.ID,
					"workflow_current_step_name": step.Name,
					"workflow_context":           runtimeContext,
				})
				metrics.ExecutionErrorsTotal.Add(ctx, 1, metric.WithAttributes(builderMetricAttrs(spec)...))
				return true
			}
			decodedOutput := decodeJSONIfPossible(subResult.Output)
			output := map[string]any{"output": decodedOutput}
			_ = rep.ReportWorkflowStep(ctx, spec.WorkspaceID, spec.ExecutionID, step.ID, step.Name, subResult.Status, stepInput, output, subResult.Error)
			result.Steps = append(result.Steps, subResult.Steps...)
			result.Output = subResult.Output
			result.Usage.PromptTokens += subResult.Usage.PromptTokens
			result.Usage.CompletionTokens += subResult.Usage.CompletionTokens
			result.Usage.TotalTokens += subResult.Usage.TotalTokens
			result.Usage.EstimatedCost += subResult.Usage.EstimatedCost
			result.Usage.LLMLatencyMs += subResult.Usage.LLMLatencyMs
			result.Usage.LLMCalls += subResult.Usage.LLMCalls
			result.Usage.ToolCalls += subResult.Usage.ToolCalls
			result.Usage.ToolErrors += subResult.Usage.ToolErrors
			if subResult.Status == "failed" || subResult.Error != "" {
				msg := subResult.Error
				if msg == "" {
					msg = "workflow agent task failed"
				}
				_ = rep.ReportWorkflowFailed(ctx, spec.WorkspaceID, spec.ExecutionID, msg, map[string]any{
					"workflow_current_step_id":   step.ID,
					"workflow_current_step_name": step.Name,
					"workflow_context":           runtimeContext,
				})
				metrics.ExecutionErrorsTotal.Add(ctx, 1, metric.WithAttributes(builderMetricAttrs(spec)...))
				return true
			}
			runtimeContext[step.ID] = decodedOutput
			runtimeContext["last_agent_output"] = decodedOutput
			currentInput = subResult.Output
			currentStepID = nextWorkflowGraphStepID(step, runtimeContext)
		case "end":
			_ = rep.ReportWorkflowStep(ctx, spec.WorkspaceID, spec.ExecutionID, step.ID, step.Name, "completed", stepInput, map[string]any{"terminal": true}, "")
			currentStepID = ""
		default:
			_ = rep.ReportWorkflowStep(ctx, spec.WorkspaceID, spec.ExecutionID, step.ID, step.Name, "completed", stepInput, map[string]any{"skipped": true}, "")
			currentStepID = nextWorkflowGraphStepID(step, runtimeContext)
		}
	}

	result.CompletedAt = time.Now().UTC()
	result.Metadata["workflow_completed"] = "true"
	_ = rep.ReportWorkflowCompleted(ctx, spec.WorkspaceID, spec.ExecutionID, map[string]any{
		"workflow_current_step_id":   "",
		"workflow_current_step_name": "",
		"workflow_context":           runtimeContext,
	})
	recordExecutionMetrics(ctx, metrics, spec, result, time.Since(startTime).Seconds())
	return true
}

func workflowStepByID(graph *job.WorkflowGraphSpec, stepID string) *job.WorkflowStepSpec {
	if graph == nil || stepID == "" {
		return nil
	}
	for i := range graph.Steps {
		if graph.Steps[i].ID == stepID {
			return &graph.Steps[i]
		}
	}
	return nil
}

func nextWorkflowGraphStepID(step *job.WorkflowStepSpec, context map[string]any) string {
	if step == nil {
		return ""
	}
	if step.NextStepID != "" {
		return step.NextStepID
	}
	for _, branch := range step.Branches {
		if branch.Condition != "" && branch.Condition != "default" && evaluateSimpleCondition(branch.Condition, context) {
			return branch.NextStepID
		}
	}
	for _, branch := range step.Branches {
		if branch.IsDefault && branch.NextStepID != "" {
			return branch.NextStepID
		}
	}
	for _, branch := range step.Branches {
		if branch.NextStepID != "" {
			return branch.NextStepID
		}
	}
	return ""
}

func decodeJSONIfPossible(raw string) any {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw
	}
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		return raw
	}
	var parsed any
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return raw
	}
	return parsed
}

func evaluateSimpleCondition(condition string, context map[string]any) bool {
	condition = strings.TrimSpace(condition)
	switch condition {
	case "", "default":
		return false
	case "true":
		return true
	case "false":
		return false
	}

	for _, operator := range []string{"==", "!=", ">=", "<=", ">", "<"} {
		if strings.Contains(condition, operator) {
			parts := strings.SplitN(condition, operator, 2)
			if len(parts) != 2 {
				return false
			}
			left := resolveContextValue(strings.TrimSpace(parts[0]), context)
			right := parseConditionValue(strings.TrimSpace(parts[1]))
			return compareValues(left, right, operator)
		}
	}

	value := resolveContextValue(condition, context)
	if boolValue, ok := value.(bool); ok {
		return boolValue
	}
	return value != nil && fmt.Sprint(value) != "" && fmt.Sprint(value) != "0"
}

func resolveContextValue(path string, context map[string]any) any {
	if context == nil || path == "" {
		return nil
	}
	parts := strings.Split(path, ".")
	var current any = context
	for _, part := range parts {
		asMap, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = asMap[part]
	}
	return current
}

func parseConditionValue(value string) any {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	switch value {
	case "true":
		return true
	case "false":
		return false
	}
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	}
	return value
}

func compareValues(left any, right any, operator string) bool {
	switch l := left.(type) {
	case int:
		return compareFloats(float64(l), right, operator)
	case int32:
		return compareFloats(float64(l), right, operator)
	case int64:
		return compareFloats(float64(l), right, operator)
	case float32:
		return compareFloats(float64(l), right, operator)
	case float64:
		return compareFloats(l, right, operator)
	case bool:
		r, ok := right.(bool)
		if !ok {
			return false
		}
		switch operator {
		case "==":
			return l == r
		case "!=":
			return l != r
		default:
			return false
		}
	default:
		ls := fmt.Sprint(left)
		rs := fmt.Sprint(right)
		switch operator {
		case "==":
			return ls == rs
		case "!=":
			return ls != rs
		default:
			return false
		}
	}
}

func compareFloats(left float64, right any, operator string) bool {
	var rightFloat float64
	switch r := right.(type) {
	case int:
		rightFloat = float64(r)
	case int32:
		rightFloat = float64(r)
	case int64:
		rightFloat = float64(r)
	case float32:
		rightFloat = float64(r)
	case float64:
		rightFloat = r
	default:
		parsed, err := strconv.ParseFloat(fmt.Sprint(right), 64)
		if err != nil {
			return false
		}
		rightFloat = parsed
	}

	switch operator {
	case "==":
		return left == rightFloat
	case "!=":
		return left != rightFloat
	case ">":
		return left > rightFloat
	case "<":
		return left < rightFloat
	case ">=":
		return left >= rightFloat
	case "<=":
		return left <= rightFloat
	default:
		return false
	}
}
