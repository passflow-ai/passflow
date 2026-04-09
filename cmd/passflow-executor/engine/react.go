package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/passflow-ai/passflow/cmd/passflow-executor/job"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/llm"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/pkg/telemetry"
	"github.com/passflow-ai/passflow/cmd/passflow-executor/tools"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// RunReactLoop executes the agent in a Reason+Act loop.
//
// Loop:
//  1. LLM receives messages + available tools
//  2. If LLM returns tool calls → execute them, append results, repeat
//  3. If LLM returns a plain text response → done
//  4. Hard stop at spec.MaxIterations to prevent infinite loops
//
// opts is optional (zero value is valid). When opts.ACF is non-nil, system
// tool calls (Bash, Read, Write, …) are routed through the ACF layer instead
// of the regular toolExecutor, and opts.SystemPromptOverride replaces
// spec.SystemPrompt as the system message.
func RunReactLoop(ctx context.Context, spec *job.Spec, llmClient llm.Client, toolExecutor *tools.Executor, opts ...Options) (*ExecutionResult, error) {
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
			attribute.String("execution.mode", "react_loop"),
			attribute.Int("execution.max_iterations", spec.MaxIterations),
		),
	)
	defer span.End()

	result := &ExecutionResult{
		ExecutionID: spec.ExecutionID,
		AgentID:     spec.AgentID,
		TaskID:      spec.TaskID,
		Metadata:    spec.BuilderMetadata,
		Status:      "completed",
		DecisionLog: NewDecisionLog(spec.ExecutionID, spec.AgentID, spec.ModelProvider, spec.ModelID),
	}

	maxIter := spec.MaxIterations
	if maxIter <= 0 {
		maxIter = job.DefaultMaxIterations
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

	for iter := 0; iter < maxIter; iter++ {
		select {
		case <-ctx.Done():
			result.Status = "failed"
			result.Error = "context cancelled"
			result.CompletedAt = time.Now()
			span.SetStatus(codes.Error, "context cancelled")
			return result, nil
		default:
		}

		llmStart := time.Now()
		_, llmSpan := tracer.Start(ctx, "llm_call",
			trace.WithAttributes(
				attribute.String("llm.provider", spec.ModelProvider),
				attribute.String("llm.model", spec.ModelID),
				attribute.Int("llm.iteration", iter),
			),
		)

		resp, err := llmClient.Complete(ctx, messages, toolDefs)
		llmDuration := time.Since(llmStart)

		if err != nil {
			llmSpan.RecordError(err)
			llmSpan.SetStatus(codes.Error, err.Error())
			llmSpan.End()
			result.Status = "failed"
			result.Error = fmt.Sprintf("LLM call failed at iteration %d: %v", iter, err)
			result.CompletedAt = time.Now()
			result.DecisionLog.Complete()
			span.SetStatus(codes.Error, result.Error)
			return result, nil
		}

		llmSpan.SetAttributes(
			attribute.Int("llm.tokens.prompt", resp.Usage.PromptTokens),
			attribute.Int("llm.tokens.completion", resp.Usage.CompletionTokens),
			attribute.Int("llm.tokens.total", resp.Usage.TotalTokens),
		)
		llmSpan.End()

		recordLLMUsage(&result.Usage, resp.Usage, llmDuration)

		// Create decision log entry for this iteration
		decisionEntry := DecisionLogEntry{
			Iteration:   iter,
			Timestamp:   llmStart,
			Prompt:      messages,
			LLMResponse: resp.Content,
			ToolCalls:   resp.ToolCalls,
			TokenUsage:  resp.Usage,
			DurationMs:  llmDuration.Milliseconds(),
		}

		result.Steps = append(result.Steps, StepResult{
			Index:     stepIdx,
			Type:      "llm_call",
			Name:      "llm",
			Input:     lastUserMessage(messages),
			Output:    resp.Content,
			Timestamp: llmStart,
		})
		stepIdx++

		if len(resp.ToolCalls) == 0 {
			// No tool calls means the agent is done
			result.DecisionLog.AddDecision(decisionEntry)
			result.Output = resp.Content
			break
		}

		messages = append(messages, llm.Message{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		toolResults := executeToolCalls(ctx, resp.ToolCalls, toolExecutor, opt.ACF)

		// Capture tool results in decision log
		decisionToolResults := make([]DecisionToolResult, len(toolResults))

		for i, tr := range toolResults {
			toolExecStart := time.Now()
			_, toolSpan := tracer.Start(ctx, "tool_execution",
				trace.WithAttributes(
					attribute.String("tool.name", tr.ToolName),
					attribute.Int("tool.iteration", iter),
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

			// Record tool result in decision log
			decisionToolResults[i] = DecisionToolResult{
				ToolCallID: tr.ToolCallID,
				ToolName:   tr.ToolName,
				Output:     tr.Output,
				Error:      tr.Error,
				DurationMs: time.Since(toolExecStart).Milliseconds(),
			}

			messages = append(messages, llm.Message{
				Role:       "tool",
				Content:    tr.Output,
				ToolCallID: tr.ToolCallID,
			})
		}

		// Add decision entry with tool results
		decisionEntry.ToolResults = decisionToolResults
		result.DecisionLog.AddDecision(decisionEntry)

		if iter == maxIter-1 {
			result.Status = "max_iterations"
			result.Error = fmt.Sprintf("reached max iterations (%d) without final response", maxIter)
			span.SetAttributes(attribute.String("execution.status", "max_iterations"))
		}
	}

	span.SetAttributes(
		attribute.String("execution.status", result.Status),
		attribute.Int("execution.steps", len(result.Steps)),
		attribute.Int("execution.tokens.total", result.Usage.TotalTokens),
	)
	result.CompletedAt = time.Now()
	result.DecisionLog.Complete()
	return result, nil
}

// executeToolCalls runs every tool call in resp.ToolCalls. System tool calls
// (when acf is non-nil and acf.IsSystemTool returns true) are routed through
// the ACF layer; all other calls go to the regular toolExecutor.
//
// Calls to the regular executor are batched and run concurrently via
// ExecuteAll; system tool calls are executed one at a time in order because
// they mutate filesystem state and order matters.
func executeToolCalls(ctx context.Context, calls []llm.ToolCall, exec *tools.Executor, acf SystemToolRouter) []tools.Result {
	results := make([]tools.Result, len(calls))

	// Separate calls into two buckets. We preserve index positions so we can
	// write results back in order regardless of concurrent execution.
	type indexedCall struct {
		idx  int
		call llm.ToolCall
	}

	var regularCalls []indexedCall
	var systemCalls []indexedCall

	for i, call := range calls {
		if acf != nil && acf.IsSystemTool(call.Name) {
			systemCalls = append(systemCalls, indexedCall{i, call})
		} else {
			regularCalls = append(regularCalls, indexedCall{i, call})
		}
	}

	// Execute regular (HTTP / integration) calls concurrently via the executor.
	if len(regularCalls) > 0 {
		subset := make([]llm.ToolCall, len(regularCalls))
		for j, ic := range regularCalls {
			subset[j] = ic.call
		}
		subResults := exec.ExecuteAll(ctx, subset)
		for j, ic := range regularCalls {
			results[ic.idx] = subResults[j]
		}
	}

	// Execute system tool calls sequentially (filesystem operations are ordered).
	for _, ic := range systemCalls {
		results[ic.idx] = acf.Execute(ctx, ic.call)
	}

	return results
}
