package main

import (
	"fmt"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	pb "github.com/grpc-ecosystem/go-grpc-prometheus/examples/grpc-server-with-prometheus/protobuf"
)

func main() {

	// Create a insecure gRPC channel to communicate with the server.
	conn, err := grpc.Dial(
		fmt.Sprintf("localhost:%v", 9093),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
	)

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err == nil {
			conn.Close()
		}
	}()

	client := pb.NewDemoServiceClient(conn)
	// Call “SayHello” method and wait for response from gRPC Server.
	resp, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "test"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp)
}
