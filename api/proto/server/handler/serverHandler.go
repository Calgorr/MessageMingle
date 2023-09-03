package handler

import (
	"context"
	pb "therealbroker/api/proto/protoGen"
	prm "therealbroker/api/proto/server/prometheus"
	"therealbroker/pkg/broker"
	"time"
)

type BrokerServer struct {
	pb.UnimplementedBrokerServer
	BrokerInstance broker.Broker
}

func (s *BrokerServer) Publish(ctx context.Context, request *pb.PublishRequest) (*pb.PublishResponse, error) {
	startTime := time.Now()
	defer prm.MethodDuration.WithLabelValues("Publish").Observe(time.Since(startTime).Seconds())
	msg := broker.Message{
		Body:       string(request.GetBody()),
		Expiration: time.Duration(request.GetExpirationSeconds()),
	}
	id, err := s.BrokerInstance.Publish(ctx, request.GetSubject(), msg)
	if err != nil {
		prm.MethodCount.WithLabelValues("Publish", "failed").Inc()
		return nil, err
	}
	prm.MethodCount.WithLabelValues("Publish", "success").Inc()
	return &pb.PublishResponse{Id: int32(id)}, nil
}

func (s *BrokerServer) Subscribe(request *pb.SubscribeRequest, server pb.Broker_SubscribeServer) error {
	startTime := time.Now()
	defer prm.MethodDuration.WithLabelValues("Publish").Observe(time.Since(startTime).Seconds())
	ch, err := s.BrokerInstance.Subscribe(server.Context(), request.GetSubject())
	prm.ActiveSubscribers.Inc()
	if err != nil {
		prm.MethodCount.WithLabelValues("Subscribe", "failed").Inc()
		return err
	}
	ctx := server.Context()
	for {
		select {
		case <-ctx.Done():
			prm.ActiveSubscribers.Dec()
			prm.MethodCount.WithLabelValues("Subscribe", "success").Inc()
			return nil
		case msg, open := <-ch:
			if !open {
				prm.ActiveSubscribers.Dec()
				prm.MethodCount.WithLabelValues("Subscribe", "success").Inc()
				return err
			}
			if err := server.Send(&pb.MessageResponse{Body: []byte(msg.Body)}); err != nil {
				prm.MethodCount.WithLabelValues("Subscribe", "failed").Inc()
				return err
			}
		}
	}
}

func (s *BrokerServer) Fetch(ctx context.Context, request *pb.FetchRequest) (*pb.MessageResponse, error) {
	startTime := time.Now()
	defer prm.MethodDuration.WithLabelValues("Publish").Observe(time.Since(startTime).Seconds())
	msg, err := s.BrokerInstance.Fetch(ctx, request.GetSubject(), int(request.GetId()))
	if err != nil {
		prm.MethodCount.WithLabelValues("Fetch", "failed").Inc()
		return nil, err
	}
	prm.MethodCount.WithLabelValues("Fetch", "success").Inc()
	return &pb.MessageResponse{Body: []byte(msg.Body)}, nil
}