package grpc_prometheus_test

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	pb "github.com/grpc-ecosystem/go-grpc-prometheus/examples/grpc-server-with-prometheus/protobuf"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
)

func ExampleStatsHandler() {
	// Listen an actual port.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	assert(err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reg := prometheus.NewRegistry()
	sts := grpc_prometheus.NewStatsHandler(
		grpc_prometheus.NewRequestsTotalStatsHandler(
			grpc_prometheus.Server,
			grpc_prometheus.NewRequestsTotalCounterVec(grpc_prometheus.Server),
		),
		grpc_prometheus.NewRequestsTotalStatsHandler(
			grpc_prometheus.Server,
			grpc_prometheus.NewRequestsTotalCounterVecV1(grpc_prometheus.Server),
		),
	)
	srv := grpc.NewServer(grpc.StatsHandler(sts))
	imp := newDemoServer()

	pb.RegisterDemoServiceServer(srv, imp)
	reg.MustRegister(sts)

	go func() {
		if err := srv.Serve(lis); err != grpc.ErrServerStopped {
			assert(err)
		}
	}()

	con, err := grpc.DialContext(ctx, lis.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())
	assert(err)

	pb.NewDemoServiceClient(con).SayHello(ctx, &pb.HelloRequest{Name: "example"})

	srv.GracefulStop()

	mf, err := reg.Gather()
	assert(err)

	for _, m := range mf {
		fmt.Println(m.GetName())
	}

	// Output:
	// grpc_server_received_requests_total
	// grpc_server_started_total
}

func assert(err error) {
	if err != nil {
		log.Println("ERR:", err)
		os.Exit(1)
	}
}

// demoServiceServer defines a Server.
type demoServiceServer struct{}

func newDemoServer() *demoServiceServer {
	return &demoServiceServer{}
}

// SayHello implements a interface defined by protobuf.
func (s *demoServiceServer) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{Message: fmt.Sprintf("Hello %s", request.Name)}, nil
}
