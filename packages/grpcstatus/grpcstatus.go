package grpcstatus

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Since error can be wrapped and the `FromError` function only checks for `GRPCStatus` function
// and as a fallback uses the `Unknown` gRPC status we need to unwrap the error if possible to get the original status.
// Eventually should be implemented in the go-grpc status function `FromError`. See https://github.com/grpc/grpc-go/issues/2934
func FromError(err error) (s *status.Status, ok bool) {
	s, ok = status.FromError(err)
	if ok {
		return s, true
	}

	// Try to unwrap gRPC status
	s, ok = unwrapGRPCStatus(err)
	if ok {
		return s, true
	}

	// We failed to unwrap any GRPSStatus so return default `Unknown`
	return status.New(codes.Unknown, err.Error()), false
}
