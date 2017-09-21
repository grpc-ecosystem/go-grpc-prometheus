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

func WithSubsystem(subsystem string) CounterOption {
	return func(o *prom.CounterOpts) {
		o.Subsystem = subsystem
	}
}

func WithNamespace(ns string) CounterOption {
	return func(o *prom.CounterOpts) {
		o.Namespace = ns
	}
}
