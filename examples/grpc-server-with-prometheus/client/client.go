package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	pb "github.com/grpc-ecosystem/go-grpc-prometheus/examples/grpc-server-with-prometheus/protobuf"
	"github.com/grpc-ecosystem/go-grpc-prometheus/examples/grpc-server-with-prometheus/util"
)

func main() {
	// Create a gRPC channel to communicate with the server
	conn, err := grpc.Dial(
		"localhost:"+strconv.FormatInt(util.SERVER_PORT, 10),
		grpc.WithInsecure(),
		// For grpc_prometheus
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor))
	defer conn.Close()

	if err != nil {
		log.Fatal(err)
	}

	client := pb.NewDemoServiceClient(conn)

	// Call “SayHello” method and wait for response from gRPC Server
	resp, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "test"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp)
}
