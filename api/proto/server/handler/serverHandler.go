package handler

import (
	"context"
	pb "therealbroker/api/proto/protoGen"
	"therealbroker/internal/exporter"
	prm "therealbroker/internal/prometheus"
	"therealbroker/pkg/broker"
	"time"

	"go.opentelemetry.io/otel"
)

type BrokerServer struct {
	pb.UnimplementedBrokerServer
	BrokerInstance broker.Broker
}

func (s *BrokerServer) Publish(ctx context.Context, request *pb.PublishRequest) (*pb.PublishResponse, error) {
	spanCtx, span := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "publish method")
	defer span.End()
	startTime := time.Now()
	defer prm.MethodDuration.WithLabelValues("Publish").Observe(time.Since(startTime).Seconds())
	msg := broker.Message{
		Body:       string(request.GetBody()),
		Expiration: time.Duration(request.GetExpirationSeconds()),
	}
	id, err := s.BrokerInstance.Publish(spanCtx, request.GetSubject(), msg)
	if err != nil {
		prm.MethodCount.WithLabelValues("Publish", "failed").Inc()
		return nil, err
	}
	prm.MethodCount.WithLabelValues("Publish", "success").Inc()
	return &pb.PublishResponse{Id: int32(id)}, nil
}

func (s *BrokerServer) Subscribe(request *pb.SubscribeRequest, server pb.Broker_SubscribeServer) error {
	spanCtx, span := otel.Tracer(exporter.DefaultServiceName).Start(server.Context(), "Subscribe method")
	defer span.End()
	startTime := time.Now()
	defer prm.MethodDuration.WithLabelValues("Publish").Observe(time.Since(startTime).Seconds())
	ch, err := s.BrokerInstance.Subscribe(spanCtx, request.GetSubject())
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
	spanCtx, span := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "Fetch method")
	defer span.End()
	startTime := time.Now()
	defer prm.MethodDuration.WithLabelValues("Publish").Observe(time.Since(startTime).Seconds())
	msg, err := s.BrokerInstance.Fetch(spanCtx, request.GetSubject(), int(request.GetId()))
	if err != nil {
		prm.MethodCount.WithLabelValues("Fetch", "failed").Inc()
		return nil, err
	}
	prm.MethodCount.WithLabelValues("Fetch", "success").Inc()
	return &pb.MessageResponse{Body: []byte(msg.Body)}, nil
}
