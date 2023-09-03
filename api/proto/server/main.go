package main

import (
	"log"
	"net"
	"net/http"

	pb "therealbroker/api/proto/protoGen"
	"therealbroker/api/proto/server/handler"
	"therealbroker/api/proto/server/prometheus"
	"therealbroker/internal/broker"

	"google.golang.org/grpc"
)

func main() {
	http.Handle("/metrics", prometheus.PrometheusHandler())
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
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
