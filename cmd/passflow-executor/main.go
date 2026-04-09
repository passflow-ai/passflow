package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jaak-ai/passflow-agent-executor/acf"
	"github.com/jaak-ai/passflow-agent-executor/config"
	"github.com/jaak-ai/passflow-agent-executor/engine"
	"github.com/jaak-ai/passflow-agent-executor/job"
	"github.com/jaak-ai/passflow-agent-executor/llm"
	"github.com/jaak-ai/passflow-agent-executor/llm/factory"
	"github.com/jaak-ai/passflow-agent-executor/llm/pricing"
	"github.com/jaak-ai/passflow-agent-executor/llm/registry"
	"github.com/jaak-ai/passflow-agent-executor/llm/secrets"
	"github.com/jaak-ai/passflow-agent-executor/pkg/telemetry"
	"github.com/jaak-ai/passflow-agent-executor/reporter"
	"github.com/jaak-ai/passflow-agent-executor/tools"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// jobReporter is the subset of *reporter.Reporter used by processJob.
// Abstracting it as an interface enables unit testing with stubs.
type jobReporter interface {
	ReportStarted(ctx context.Context, spec job.Spec) error
	ReportFailed(ctx context.Context, spec job.Spec, errMsg string) error
	ReportCompleted(ctx context.Context, spec job.Spec, result *engine.ExecutionResult) error
	ReportWorkflowStarted(ctx context.Context, workspaceID, executionID string, triggerData map[string]any) error
	ReportWorkflowStep(ctx context.Context, workspaceID, executionID, stepID, stepName, status string, input map[string]any, output map[string]any, errMsg string) error
	ReportWorkflowCompleted(ctx context.Context, workspaceID, executionID string, triggerData map[string]any) error
	ReportWorkflowFailed(ctx context.Context, workspaceID, executionID, errMsg string, triggerData map[string]any) error
}

// llmFactoryFunc is a function that creates an llm.Client for the given
// provider and model. Injecting it into processJob allows tests to substitute
// a stub without network access.
type llmFactoryFunc func(ctx context.Context, provider, modelID, workspaceID string, cfg *config.Config) (llm.Client, error)

// defaultLLMFactory is the production factory: creates the appropriate client
// based on provider, with optional fallback wrapping controlled by cfg.
// It uses the control-plane to resolve workspace-level credentials.
func defaultLLMFactory(ctx context.Context, provider, modelID, workspaceID string, cfg *config.Config) (llm.Client, error) {
	// Create control-plane source for workspace credentials
	var workspaceSource secrets.WorkspaceCredentialSource
	if cfg.APIURL != "" && cfg.ServiceToken != "" {
		workspaceSource = secrets.NewControlPlaneSource(cfg.APIURL, cfg.ServiceToken)
	}

	// Create resolver with workspace credentials support
	resolver := secrets.NewResolver(nil, workspaceSource)

	// Create factory with the resolver
	reg := registry.New()
	f := factory.NewFactory(reg, resolver)

	// Create client config
	factoryCfg := factory.FactoryConfig{
		EnableFallback: cfg.EnableLLMFallback,
		MaxRetries:     cfg.LLMMaxRetries,
	}

	// Create client, passing workspaceID for credential resolution
	return f.NewClientWithWorkspace(ctx, provider, modelID, workspaceID, factoryCfg)
}

// pendingMsg bundles a parsed job spec with the Redis message ID so that
// the worker can acknowledge the message only after successful processing.
type pendingMsg struct {
	msgID string
	spec  job.Spec
}

func main() {
	// Initialize telemetry
	telemetryCfg := telemetry.LoadConfig()
	logger := telemetry.NewLogger(telemetryCfg.ServiceName)
	defer logger.Sync()

	logger.Info("Passflow Agent Executor starting")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownTelemetry, err := telemetry.InitProvider(ctx, telemetryCfg)
	if err != nil {
		logger.Fatal("Failed to initialize telemetry", zap.Error(err))
	}
	defer shutdownTelemetry(context.Background())

	metrics, err := telemetry.InitMetrics()
	if err != nil {
		logger.Fatal("Failed to initialize metrics", zap.Error(err))
	}

	logger.Info("Telemetry initialized",
		zap.String("service", telemetryCfg.ServiceName),
		zap.String("endpoint", telemetryCfg.OTLPEndpoint),
		zap.Bool("traces", telemetryCfg.TracesEnabled),
		zap.Bool("metrics", telemetryCfg.MetricsEnabled),
	)

	cfg := config.Load()
	if cfg.ServiceToken == "" {
		logger.Warn("CHRONOS_SERVICE_TOKEN is not set - API reporting will likely fail")
	}

	// Graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		logger.Info("Shutdown signal received")
		cancel()
	}()

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Fatal("Failed to connect to Redis",
			zap.String("url", redactURL(cfg.RedisURL)),
			zap.Error(err))
	}
	logger.Info("Connected to Redis", zap.String("url", redactURL(cfg.RedisURL)))

	// Ensure the consumer group exists
	if err := ensureConsumerGroup(ctx, rdb, cfg.StreamKey, cfg.ConsumerGroup); err != nil {
		logger.Fatal("Failed to create consumer group", zap.Error(err))
	}

	rep := reporter.New(cfg.APIURL, cfg.ServiceToken)
	shadowRep := reporter.NewShadowReporter(rdb)

	// Worker pool — channel now carries pendingMsg so workers receive both
	// the spec and the message ID needed for acknowledgment.
	jobCh := make(chan pendingMsg, cfg.Workers*2)
	var wg sync.WaitGroup

	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for pm := range jobCh {
				ok := processJob(ctx, cfg, rep, pm.spec, logger, metrics, defaultLLMFactory, shadowRep)
				if ok {
					// ACK only after successful processing to guarantee
					// at-least-once delivery semantics.
					if err := rdb.XAck(ctx, cfg.StreamKey, cfg.ConsumerGroup, pm.msgID).Err(); err != nil {
						logger.Error("Failed to XAck message",
							zap.String("msg_id", pm.msgID),
							zap.Error(err))
					}
				}
				// On failure the message remains in the PEL and will be
				// re-delivered during the next PEL recovery pass.
			}
		}(i)
	}

	logger.Info("Worker pool started",
		zap.Int("workers", cfg.Workers),
		zap.String("stream", cfg.StreamKey),
		zap.String("group", cfg.ConsumerGroup))

	// PEL recovery: re-process any messages that were delivered to this
	// consumer before a previous crash but never acknowledged.
	recoverPendingMessages(ctx, rdb, cfg, rep, jobCh, logger, metrics)

	// Main read loop — uses ">" to receive only new (undelivered) messages.
	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping consumer loop")
			close(jobCh)
			wg.Wait()
			logger.Info("All workers done. Goodbye.")
			return
		default:
		}

		msgs, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    cfg.ConsumerGroup,
			Consumer: cfg.ConsumerID,
			Streams:  []string{cfg.StreamKey, ">"},
			Count:    int64(cfg.Workers),
			Block:    2 * time.Second,
		}).Result()

		if err != nil {
			if err == redis.Nil || err.Error() == "redis: nil" {
				continue
			}
			if ctx.Err() != nil {
				continue
			}
			logger.Error("XReadGroup error", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		for _, stream := range msgs {
			for _, msg := range stream.Messages {
				spec, err := parseJobSpec(msg)
				if err != nil {
					logger.Error("Failed to parse job spec",
						zap.String("msg_id", msg.ID),
						zap.Error(err))
					// Unrecoverable parse error — ACK immediately so the
					// message is not replayed indefinitely.
					if ackErr := rdb.XAck(ctx, cfg.StreamKey, cfg.ConsumerGroup, msg.ID).Err(); ackErr != nil {
						logger.Error("Failed to XAck unparseable message",
							zap.String("msg_id", msg.ID),
							zap.Error(ackErr))
					}
					continue
				}

				// Enqueue for processing; ACK happens inside the worker
				// after processJob returns successfully.
				select {
				case jobCh <- pendingMsg{msgID: msg.ID, spec: spec}:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// recoverPendingMessages reads all messages in the PEL (Pending Entries List)
// for this consumer — messages that were delivered before a previous crash but
// never acknowledged — and re-queues them for processing.
//
// It uses XREADGROUP with "0" (instead of ">") which returns entries already
// owned by this consumer. It pages through the PEL until no entries remain.
func recoverPendingMessages(
	ctx context.Context,
	rdb *redis.Client,
	cfg *config.Config,
	rep *reporter.Reporter,
	jobCh chan<- pendingMsg,
	logger *zap.Logger,
	metrics *telemetry.Metrics,
) {
	logger.Info("Checking PEL for pending messages to recover")

	recovered := 0
	// Read in batches starting from ID "0" until the PEL is exhausted.
	startID := "0"
	for {
		if ctx.Err() != nil {
			return
		}

		msgs, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    cfg.ConsumerGroup,
			Consumer: cfg.ConsumerID,
			Streams:  []string{cfg.StreamKey, startID},
			Count:    int64(cfg.Workers * 2),
			// No blocking — we want only what is already pending.
			Block: 0,
		}).Result()

		if err != nil {
			if err == redis.Nil || err.Error() == "redis: nil" {
				// No pending entries; PEL is empty.
				break
			}
			logger.Error("Error reading PEL during recovery", zap.Error(err))
			break
		}

		batchEmpty := true
		for _, stream := range msgs {
			if len(stream.Messages) == 0 {
				continue
			}
			batchEmpty = false
			for _, msg := range stream.Messages {
				// Advance cursor so the next iteration starts after this batch.
				startID = msg.ID

				spec, err := parseJobSpec(msg)
				if err != nil {
					logger.Error("Failed to parse pending job spec during recovery",
						zap.String("msg_id", msg.ID),
						zap.Error(err))
					// Unrecoverable; ACK so it leaves the PEL.
					if ackErr := rdb.XAck(ctx, cfg.StreamKey, cfg.ConsumerGroup, msg.ID).Err(); ackErr != nil {
						logger.Error("Failed to XAck unparseable pending message",
							zap.String("msg_id", msg.ID),
							zap.Error(ackErr))
					}
					continue
				}

				logger.Info("Recovering pending message",
					zap.String("msg_id", msg.ID),
					zap.String("execution_id", spec.ExecutionID))

				select {
				case jobCh <- pendingMsg{msgID: msg.ID, spec: spec}:
					recovered++
				case <-ctx.Done():
					return
				}
			}
		}

		if batchEmpty {
			break
		}
	}

	if recovered > 0 {
		logger.Info("PEL recovery complete", zap.Int("recovered", recovered))
	} else {
		logger.Info("No pending messages to recover")
	}
}

// processJob runs the agent and reports the result.
//
// ACK/NACK policy (return value):
//   - true  → ACK: remove from PEL. Used for permanent outcomes (valid result,
//     permanent spec error, engine error already reported to API).
//   - false → NACK: leave in PEL for re-delivery. Used for transient failures
//     where retrying the job could succeed (LLM client creation failure,
//     network error when reporting the final result).
//
// The llmFactory parameter is injected so that unit tests can substitute a
// stub without requiring live credentials or network access.
func processJob(
	ctx context.Context,
	cfg *config.Config,
	rep jobReporter,
	spec job.Spec,
	logger *zap.Logger,
	metrics *telemetry.Metrics,
	llmFactory llmFactoryFunc,
	shadowRep *reporter.ShadowReporter,
) bool {
	// Extract trace context from job spec
	ctx = telemetry.ExtractTraceContextFromSpec(ctx, spec.TraceID, spec.SpanID)
	execLogger := telemetry.LoggerWithTraceContext(ctx, logger)
	builderLogFields := builderLogFields(spec)

	logFields := []zap.Field{
		zap.String("execution_id", spec.ExecutionID),
		zap.String("agent_id", spec.AgentID),
		zap.String("mode", spec.Mode),
		zap.String("model_provider", spec.ModelProvider),
		zap.String("model_id", spec.ModelID),
	}
	logFields = append(logFields, builderLogFields...)
	execLogger.Info("Starting job", logFields...)

	startTime := time.Now()

	if msg := spec.Validate(); msg != "" {
		execLogger.Error("Invalid spec", zap.String("reason", msg))
		rep.ReportFailed(ctx, spec, "invalid job spec: "+msg)
		metrics.ExecutionErrorsTotal.Add(ctx, 1, metric.WithAttributes(builderMetricAttrs(spec)...))
		// Validation failure is a permanent error; ACK to avoid infinite retry.
		return true
	}

	if err := rep.ReportStarted(ctx, spec); err != nil {
		execLogger.Warn("Failed to report started", zap.Error(err))
	}

	// Create LLM client via the injected factory. Failure here is treated as
	// transient (credentials may be temporarily unavailable, secret rotation
	// in progress, etc.) so we NACK to allow re-delivery.
	llmClient, llmErr := llmFactory(ctx, spec.ModelProvider, spec.ModelID, spec.WorkspaceID, cfg)
	if llmErr != nil {
		msg := fmt.Sprintf("failed to create LLM client: %v", llmErr)
		execLogger.Error(msg)
		rep.ReportFailed(ctx, spec, msg)
		metrics.ExecutionErrorsTotal.Add(ctx, 1, metric.WithAttributes(builderMetricAttrs(spec)...))
		// Transient failure — NACK so the job is re-delivered after the
		// operator fixes credentials or the secret becomes available.
		return false
	}

	// Release any resources held by the LLM client (e.g., Gemini gRPC
	// connections) when the job is done, regardless of outcome.
	if closer, ok := llmClient.(interface{ Close() error }); ok {
		defer func() {
			if err := closer.Close(); err != nil {
				execLogger.Warn("Failed to close LLM client", zap.Error(err))
			}
		}()
	}

	toolExec := tools.NewWithRuntime(spec.Tools, tools.RuntimeContext{
		WorkspaceID: spec.WorkspaceID,
		AgentID:     spec.AgentID,
		ExecutionID: spec.ExecutionID,
	})

	// Shadow mode: attach interceptor to the tool executor so all tool calls
	// are routed through the shadow policy (mock / passthrough / abort).
	if spec.IsShadow() {
		execLogger.Info("Shadow execution mode",
			zap.String("original_execution_id", spec.ShadowConfig.OriginalExecutionID),
			zap.String("analysis_batch_id", spec.ShadowConfig.AnalysisBatchID),
			zap.String("replay_policy", spec.ShadowConfig.ReplayPolicy))

		policy := tools.NewDefaultShadowPolicy()
		interceptor := tools.NewShadowInterceptor(policy, spec.ShadowConfig.MockedToolOutputs, spec.ShadowConfig.ReplayPolicy)
		toolExec.SetShadowMode(interceptor)

		// Override system prompt if the shadow config provides one.
		if spec.ShadowConfig.InjectedSystemPrompt != "" {
			spec.SystemPrompt = spec.ShadowConfig.InjectedSystemPrompt
		}
	}

	// Prepare ACF context for agents imported via the Agent Compatibility
	// Framework. PrepareACFExecution returns nil for non-ACF jobs so the
	// nil-check below is the only gate needed.
	workDir := os.TempDir() // In K8s pods this is the ephemeral workspace mount.
	acfCtx := acf.PrepareACFExecution(&spec, workDir, execLogger)

	// Build engine options. For ACF agents these carry the translated system
	// prompt and the system tool router; for non-ACF jobs opts is zero-valued
	// and the loop behaves exactly as before.
	var opts engine.Options
	if acfCtx != nil {
		opts.ACF = acf.NewEngineAdapter(acfCtx)
		opts.SystemPromptOverride = acfCtx.SystemPrompt
		execLogger.Info("ACF execution prepared",
			zap.String("provider", acfCtx.Provider),
			zap.Int("system_prompt_len", len(acfCtx.SystemPrompt)))
	}

	var result *engine.ExecutionResult
	var engineErr error

	switch spec.Mode {
	case job.ModeWorkflow:
		return processWorkflowGraphJob(ctx, rep, spec, llmClient, toolExec, execLogger, metrics, startTime)
	case job.ModeReactLoop:
		result, engineErr = engine.RunReactLoop(ctx, &spec, llmClient, toolExec, opts)
	default:
		result, engineErr = engine.RunOneShot(ctx, &spec, llmClient, toolExec, opts)
	}

	duration := time.Since(startTime)
	durationSec := duration.Seconds()

	// Shadow mode: publish result to the shadow results stream and skip
	// the normal passflow-api reporting path entirely.
	if spec.IsShadow() {
		return handleShadowResult(ctx, execLogger, spec, result, engineErr, toolExec, duration, shadowRep, metrics)
	}

	if engineErr != nil {
		msg := fmt.Sprintf("engine error: %v", engineErr)
		execLogger.Error(msg)
		rep.ReportFailed(ctx, spec, msg)
		metrics.ExecutionErrorsTotal.Add(ctx, 1, metric.WithAttributes(builderMetricAttrs(spec)...))
		metrics.ExecutionDuration.Record(ctx, durationSec, metric.WithAttributes(builderMetricAttrs(spec)...))
		// Engine errors are reported as failures to the API; ACK so the
		// message does not loop forever. The API holds the authoritative
		// failed state.
		return true
	}

	execLogger.Info("Completed",
		zap.String("status", result.Status),
		zap.Int("steps", len(result.Steps)),
		zap.Int("tokens", result.Usage.TotalTokens),
		zap.String("error", result.Error))

	enrichExecutionResult(spec, result, duration)

	// Record metrics
	recordExecutionMetrics(ctx, metrics, spec, result, durationSec)

	if err := rep.ReportCompleted(ctx, spec, result); err != nil {
		execLogger.Warn("Failed to report completion — NACK to retry", zap.Error(err))
		// Network error reporting the result: NACK so the job is re-delivered
		// and the report is attempted again. The execution engine won't run
		// again because the result is already in memory; only the HTTP report
		// is retried via re-delivery.
		return false
	}

	return true
}

// handleShadowResult builds a ShadowResultPayload from the execution outcome
// and publishes it to the shadow results Redis Stream. Shadow executions never
// report to passflow-api; the result is consumed by the shadow analyzer service.
func handleShadowResult(
	ctx context.Context,
	logger *zap.Logger,
	spec job.Spec,
	result *engine.ExecutionResult,
	engineErr error,
	toolExec *tools.Executor,
	duration time.Duration,
	shadowRep *reporter.ShadowReporter,
	metrics *telemetry.Metrics,
) bool {
	shadowStats := toolExec.ShadowStats()

	status := "completed"
	var errMsg string
	var output string

	if engineErr != nil {
		status = "failed"
		errMsg = engineErr.Error()
	} else if result != nil {
		status = result.Status
		output = result.Output
		if result.Error != "" {
			errMsg = result.Error
		}
	}

	// If any tool was aborted, mark the entire shadow run as aborted.
	if len(shadowStats.AbortedTools) > 0 {
		status = "aborted"
		if errMsg == "" {
			errMsg = fmt.Sprintf("shadow mode aborted: missing mocks for tools %v", shadowStats.AbortedTools)
		}
	}

	// Build metrics from execution result.
	var sm reporter.ShadowMetrics
	sm.TotalDurationMs = duration.Milliseconds()
	sm.MockedToolsCount = shadowStats.MockedCount
	sm.PassthroughToolsCount = shadowStats.PassthroughCount

	if result != nil {
		sm.TotalTokens = result.Usage.TotalTokens
		sm.PromptTokens = result.Usage.PromptTokens
		sm.CompletionTokens = result.Usage.CompletionTokens
		sm.EstimatedCostUSD = result.Usage.EstimatedCost
		sm.Iterations = len(result.Steps)

		for _, step := range result.Steps {
			if step.Type == "tool_call" {
				sm.ToolCallsCount++
				if step.Error != "" {
					sm.ToolErrorsCount++
				}
			}
		}
	}

	// Serialize steps to JSON.
	var stepsJSON json.RawMessage
	if result != nil && len(result.Steps) > 0 {
		stepsJSON, _ = json.Marshal(result.Steps)
	} else {
		stepsJSON = json.RawMessage("[]")
	}

	idempotencyKey := fmt.Sprintf("%s:%s:%s",
		spec.ShadowConfig.AnalysisBatchID,
		spec.ShadowConfig.OriginalExecutionID,
		spec.ExecutionID)

	payload := &reporter.ShadowResultPayload{
		ShadowExecutionID:   spec.ExecutionID,
		OriginalExecutionID: spec.ShadowConfig.OriginalExecutionID,
		AnalysisBatchID:     spec.ShadowConfig.AnalysisBatchID,
		TenantID:            spec.WorkspaceID,
		AgentID:             spec.AgentID,
		AgentVersion:        spec.ModelID, // version is the model variant used
		TraceID:             spec.ShadowConfig.TraceID,
		IdempotencyKey:      idempotencyKey,
		Status:              status,
		Output:              output,
		Metrics:             sm,
		Steps:               stepsJSON,
		Error:               errMsg,
	}

	if shadowRep == nil {
		logger.Error("Shadow reporter is nil — cannot publish shadow result",
			zap.String("execution_id", spec.ExecutionID))
		metrics.ExecutionErrorsTotal.Add(ctx, 1, metric.WithAttributes(builderMetricAttrs(spec)...))
		return true // ACK to avoid infinite loop
	}

	if err := shadowRep.PublishResult(ctx, payload); err != nil {
		logger.Error("Failed to publish shadow result — NACK to retry",
			zap.String("execution_id", spec.ExecutionID),
			zap.Error(err))
		metrics.ExecutionErrorsTotal.Add(ctx, 1, metric.WithAttributes(builderMetricAttrs(spec)...))
		return false // NACK: publishing failure is transient
	}

	logger.Info("Shadow result published",
		zap.String("execution_id", spec.ExecutionID),
		zap.String("status", status),
		zap.Int("mocked_tools", shadowStats.MockedCount),
		zap.Int("passthrough_tools", shadowStats.PassthroughCount))

	metrics.ExecutionsTotal.Add(ctx, 1, metric.WithAttributes(builderMetricAttrs(spec)...))
	metrics.ExecutionDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(builderMetricAttrs(spec)...))
	if result != nil {
		enrichExecutionResult(spec, result, duration)
		metrics.TokensTotal.Add(ctx, int64(result.Usage.TotalTokens), metric.WithAttributes(builderMetricAttrs(spec)...))
		recordToolMetrics(ctx, metrics, spec, result)
		recordLLMMetrics(ctx, metrics, spec, result)
		if result.Usage.EstimatedCost > 0 {
			metrics.CostTotal.Add(ctx, result.Usage.EstimatedCost, metric.WithAttributes(builderMetricAttrs(spec)...))
		}
	}

	return true
}

func builderMetricAttrs(spec job.Spec) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("workspace_id", spec.WorkspaceID),
		attribute.String("agent_id", spec.AgentID),
		attribute.String("model_provider", spec.ModelProvider),
		attribute.String("model_id", spec.ModelID),
	}
	if blueprintID := spec.BuilderMetadata["blueprint_id"]; blueprintID != "" {
		attrs = append(attrs, attribute.String("blueprint_id", blueprintID))
	}
	if workflowID := spec.BuilderMetadata["workflow_id"]; workflowID != "" {
		attrs = append(attrs, attribute.String("workflow_id", workflowID))
	}
	if compiledWorkflowID := spec.BuilderMetadata["compiled_workflow_id"]; compiledWorkflowID != "" {
		attrs = append(attrs, attribute.String("compiled_workflow_id", compiledWorkflowID))
	}
	return attrs
}

func enrichExecutionResult(spec job.Spec, result *engine.ExecutionResult, duration time.Duration) {
	if result == nil {
		return
	}
	if result.Metadata == nil {
		result.Metadata = map[string]string{}
	}
	if result.Usage.EstimatedCost <= 0 {
		cost, source := pricing.EstimateCost(
			spec.ModelID,
			int64(result.Usage.PromptTokens),
			int64(result.Usage.CompletionTokens),
		)
		result.Usage.EstimatedCost = cost
		if source != "" {
			result.Metadata["cost_estimate_source"] = source
		}
	}
	result.Metadata["model_provider"] = spec.ModelProvider
	result.Metadata["model_id"] = spec.ModelID
	result.Metadata["llm_calls"] = fmt.Sprintf("%d", result.Usage.LLMCalls)
	result.Metadata["tool_calls"] = fmt.Sprintf("%d", result.Usage.ToolCalls)
	result.Metadata["tool_errors"] = fmt.Sprintf("%d", result.Usage.ToolErrors)
	result.Metadata["llm_latency_ms"] = fmt.Sprintf("%d", result.Usage.LLMLatencyMs)
	result.Metadata["execution_duration_ms"] = fmt.Sprintf("%d", duration.Milliseconds())
	if result.Usage.EstimatedCost > 0 {
		result.Metadata["estimated_cost_usd"] = fmt.Sprintf("%.6f", result.Usage.EstimatedCost)
	}
}

func recordExecutionMetrics(ctx context.Context, metrics *telemetry.Metrics, spec job.Spec, result *engine.ExecutionResult, durationSec float64) {
	baseAttrs := builderMetricAttrs(spec)
	metrics.ExecutionsTotal.Add(ctx, 1, metric.WithAttributes(baseAttrs...))
	metrics.ExecutionDuration.Record(ctx, durationSec, metric.WithAttributes(baseAttrs...))
	metrics.TokensTotal.Add(ctx, int64(result.Usage.TotalTokens), metric.WithAttributes(baseAttrs...))
	recordLLMMetrics(ctx, metrics, spec, result)
	recordToolMetrics(ctx, metrics, spec, result)
	if result.Usage.EstimatedCost > 0 {
		metrics.CostTotal.Add(ctx, result.Usage.EstimatedCost, metric.WithAttributes(baseAttrs...))
	}
}

func recordLLMMetrics(ctx context.Context, metrics *telemetry.Metrics, spec job.Spec, result *engine.ExecutionResult) {
	if result == nil || result.Usage.LLMCalls == 0 {
		return
	}
	metrics.LLMCallsTotal.Add(ctx, int64(result.Usage.LLMCalls), metric.WithAttributes(builderMetricAttrs(spec)...))
}

func recordToolMetrics(ctx context.Context, metrics *telemetry.Metrics, spec job.Spec, result *engine.ExecutionResult) {
	if result == nil {
		return
	}
	toolCallsByName := map[string]int64{}
	toolErrorsByName := map[string]int64{}
	for _, step := range result.Steps {
		if step.Type != "tool_call" {
			continue
		}
		toolCallsByName[step.Name]++
		if step.Error != "" {
			toolErrorsByName[step.Name]++
		}
	}
	for toolName, count := range toolCallsByName {
		attrs := append(builderMetricAttrs(spec), attribute.String("tool_name", toolName))
		metrics.ToolCallsTotal.Add(ctx, count, metric.WithAttributes(attrs...))
	}
	for toolName, count := range toolErrorsByName {
		attrs := append(builderMetricAttrs(spec), attribute.String("tool_name", toolName))
		metrics.ToolErrorsTotal.Add(ctx, count, metric.WithAttributes(attrs...))
	}
}

func builderLogFields(spec job.Spec) []zap.Field {
	fields := []zap.Field{}
	if blueprintID := spec.BuilderMetadata["blueprint_id"]; blueprintID != "" {
		fields = append(fields, zap.String("blueprint_id", blueprintID))
	}
	if workflowID := spec.BuilderMetadata["workflow_id"]; workflowID != "" {
		fields = append(fields, zap.String("workflow_id", workflowID))
	}
	if compiledWorkflowID := spec.BuilderMetadata["compiled_workflow_id"]; compiledWorkflowID != "" {
		fields = append(fields, zap.String("compiled_workflow_id", compiledWorkflowID))
	}
	return fields
}

// redactURL replaces the password component of rawURL with [REDACTED] so that
// connection strings can be safely written to log output without exposing
// credentials. If rawURL cannot be parsed, "[REDACTED]" is returned in full
// to avoid accidentally leaking malformed but potentially sensitive strings.
func redactURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "[REDACTED]"
	}
	if parsed.User == nil {
		return rawURL
	}
	password, hasPassword := parsed.User.Password()
	if !hasPassword || password == "" {
		return rawURL
	}
	// Replace only the password portion in the raw string to avoid
	// url.UserPassword's percent-encoding of the placeholder text.
	username := parsed.User.Username()
	var userInfo string
	if username != "" {
		userInfo = username + ":[REDACTED]"
	} else {
		userInfo = ":[REDACTED]"
	}
	// Reconstruct: scheme://userInfo@host/path?query#fragment
	result := parsed.Scheme + "://" + userInfo + "@" + parsed.Host
	if parsed.Path != "" {
		result += parsed.Path
	}
	if parsed.RawQuery != "" {
		result += "?" + parsed.RawQuery
	}
	if parsed.Fragment != "" {
		result += "#" + parsed.Fragment
	}
	return result
}

// ensureConsumerGroup creates the Redis consumer group if it doesn't exist.
func ensureConsumerGroup(ctx context.Context, rdb *redis.Client, stream, group string) error {
	err := rdb.XGroupCreateMkStream(ctx, stream, group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create consumer group %q on stream %q: %w", group, stream, err)
	}
	return nil
}

// parseJobSpec extracts and unmarshals the job spec from a Redis stream message.
func parseJobSpec(msg redis.XMessage) (job.Spec, error) {
	payload, ok := msg.Values["payload"]
	if !ok {
		return job.Spec{}, fmt.Errorf("message missing 'payload' field")
	}

	payloadStr, ok := payload.(string)
	if !ok {
		return job.Spec{}, fmt.Errorf("payload is not a string")
	}

	var spec job.Spec
	if err := json.Unmarshal([]byte(payloadStr), &spec); err != nil {
		return job.Spec{}, fmt.Errorf("failed to unmarshal job spec: %w", err)
	}

	return spec, nil
}
