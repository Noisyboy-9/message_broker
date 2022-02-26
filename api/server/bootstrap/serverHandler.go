package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"therealbroker/api/pb/api/proto"
	"therealbroker/pkg/broker"
	"therealbroker/pkg/models"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Server struct {
	proto.UnimplementedBrokerServer
	BrokerInstance  broker.Broker
	Database        *pgxpool.Pool
	DatabaseContext context.Context
}

func (s Server) Publish(ctx context.Context, request *proto.PublishRequest) (*proto.PublishResponse, error) {
	publishStartTime := time.Now()
	log.Println("Getting publish request")
	defer log.Println("Finish handling publish request")

	topic := models.GetOrCreateTopicByName(s.Database, s.DatabaseContext, request.Subject)
	msg := models.CreateMessage(s.Database, s.DatabaseContext, topic, string(request.Body), request.ExpirationSeconds)

	publishId, err := s.BrokerInstance.Publish(ctx, topic, msg)

	publishDuration := time.Since(publishStartTime)
	MethodDuration.WithLabelValues("publish_duration").Observe(float64(publishDuration) / float64(time.Millisecond))

	if err != nil {
		MethodCount.WithLabelValues("publish", "failed").Inc()
		return nil, err
	}

	MethodCount.WithLabelValues("publish", "successful").Inc()
	return &proto.PublishResponse{Id: int32(publishId)}, nil
}

func (s Server) Subscribe(request *proto.SubscribeRequest, server proto.Broker_SubscribeServer) error {
	fmt.Println("Subscriber request received.")
	var subscribeError error

	SubscribedChannel, err := s.BrokerInstance.Subscribe(
		server.Context(),
		models.GetTopicByName(s.Database, s.DatabaseContext, request.Subject),
	)

	if err != nil {
		MethodCount.WithLabelValues("subscribe", "failed").Inc()
		return err
	}

	ActiveSubscribersGauge.Inc()
	ctx := server.Context()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			select {
			case msg, closed := <-SubscribedChannel:
				if closed {
					ActiveSubscribersGauge.Dec()
					wg.Done()
					return
				}

				if err := server.Send(&(proto.MessageResponse{Body: []byte(msg.Body)})); err != nil {
					subscribeError = err
				}
			case <-ctx.Done():
				ActiveSubscribersGauge.Dec()
				subscribeError = errors.New("context timeout reached")
				wg.Done()
				return
			}
		}
	}()

	wg.Wait()
	if subscribeError == nil {
		MethodCount.WithLabelValues("subscribe", "successful").Inc()
	} else {
		MethodCount.WithLabelValues("subscribe", "failed").Inc()
	}
	return subscribeError
}

func (s Server) Fetch(ctx context.Context, request *proto.FetchRequest) (*proto.MessageResponse, error) {
	fetchStartTime := time.Now()

	log.Println("Getting fetch request")
	defer log.Println("Finish handling fetch request")

	topic := models.GetTopicByName(s.Database, s.DatabaseContext, request.Subject)
	msg, err := s.BrokerInstance.Fetch(ctx, topic, int(request.Id))

	if err != nil {
		MethodCount.WithLabelValues("fetch", "failed").Inc()
		return nil, err
	}

	fetchDuration := time.Since(fetchStartTime)
	MethodDuration.WithLabelValues("fetch_duration").Observe(float64(fetchDuration) / float64(time.Millisecond))
	MethodCount.WithLabelValues("fetch", "successful").Inc()

	return &proto.MessageResponse{Body: []byte(msg.Body)}, nil
}