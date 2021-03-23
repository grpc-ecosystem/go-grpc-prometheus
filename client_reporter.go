// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

// Forked originally form https://github.com/grpc-ecosystem/go-grpc-prometheus/
// the very same thing with https://github.com/grpc-ecosystem/go-grpc-prometheus/pull/88 integrated
// for the additional functionality to monitore bytes received and send from clients or servers
// eveything that is in between a "---- PR-88 ---- {"  and   "---- PR-88 ---- }" comment is the new addition from the PR88.

package grpc_prometheus

import (
	"fmt" // PR-88

	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
)

type clientReporter struct {
	metrics     *ClientMetrics
	rpcType     grpcType
	serviceName string
	methodName  string
	startTime   time.Time
}

func newClientReporter(m *ClientMetrics, rpcType grpcType, fullMethod string) *clientReporter {
	r := &clientReporter{
		metrics: m,
		rpcType: rpcType,
	}
	if r.metrics.clientHandledHistogramEnabled {
		r.startTime = time.Now()
	}
	r.serviceName, r.methodName = splitMethodName(fullMethod)
	r.metrics.clientStartedCounter.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
	return r
}

// ---- PR-88 ---- {

func newClientReporterForStatsHanlder(startTime time.Time, m *ClientMetrics, fullMethod string) *clientReporter {
	r := &clientReporter{
		metrics:   m,
		rpcType:   Unary,
		startTime: startTime,
	}
	r.serviceName, r.methodName = splitMethodName(fullMethod)
	return r
}

// ---- PR-88 ---- }

// timer is a helper interface to time functions.
type timer interface {
	ObserveDuration() time.Duration
}

type noOpTimer struct {
}

func (noOpTimer) ObserveDuration() time.Duration {
	return 0
}

var emptyTimer = noOpTimer{}

func (r *clientReporter) ReceiveMessageTimer() timer {
	if r.metrics.clientStreamRecvHistogramEnabled {
		hist := r.metrics.clientStreamRecvHistogram.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName)
		return prometheus.NewTimer(hist)
	}

	return emptyTimer
}

// ---- PR-88 ---- {

func (r *clientReporter) StartedConn() {
	r.metrics.clientStartedCounter.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
}

// ---- PR-88 ---- }

func (r *clientReporter) ReceivedMessage() {
	r.metrics.clientStreamMsgReceived.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
}

// ---- PR-88 ---- {

// ReceivedMessageSize counts the size of received messages on client-side
func (r *clientReporter) ReceivedMessageSize(rpcStats grpcStats, size float64) {
	if rpcStats == Payload {
		r.ReceivedMessage()
	}

	if r.metrics.clientMsgSizeReceivedHistogramEnabled {
		r.metrics.clientMsgSizeReceivedHistogram.WithLabelValues(r.serviceName, r.methodName, string(rpcStats)).Observe(size)
	}
}

// ---- PR-88 ---- }

func (r *clientReporter) SendMessageTimer() timer {
	if r.metrics.clientStreamSendHistogramEnabled {
		hist := r.metrics.clientStreamSendHistogram.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName)
		return prometheus.NewTimer(hist)
	}

	return emptyTimer
}

func (r *clientReporter) SentMessage() {
	r.metrics.clientStreamMsgSent.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
}

// ---- PR-88 ---- {

func (r *clientReporter) SentMessageSize(rpcStats grpcStats, size float64) {
	if rpcStats == Payload {
		r.SentMessage()
	}

	if r.metrics.clientMsgSizeSentHistogramEnabled {
		r.metrics.clientMsgSizeSentHistogram.WithLabelValues(r.serviceName, r.methodName, string(rpcStats)).Observe(size)
	}
}

// ---- PR-88 ---- }

func (r *clientReporter) Handled(code codes.Code) {
	r.metrics.clientHandledCounter.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName, code.String()).Inc()
	if r.metrics.clientHandledHistogramEnabled {
		fmt.Printf("client handled count + 1: %v,%f\n", code, time.Since(r.startTime).Seconds()) // PR88
		r.metrics.clientHandledHistogram.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Observe(time.Since(r.startTime).Seconds())
	}
}
