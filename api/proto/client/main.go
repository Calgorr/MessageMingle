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
	var wg *sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		Publish(ctx, client)
	}()
	go func() {
		defer wg.Done()
		Subscribe(ctx, client)
	}()
	go func() {
		defer wg.Done()
		Fetch(ctx, client)
	}()
	wg.Wait()
}

func Publish(ctx context.Context, client pb.BrokerClient) {
	ticker := time.NewTicker(20 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, err := client.Publish(ctx, &pb.PublishRequest{
				Subject:           "test",
				Body:              []byte(randomString(10)),
				ExpirationSeconds: 0,
			})
			if err != nil {
				log.Printf("Publish error: %v", err)
			}
		}
	}

}

func Subscribe(ctx context.Context, client pb.BrokerClient) {
	stream, err := client.Subscribe(ctx, &pb.SubscribeRequest{Subject: "test"})
	if err != nil {
		log.Fatalf("Subscribe error: %v", err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := stream.Recv()
			if err == io.EOF {
				break
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
