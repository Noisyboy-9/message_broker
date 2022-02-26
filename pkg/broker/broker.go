package broker

import (
	"context"
	"io"

	"therealbroker/pkg/models"
)

// The whole implementation should be thread-safe
// If any problem occurred, return the proper error based on errors.go

type Broker interface {
	io.Closer
	// Publish returns an int as the id of message published.
	// It should preserve the order. So if we are publishing messages
	// A, B and C, all subscribers should get these messages as
	// A, B and C.
	Publish(ctx context.Context, topic *models.Topic, msg *models.Message) (int, error)

	// Subscribe listens to every publish, and returns the messages to all
	// subscribed clients ( channels ).
	// If the context is cancelled, you have to stop sending messages
	// to this subscriber. Do nothing on time-out
	Subscribe(ctx context.Context, topic *models.Topic) (<-chan models.Message, error)

	// Fetch enables us to retrieve a message that is already published, if
	// it's not expired yet.
	Fetch(ctx context.Context, topic *models.Topic, id int) (models.Message, error)
}
