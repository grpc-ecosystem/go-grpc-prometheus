package main

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/grpc-ecosystem/go-grpc-prometheus"

	pb "github.com/grpc-ecosystem/go-grpc-prometheus/examples/grpc-server-with-prometheus/protobuf"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"os"
	"os/signal"
	"syscall"
)

var (
	// Create a metrics registry.
	reg = prometheus.NewRegistry()

	// Create some standard server metrics.
	grpcMetrics = grpc_prometheus.NewClientMetrics()

	// Create a customized counter metric.
	customizedCounterMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "demo_client_say_hello_method_handle_count",
		Help: "Total number of RPCs call from the client.",
	}, []string{"ClientName1"})
)

func init() {
	// Register standard server metrics and customized metrics to registry.
	reg.MustRegister(grpcMetrics, customizedCounterMetric)
	customizedCounterMetric.WithLabelValues("ClientTest1")
}
func main() {
	// Create a insecure gRPC channel to communicate with the server.
	conn, err := grpc.Dial(
		fmt.Sprintf("localhost:%v", 9093),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpcMetrics.UnaryClientInterceptor()),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create a HTTP server for prometheus.
	httpServer := &http.Server{Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}), Addr: fmt.Sprintf("0.0.0.0:%d", 9094)}


	defer conn.Close()

	// Create a gRPC server client.
	client := pb.NewDemoServiceClient(conn)
	// Call “SayHello” method and wait for response from gRPC Server.
	request:=&pb.HelloRequest{Name: "Test1"}
	for i:=0;i<5;i++{
		customizedCounterMetric.WithLabelValues(request.Name).Inc()
		_, err := client.SayHello(context.Background(),request)
		if err != nil {
			log.Fatal(err)
		}
	}



	// Start your http server for prometheus.
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatal("Unable to start a http server.")
		}
	}()
	//fmt.Println(resp)
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-exit:
		httpServer.Close()
	}

}
