package grpc_prometheus

import (
	"context"

	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

type serverStatsHandler struct {
	serverMetrics *ServerMetrics
}

// TagRPC implements the stats.Hanlder interface.
func (h *serverStatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	rpcInfo := newRPCInfo(info.FullMethodName)
	return context.WithValue(ctx, &rpcInfoKey, rpcInfo)
}

// HandleRPC implements the stats.Hanlder interface.
func (h *serverStatsHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	v, ok := ctx.Value(&rpcInfoKey).(*rpcInfo)
	if !ok {
		return
	}
	monitor := newServerReporterForStatsHanlder(v.startTime, h.serverMetrics, v.fullMethodName)
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
		monitor.SentMessageSize(Payload, float64(s.WireLength))
	case *stats.OutTrailer:
		monitor.SentMessageSize(Tailer, float64(s.WireLength))
	}
}

// TagConn implements the stats.Hanlder interface.
func (h *serverStatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	return ctx
}

// HandleConn implements the stats.Hanlder interface.
func (h *serverStatsHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
}
