// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_prometheus

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	pb_testproto "github.com/grpc-ecosystem/go-grpc-prometheus/examples/testproto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// client metrics must satisfy the Collector interface
	_ prometheus.Collector = NewClientMetrics()
)

func TestClientStatsHandlerSuite(t *testing.T) {
	suite.Run(t, &ClientStatsHandlerTestSuite{})
}

type ClientStatsHandlerTestSuite struct {
	suite.Suite

	serverListener net.Listener
	server         *grpc.Server
	clientConn     *grpc.ClientConn
	testClient     pb_testproto.TestServiceClient
	ctx            context.Context
	cancel         context.CancelFunc
}

func (s *ClientStatsHandlerTestSuite) SetupSuite() {
	var err error

	EnableClientHandlingTimeHistogram()

	EnableClientMsgSizeSentBytesHistogram()

	EnableClientMsgSizeReceivedBytesHistogram()

	s.serverListener, err = net.Listen("tcp", "127.0.0.1:0")
	require.NoError(s.T(), err, "must be able to allocate a port for serverListener")

	// This is the point where we hook up the interceptor
	s.server = grpc.NewServer()
	pb_testproto.RegisterTestServiceServer(s.server, &testService{t: s.T()})

	go func() {
		s.server.Serve(s.serverListener)
	}()

	s.clientConn, err = grpc.Dial(
		s.serverListener.Addr().String(),
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithStatsHandler(ClientStatsHandler),
		grpc.WithTimeout(2*time.Second))
	require.NoError(s.T(), err, "must not error on client Dial")
	s.testClient = pb_testproto.NewTestServiceClient(s.clientConn)
}

func (s *ClientStatsHandlerTestSuite) SetupTest() {
	// Make all RPC calls last at most 2 sec, meaning all async issues or deadlock will not kill tests.
	s.ctx, s.cancel = context.WithTimeout(context.TODO(), 2*time.Second)

	// Make sure every test starts with same fresh, intialized metric state.
	DefaultClientMetrics.clientStartedCounter.Reset()
	DefaultClientMetrics.clientHandledCounter.Reset()
	DefaultClientMetrics.clientHandledHistogram.Reset()
	DefaultClientMetrics.clientStreamMsgReceived.Reset()
	DefaultClientMetrics.clientStreamMsgSent.Reset()
	DefaultClientMetrics.clientMsgSizeReceivedHistogram.Reset()
	DefaultClientMetrics.clientMsgSizeSentHistogram.Reset()

}

func (s *ClientStatsHandlerTestSuite) TearDownSuite() {
	if s.serverListener != nil {
		s.server.Stop()
		s.T().Logf("stopped grpc.Server at: %v", s.serverListener.Addr().String())
		s.serverListener.Close()

	}
	if s.clientConn != nil {
		s.clientConn.Close()
	}
}

func (s *ClientStatsHandlerTestSuite) TearDownTest() {
	s.cancel()
}

func (s *ClientStatsHandlerTestSuite) TestUnaryIncrementsMetrics() {
	_, err := s.testClient.PingEmpty(s.ctx, &pb_testproto.Empty{}) // should return with code=OK
	require.NoError(s.T(), err)
	requireValue(s.T(), 1, DefaultClientMetrics.clientStartedCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingEmpty"))
	requireValue(s.T(), 1, DefaultClientMetrics.clientHandledCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingEmpty", "OK"))
	requireValueHistCount(s.T(), 1, DefaultClientMetrics.clientHandledHistogram.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingEmpty"))

	requireValueHistCount(s.T(), 1, DefaultClientMetrics.clientMsgSizeSentHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingEmpty", Payload.String()))
	requireValueHistCount(s.T(), 1, DefaultClientMetrics.clientMsgSizeReceivedHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingEmpty", Header.String()))
	requireValueHistCount(s.T(), 1, DefaultClientMetrics.clientMsgSizeReceivedHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingEmpty", Payload.String()))
	requireValueHistCount(s.T(), 1, DefaultClientMetrics.clientMsgSizeReceivedHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingEmpty", Tailer.String()))

	_, err = s.testClient.PingError(s.ctx, &pb_testproto.PingRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)}) // should return with code=FailedPrecondition
	require.Error(s.T(), err)
	requireValue(s.T(), 1, DefaultClientMetrics.clientStartedCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingError"))
	requireValue(s.T(), 1, DefaultClientMetrics.clientHandledCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingError", "FailedPrecondition"))
	requireValueHistCount(s.T(), 1, DefaultClientMetrics.clientHandledHistogram.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingError"))

	requireValueHistCount(s.T(), 1, DefaultClientMetrics.clientMsgSizeSentHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingError", Payload.String()))
	requireValueHistCount(s.T(), 1, DefaultClientMetrics.clientMsgSizeReceivedHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingError", Tailer.String()))
}

func (s *ClientStatsHandlerTestSuite) TestStreamingIncrementsMetrics() {
	ss, _ := s.testClient.PingList(s.ctx, &pb_testproto.PingRequest{}) // should return with code=OK
	// Do a read, just for kicks.
	count := 0
	for {
		_, err := ss.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading pingList shouldn't fail")
		count++
	}
	require.EqualValues(s.T(), countListResponses, count, "Number of received msg on the wire must match")

	requireValue(s.T(), 1, DefaultClientMetrics.clientStartedCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingList"))
	requireValue(s.T(), 1, DefaultClientMetrics.clientHandledCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingList", "OK"))
	requireValue(s.T(), countListResponses, DefaultClientMetrics.clientStreamMsgReceived.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingList"))
	requireValue(s.T(), 1, DefaultClientMetrics.clientStreamMsgSent.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingList"))
	requireValueWithRetryHistCount(s.ctx, s.T(), 1, DefaultClientMetrics.clientHandledHistogram.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingList"))

	requireValueWithRetryHistCount(s.ctx, s.T(), 1, DefaultClientMetrics.clientMsgSizeSentHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingList", Payload.String()))
	requireValueWithRetryHistCount(s.ctx, s.T(), 1, DefaultClientMetrics.clientMsgSizeReceivedHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingList", Header.String()))
	requireValueWithRetryHistCount(s.ctx, s.T(), countListResponses, DefaultClientMetrics.clientMsgSizeReceivedHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingList", Payload.String()))
	requireValueWithRetryHistCount(s.ctx, s.T(), 1, DefaultClientMetrics.clientMsgSizeReceivedHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingList", Tailer.String()))

	ss, err := s.testClient.PingList(s.ctx, &pb_testproto.PingRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)}) // should return with code=FailedPrecondition
	require.NoError(s.T(), err, "PingList must not fail immediately")

	// Do a read, just to progate errors.
	_, err = ss.Recv()
	st, _ := status.FromError(err)
	require.Equal(s.T(), codes.FailedPrecondition, st.Code(), "Recv must return FailedPrecondition, otherwise the test is wrong")

	requireValue(s.T(), 2, DefaultClientMetrics.clientStartedCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingList"))
	requireValue(s.T(), 1, DefaultClientMetrics.clientHandledCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingList", "FailedPrecondition"))
	requireValueWithRetryHistCount(s.ctx, s.T(), 2, DefaultClientMetrics.clientHandledHistogram.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingList"))

	requireValueWithRetryHistCount(s.ctx, s.T(), 2, DefaultClientMetrics.clientMsgSizeSentHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingList", Payload.String()))
	requireValueWithRetryHistCount(s.ctx, s.T(), 2, DefaultClientMetrics.clientMsgSizeReceivedHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingList", Tailer.String()))
}
