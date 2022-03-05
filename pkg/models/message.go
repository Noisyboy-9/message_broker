package models

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Message struct {
	Model
	Id      int
	TopicID int
	Body    string
}

func (msg *Message) Save(b *strings.Builder) *Message {
	query := fmt.Sprintf(
		"(%d, %d, '%s'),",
		msg.Id,
		msg.TopicID,
		"some body for testing",
	)
	b.WriteString(query)

	return msg
}

func CreateMessage(db *pgxpool.Pool, ctx context.Context, topic *Topic, body string, lastId *int, lock *sync.Mutex, builder *strings.Builder) *Message {
	lock.Lock()
	*lastId += 1
	message := &Message{
		Id:      *lastId,
		Model:   Model{db: db, dbCtx: ctx},
		TopicID: topic.Id,
		Body:    body,
	}
	lock.Unlock()

	return message.Save(builder)
}
