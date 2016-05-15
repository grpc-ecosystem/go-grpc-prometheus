// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_prometheus

import (
	"time"

	"google.golang.org/grpc/codes"

	prom "github.com/prometheus/client_golang/prometheus"
)

type rpcType string

const (
	Unary rpcType = "unary"
	ClientStream rpcType = "client_stream"
	ServerStream rpcType = "server_stream"
	BidiStream rpcType = "bidi_stream"
)

var (
	serverStartedCounter = prom.NewCounterVec(
		prom.CounterOpts{
			Namespace: "grpc",
			Subsystem: "server",
			Name:      "rpc_started_total",
			Help:      "Total number of RPCs started on the server.",
		}, []string{"type", "service", "method"})

	serverHandledCounter = prom.NewCounterVec(
		prom.CounterOpts{
			Namespace: "grpc",
			Subsystem: "server",
			Name:      "rpc_handled_total",
			Help:      "Total number of RPCs completed on the server, regardless of success or failure.",
		}, []string{"type", "service", "method", "code"})

	serverStreamMsgReceived = prom.NewCounterVec(
		prom.CounterOpts{
			Namespace: "grpc",
			Subsystem: "server",
			Name:      "rpc_msg_received_total",
			Help:      "Total number of RPC stream messages received on the server.",
		}, []string{"type", "service", "method"})

	serverStreamMsgSent = prom.NewCounterVec(
		prom.CounterOpts{
			Namespace: "grpc",
			Subsystem: "server",
			Name:      "rpc_msg_sent_total",
			Help:      "Total number of RPC stream messages sent by the server.",
		}, []string{"type", "service", "method"})

	serverHandledHistogramEnabled = false
	serverHandledHistogram = prom.NewHistogramVec(
		prom.HistogramOpts{
			Namespace: "grpc",
			Subsystem: "server",
			Name:      "rpc_handling_seconds",
			Help:      "Histogram of response latency (seconds) of RPC that had been application-level handled by the server.",
			Buckets:   prom.DefBuckets,
		}, []string{"type", "service", "method"})
)

func init() {
	prom.MustRegister(serverStartedCounter)
	prom.MustRegister(serverHandledCounter)
	prom.MustRegister(serverStreamMsgReceived)
	prom.MustRegister(serverStreamMsgSent)
}

// EnableHandlingTimeHistogram turns on recording of handling time of RPCs.
// Histogram metrics can be very expensive for Prometheus to retain and query.
func EnableHandlingTimeHistogram() {
	if !serverHandledHistogramEnabled {
		prom.Register(serverHandledHistogram)
	}
	serverHandledHistogramEnabled = true
}

type serverReporter struct {
	rpcType     rpcType
	serviceName string
	methodName  string
	startTime   time.Time
}

func newServerReporter(rpcType rpcType, fullMethod string) *serverReporter {
	r := &serverReporter{rpcType: rpcType}
	if serverHandledHistogramEnabled {
		r.startTime = time.Now()
	}
	r.serviceName, r.methodName = splitMethodName(fullMethod)
	serverStartedCounter.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
	return r
}

func (r *serverReporter) ReceivedMessage() {
	serverStreamMsgReceived.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
}

func (r *serverReporter) SentMessage() {
	serverStreamMsgSent.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
}

func (r *serverReporter) Handled(code codes.Code) {
	serverHandledCounter.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName, code.String()).Inc()
	if serverHandledHistogramEnabled {
		serverHandledHistogram.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Observe(time.Since(r.startTime).Seconds())
	}
}
