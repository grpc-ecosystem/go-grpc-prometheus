package grpc_prometheus

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

// CollectorOption lets you add options to monitor using With* funcs.
type CollectorOption func(*collectorOptions)

type collectorOptions struct {
	namespace, subsystem string
	constLabels          prom.Labels
	buckets              []float64
}

type counterOptions []CollectorOption

// TODO: remove once ServerMetrics is changed.
func (co counterOptions) apply(base prom.Opts) prom.Opts {
	var opts collectorOptions
	for _, f := range co {
		f(&opts)
	}
	if opts.namespace != "" {
		base.Namespace = opts.namespace
	}
	if opts.subsystem != "" {
		base.Subsystem = opts.subsystem
	}
	if opts.constLabels != nil {
		base.ConstLabels = opts.constLabels
	}
	return base
}

// WithConstLabels allows you to add ConstLabels to Counter monitor.
func WithConstLabels(labels prom.Labels) CollectorOption {
	return func(o *collectorOptions) {
		o.constLabels = labels
	}
}

// WithNamespace allows to change default subsystem.
func WithNamespace(namespace string) CollectorOption {
	return func(o *collectorOptions) {
		o.namespace = namespace
	}
}

// WithSubsystem allows to change default subsystem.
func WithSubsystem(subsystem string) CollectorOption {
	return func(o *collectorOptions) {
		o.subsystem = subsystem
	}
}

// WithBuckets allows you to specify custom bucket ranges for histograms.
func WithBuckets(buckets []float64) CollectorOption {
	return func(o *collectorOptions) {
		o.buckets = buckets
	}
}
