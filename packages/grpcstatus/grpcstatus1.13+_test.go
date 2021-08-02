// +build go1.13

package grpcstatus

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNativeErrorUnwrapping(t *testing.T) {
	gRPCCode := codes.FailedPrecondition
	gRPCError := status.Errorf(gRPCCode, "Userspace error.")
	expectedGRPCStatus := status.Convert(gRPCError)
	testedErrors := []error{
		fmt.Errorf("go native wrapped error: %w", gRPCError),
	}

	for _, e := range testedErrors {
		resultingStatus, ok := FromError(e)
		require.True(t, ok)
		require.Equal(t, expectedGRPCStatus, resultingStatus)
	}
}
