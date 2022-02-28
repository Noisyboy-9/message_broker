package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"therealbroker/api/pb/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	subject = "sub1"
)

func main() {
	connection, err := grpc.Dial("localhost:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))

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

	var wg sync.WaitGroup
	ticker := time.NewTicker(150 * time.Microsecond) // 0.5 billion request in 20 minutes

	doneIndicator := make(chan bool)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-doneIndicator:
				return
			case <-ticker.C:
				body := fmt.Sprintf("some text for testing : %v", time.Now())
				go pushToSubject(client, ctx, subject, body, int(10*time.Hour))
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(20 * time.Minute)
		ticker.Stop()
		doneIndicator <- true
	}()
	wg.Wait()
}

func pushToSubject(client proto.BrokerClient, ctx context.Context, subject string, body string, expire int) {
	response, err := client.Publish(ctx, &proto.PublishRequest{
		Subject:           subject,
		Body:              []byte(body),
		ExpirationSeconds: int32(expire),
	})

	if err != nil {
		log.Fatalf("publish to subject failed: %s", err)
	}

	log.Println("response Id: ", response.Id)
}
