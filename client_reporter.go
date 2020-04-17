// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_prometheus

import (
	"fmt"
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

func newClientReporterForStatsHanlder(startTime time.Time, m *ClientMetrics, fullMethod string) *clientReporter {
	r := &clientReporter{
		metrics:   m,
		rpcType:   Unary,
		startTime: startTime,
	}
	r.serviceName, r.methodName = splitMethodName(fullMethod)
	return r
}

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

func (r *clientReporter) StartedConn() {
	r.metrics.clientStartedCounter.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
}

func (r *clientReporter) ReceivedMessage() {
	r.metrics.clientStreamMsgReceived.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
}

// ReceivedMessageSize counts the size of received messages on client-side
func (r *clientReporter) ReceivedMessageSize(rpcStats grpcStats, size float64) {
	if rpcStats == Payload {
		r.ReceivedMessage()
	}

	if r.metrics.clientMsgSizeReceivedHistogramEnabled {
		r.metrics.clientMsgSizeReceivedHistogram.WithLabelValues(r.serviceName, r.methodName, string(rpcStats)).Observe(size)
	}
}

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

// SentMessageSize counts the size of sent messages on client-side
func (r *clientReporter) SentMessageSize(rpcStats grpcStats, size float64) {
	if rpcStats == Payload {
		r.SentMessage()
	}

	if r.metrics.clientMsgSizeSentHistogramEnabled {
		r.metrics.clientMsgSizeSentHistogram.WithLabelValues(r.serviceName, r.methodName, string(rpcStats)).Observe(size)
	}
}

func (r *clientReporter) Handled(code codes.Code) {
	r.metrics.clientHandledCounter.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName, code.String()).Inc()
	if r.metrics.clientHandledHistogramEnabled {
		fmt.Printf("client handled count + 1: %v,%f\n", code, time.Since(r.startTime).Seconds())
		r.metrics.clientHandledHistogram.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Observe(time.Since(r.startTime).Seconds())
	}
}
