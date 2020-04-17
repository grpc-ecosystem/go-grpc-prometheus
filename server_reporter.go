// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_prometheus

import (
	"time"

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

func newServerReporterForStatsHanlder(startTime time.Time, m *ServerMetrics, fullMethod string) *serverReporter {
	r := &serverReporter{
		metrics:   m,
		rpcType:   Unary,
		startTime: startTime,
	}
	r.serviceName, r.methodName = splitMethodName(fullMethod)
	return r
}

func (r *serverReporter) StartedConn() {
	r.metrics.serverStartedCounter.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
}

func (r *serverReporter) ReceivedMessage() {
	r.metrics.serverStreamMsgReceived.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
}

// ReceivedMessageSize counts the size of received messages on server-side
func (r *serverReporter) ReceivedMessageSize(rpcStats grpcStats, size float64) {
	if rpcStats == Payload {
		r.ReceivedMessage()
	}
	if r.metrics.serverMsgSizeReceivedHistogramEnabled {
		r.metrics.serverMsgSizeReceivedHistogram.WithLabelValues(r.serviceName, r.methodName, rpcStats.String()).Observe(size)
	}
}

func (r *serverReporter) SentMessage() {
	r.metrics.serverStreamMsgSent.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
}

// SentMessageSize counts the size of sent messages on server-side
func (r *serverReporter) SentMessageSize(rpcStats grpcStats, size float64) {
	if rpcStats == Payload {
		r.SentMessage()
	}
	if r.metrics.serverMsgSizeSentHistogramEnabled {
		r.metrics.serverMsgSizeSentHistogram.WithLabelValues(r.serviceName, r.methodName, rpcStats.String()).Observe(size)
	}
}

func (r *serverReporter) Handled(code codes.Code) {
	r.metrics.serverHandledCounter.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName, code.String()).Inc()
	if r.metrics.serverHandledHistogramEnabled {
		r.metrics.serverHandledHistogram.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Observe(time.Since(r.startTime).Seconds())
	}
}
