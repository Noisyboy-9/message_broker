package broker

import (
	"context"
	"sync"

	"therealbroker/pkg/broker"
	"therealbroker/pkg/models"
)

type Module struct {
	subscribers   map[*models.Topic][]chan models.Message
	IsClosed      bool
	ListenersLock sync.Mutex
}

func NewModule() broker.Broker {
	return &Module{
		subscribers:   make(map[*models.Topic][]chan models.Message),
		IsClosed:      false,
		ListenersLock: sync.Mutex{},
	}
}

func (m *Module) Close() error {
	if m.IsClosed {
		return broker.ErrUnavailable
	}

	m.IsClosed = true
	return nil
}

func (m *Module) Publish(ctx context.Context, topic *models.Topic, msg *models.Message) (int, error) {
	if m.IsClosed {
		return -1, broker.ErrUnavailable
	}

	for _, listener := range m.subscribers[topic] {
		go func(listener chan models.Message) {
			if cap(listener) != len(listener) {
				listener <- *msg
			}
		}(listener)
	}

	return msg.Id, nil
}

func (m *Module) Subscribe(ctx context.Context, topic *models.Topic) (<-chan models.Message, error) {
	if m.IsClosed {
		return nil, broker.ErrUnavailable
	}

	select {
	case <-ctx.Done():
		return nil, broker.ErrExpiredID
	default:
		newChannel := make(chan models.Message, 100)
		m.ListenersLock.Lock()
		m.subscribers[topic] = append(make([]chan models.Message, 0), newChannel)
		m.ListenersLock.Unlock()

		return newChannel, nil
	}
}

func (m *Module) Fetch(_ context.Context, topic *models.Topic, id int) (models.Message, error) {
	if m.IsClosed {
		return models.Message{}, broker.ErrUnavailable
	}

	for _, msg := range topic.Messages() {
		if msg.Id == id {
			return *msg, nil
		}
	}

	return models.Message{}, broker.ErrInvalidID
}
