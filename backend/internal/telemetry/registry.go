package telemetry

import (
	"context"
	"fmt"

	"github.com/leshless/pet/cub/internal/config"
	"github.com/leshless/pet/cub/internal/environment"
	"github.com/prometheus/client_golang/prometheus"
)

// @PrivatePointerInstance
type gauge struct {
	promGauge prometheus.Gauge
}

var _ Gauge = (*gauge)(nil)

func (g *gauge) Set(value float64) {
	g.promGauge.Set(value)
}

func (g *gauge) Inc() {
	g.promGauge.Inc()
}

func (g *gauge) Dec() {
	g.promGauge.Dec()
}

func (g *gauge) Add(delta float64) {
	g.promGauge.Add(delta)
}

func (g *gauge) Sub(delta float64) {
	g.promGauge.Sub(delta)
}

// @PrivatePointerInstance
type gaugeNop struct{}

var _ Gauge = (*gaugeNop)(nil)

func (g *gaugeNop) Set(value float64) {}

func (g *gaugeNop) Inc() {}

func (g *gaugeNop) Dec() {}

func (g *gaugeNop) Add(delta float64) {}

func (g *gaugeNop) Sub(delta float64) {}

// @PrivatePointerInstance
type counter struct {
	promCounter prometheus.Counter
}

var _ Counter = (*counter)(nil)

func (c *counter) Inc() {
	c.promCounter.Inc()
}

func (c *counter) Add(delta float64) {
	c.promCounter.Add(delta)
}

// @PrivatePointerInstance
type counterNop struct{}

var _ Counter = (*counterNop)(nil)

func (c *counterNop) Inc() {}

func (c *counterNop) Add(delta float64) {}

// @PrivatePointerInstance
type histogram struct {
	promHistogram prometheus.Observer
}

var _ Histogram = (*histogram)(nil)

func (h *histogram) Observe(value float64) {
	h.promHistogram.Observe(value)
}

// @PrivatePointerInstance
type histogramNop struct{}

var _ Histogram = (*histogramNop)(nil)

func (h *histogramNop) Observe(value float64) {}

// @PrivatePointerInstance
type summary struct {
	promSummary prometheus.Observer
}

var _ Summary = (*summary)(nil)

func (s *summary) Observe(value float64) {
	s.promSummary.Observe(value)
}

// @PrivatePointerInstance
type summaryNop struct{}

var _ Summary = (*summaryNop)(nil)

func (s *summaryNop) Observe(value float64) {}

type registry struct {
	promRegistry *prometheus.Registry

	metricLabels map[string][]string

	gauges     map[string]*prometheus.GaugeVec
	counters   map[string]*prometheus.CounterVec
	histograms map[string]*prometheus.HistogramVec
	summaries  map[string]*prometheus.SummaryVec
}

var _ Registry = (*registry)(nil)

func (r *registry) Gauge(ctx context.Context, metric gaugeMetric, labels ...label) Gauge {
	labels = allLabels(ctx, labels)

	gauge, ok := r.gauges[string(metric)]
	if !ok {
		return newGaugeNop()
	}

	metricLabels, ok := r.metricLabels[string(metric)]
	if !ok {
		return newGaugeNop()
	}

	promLabels := getLabelsForMetric(metricLabels, labels)

	gaugeWithLabels, err := gauge.GetMetricWith(promLabels)
	if err != nil {
		return newGaugeNop()
	}

	return newGauge(gaugeWithLabels)
}

func (r *registry) Counter(ctx context.Context, metric counterMetric, labels ...label) Counter {
	labels = allLabels(ctx, labels)

	counter, ok := r.counters[string(metric)]
	if !ok {
		return newCounterNop()
	}

	metricLabels, ok := r.metricLabels[string(metric)]
	if !ok {
		return newCounterNop()
	}

	promLabels := getLabelsForMetric(metricLabels, labels)

	counterWithLabels, err := counter.GetMetricWith(promLabels)
	if err != nil {
		return newCounterNop()
	}

	return newCounter(counterWithLabels)
}

func (r *registry) Histogram(ctx context.Context, metric histogramMetric, labels ...label) Histogram {
	labels = allLabels(ctx, labels)

	histogram, ok := r.histograms[string(metric)]
	if !ok {
		return newHistogramNop()
	}

	metricLabels, ok := r.metricLabels[string(metric)]
	if !ok {
		return newHistogramNop()
	}

	promLabels := getLabelsForMetric(metricLabels, labels)

	histogramWithLabels, err := histogram.GetMetricWith(promLabels)
	if err != nil {
		return newHistogramNop()
	}

	return newHistogram(histogramWithLabels)
}

func (r *registry) Summary(ctx context.Context, metric summaryMetric, labels ...label) Summary {
	labels = allLabels(ctx, labels)

	summary, ok := r.summaries[string(metric)]
	if !ok {
		return newSummaryNop()
	}

	metricLabels, ok := r.metricLabels[string(metric)]
	if !ok {
		return newSummaryNop()
	}

	promLabels := getLabelsForMetric(metricLabels, labels)

	summaryWithLabels, err := summary.GetMetricWith(promLabels)
	if err != nil {
		return newSummaryNop()
	}

	return newSummary(summaryWithLabels)
}

func getLabelsForMetric(metricLabels []string, allLabels []label) map[string]string {
	labels := make(map[string]string)

	for _, metricLabel := range metricLabels {
		labels[metricLabel] = ""
	}

	for _, label := range allLabels {
		if _, ok := labels[label.key]; !ok {
			continue
		}

		labels[label.key] = fmt.Sprintf("%v", label.value)
	}

	return labels
}

func InitRegistry(
	configHolder config.Holder,
	environmentHolder environment.Holder,
	logger Logger,
) (*registry, error) {
	logger.Info(context.Background(), "initializing metrics registry")

	serviceConfig := configHolder.Config().Service
	monitoringConfig := configHolder.Config().Monitoring
	environment := environmentHolder.Environment()

	baseLabels := []label{
		Environment(serviceConfig.Environment),
		Host(environment.HostName),
	}

	basePromLabels := []string{EnvironmentLabel, HostLabel}

	promRegistry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = promRegistry
	prometheus.DefaultGatherer = promRegistry

	gauges := make(map[string]*prometheus.GaugeVec)
	counters := make(map[string]*prometheus.CounterVec)
	histograms := make(map[string]*prometheus.HistogramVec)
	summaries := make(map[string]*prometheus.SummaryVec)
	metricLabels := make(map[string][]string)

	for _, gaugeConfig := range monitoringConfig.Gauges {
		gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: serviceConfig.Name,
			Name:      gaugeConfig.Name,
		}, append(basePromLabels, gaugeConfig.Labels...))

		gauge, err := gauge.CurryWith(labelsToProm(baseLabels))
		if err != nil {
			return nil, fmt.Errorf("adding base labels to gauge: %w", err)
		}

		err = promRegistry.Register(gauge)
		if err != nil {
			return nil, fmt.Errorf("registering gauge metric: %w", err)
		}

		gauges[gaugeConfig.Name] = gauge
		metricLabels[gaugeConfig.Name] = gaugeConfig.Labels
	}

	for _, counterConfig := range monitoringConfig.Counters {
		counter := prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: serviceConfig.Name,
			Name:      counterConfig.Name,
		}, append(basePromLabels, counterConfig.Labels...))

		counter, err := counter.CurryWith(labelsToProm(baseLabels))
		if err != nil {
			return nil, fmt.Errorf("adding base labels to counter: %w", err)
		}

		err = promRegistry.Register(counter)
		if err != nil {
			return nil, fmt.Errorf("registering counter metric: %w", err)
		}

		counters[counterConfig.Name] = counter
		metricLabels[counterConfig.Name] = counterConfig.Labels
	}

	for _, histogramConfig := range monitoringConfig.Histograms {
		histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: serviceConfig.Name,
			Name:      histogramConfig.Name,
			Buckets:   histogramConfig.Buckets,
		}, append(basePromLabels, histogramConfig.Labels...))

		observer, err := histogram.CurryWith(labelsToProm(baseLabels))
		if err != nil {
			return nil, fmt.Errorf("adding base labels to histogram: %w", err)
		}

		histogram = observer.(*prometheus.HistogramVec)

		err = promRegistry.Register(histogram)
		if err != nil {
			return nil, fmt.Errorf("registering histogram metric: %w", err)
		}

		histograms[histogramConfig.Name] = histogram
		metricLabels[histogramConfig.Name] = histogramConfig.Labels
	}

	for _, summaryConfig := range monitoringConfig.Summaries {
		summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  serviceConfig.Name,
			Name:       summaryConfig.Name,
			Objectives: summaryConfig.Objectives,
			MaxAge:     summaryConfig.MaxAge,
			AgeBuckets: summaryConfig.AgeBuckets,
		}, append(basePromLabels, summaryConfig.Labels...))

		observer, err := summary.CurryWith(labelsToProm(baseLabels))
		if err != nil {
			return nil, fmt.Errorf("adding base labels to summary: %w", err)
		}

		summary = observer.(*prometheus.SummaryVec)

		err = promRegistry.Register(summary)
		if err != nil {
			return nil, fmt.Errorf("registering summary metric: %w", err)
		}

		summaries[summaryConfig.Name] = summary
		metricLabels[summaryConfig.Name] = summaryConfig.Labels
	}

	logger.Info(context.Background(), "metrics registry successfully initialized")

	return &registry{
		promRegistry: promRegistry,
		gauges:       gauges,
		counters:     counters,
		histograms:   histograms,
		summaries:    summaries,
		metricLabels: metricLabels,
	}, nil
}
