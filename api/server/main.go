package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"therealbroker/api/pb/api/proto"
	broker2 "therealbroker/internal/broker"
	"therealbroker/pkg/broker"
	"therealbroker/pkg/message"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

var (
	brokerPort             = 9000
	internalPrometheusPort = 8000
)

type Server struct {
	proto.UnimplementedBrokerServer
	brokerInstance broker.Broker
}

var (
	MethodCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rpc_request_get_method_count",
			Help: "Count of each rpc call.",
		},
		[]string{"method", "status"},
	)

	MethodDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "rpc_request_get_method_duration",
			Help:       "Measure the duration of each RPC request.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"rpc_duration"},
	)

	ActiveSubscribersGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_subscribers",
			Help: "Number of active subscribers",
		})
)

func (s Server) Publish(ctx context.Context, request *proto.PublishRequest) (*proto.PublishResponse, error) {
	publishStartTime := time.Now()

	log.Println("Getting publish request")
	defer log.Println("Finish handling publish request")

	publishId, err := s.brokerInstance.Publish(ctx, request.Subject, message.Message{
		Body:       string(request.Body),
		Expiration: time.Duration(request.ExpirationSeconds),
	})

	publishDuration := time.Since(publishStartTime)
	MethodDuration.WithLabelValues("publish_duration").Observe(float64(publishDuration) / float64(time.Millisecond))

	if err != nil {
		MethodCount.WithLabelValues("publish", "failed").Inc()
		return nil, err
	} else {
		MethodCount.WithLabelValues("publish", "successful").Inc()
	}

	return &proto.PublishResponse{Id: int32(publishId)}, nil
}

func (s Server) Subscribe(request *proto.SubscribeRequest, server proto.Broker_SubscribeServer) error {
	fmt.Println("Subscriber request received.")
	var subscribeError error

	SubscribedChannel, err := s.brokerInstance.Subscribe(server.Context(), request.Subject)
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

	msg, err := s.brokerInstance.Fetch(ctx, request.Subject, int(request.Id))

	if err != nil {
		MethodCount.WithLabelValues("fetch", "failed").Inc()
		return nil, err
	}

	fetchDuration := time.Since(fetchStartTime)
	MethodDuration.WithLabelValues("fetch_duration").Observe(float64(fetchDuration) / float64(time.Millisecond))
	MethodCount.WithLabelValues("fetch", "successful").Inc()

	return &proto.MessageResponse{
		Body: []byte(msg.Body),
	}, nil
}

func startPrometheusServer() {
	if (prometheus.Register(MethodCount) != nil) || (prometheus.Register(MethodDuration) != nil) ||
		(prometheus.Register(ActiveSubscribersGauge) != nil) {
		fmt.Println("Some error with prometheus")
		return
	}

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", internalPrometheusPort), nil))
}

func main() {
	go startPrometheusServer()
	log.Println("************** Sina Shariati Broker - Bale 1400 winter bootcamp*************")

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", brokerPort))

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
