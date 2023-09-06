package main

import (
	"context"
	"io"
	"log"
	"math/rand"
	"sync"
	"time"

	pb "therealbroker/api/proto/protoGen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewBrokerClient(conn)
	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan bool, 1)
	go func() {
		defer wg.Done()
		Publish(ctx, client, done)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(20 * time.Minute)
		done <- true
	}()
	wg.Wait()
}

func Publish(ctx context.Context, client pb.BrokerClient, done chan bool) {
	var wg sync.WaitGroup
	ticker := time.NewTicker(50 * time.Microsecond) // 20k msg/s 1,2M msg/min 24M msg in 20 min
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
				_, err := client.Publish(ctx, &pb.PublishRequest{Subject: "test", Body: []byte(randomString(10)), ExpirationSeconds: 10})
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

func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
