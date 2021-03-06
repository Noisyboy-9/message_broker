package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

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
	// exporter.Register()
}

func main() {
	go log.Println("************** Sina Shariati Broker - Bale 1400 winter bootcamp*************")
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", brokerPort))

	if err != nil {
		log.Fatalf("Cant start listener: %v", err)
	}

	server := grpc.NewServer()

	batchNotifier := make(chan struct{})
	batchNotifierPointer := &batchNotifier

	batchBuilder := &strings.Builder{}
	batchBuilder.WriteString("INSERT INTO messages(id, topic_id, body) VALUES ")
	batchBuilder.Grow(5 * 1e8)

	go func() {
		for {
			newChannel := make(chan struct{})
			time.Sleep(100 * time.Millisecond)

			if batchBuilder.Len() == 48 {
				continue
			}

			query := batchBuilder.String()[:len(batchBuilder.String())-1]
			query += ";"

			// _, err := db.Exec(dbContext, query)
			// if err != nil {
			// 	log.Fatalf("some err in batcher: %v", err)
			//
			// }
			close(*batchNotifierPointer)
			*batchNotifierPointer = newChannel

			// *batchBuilder = strings.Builder{}
			// batchBuilder.WriteString("INSERT INTO messages(id, topic_id, body) VALUES ")
		}
	}()

	proto.RegisterBrokerServer(
		server,
		&bootstrap.Server{
			BrokerInstance:       broker2.NewModule(),
			Database:             db,
			DatabaseContext:      dbContext,
			LastPublishLock:      &sync.Mutex{},
			LastTopicLock:        &sync.Mutex{},
			LastPublishId:        0,
			LastTopicId:          0,
			MessageBatchString:   batchBuilder,
			BatchNotifierPointer: batchNotifierPointer,
		},
	)

	log.Printf("Server starting at: %s", listener.Addr())

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Server serve failed: %v", err)
	}
}
