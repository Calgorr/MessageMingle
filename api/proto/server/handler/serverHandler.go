package handler

import (
	"context"
	pb "therealbroker/api/proto/protoGen"
	"therealbroker/pkg/broker"
	"time"
)

type brokerServer struct {
	pb.UnimplementedBrokerServer
	BrokerInstance broker.Broker
}

func (s *brokerServer) Publish(ctx context.Context, request *pb.PublishRequest) (*pb.PublishResponse, error) {
	msg := broker.Message{
		Body:       string(request.GetBody()),
		Expiration: time.Duration(request.GetExpirationSeconds()),
	}
	id, err := s.BrokerInstance.Publish(ctx, request.GetSubject(), msg)
	if err != nil {
		return nil, err
	}
	return &pb.PublishResponse{Id: int32(id)}, nil
}

func (s *brokerServer) Subscribe(*pb.SubscribeRequest, pb.Broker_SubscribeServer) error {
	return nil
}

func (s *brokerServer) Fetch(context.Context, *pb.FetchRequest) (*pb.MessageResponse, error) {
	return nil, nil
}
