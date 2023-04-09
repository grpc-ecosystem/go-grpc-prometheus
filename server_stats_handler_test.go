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
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func TestServerStatsHandlerSuite(t *testing.T) {
	suite.Run(t, &ServerStatsHandlerTestSuite{})
}

type ServerStatsHandlerTestSuite struct {
	suite.Suite

	serverListener net.Listener
	server         *grpc.Server
	clientConn     *grpc.ClientConn
	testClient     pb_testproto.TestServiceClient
	ctx            context.Context
	cancel         context.CancelFunc
}

func (s *ServerStatsHandlerTestSuite) SetupSuite() {
	var err error

	EnableHandlingTimeHistogram()

	EnableServerMsgSizeReceivedBytesHistogram()

	EnableServerMsgSizeSentBytesHistogram()

	s.serverListener, err = net.Listen("tcp", "127.0.0.1:0")
	require.NoError(s.T(), err, "must be able to allocate a port for serverListener")

	// This is the point where we hook up the interceptor
	s.server = grpc.NewServer(
		grpc.StatsHandler(ServerStatsHandler),
	)
	pb_testproto.RegisterTestServiceServer(s.server, &testService{t: s.T()})

	go func() {
		s.server.Serve(s.serverListener)
	}()

	s.clientConn, err = grpc.Dial(s.serverListener.Addr().String(), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2*time.Second))
	require.NoError(s.T(), err, "must not error on client Dial")
	s.testClient = pb_testproto.NewTestServiceClient(s.clientConn)
}

func (s *ServerStatsHandlerTestSuite) SetupTest() {
	// Make all RPC calls last at most 2 sec, meaning all async issues or deadlock will not kill tests.
	s.ctx, s.cancel = context.WithTimeout(context.TODO(), 2*time.Second)

	// Make sure every test starts with same fresh, intialized metric state.
	DefaultServerMetrics.serverStartedCounter.Reset()
	DefaultServerMetrics.serverHandledCounter.Reset()
	DefaultServerMetrics.serverHandledHistogram.Reset()
	DefaultServerMetrics.serverStreamMsgReceived.Reset()
	DefaultServerMetrics.serverStreamMsgSent.Reset()
	DefaultServerMetrics.serverMsgSizeReceivedHistogram.Reset()
	DefaultServerMetrics.serverMsgSizeSentHistogram.Reset()
	Register(s.server)
}

func (s *ServerStatsHandlerTestSuite) TearDownSuite() {
	if s.serverListener != nil {
		s.server.Stop()
		s.T().Logf("stopped grpc.Server at: %v", s.serverListener.Addr().String())
		s.serverListener.Close()

	}
	if s.clientConn != nil {
		s.clientConn.Close()
	}
}

func (s *ServerStatsHandlerTestSuite) TearDownTest() {
	s.cancel()
}

func (s *ServerStatsHandlerTestSuite) TestUnaryIncrementsMetrics() {
	_, err := s.testClient.PingEmpty(s.ctx, &pb_testproto.Empty{}) // should return with code=OK
	require.NoError(s.T(), err)
	requireValue(s.T(), 1, DefaultServerMetrics.serverStartedCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingEmpty"))
	requireValue(s.T(), 1, DefaultServerMetrics.serverHandledCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingEmpty", "OK"))
	requireValueHistCount(s.T(), 1, DefaultServerMetrics.serverHandledHistogram.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingEmpty"))

	requireValueHistCount(s.T(), 1, DefaultServerMetrics.serverMsgSizeReceivedHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingEmpty", Header.String()))
	requireValueHistCount(s.T(), 1, DefaultServerMetrics.serverMsgSizeReceivedHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingEmpty", Payload.String()))
	requireValueHistCount(s.T(), 1, DefaultServerMetrics.serverMsgSizeSentHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingEmpty", Payload.String()))
	requireValueHistCount(s.T(), 1, DefaultServerMetrics.serverMsgSizeSentHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingEmpty", Tailer.String()))

	_, err = s.testClient.PingError(s.ctx, &pb_testproto.PingRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)}) // should return with code=FailedPrecondition
	require.Error(s.T(), err)
	requireValue(s.T(), 1, DefaultServerMetrics.serverStartedCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingError"))
	requireValue(s.T(), 1, DefaultServerMetrics.serverHandledCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingError", "FailedPrecondition"))
	requireValueHistCount(s.T(), 1, DefaultServerMetrics.serverHandledHistogram.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingError"))

	requireValueHistCount(s.T(), 1, DefaultServerMetrics.serverMsgSizeReceivedHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingEmpty", Header.String()))
	requireValueHistCount(s.T(), 1, DefaultServerMetrics.serverMsgSizeReceivedHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingEmpty", Payload.String()))
	requireValueHistCount(s.T(), 1, DefaultServerMetrics.serverMsgSizeSentHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingEmpty", Tailer.String()))

}

func (s *ServerStatsHandlerTestSuite) TestStreamingIncrementsMetrics() {
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

	requireValueWithRetry(s.ctx, s.T(), 1,
		DefaultServerMetrics.serverStartedCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingList"))
	requireValueWithRetry(s.ctx, s.T(), 1,
		DefaultServerMetrics.serverHandledCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingList", "OK"))
	requireValueWithRetry(s.ctx, s.T(), countListResponses,
		DefaultServerMetrics.serverStreamMsgSent.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingList"))
	requireValueWithRetry(s.ctx, s.T(), 1,
		DefaultServerMetrics.serverStreamMsgReceived.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingList"))
	requireValueHistCount(s.T(), 1,
		DefaultServerMetrics.serverHandledHistogram.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingList"))

	requireValueHistCount(s.T(), 1, DefaultServerMetrics.serverMsgSizeReceivedHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingList", Header.String()))
	requireValueHistCount(s.T(), 1, DefaultServerMetrics.serverMsgSizeReceivedHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingList", Payload.String()))
	requireValueHistCount(s.T(), countListResponses, DefaultServerMetrics.serverMsgSizeSentHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingList", Payload.String()))
	requireValueHistCount(s.T(), 1, DefaultServerMetrics.serverMsgSizeSentHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingList", Tailer.String()))

	_, err := s.testClient.PingList(s.ctx, &pb_testproto.PingRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)}) // should return with code=FailedPrecondition
	require.NoError(s.T(), err, "PingList must not fail immediately")

	requireValueWithRetry(s.ctx, s.T(), 2,
		DefaultServerMetrics.serverStartedCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingList"))
	requireValueWithRetry(s.ctx, s.T(), 1,
		DefaultServerMetrics.serverHandledCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingList", "FailedPrecondition"))
	requireValueHistCount(s.T(), 2,
		DefaultServerMetrics.serverHandledHistogram.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingList"))

	requireValueHistCount(s.T(), 2, DefaultServerMetrics.serverMsgSizeReceivedHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingList", Header.String()))
	requireValueHistCount(s.T(), 2, DefaultServerMetrics.serverMsgSizeReceivedHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingList", Payload.String()))
	requireValueHistCount(s.T(), 2, DefaultServerMetrics.serverMsgSizeSentHistogram.WithLabelValues("mwitkow.testproto.TestService", "PingList", Tailer.String()))

}
