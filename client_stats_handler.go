package grpc_prometheus

import (
	"context"

	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

type clientStatsHandler struct {
	clientMetrics *ClientMetrics
}

// TagRPC implements the stats.Hanlder interface.
func (h *clientStatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	rpcInfo := newRPCInfo(info.FullMethodName)
	return context.WithValue(ctx, &rpcInfoKey, rpcInfo)
}

// HandleRPC implements the stats.Hanlder interface.
func (h *clientStatsHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	v, ok := ctx.Value(&rpcInfoKey).(*rpcInfo)
	if !ok {
		return
	}
	monitor := newClientReporterForStatsHanlder(v.startTime, h.clientMetrics, v.fullMethodName)
	switch s := s.(type) {
	case *stats.Begin:
		v.startTime = s.BeginTime
		monitor.StartedConn()
	case *stats.End:
		monitor.Handled(status.Code(s.Error))
	case *stats.InHeader:
		monitor.ReceivedMessageSize(Header, float64(s.WireLength))
	case *stats.InPayload:
		// TODO: remove the +5 offset on wire length here, which is a temporary stand-in for the missing grpc framing offset
		//  See: https://github.com/grpc/grpc-go/issues/1647
		monitor.ReceivedMessageSize(Payload, float64(s.WireLength+5))
	case *stats.InTrailer:
		monitor.ReceivedMessageSize(Tailer, float64(s.WireLength))
	case *stats.OutHeader:
		// TODO: Add the sent header message size stats, if the wire length of the send header is provided
	case *stats.OutPayload:
		// TODO(tonywang): response latency (seconds) of the gRPC single message send
		monitor.SentMessageSize(Payload, float64(s.WireLength))
	case *stats.OutTrailer:
		monitor.SentMessageSize(Tailer, float64(s.WireLength))
	}
}

// TagConn implements the stats.Hanlder interface.
func (h *clientStatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	return ctx
}

// HandleConn implements the stats.Hanlder interface.
func (h *clientStatsHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
}
