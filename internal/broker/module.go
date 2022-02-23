package broker

import (
	"context"
	"sync"
	"time"

	"therealbroker/pkg/broker"
	"therealbroker/pkg/models"
)

type Module struct {
	Listeners                 map[string][]chan models.Message
	MessagesPerSubject        map[string][]models.Message
	MessageExpirationTime     map[models.Message]time.Time
	lastPublishId             int
	IsClosed                  bool
	ListenersLock             sync.Mutex
	MessagesPerSubjectLock    sync.Mutex
	MessageExpirationTimeLock sync.Mutex
}

func NewModule() broker.Broker {
	return &Module{
		Listeners:                 make(map[string][]chan models.Message),
		MessagesPerSubject:        make(map[string][]models.Message),
		MessageExpirationTime:     make(map[models.Message]time.Time),
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

func (m *Module) Publish(ctx context.Context, topic *models.Topic, msg *models.Message) (int, error) {
	if m.IsClosed {
		return -1, broker.ErrUnavailable
	}
	for _, listener := range m.Listeners[topic.Name] {
		go func(listener chan models.Message) {
			if cap(listener) != len(listener) {
				listener <- *msg
			}
		}(listener)
	}

	return msg.Id, nil
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan models.Message, error) {
	if m.IsClosed {
		return nil, broker.ErrUnavailable
	}

	select {
	case <-ctx.Done():
		return nil, broker.ErrExpiredID
	default:
		newChannel := make(chan models.Message, 100)

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
		m.Listeners[subject] = append(make([]chan models.Message, 0), newChannel)
		m.ListenersLock.Unlock()

		m.MessagesPerSubjectLock.Lock()
		m.MessagesPerSubject[subject] = make([]models.Message, 0)
		m.MessagesPerSubjectLock.Unlock()

		return newChannel, nil
	}
}

func (m *Module) Fetch(_ context.Context, subject string, id int) (models.Message, error) {
	if m.IsClosed {
		return models.Message{}, broker.ErrUnavailable
	}

	for _, msg := range m.MessagesPerSubject[subject] {
		if msg.Id == id {
			// found the message check if it is expired or not
			expireTime := m.MessageExpirationTime[msg]

			if expireTime.After(time.Now()) {
				return msg, nil
			}

			return models.Message{}, broker.ErrExpiredID
		}
	}

	return models.Message{}, broker.ErrInvalidID
}
