package main

import (
	"log"
	"net"

	pb "therealbroker/api/proto/protoGen"
	"therealbroker/api/proto/server/handler"
	"therealbroker/internal/broker"

	"google.golang.org/grpc"
)

func main() {
	// go func() {
	// 	prometheusServerStart()
	// }()
	// go func() {
	// 	err := exporter.Register()
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	fmt.Println("exporter registered")
	// }()
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	pb.RegisterBrokerServer(server, &handler.BrokerServer{
		BrokerInstance: broker.NewModule(),
	})
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Server serve failed: %v", err)
	}
}
