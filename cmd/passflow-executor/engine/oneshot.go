package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/jaak-ai/passflow-agent-executor/job"
	"github.com/jaak-ai/passflow-agent-executor/llm"
	"github.com/jaak-ai/passflow-agent-executor/pkg/telemetry"
	"github.com/jaak-ai/passflow-agent-executor/tools"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// RunOneShot executes a single LLM call (no agentic loop).
// The agent receives the system prompt + user input, optionally calls tools
// in a single round-trip, and returns the final response.
//
// opts is optional (zero value is valid). When opts.ACF is non-nil, system
// tool calls are routed through the ACF layer, and opts.SystemPromptOverride
// replaces spec.SystemPrompt as the system message.
func RunOneShot(ctx context.Context, spec *job.Spec, llmClient llm.Client, toolExecutor *tools.Executor, opts ...Options) (*ExecutionResult, error) {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	tracer := telemetry.GetTracer()
	ctx, span := tracer.Start(ctx, "execution",
		trace.WithAttributes(
			attribute.String("execution.id", spec.ExecutionID),
			attribute.String("workspace.id", spec.WorkspaceID),
			attribute.String("agent.id", spec.AgentID),
			attribute.String("execution.mode", "one_shot"),
		),
	)
	defer span.End()

	result := &ExecutionResult{
		ExecutionID: spec.ExecutionID,
		AgentID:     spec.AgentID,
		TaskID:      spec.TaskID,
		Metadata:    spec.BuilderMetadata,
		Status:      "completed",
	}

	// Apply system prompt override when ACF provides a translated prompt.
	effectiveSpec := spec
	if opt.SystemPromptOverride != "" {
		copy := *spec
		copy.SystemPrompt = opt.SystemPromptOverride
		effectiveSpec = &copy
	}

	messages := buildInitialMessages(effectiveSpec)

	// Merge tool definitions: workspace tools first, then ACF system tools.
	toolDefs := toolExecutor.Definitions()
	if opt.ACF != nil {
		toolDefs = append(toolDefs, opt.ACF.GetSystemToolDefs()...)
	}

	stepIdx := 0

	llmStart := time.Now()
	_, llmSpan := tracer.Start(ctx, "llm_call",
		trace.WithAttributes(
			attribute.String("llm.provider", spec.ModelProvider),
			attribute.String("llm.model", spec.ModelID),
			attribute.Int("llm.iteration", 0),
		),
	)

	resp, err := llmClient.Complete(ctx, messages, toolDefs)
	if err != nil {
		llmSpan.RecordError(err)
		llmSpan.SetStatus(codes.Error, err.Error())
		llmSpan.End()
		result.Status = "failed"
		result.Error = fmt.Sprintf("LLM call failed: %v", err)
		span.SetStatus(codes.Error, result.Error)
		return result, nil
	}

	llmSpan.SetAttributes(
		attribute.Int("llm.tokens.prompt", resp.Usage.PromptTokens),
		attribute.Int("llm.tokens.completion", resp.Usage.CompletionTokens),
		attribute.Int("llm.tokens.total", resp.Usage.TotalTokens),
	)
	llmSpan.End()

	result.Steps = append(result.Steps, StepResult{
		Index:     stepIdx,
		Type:      "llm_call",
		Name:      "llm",
		Input:     spec.Input,
		Output:    resp.Content,
		Timestamp: llmStart,
	})
	stepIdx++

	recordLLMUsage(&result.Usage, resp.Usage, time.Since(llmStart))

	if len(resp.ToolCalls) > 0 {
		messages = append(messages, llm.Message{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		toolResults := executeToolCalls(ctx, resp.ToolCalls, toolExecutor, opt.ACF)
		for _, tr := range toolResults {
			_, toolSpan := tracer.Start(ctx, "tool_execution",
				trace.WithAttributes(
					attribute.String("tool.name", tr.ToolName),
					attribute.Int("tool.iteration", 0),
				),
			)

			if tr.Error != "" {
				toolSpan.SetAttributes(attribute.String("tool.status", "error"))
				toolSpan.SetStatus(codes.Error, tr.Error)
			} else {
				toolSpan.SetAttributes(attribute.String("tool.status", "success"))
			}
			toolSpan.End()

			result.Steps = append(result.Steps, StepResult{
				Index:     stepIdx,
				Type:      "tool_call",
				Name:      tr.ToolName,
				Input:     findToolArgs(resp.ToolCalls, tr.ToolCallID),
				Output:    tr.Output,
				Error:     tr.Error,
				Timestamp: time.Now(),
			})
			stepIdx++
			recordToolUsage(&result.Usage, tr.Error)

			messages = append(messages, llm.Message{
				Role:       "tool",
				Content:    tr.Output,
				ToolCallID: tr.ToolCallID,
			})
		}

		finalLLMStart := time.Now()
		_, finalLlmSpan := tracer.Start(ctx, "llm_call",
			trace.WithAttributes(
				attribute.String("llm.provider", spec.ModelProvider),
				attribute.String("llm.model", spec.ModelID),
				attribute.Int("llm.iteration", 1),
				attribute.Bool("llm.is_followup", true),
			),
		)

		finalResp, err := llmClient.Complete(ctx, messages, nil)
		if err != nil {
			finalLlmSpan.RecordError(err)
			finalLlmSpan.SetStatus(codes.Error, err.Error())
			finalLlmSpan.End()
			result.Status = "failed"
			result.Error = fmt.Sprintf("follow-up LLM call failed: %v", err)
			span.SetStatus(codes.Error, result.Error)
			return result, nil
		}

		finalLlmSpan.SetAttributes(
			attribute.Int("llm.tokens.prompt", finalResp.Usage.PromptTokens),
			attribute.Int("llm.tokens.completion", finalResp.Usage.CompletionTokens),
			attribute.Int("llm.tokens.total", finalResp.Usage.TotalTokens),
		)
		finalLlmSpan.End()

		recordLLMUsage(&result.Usage, finalResp.Usage, time.Since(finalLLMStart))
		result.Steps = append(result.Steps, StepResult{
			Index:     stepIdx,
			Type:      "llm_call",
			Name:      "llm",
			Input:     "(tool results)",
			Output:    finalResp.Content,
			Timestamp: time.Now(),
		})

		result.Output = finalResp.Content
	} else {
		result.Output = resp.Content
	}

	span.SetAttributes(
		attribute.String("execution.status", result.Status),
		attribute.Int("execution.steps", len(result.Steps)),
		attribute.Int("execution.tokens.total", result.Usage.TotalTokens),
	)
	result.CompletedAt = time.Now()
	return result, nil
}
