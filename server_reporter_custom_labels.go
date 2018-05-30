// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_prometheus

import (
	"time"

	"google.golang.org/grpc/codes"
	"context"
)

type serverReporterCustomLabels struct {
	metrics     *ServerMetricsCustomLabels
	rpcType     grpcType
	serviceName string
	methodName  string
	startTime   time.Time
	labelValues []string
}

func newServerReporterCustomLabels(ctx context.Context, m *ServerMetricsCustomLabels, rpcType grpcType, fullMethod string) *serverReporterCustomLabels {
	r := &serverReporterCustomLabels{
		metrics: m,
		rpcType: rpcType,
	}

	if r.metrics.serverHandledHistogramEnabled {
		r.startTime = time.Now()
	}

	r.serviceName, r.methodName = splitMethodName(fullMethod)
	labelValues := make([]string, len(m.customLabels), len(m.customLabels))
	copy(labelValues, []string{string(r.rpcType), r.serviceName, r.methodName})
	customLabelValues := m.customLabelsProvider.LabelsFromContext(ctx)
	for i := len(defaultLables); i < len(m.customLabels); i++ {
		si := i - len(defaultLables) //shifted index
		if si >= len(customLabelValues) {
			break
		}
		labelValues[i] = customLabelValues[si]
	}
	r.labelValues = labelValues

	r.metrics.serverStartedCounter.WithLabelValues(r.labelValues...).Inc()
	return r
}

func (r *serverReporterCustomLabels) ReceivedMessage() {
	r.metrics.serverStreamMsgReceived.WithLabelValues(r.labelValues...).Inc()
}

func (r *serverReporterCustomLabels) SentMessage() {
	r.metrics.serverStreamMsgSent.WithLabelValues(r.labelValues...).Inc()
}

func (r *serverReporterCustomLabels) Handled(code codes.Code) {
	lvs := r.labelValues
	lvs = append(lvs, code.String())
	r.metrics.serverHandledCounter.WithLabelValues(lvs...).Inc()
	if r.metrics.serverHandledHistogramEnabled {
		r.metrics.serverHandledHistogram.WithLabelValues(r.labelValues...).Observe(time.Since(r.startTime).Seconds())
	}
}
