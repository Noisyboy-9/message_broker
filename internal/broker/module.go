package broker

import (
	"context"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
	"therealbroker/pkg/broker"
	"therealbroker/pkg/message"
)

type Module struct {
	Listeners                 map[string][]chan message.Message
	MessagesPerSubject        map[string][]message.Message
	MessageExpirationTime     map[message.Message]time.Time
	lastPublishId             int
	IsClosed                  bool
	ListenersLock             sync.Mutex
	MessagesPerSubjectLock    sync.Mutex
	MessageExpirationTimeLock sync.Mutex
}

var (
	config = jaegercfg.Configuration{
		ServiceName: "inside module",
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}
	jLogger         = jaegerlog.StdLogger
	jMetricsFactory = metrics.NullFactory
)

func NewModule() broker.Broker {
	return &Module{
		Listeners:                 make(map[string][]chan message.Message),
		MessagesPerSubject:        make(map[string][]message.Message),
		MessageExpirationTime:     make(map[message.Message]time.Time),
		IsClosed:                  false,
		ListenersLock:             sync.Mutex{},
		MessagesPerSubjectLock:    sync.Mutex{},
		MessageExpirationTimeLock: sync.Mutex{},
	}
}

func (m *Module) Close() error {
	if m.IsClosed {
		return broker.ErrUnavailable
	}

	m.IsClosed = true
	return nil
}

func (m *Module) Publish(_ context.Context, subject string, msg message.Message) (int, error) {
	if m.IsClosed {
		return -1, broker.ErrUnavailable
	}
	tracer, closer, err := config.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		return 0, err
	}
	defer closer.Close()

	opentracing.SetGlobalTracer(tracer)

	expirationSpan := tracer.StartSpan("message expiration setting")
	m.MessageExpirationTimeLock.Lock()
	m.MessageExpirationTime[msg] = time.Now().Add(msg.Expiration)
	msg.SetId(m.lastPublishId + 1)
	m.lastPublishId += 1
	m.MessageExpirationTimeLock.Unlock()
	expirationSpan.Finish()

	messagePerSubjectSpan := tracer.StartSpan("setting message per subject")
	m.MessagesPerSubjectLock.Lock()
	m.MessagesPerSubject[subject] = append(m.MessagesPerSubject[subject], msg)
	m.MessagesPerSubjectLock.Unlock()
	messagePerSubjectSpan.Finish()

	pushingMessageToChannelsSpan := tracer.StartSpan("push message to channels span")
	var wg sync.WaitGroup
	for _, listener := range m.Listeners[subject] {
		wg.Add(1)
		go func(listener chan message.Message) {
			defer wg.Done()
			listener <- msg
		}(listener)
	}
	wg.Wait()
	pushingMessageToChannelsSpan.Finish()

	return msg.GetId(), nil
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan message.Message, error) {
	if m.IsClosed {
		return nil, broker.ErrUnavailable
	}

	select {
	case <-ctx.Done():
		return nil, broker.ErrExpiredID
	default:
		newChannel := make(chan message.Message, 100)

		m.ListenersLock.Lock()
		_, exist := m.Listeners[subject]
		m.ListenersLock.Unlock()

		if exist {
			m.ListenersLock.Lock()
			m.Listeners[subject] = append(m.Listeners[subject], newChannel)
			m.ListenersLock.Unlock()

			return newChannel, nil
		}

		m.ListenersLock.Lock()
		m.Listeners[subject] = append(make([]chan message.Message, 0), newChannel)
		m.ListenersLock.Unlock()

		m.MessagesPerSubjectLock.Lock()
		m.MessagesPerSubject[subject] = make([]message.Message, 0)
		m.MessagesPerSubjectLock.Unlock()

		return newChannel, nil
	}
}

func (m *Module) Fetch(_ context.Context, subject string, id int) (message.Message, error) {
	if m.IsClosed {
		return message.Message{}, broker.ErrUnavailable
	}

	for _, msg := range m.MessagesPerSubject[subject] {
		if msg.GetId() == id {
			// found the message check if it is expired or not
			expireTime := m.MessageExpirationTime[msg]

			if expireTime.After(time.Now()) {
				return msg, nil
			}

			return message.Message{}, broker.ErrExpiredID
		}
	}

	return message.Message{}, broker.ErrInvalidID
}
