package broker

import (
	"context"
	"sync"
	"time"

	"therealbroker/pkg/broker"
	"therealbroker/pkg/message"
)

type Module struct {
	listeners                 map[string][]chan message.Message
	messagesPerSubject        map[string][]message.Message
	messageExpirationTime     map[message.Message]time.Time
	isClosed                  bool
	listenersLock             sync.Mutex
	messagesPerSubjectLock    sync.Mutex
	messageExpirationTimeLock sync.Mutex
}

func NewModule() broker.Broker {
	return &Module{
		listeners:                 make(map[string][]chan message.Message),
		messagesPerSubject:        make(map[string][]message.Message),
		messageExpirationTime:     make(map[message.Message]time.Time),
		isClosed:                  false,
		listenersLock:             sync.Mutex{},
		messagesPerSubjectLock:    sync.Mutex{},
		messageExpirationTimeLock: sync.Mutex{},
	}
}

func (m *Module) Close() error {
	if m.isClosed {
		return broker.ErrUnavailable
	}

	m.isClosed = true
	return nil
}

func (m *Module) Publish(_ context.Context, subject string, msg message.Message) (int, error) {
	if m.isClosed {
		return -1, broker.ErrUnavailable
	}

	m.messageExpirationTimeLock.Lock()
	m.messageExpirationTime[msg] = time.Now().Add(msg.Expiration)
	m.messageExpirationTimeLock.Unlock()

	m.messagesPerSubjectLock.Lock()
	m.messagesPerSubject[subject] = append(m.messagesPerSubject[subject], msg)
	m.messagesPerSubjectLock.Unlock()

	var wg sync.WaitGroup

	for _, listener := range m.listeners[subject] {
		wg.Add(1)
		go func(listener chan message.Message) {
			defer wg.Done()
			listener <- msg
		}(listener)
	}
	wg.Wait()
	return msg.GetId(), nil
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan message.Message, error) {
	if m.isClosed {
		return nil, broker.ErrUnavailable
	}

	select {
	case <-ctx.Done():
		return nil, broker.ErrExpiredID
	default:
		newChannel := make(chan message.Message, 100)

		m.listenersLock.Lock()
		_, exist := m.listeners[subject]
		m.listenersLock.Unlock()

		if exist {
			m.listenersLock.Lock()
			m.listeners[subject] = append(m.listeners[subject], newChannel)
			m.listenersLock.Unlock()

			return newChannel, nil
		}

		m.listenersLock.Lock()
		m.listeners[subject] = append(make([]chan message.Message, 0), newChannel)
		m.listenersLock.Unlock()

		m.messagesPerSubjectLock.Lock()
		m.messagesPerSubject[subject] = make([]message.Message, 0)
		m.messagesPerSubjectLock.Unlock()

		return newChannel, nil
	}
}

func (m *Module) Fetch(_ context.Context, subject string, id int) (message.Message, error) {
	if m.isClosed {
		return message.Message{}, broker.ErrUnavailable
	}

	for _, msg := range m.messagesPerSubject[subject] {
		if msg.GetId() == id {
			// found the message check if it is expired or not
			expireTime := m.messageExpirationTime[msg]

			if expireTime.After(time.Now()) {
				return msg, nil
			}

			return message.Message{}, broker.ErrExpiredID
		}
	}

	return message.Message{}, broker.ErrInvalidID
}
