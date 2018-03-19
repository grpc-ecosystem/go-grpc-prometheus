package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	pb "github.com/grpc-ecosystem/go-grpc-prometheus/examples/grpc-server-with-prometheus/protobuf"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// DemoServiceServer defines a Server.
type DemoServiceServer struct{}

func newDemoServer() *DemoServiceServer {
	return &DemoServiceServer{}
}

// SayHello implements a interface defined by protobuf.
func (s *DemoServiceServer) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloResponse, error) {
	pushCounter.Inc()
	return &pb.HelloResponse{Message: fmt.Sprintf("Hello %s", request.Name)}, nil
}

// Create a counter metric
var (
	pushCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "repository_pushes",
		Help: "Number of pushes to external repository.",
	})
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(pushCounter)
}

func main() {
	// Listen an actual port.
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9093))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create a HTTP server for prometheus.
	httpServer := &http.Server{Handler: promhttp.Handler(), Addr: fmt.Sprintf("0.0.0.0:%d", 9092)}

	// Create a gRPC Server with gRPC interceptor.
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)

	// Register your service.
	pb.RegisterDemoServiceServer(grpcServer, newDemoServer())

	// Enable Histogram for gRPC-Prometheus interceptor.
	grpc_prometheus.EnableHandlingTimeHistogram()

	// Register your gRPC Server to gRPC-Prometheus interceptor.
	grpc_prometheus.Register(grpcServer)

	// Start your http server for prometheus.
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatal("Unable to start a http server.")
		}
	}()

	// Start your gRPC server.
	log.Fatal(grpcServer.Serve(lis))
}
