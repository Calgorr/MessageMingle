package handler

import (
	"context"
	pb "therealbroker/api/proto/protoGen"
	"therealbroker/pkg/broker"
	"time"
)

type BrokerServer struct {
	pb.UnimplementedBrokerServer
	BrokerInstance broker.Broker
}

func (s *BrokerServer) Publish(ctx context.Context, request *pb.PublishRequest) (*pb.PublishResponse, error) {
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

func (s *BrokerServer) Subscribe(request *pb.SubscribeRequest, server pb.Broker_SubscribeServer) error {
	ch, err := s.BrokerInstance.Subscribe(server.Context(), request.GetSubject())
	if err != nil {
		return err
	}
	ctx := server.Context()
	for {
		select {
		case <-ctx.Done():
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

func (s *BrokerServer) Fetch(ctx context.Context, request *pb.FetchRequest) (*pb.MessageResponse, error) {
	msg, err := s.BrokerInstance.Fetch(ctx, request.GetSubject(), int(request.GetId()))
	if err != nil {
		return nil, err
	}
	return &pb.MessageResponse{Body: []byte(msg.Body)}, nil
}
