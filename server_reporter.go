// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

// Forked originally form https://github.com/grpc-ecosystem/go-grpc-prometheus/
// the very same thing with https://github.com/grpc-ecosystem/go-grpc-prometheus/pull/88 integrated
// for the additional functionality to monitore bytes received and send from clients or servers
// eveything that is in between a "---- PR-88 ---- {"  and   "---- PR-88 ---- }" comment is the new addition from the PR88.

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

// ---- PR-88 ---- {

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

// ---- PR-88 ---- }

func (r *serverReporter) ReceivedMessage() {
	r.metrics.serverStreamMsgReceived.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
}

// ---- PR-88 ---- {

// ReceivedMessageSize counts the size of received messages on server-side
func (r *serverReporter) ReceivedMessageSize(rpcStats grpcStats, size float64) {
	if rpcStats == Payload {
		r.ReceivedMessage()
	}
	if r.metrics.serverMsgSizeReceivedHistogramEnabled {
		r.metrics.serverMsgSizeReceivedHistogram.WithLabelValues(r.serviceName, r.methodName, rpcStats.String()).Observe(size)
	}
}

// ---- PR-88 ---- }

func (r *serverReporter) SentMessage() {
	r.metrics.serverStreamMsgSent.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
}

// ---- PR-88 ---- {

// SentMessageSize counts the size of sent messages on server-side
func (r *serverReporter) SentMessageSize(rpcStats grpcStats, size float64) {
	if rpcStats == Payload {
		r.SentMessage()
	}
	if r.metrics.serverMsgSizeSentHistogramEnabled {
		r.metrics.serverMsgSizeSentHistogram.WithLabelValues(r.serviceName, r.methodName, rpcStats.String()).Observe(size)
	}
}

// ---- PR-88 ---- }

func (r *serverReporter) Handled(code codes.Code) {
	r.metrics.serverHandledCounter.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName, code.String()).Inc()
	if r.metrics.serverHandledHistogramEnabled {
		r.metrics.serverHandledHistogram.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Observe(time.Since(r.startTime).Seconds())
	}
}
