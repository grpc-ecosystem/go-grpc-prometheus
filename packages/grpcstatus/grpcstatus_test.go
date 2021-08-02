package grpcstatus

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Own implementation of pkg/errors withStack to avoid additional dependency
type wrappedError struct {
	cause error
	msg   string
}

func (w *wrappedError) Error() string { return w.msg + ": " + w.cause.Error() }

func (w *wrappedError) Unwrap() error { return w.cause }

func TestErrorUnwrapping(t *testing.T) {
	gRPCCode := codes.FailedPrecondition
	gRPCError := status.Errorf(gRPCCode, "Userspace error.")
	expectedGRPCStatus := status.Convert(gRPCError)
	testedErrors := []error{
		gRPCError,
		&wrappedError{cause: gRPCError, msg: "pkg/errors wrapped error: "},
	}

	for _, e := range testedErrors {
		resultingStatus, ok := FromError(e)
		require.True(t, ok)
		require.Equal(t, expectedGRPCStatus, resultingStatus)
	}
}
