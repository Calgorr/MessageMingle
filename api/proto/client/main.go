package main

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	pb "therealbroker/api/proto/protoGen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ticker := time.NewTicker(200 * time.Millisecond)
	for {
		select {
		case <-context.Background().Done():
			return
		case <-ticker.C:
			go func() {
				go PublishNewConnection()
			}()
		}
	}
}

func Publish(ctx context.Context, client pb.BrokerClient, done chan bool) {
	var wg sync.WaitGroup
	ticker := time.NewTicker(20 * time.Microsecond)
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			wg.Wait()
			return
		case <-ticker.C:
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := client.Publish(ctx, &pb.PublishRequest{Subject: "test", Body: []byte("randomString(10)"), ExpirationSeconds: 10})
				if err != nil {
					log.Printf("Publish error: %v", err)
				}
			}()
		}
	}
}

func Subscribe(ctx context.Context, client pb.BrokerClient) {
	stream, err := client.Subscribe(ctx, &pb.SubscribeRequest{Subject: "test"})
	if err != nil {
		log.Fatalf("Subscribe error: %v", err)
	}
out:
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := stream.Recv()
			if err == io.EOF {
				break out
			}
			if err != nil {
				log.Fatalf("Subscribe error: %v", err)
			}
			log.Println(msg)
		}
	}
}

func Fetch(ctx context.Context, client pb.BrokerClient) {
	msg, err := client.Fetch(ctx, &pb.FetchRequest{Subject: "test", Id: 1})
	if err != nil {
		log.Fatalf("Fetch error: %v", err)
	}
	log.Println(msg)
}

func PublishNewConnection() {
	conn, err := grpc.Dial("192.168.59.100:30456", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewBrokerClient(conn)
	ctx := context.Background()
	_, err = client.Publish(ctx, &pb.PublishRequest{Subject: "test", Body: []byte("randomString(10)"), ExpirationSeconds: 10})
	if err != nil {
		log.Printf("Publish error: %v", err)
	}
}
