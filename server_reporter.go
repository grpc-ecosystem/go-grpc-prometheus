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
	Unary     rpcType = "unary"
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
			Help:      "Total number of RPCs started by the server.",
		}, []string{"type", "service", "method"})

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

	serverHandledHistogram = prom.NewHistogramVec(
		prom.HistogramOpts{
			Namespace: "grpc",
			Subsystem: "server",
			Name:      "rpc_handled",
			Help:      "Histogram of response latency of RPC that had been application-level handled by the server.",
			Buckets:   prom.DefBuckets,
		}, []string{"type", "service", "method", "code"})
)

func init() {
	prom.MustRegister(serverStartedCounter)
	prom.MustRegister(serverStreamMsgReceived)
	prom.MustRegister(serverStreamMsgSent)
	prom.MustRegister(serverHandledHistogram)
}

type serverReporter struct {
	rpcType     rpcType
	serviceName string
	methodName  string
	startTime   time.Time
}

func newServerReporter(rpcType rpcType, fullMethod string) *serverReporter {
	r := &serverReporter{rpcType: rpcType, startTime: time.Now()}
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
	serverHandledHistogram.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName, code.String()).Observe(time.Since(r.startTime).Seconds())
}
