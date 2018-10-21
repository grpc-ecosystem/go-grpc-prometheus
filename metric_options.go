package grpc_prometheus

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

type options struct {
	namespace, subsystem string
	constLabels          prom.Labels
	buckets              []float64
}

// A CollectorOption lets you add options to monitor using With* funcs.
type CollectorOption func(*options)

type counterOptions []CollectorOption

func (co counterOptions) apply(base prom.Opts) prom.Opts {
	var opts options
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
	return func(o *options) {
		o.constLabels = labels
	}
}

// WithSubsystem allows to change default subsystem.
func WithNamespace(namespace string) CollectorOption {
	return func(o *options) {
		o.namespace = namespace
	}
}

// WithSubsystem allows to change default subsystem.
func WithSubsystem(subsystem string) CollectorOption {
	return func(o *options) {
		o.subsystem = subsystem
	}
}

// WithBuckets allows you to specify custom bucket ranges for histograms.
func WithBuckets(buckets []float64) CollectorOption {
	return func(o *options) {
		o.buckets = buckets
	}
}
