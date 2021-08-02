// +build !go1.13

package grpcstatus

import (
	"google.golang.org/grpc/status"
)

func unwrapGRPCStatus(err error) (*status.Status, bool) {
	return nil, false
}
