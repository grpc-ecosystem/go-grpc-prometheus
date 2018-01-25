package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/soheilhy/cmux"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	pb "github.com/grpc-ecosystem/go-grpc-prometheus/examples/grpc-server-with-prometheus/protobuf"
	"github.com/grpc-ecosystem/go-grpc-prometheus/examples/grpc-server-with-prometheus/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define a Server
type DemoServiceServer struct{}

func newDemoServer() *DemoServiceServer {
	return &DemoServiceServer{}
}

// Implement a interface defined by protobuf
func (s *DemoServiceServer) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{Message: fmt.Sprintf("hello %s", request.Name)}, nil
}

func main() {
	// Define a port
	port := util.SERVER_PORT

	// Listen an actual port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create a new cmux instance
	cmux_ins := cmux.New(lis)

	// Create matcher for specific protocol, such as HTTP, gRPC
	grpcL := cmux_ins.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	httpL := cmux_ins.Match(cmux.HTTP1Fast())

	// Create HTTP server for Prometheus
	httpServer := &http.Server{Handler: promhttp.Handler()}

	// Create a new gRPC Server with gRPC interceptor
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)

	// Register your service
	pb.RegisterDemoServiceServer(grpcServer, newDemoServer())

	// Enable Histogram for gRPC-Prometheus interceptor
	grpc_prometheus.EnableHandlingTimeHistogram()

	// Create a counter metric
	pushCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "repository_pushes",
		Help: "Number of pushes to external repository.",
	})

	// Register it!
	err = prometheus.Register(pushCounter)
	if err != nil {
		fmt.Println("Push counter couldn't be registered AGAIN, no counting will happen:", err)
		return
	}

	// Try to create some time series
	pushCounter.Inc()
	pushCounter.Inc()

	// Register your gRPC Server to gRPC-Prometheus interceptor
	grpc_prometheus.Register(grpcServer)

	// Use “fake” listener
	go httpServer.Serve(httpL)
	go grpcServer.Serve(grpcL)

	// Start your server
	cmux_ins.Serve()
}
