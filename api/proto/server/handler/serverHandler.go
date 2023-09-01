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

func (s *brokerServer) Subscribe(request *pb.SubscribeRequest, server pb.Broker_SubscribeServer) error {
	ch, err := s.BrokerInstance.Subscribe(server.Context(), request.GetSubject())
	if err != nil {
		return err
	}
	for {
		select {
		case <-server.Context().Done():
			return nil
		case msg, closed := <-ch:
			if closed {
				return nil
			}
			if err := server.Send(&pb.MessageResponse{Body: []byte(msg.Body)}); err != nil {
				return err
			}
		}
	}
}

func (s *brokerServer) Fetch(xtx context.Context, request *pb.FetchRequest) (*pb.MessageResponse, error) {
	return nil, nil
}
