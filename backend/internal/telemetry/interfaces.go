package telemetry

import "context"

type Logger interface {
	Sync() error
	Debug(ctx context.Context, message string, labels ...label)
	Info(ctx context.Context, message string, labels ...label)
	Warn(ctx context.Context, message string, labels ...label)
	Error(ctx context.Context, message string, labels ...label)
	With(labels ...label) Logger
}

type Gauge interface {
	Set(value float64)
	Inc()
	Dec()
	Add(delta float64)
	Sub(delta float64)
}

type Counter interface {
	Inc()
	Add(delta float64)
}

type Histogram interface {
	Observe(value float64)
}

type Summary interface {
	Observe(value float64)
}

type Registry interface {
	Gauge(ctx context.Context, metric gaugeMetric, labels ...label) Gauge
	Counter(ctx context.Context, metric counterMetric, labels ...label) Counter
	Histogram(ctx context.Context, metric histogramMetric, labels ...label) Histogram
	Summary(ctx context.Context, metric summaryMetric, labels ...label) Summary
}

// @PublicValueInstance
type Telemetry struct {
	Logger   Logger
	Registry Registry
}
