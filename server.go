// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

// gRPC Prometheus monitoring interceptors for server-side gRPC.

// Forked originally form https://github.com/grpc-ecosystem/go-grpc-prometheus/
// the very same thing with https://github.com/grpc-ecosystem/go-grpc-prometheus/pull/88 integrated
// for the additional functionality to monitore bytes received and send from clients or servers
// eveything that is in between a "---- PR-88 ---- {"  and   "---- PR-88 ---- }" comment is the new addition from the PR88.

package grpc_prometheus

import (
	prom "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

var (
	// DefaultServerMetrics is the default instance of ServerMetrics. It is
	// intended to be used in conjunction the default Prometheus metrics
	// registry.
	DefaultServerMetrics = NewServerMetrics()

	// UnaryServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Unary RPCs.
	UnaryServerInterceptor = DefaultServerMetrics.UnaryServerInterceptor()

	// StreamServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Streaming RPCs.
	StreamServerInterceptor = DefaultServerMetrics.StreamServerInterceptor()

	// ServerStatsHandler is a gRPC server-side stats.Handler that provides Prometheus monitoring
	ServerStatsHandler = DefaultServerMetrics.NewServerStatsHandler() //new
)

func init() {
	prom.MustRegister(DefaultServerMetrics.serverStartedCounter)
	prom.MustRegister(DefaultServerMetrics.serverHandledCounter)
	prom.MustRegister(DefaultServerMetrics.serverStreamMsgReceived)
	prom.MustRegister(DefaultServerMetrics.serverStreamMsgSent)
}

// Register takes a gRPC server and pre-initializes all counters to 0. This
// allows for easier monitoring in Prometheus (no missing metrics), and should
// be called *after* all services have been registered with the server. This
// function acts on the DefaultServerMetrics variable.
func Register(server *grpc.Server) {
	DefaultServerMetrics.InitializeMetrics(server)
}

// EnableHandlingTimeHistogram turns on recording of handling time
// of RPCs. Histogram metrics can be very expensive for Prometheus
// to retain and query. This function acts on the DefaultServerMetrics
// variable and the default Prometheus metrics registry.
func EnableHandlingTimeHistogram(opts ...HistogramOption) {
	DefaultServerMetrics.EnableHandlingTimeHistogram(opts...)
	prom.Register(DefaultServerMetrics.serverHandledHistogram)
}

// ---- PR-88 ---- {

// EnableServerMsgSizeReceivedBytesHistogram turns on recording of handling time
// of RPCs. Histogram metrics can be very expensive for Prometheus
// to retain and query. This function acts on the DefaultServerMetrics
// variable and the default Prometheus metrics registry.
func EnableServerMsgSizeReceivedBytesHistogram(opts ...HistogramOption) {
	DefaultServerMetrics.EnableMsgSizeReceivedBytesHistogram(opts...)
	prom.Register(DefaultServerMetrics.serverMsgSizeReceivedHistogram)
}

// EnableServerMsgSizeSentBytesHistogram turns on recording of handling time
// of RPCs. Histogram metrics can be very expensive for Prometheus
// to retain and query. This function acts on the DefaultServerMetrics
// variable and the default Prometheus metrics registry.
func EnableServerMsgSizeSentBytesHistogram(opts ...HistogramOption) {
	DefaultServerMetrics.EnableMsgSizeSentBytesHistogram(opts...)
	prom.Register(DefaultServerMetrics.serverMsgSizeSentHistogram)
}

// ---- PR-88 ---- }
