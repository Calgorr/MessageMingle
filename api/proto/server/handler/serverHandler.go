package handler

import (
	"context"
	"fmt"
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

var (
	selfIP    = os.Getenv("POD_IP")
	selfIPlen = len(selfIP)
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
	fmt.Println("From pod " + selfIP)
	spanCtx, span := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "publish method")
	defer span.End()
	startTime := time.Now()
	defer prm.MethodDuration.WithLabelValues("Publish").Observe(time.Since(startTime).Seconds())
	ips, err := s.redisClient.GetPodsIPBySubject(request.GetSubject())
	if err != nil {
		prm.MethodCount.WithLabelValues("Publish", "failed", selfIP[selfIPlen-1:]).Inc()
		return nil, err
	}
	for _, ip := range ips {
		if ip == selfIP {
			_, err := s.BrokerInstance.Publish(spanCtx, request.GetSubject(), broker.Message{
				Body:       string(request.GetBody()),
				Expiration: time.Duration(request.GetExpirationSeconds()),
			})
			if err != nil {
				prm.MethodCount.WithLabelValues("Publish", "failed", selfIP[selfIPlen-1:]).Inc()
			} else {
				prm.MethodCount.WithLabelValues("Publish", "success", selfIP[selfIPlen-1:]).Inc()
			}
			continue
		} else {
			fmt.Println("Publishing to " + ip + " From pod " + selfIP)
			_, err := forwardPublishRequest(spanCtx, request, ip)
			if err != nil {
				prm.MethodCount.WithLabelValues("Publish", "failed", selfIP[selfIPlen-1:]).Inc()
			}
			_, err = s.BrokerInstance.SaveMessage(spanCtx, broker.Message{
				Body:       string(request.GetBody()),
				Expiration: time.Duration(request.GetExpirationSeconds()),
			}, request.GetSubject())
			if err != nil {
				prm.MethodCount.WithLabelValues("Publish", "failed", selfIP[selfIPlen-1:]).Inc()
			} else {
				prm.MethodCount.WithLabelValues("Publish", "success", selfIP[selfIPlen-1:]).Inc()
			}
			continue
		}
	}
	if len(ips) == 0 {
		id, err := s.BrokerInstance.Publish(spanCtx, request.GetSubject(), broker.Message{
			Body:       string(request.GetBody()),
			Expiration: time.Duration(request.GetExpirationSeconds()),
		})
		if err != nil {
			prm.MethodCount.WithLabelValues("Publish", "failed", selfIP[selfIPlen-1:]).Inc()
			return nil, err
		}
		prm.MethodCount.WithLabelValues("Publish", "success", selfIP[selfIPlen-1:]).Inc()
		return &pb.PublishResponse{Id: int32(id)}, nil
	}
	return &pb.PublishResponse{Id: int32(0)}, nil
}

func (s *BrokerServer) Subscribe(request *pb.SubscribeRequest, server pb.Broker_SubscribeServer) error {
	spanCtx, span := otel.Tracer(exporter.DefaultServiceName).Start(server.Context(), "Subscribe method")
	defer span.End()
	startTime := time.Now()
	defer prm.MethodDuration.WithLabelValues("Publish").Observe(time.Since(startTime).Seconds())
	ctx := server.Context()
	if _, err := s.redisClient.GetPodIPBySubjectAndIP(request.GetSubject(), selfIP); err != nil {
		s.redisClient.SetPodIPBySubjectAndIP(request.GetSubject(), selfIP)
	}
	ch, err := s.BrokerInstance.Subscribe(spanCtx, request.GetSubject())
	prm.ActiveSubscribers.Inc()
	if err != nil {
		prm.MethodCount.WithLabelValues("Subscribe", "failed").Inc()
		return err
	}
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
			fmt.Println("From pod " + selfIP + " Subject : " + request.GetSubject() + " message : " + msg.Body)
			if err := server.Send(&pb.MessageResponse{Body: []byte(msg.Body)}); err != nil {
				prm.MethodCount.WithLabelValues("Subscribe", "failed").Inc()
				return err
			}
		}
	}
}

func (s *BrokerServer) Fetch(ctx context.Context, request *pb.FetchRequest) (*pb.MessageResponse, error) {
	fmt.Println("From pod " + selfIP)
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

func (s *BrokerServer) InternalPublish(ctx context.Context, request *pb.PublishRequest) (*pb.PublishResponse, error) {
	spanCtx, span := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "InternalPublish method")
	defer span.End()
	startTime := time.Now()
	defer prm.MethodDuration.WithLabelValues("Publish").Observe(time.Since(startTime).Seconds())
	msg := broker.Message{
		Body:       string(request.GetBody()),
		Expiration: time.Duration(request.GetExpirationSeconds()),
	}
	id, err := s.BrokerInstance.PublishInternal(spanCtx, request.GetSubject(), msg)
	if err != nil {
		prm.MethodCount.WithLabelValues("Publish", "failed").Inc()
		return nil, err
	}
	prm.MethodCount.WithLabelValues("Publish", "success").Inc()
	return &pb.PublishResponse{Id: int32(id)}, nil
}

func forwardPublishRequest(ctx context.Context, request *pb.PublishRequest, remoteServerAddr string) (*pb.PublishResponse, error) {
	conn, err := grpc.Dial(remoteServerAddr+":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	remoteClient := pb.NewBrokerClient(conn)
	remoteResponse, err := remoteClient.InternalPublish(ctx, request)
	if err != nil {
		return nil, err
	}
	return remoteResponse, nil
}
