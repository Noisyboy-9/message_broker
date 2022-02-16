package main

import (
	"context"
	"log"
	"net"
	"time"

	"therealbroker/api/pb/api/proto"
	broker2 "therealbroker/internal/broker"
	"therealbroker/pkg/broker"
	"therealbroker/pkg/message"

	"google.golang.org/grpc"
)

type Server struct {
	proto.UnimplementedBrokerServer
	brokerInstance broker.Broker
}

func (s Server) Publish(ctx context.Context, request *proto.PublishRequest) (*proto.PublishResponse, error) {
	log.Println("Getting publish request")

	publishId, err := s.brokerInstance.Publish(ctx, request.Subject, message.Message{
		Body:       string(request.Body),
		Expiration: time.Duration(request.ExpirationSeconds),
	})

	if err != nil {
		return nil, err
	}

	return &proto.PublishResponse{Id: int32(publishId)}, nil
}

func (s Server) Subscribe(request *proto.SubscribeRequest, server proto.Broker_SubscribeServer) error {
	// TODO implement me
	panic("implement me")
}

func (s Server) Fetch(ctx context.Context, request *proto.FetchRequest) (*proto.MessageResponse, error) {
	// TODO implement me
	panic("implement me")
}

func main() {
	log.Println("************** Sina Shariati Broker - Bale 1400 winter bootcamp")
	listener, err := net.Listen("tcp", "127.0.0.1:8080")

	if err != nil {
		log.Fatalf("Cant start listener: %v", err)
	}

	server := grpc.NewServer()
	proto.RegisterBrokerServer(server, &Server{
		brokerInstance: broker2.NewModule(),
	})

	log.Printf("Server starting at: %s", listener.Addr())

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Server serve failed: %v", err)
	}
}
