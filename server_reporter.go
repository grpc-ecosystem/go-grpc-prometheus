// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
)

type serverReporter struct {
	metrics     *ServerMetrics
	rpcType     grpcType
	serviceName string
	methodName  string
	startTime   time.Time
}

func newServerReporter(m *ServerMetrics, rpcType grpcType, fullMethod string) *serverReporter {
	r := &serverReporter{
		metrics: m,
		rpcType: rpcType,
	}
	if r.metrics.serverHandledHistogramEnabled {
		r.startTime = time.Now()
	}
	r.serviceName, r.methodName = splitMethodName(fullMethod)
	r.metrics.serverStartedCounter.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
	return r
}

func (r *serverReporter) ReceiveMessageTimer() timer {
	if r.metrics.serverStreamRecvHistogramEnabled {
		hist := r.metrics.serverStreamRecvHistogram.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName)
		return prometheus.NewTimer(hist)
	}

	return emptyTimer
}

func (r *serverReporter) ReceivedMessage() {
	r.metrics.serverStreamMsgReceived.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
}

func (r *serverReporter) SendMessageTimer() timer {
	if r.metrics.serverStreamSendHistogramEnabled {
		hist := r.metrics.serverStreamSendHistogram.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName)
		return prometheus.NewTimer(hist)
	}

	return emptyTimer
}

func (r *serverReporter) SentMessage() {
	r.metrics.serverStreamMsgSent.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
}

func (r *serverReporter) Handled(code codes.Code) {
	r.metrics.serverHandledCounter.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName, code.String()).Inc()
	if r.metrics.serverHandledHistogramEnabled {
		r.metrics.serverHandledHistogram.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Observe(time.Since(r.startTime).Seconds())
	}
}
