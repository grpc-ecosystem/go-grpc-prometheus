package grpc_prometheus

import (
	"io"

	prom "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ClientMetrics represents a collection of monitor to be registered on a
// Prometheus monitor registry for a gRPC client.
type ClientMetrics struct {
	*clientMonitor
}

// NewClientMetrics returns a ClientMetrics object. Use a new instance of
// ClientMetrics when not using the default Prometheus monitor registry, for
// example when wanting to control which monitor are added to a registry as
// opposed to automatically adding monitor via init functions.
func NewClientMetrics(opts ...CollectorOption) *ClientMetrics {
	var collectorOptions collectorOptions
	for _, fn := range opts {
		fn(&collectorOptions)
	}

	if collectorOptions.namespace == "" {
		collectorOptions.namespace = namespace
	}
	if collectorOptions.subsystem == "" {
		collectorOptions.subsystem = "client"
	}

	return &ClientMetrics{
		clientMonitor: initClientMonitor(
			prom.Opts{
				Namespace:   collectorOptions.namespace,
				Subsystem:   collectorOptions.subsystem,
				ConstLabels: collectorOptions.constLabels,
			},
			collectorOptions.buckets,
		),
	}
}

// UnaryInterceptor is a gRPC client-side interceptor that provides Prometheus monitoring for Unary RPCs.
func (m *ClientMetrics) UnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		rep := newClientReporter(m.clientMonitor, Unary, method)
		rep.outgoingRequest()
		rep.outgoingMessage() // TODO: necessary?

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			rep.incomingMessage()
		}
		st, _ := status.FromError(err)
		rep.incomingResponse(st.Code())
		return err
	}
}

// StreamInterceptor is a gRPC client-side interceptor that provides Prometheus monitoring for Streaming RPCs.
func (m *ClientMetrics) StreamInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		rep := newClientReporter(m.clientMonitor, clientStreamType(desc), method)
		rep.outgoingRequest()

		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			st, _ := status.FromError(err)
			rep.incomingResponse(st.Code())
			return nil, err
		}
		return &monitoredClientStream{clientStream, rep}, nil
	}
}

func clientStreamType(desc *grpc.StreamDesc) grpcType {
	if desc.ClientStreams && !desc.ServerStreams {
		return ClientStream
	} else if !desc.ClientStreams && desc.ServerStreams {
		return ServerStream
	}
	return BidiStream
}

// monitoredClientStream wraps grpc.ClientStream allowing each Sent/Recv of message to increment counters.
type monitoredClientStream struct {
	grpc.ClientStream
	monitor *clientReporter
}

func (s *monitoredClientStream) SendMsg(m interface{}) error {
	err := s.ClientStream.SendMsg(m)
	if err == nil {
		s.monitor.outgoingMessage()
	}
	return err
}

func (s *monitoredClientStream) RecvMsg(m interface{}) error {
	err := s.ClientStream.RecvMsg(m)
	if err == nil {
		s.monitor.incomingMessage()
	} else if err == io.EOF {
		s.monitor.incomingResponse(codes.OK)
	} else {
		st, _ := status.FromError(err)
		s.monitor.incomingResponse(st.Code())
	}
	return err
}
