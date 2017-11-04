package grpc_prometheus

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

type CounterOption func(opts *prom.CounterOpts)

type counterOptions []CounterOption

func (co counterOptions) apply(o prom.CounterOpts) prom.CounterOpts {
	for _, f := range co {
		f(&o)
	}
	return o
}

func WithConstLabels(labels prom.Labels) CounterOption {
	return func(o *prom.CounterOpts) {
		o.ConstLabels = labels
	}
}

type HistogramOption func(*prom.HistogramOpts)

// WithHistogramBuckets allows you to specify custom bucket ranges for histograms if EnableHandlingTimeHistogram is on.
func WithHistogramBuckets(buckets []float64) HistogramOption {
	return func(o *prom.HistogramOpts) { o.Buckets = buckets }
}

func WithHistogramConstLabels(labels prom.Labels) HistogramOption {
	return func(o *prom.HistogramOpts) {
		o.ConstLabels = labels
	}
}
