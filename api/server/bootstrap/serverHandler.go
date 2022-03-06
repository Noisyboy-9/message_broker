package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"therealbroker/api/pb/api/proto"
	"therealbroker/pkg/broker"
	"therealbroker/pkg/models"

	"go.opentelemetry.io/otel"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Server struct {
	proto.UnimplementedBrokerServer
	BrokerInstance  broker.Broker
	Database        *pgxpool.Pool
	DatabaseContext context.Context

	LastPublishLock *sync.Mutex
	LastTopicLock   *sync.Mutex
	LastPublishId   int
	LastTopicId     int

	MessageBatchString   *strings.Builder
	BatchNotifierPointer *chan struct{}
}

func (s *Server) Publish(globalContext context.Context, request *proto.PublishRequest) (*proto.PublishResponse, error) {
	log.Println("inside publish function")
	globalContext, globalSpan := otel.Tracer("Server").Start(globalContext, "publish method")
	publishStartTime := time.Now()

	_, topicSpan := otel.Tracer("Server").Start(globalContext, "Get or create topics by name")
	topic := models.GetOrCreateTopicByName(s.Database, s.DatabaseContext, request.Subject, &s.LastTopicId, s.LastTopicLock)
	topicSpan.End()

	_, msgSpan := otel.Tracer("Server").Start(globalContext, "create message")
	msg := models.CreateMessage(s.Database, s.DatabaseContext, topic, string(request.Body), &s.LastPublishId, s.LastPublishLock, s.MessageBatchString)
	msgSpan.End()

	publishId, err := s.BrokerInstance.Publish(globalContext, topic, msg)

	publishDuration := time.Since(publishStartTime)
	MethodDuration.WithLabelValues("publish_duration").Observe(float64(publishDuration) / float64(time.Nanosecond))

	if err != nil {
		MethodCount.WithLabelValues("publish", "failed").Inc()
		return nil, err
	}

	MethodCount.WithLabelValues("publish", "successful").Inc()

	globalSpan.End()

	<-*s.BatchNotifierPointer
	return &proto.PublishResponse{Id: int32(publishId)}, nil
}

func (s *Server) Subscribe(request *proto.SubscribeRequest, server proto.Broker_SubscribeServer) error {
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

func (s *Server) Fetch(ctx context.Context, request *proto.FetchRequest) (*proto.MessageResponse, error) {
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
	MethodDuration.WithLabelValues("fetch_duration").Observe(float64(fetchDuration) / float64(time.Nanosecond))
	MethodCount.WithLabelValues("fetch", "successful").Inc()

	return &proto.MessageResponse{Body: []byte(msg.Body)}, nil
}
