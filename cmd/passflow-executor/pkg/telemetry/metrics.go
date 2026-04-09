package telemetry

import (
	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	ExecutionDuration    metric.Float64Histogram
	LLMCallDuration      metric.Float64Histogram
	ToolCallDuration     metric.Float64Histogram
	QueueLag             metric.Float64Histogram
	ExecutionsTotal      metric.Int64Counter
	ExecutionErrorsTotal metric.Int64Counter
	LLMCallsTotal        metric.Int64Counter
	ToolCallsTotal       metric.Int64Counter
	ToolErrorsTotal      metric.Int64Counter
	TokensTotal          metric.Int64Counter
	CostTotal            metric.Float64Counter
	QueueDepth           metric.Int64Gauge
}

func InitMetrics() (*Metrics, error) {
	meter := GetMeter()
	m := &Metrics{}
	var err error

	m.ExecutionDuration, err = meter.Float64Histogram(
		"passflow.execution.duration",
		metric.WithDescription("Duration of agent execution in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	m.LLMCallDuration, err = meter.Float64Histogram(
		"passflow.llm.call.duration",
		metric.WithDescription("Duration of LLM API call in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	m.ToolCallDuration, err = meter.Float64Histogram(
		"passflow.tool.call.duration",
		metric.WithDescription("Duration of tool execution in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	m.QueueLag, err = meter.Float64Histogram(
		"passflow.queue.lag",
		metric.WithDescription("Time jobs wait in queue before processing"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	m.ExecutionsTotal, err = meter.Int64Counter(
		"passflow.executions.total",
		metric.WithDescription("Total number of executions"),
	)
	if err != nil {
		return nil, err
	}

	m.ExecutionErrorsTotal, err = meter.Int64Counter(
		"passflow.execution.errors.total",
		metric.WithDescription("Total number of failed executions"),
	)
	if err != nil {
		return nil, err
	}

	m.LLMCallsTotal, err = meter.Int64Counter(
		"passflow.llm.calls.total",
		metric.WithDescription("Total number of LLM calls"),
	)
	if err != nil {
		return nil, err
	}

	m.ToolCallsTotal, err = meter.Int64Counter(
		"passflow.tool.calls.total",
		metric.WithDescription("Total number of tool calls"),
	)
	if err != nil {
		return nil, err
	}

	m.ToolErrorsTotal, err = meter.Int64Counter(
		"passflow.tool.errors.total",
		metric.WithDescription("Total number of tool errors"),
	)
	if err != nil {
		return nil, err
	}

	m.TokensTotal, err = meter.Int64Counter(
		"passflow.llm.tokens.total",
		metric.WithDescription("Total tokens consumed"),
	)
	if err != nil {
		return nil, err
	}

	m.CostTotal, err = meter.Float64Counter(
		"passflow.llm.cost.usd",
		metric.WithDescription("Estimated cost in USD"),
	)
	if err != nil {
		return nil, err
	}

	m.QueueDepth, err = meter.Int64Gauge(
		"passflow.queue.depth",
		metric.WithDescription("Number of jobs waiting in queue"),
	)
	if err != nil {
		return nil, err
	}

	return m, nil
}
