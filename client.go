// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

// gRPC Prometheus monitoring interceptors for client-side gRPC.

package grpc_prometheus

var (
//// DefaultClientMetrics is the default instance of ClientMetrics. It is
//// intended to be used in conjunction the default Prometheus monitor
//// registry.
//DefaultClientMetrics = NewClientMetrics()
//
//// UnaryInterceptor is a gRPC client-side interceptor that provides Prometheus monitoring for Unary RPCs.
//UnaryClientInterceptor = DefaultClientMetrics.UnaryInterceptor()
//
//// StreamInterceptor is a gRPC client-side interceptor that provides Prometheus monitoring for Streaming RPCs.
//StreamClientInterceptor = DefaultClientMetrics.StreamInterceptor()
)

func init() {
	//prom.MustRegister(DefaultClientMetrics)
	//prom.MustRegister(DefaultClientMetrics.clientHandledCounter)
	//prom.MustRegister(DefaultClientMetrics.clientStreamMsgReceived)
	//prom.MustRegister(DefaultClientMetrics.clientStreamMsgSent)
}

// EnableClientHandlingTimeHistogram turns on recording of handling time of
// RPCs. Histogram monitor can be very expensive for Prometheus to retain and
// query. This function acts on the DefaultClientMetrics variable and the
// default Prometheus monitor registry.
//func EnableClientHandlingTimeHistogram(opts ...HistogramCollectorOption) {
//DefaultClientMetrics.EnableClientHandlingTimeHistogram(opts...)
//prom.Register(DefaultClientMetrics.clientHandledHistogram)
//}
