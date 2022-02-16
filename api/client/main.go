package main

import (
	"context"
	"log"
	"time"

	"therealbroker/api/pb/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	sub = "sub1"
)

func main() {
	connection, err := grpc.Dial("127.0.0.1:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("can't connect to server: %v", err)
	}

	defer func(clientConn *grpc.ClientConn) {
		err := clientConn.Close()
		if err != nil {
			log.Fatalf("can't close server connection: %v", err)
		}
	}(connection)

	client := proto.NewBrokerClient(connection)
	ctx := context.Background()

	for i := 0; i < 50; i++ {
		pushToSubject(client, ctx, sub, "some body for testing", 2*time.Hour)
	}

	subscribeToSubject(client, ctx, sub)
}

func subscribeToSubject(client proto.BrokerClient, ctx context.Context, subject string) {
	subClient, err := client.Subscribe(ctx, &proto.SubscribeRequest{Subject: subject})
	if err != nil {
		log.Fatalf("subscribe to broker failed: %s", err)
	}

	for {
		messageResponse, err := subClient.Recv()
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		log.Println("subscribe response", string(messageResponse.Body))
	}
}

func pushToSubject(client proto.BrokerClient, ctx context.Context, subject string, body string, expire time.Duration) {
	publishResponse, err := client.Publish(ctx, &proto.PublishRequest{
		Subject:           subject,
		Body:              []byte(body),
		ExpirationSeconds: int32(expire),
	})

	if err != nil {
		log.Fatalf("publish to subject failed: %s", err)
	}

	log.Printf("server responded with message id: %d", publishResponse.Id)
}
