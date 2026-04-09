package telemetry

import (
	"go.opentelemetry.io/otel/metric"
)

// ShadowMetrics holds OpenTelemetry instruments for shadow-mode execution
// observability (Obs-2001). These metrics allow operators to monitor shadow
// runs independently from production executions.
type ShadowMetrics struct {
	// ShadowRunsTotal counts shadow executions partitioned by outcome.
	// Use attribute "outcome" with values "success", "aborted", "error".
	ShadowRunsTotal metric.Int64Counter

	// ShadowRunsSuccess counts shadow runs that completed without abort.
	// Derive rate from this counter for the success-rate metric.
	ShadowRunsSuccess metric.Int64Counter

	// ShadowToolMocksHit counts tool invocations that returned a mocked
	// response (write tools or tools with captured output).
	ShadowToolMocksHit metric.Int64Counter

	// ShadowToolMocksMiss counts tool invocations where no mock was available
	// and the call either passed through (deterministic read) or caused an
	// abort (write / non-deterministic read).
	ShadowToolMocksMiss metric.Int64Counter

	// ShadowDuration records the wall-clock duration of shadow executions.
	ShadowDuration metric.Float64Histogram

	// ShadowTokensTotal counts LLM tokens consumed during shadow runs.
	ShadowTokensTotal metric.Int64Counter
}

// InitShadowMetrics creates and registers all shadow-mode metrics using the
// global OpenTelemetry meter. Call this alongside InitMetrics during startup.
func InitShadowMetrics() (*ShadowMetrics, error) {
	meter := GetMeter()
	m := &ShadowMetrics{}
	var err error

	m.ShadowRunsTotal, err = meter.Int64Counter(
		"passflow.shadow.runs.total",
		metric.WithDescription("Total shadow-mode executions by outcome"),
	)
	if err != nil {
		return nil, err
	}

	m.ShadowRunsSuccess, err = meter.Int64Counter(
		"passflow.shadow.runs.success",
		metric.WithDescription("Shadow runs that completed without abort"),
	)
	if err != nil {
		return nil, err
	}

	m.ShadowToolMocksHit, err = meter.Int64Counter(
		"passflow.shadow.tool.mocks.hit.total",
		metric.WithDescription("Tool invocations that returned mocked output in shadow mode"),
	)
	if err != nil {
		return nil, err
	}

	m.ShadowToolMocksMiss, err = meter.Int64Counter(
		"passflow.shadow.tool.mocks.miss.total",
		metric.WithDescription("Tool invocations with no mock available in shadow mode"),
	)
	if err != nil {
		return nil, err
	}

	m.ShadowDuration, err = meter.Float64Histogram(
		"passflow.shadow.duration",
		metric.WithDescription("Duration of shadow-mode execution in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	m.ShadowTokensTotal, err = meter.Int64Counter(
		"passflow.shadow.tokens.total",
		metric.WithDescription("LLM tokens consumed during shadow-mode runs"),
	)
	if err != nil {
		return nil, err
	}

	return m, nil
}
