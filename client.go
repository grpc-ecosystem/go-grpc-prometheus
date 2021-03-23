// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

// gRPC Prometheus monitoring interceptors for client-side gRPC.

// Forked originally form https://github.com/grpc-ecosystem/go-grpc-prometheus/
// the very same thing with https://github.com/grpc-ecosystem/go-grpc-prometheus/pull/88 integrated
// for the additional functionality to monitore bytes received and send from clients or servers
// eveything that is in between a " ---- PR-88 ---- {"  and   "---- PR-88 ---- }" comment is the new addition from the PR88.

package grpc_prometheus

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

var (
	// DefaultClientMetrics is the default instance of ClientMetrics. It is
	// intended to be used in conjunction the default Prometheus metrics
	// registry.
	DefaultClientMetrics = NewClientMetrics()

	// UnaryClientInterceptor is a gRPC client-side interceptor that provides Prometheus monitoring for Unary RPCs.
	UnaryClientInterceptor = DefaultClientMetrics.UnaryClientInterceptor()

	// StreamClientInterceptor is a gRPC client-side interceptor that provides Prometheus monitoring for Streaming RPCs.
	StreamClientInterceptor = DefaultClientMetrics.StreamClientInterceptor()

	// ---- PR-88 ---- {

	// ClientStatsHandler is a gRPC client-side stats.Handler that provides Prometheus monitoring for RPCs.
	ClientStatsHandler = DefaultClientMetrics.NewClientStatsHandler()

	// ---- PR-88 ---- }

)

func init() {
	prom.MustRegister(DefaultClientMetrics.clientStartedCounter)
	prom.MustRegister(DefaultClientMetrics.clientHandledCounter)
	prom.MustRegister(DefaultClientMetrics.clientStreamMsgReceived)
	prom.MustRegister(DefaultClientMetrics.clientStreamMsgSent)
}

// EnableClientHandlingTimeHistogram turns on recording of handling time of
// RPCs. Histogram metrics can be very expensive for Prometheus to retain and
// query. This function acts on the DefaultClientMetrics variable and the
// default Prometheus metrics registry.
func EnableClientHandlingTimeHistogram(opts ...HistogramOption) {
	DefaultClientMetrics.EnableClientHandlingTimeHistogram(opts...)
	prom.Register(DefaultClientMetrics.clientHandledHistogram)
}

// EnableClientStreamReceiveTimeHistogram turns on recording of
// single message receive time of streaming RPCs.
// This function acts on the DefaultClientMetrics variable and the
// default Prometheus metrics registry.
func EnableClientStreamReceiveTimeHistogram(opts ...HistogramOption) {
	DefaultClientMetrics.EnableClientStreamReceiveTimeHistogram(opts...)
	prom.Register(DefaultClientMetrics.clientStreamRecvHistogram)
}

// EnableClientStreamSendTimeHistogram turns on recording of
// single message send time of streaming RPCs.
// This function acts on the DefaultClientMetrics variable and the
// default Prometheus metrics registry.
func EnableClientStreamSendTimeHistogram(opts ...HistogramOption) {
	DefaultClientMetrics.EnableClientStreamSendTimeHistogram(opts...)
	prom.Register(DefaultClientMetrics.clientStreamSendHistogram)
}

// ---- PR-88 ---- {

// EnableClientMsgSizeReceivedBytesHistogram turns on recording of
// single message send time of streaming RPCs.
// This function acts on the DefaultClientMetrics variable and the
// default Prometheus metrics registry.
func EnableClientMsgSizeReceivedBytesHistogram(opts ...HistogramOption) {
	DefaultClientMetrics.EnableMsgSizeReceivedBytesHistogram(opts...)
	prom.Register(DefaultClientMetrics.clientMsgSizeReceivedHistogram)
}

// EnableClientMsgSizeSentBytesHistogram turns on recording of
// single message send time of streaming RPCs.
// This function acts on the DefaultClientMetrics variable and the
// default Prometheus metrics registry.
func EnableClientMsgSizeSentBytesHistogram(opts ...HistogramOption) {
	DefaultClientMetrics.EnableMsgSizeSentBytesHistogram(opts...)
	prom.Register(DefaultClientMetrics.clientMsgSizeSentHistogram)

}

// ---- PR-88 ---- }
