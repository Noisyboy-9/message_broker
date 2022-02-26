package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"therealbroker/api/pb/api/proto"
	"therealbroker/api/server/bootstrap"
	broker2 "therealbroker/internal/broker"
	"therealbroker/pkg/database"

	"github.com/jackc/pgx/v4/pgxpool"

	"google.golang.org/grpc"
)

var (
	brokerPort = 9000
	db         *pgxpool.Pool
	dbContext  context.Context
)

func init() {
	go bootstrap.StartPrometheusServer()
	db, dbContext = database.Setup()
}

func main() {
	go log.Println("************** Sina Shariati Broker - Bale 1400 winter bootcamp*************")
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", brokerPort))

	if err != nil {
		log.Fatalf("Cant start listener: %v", err)
	}

	server := grpc.NewServer()
	proto.RegisterBrokerServer(server, &bootstrap.Server{
		BrokerInstance:  broker2.NewModule(),
		Database:        db,
		DatabaseContext: dbContext,
	})

	log.Printf("Server starting at: %s", listener.Addr())

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Server serve failed: %v", err)
	}
}
