package handler

import (
	"context"
	"log"
	"net"
	"os"
	pb "therealbroker/api/proto/protoGen"
	"therealbroker/api/proto/server/redis"
	"therealbroker/api/proto/server/trace"
	brk "therealbroker/internal/broker"
	"therealbroker/internal/exporter"
	prm "therealbroker/internal/prometheus"
	"therealbroker/pkg/broker"
	"time"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type BrokerServer struct {
	pb.UnimplementedBrokerServer
	BrokerInstance broker.Broker
	redisClient    *redis.RedisDB
}

func StartServer() {
	go func() {
		trace.PrometheusServerStart()
	}()
	go func() {
		err := trace.JaegerRegister()
		if err != nil {
			log.Fatalf("Jaeger failed: %v", err)
		}
	}()
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	pb.RegisterBrokerServer(server, &BrokerServer{
		BrokerInstance: brk.NewModule(),
		redisClient:    redis.NewModule(),
	})
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Server serve failed: %v", err)
	}
}

func (s *BrokerServer) Publish(ctx context.Context, request *pb.PublishRequest) (*pb.PublishResponse, error) {
	spanCtx, span := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "publish method")
	defer span.End()
	startTime := time.Now()
	defer prm.MethodDuration.WithLabelValues("Publish").Observe(time.Since(startTime).Seconds())
	ip, err := s.redisClient.GetPodIPBySubject(request.GetSubject())
	if err != nil {
		s.redisClient.SetPodIPBySubject(request.GetSubject(), os.Getenv("POD_IP"))
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
	resp, err := forwardPublishRequest(spanCtx, request, ip)
	if err != nil {
		prm.MethodCount.WithLabelValues("Publish", "failed").Inc()
		return nil, err
	}
	prm.MethodCount.WithLabelValues("Publish", "success").Inc()
	return resp, nil
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

func forwardPublishRequest(ctx context.Context, request *pb.PublishRequest, remoteServerAddr string) (*pb.PublishResponse, error) {
	conn, err := grpc.Dial(remoteServerAddr+":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	remoteClient := pb.NewBrokerClient(conn)
	remoteResponse, err := remoteClient.Publish(ctx, request)
	if err != nil {
		return nil, err
	}
	return remoteResponse, nil
}
