// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_prometheus

import (
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type grpcType string

const (
	Unary        grpcType = "unary"
	ClientStream grpcType = "client_stream"
	ServerStream grpcType = "server_stream"
	BidiStream   grpcType = "bidi_stream"
)

var (
	allCodes = []codes.Code{
		codes.OK, codes.Canceled, codes.Unknown, codes.InvalidArgument, codes.DeadlineExceeded, codes.NotFound,
		codes.AlreadyExists, codes.PermissionDenied, codes.Unauthenticated, codes.ResourceExhausted,
		codes.FailedPrecondition, codes.Aborted, codes.OutOfRange, codes.Unimplemented, codes.Internal,
		codes.Unavailable, codes.DataLoss,
	}

	allStatss = []grpcStats{Header, Payload, Tailer}

	rpcInfoKey = "rpc-info"

	defMsgBytesBuckets = []float64{0, 32, 64, 128, 256, 512, 1024, 2048, 8192, 32768, 131072, 524288}
)

func splitMethodName(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/") // remove leading slash
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return "unknown", "unknown"
}

func typeFromMethodInfo(mInfo *grpc.MethodInfo) grpcType {
	if !mInfo.IsClientStream && !mInfo.IsServerStream {
		return Unary
	}
	if mInfo.IsClientStream && !mInfo.IsServerStream {
		return ClientStream
	}
	if !mInfo.IsClientStream && mInfo.IsServerStream {
		return ServerStream
	}
	return BidiStream
}

type grpcStats string

const (
	// Header indicates that the stats is the header
	Header grpcStats = "header"

	// Payload indicates that the stats is the Payload
	Payload grpcStats = "payload"

	// Tailer indicates that the stats is the Payload
	Tailer grpcStats = "tailer"
)

// String function returns the grpcStats with string format.
func (s grpcStats) String() string {
	return string(s)
}

type rpcInfo struct {
	fullMethodName string
	startTime      time.Time
}

func newRPCInfo(fullMethodName string) *rpcInfo {
	return &rpcInfo{fullMethodName: fullMethodName}
}
