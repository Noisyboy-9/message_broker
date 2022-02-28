package models

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Message struct {
	Model
	Id        int
	TopicID   int
	Body      string
	CreatedAT time.Time
	ExpiredAt time.Time
	DeletedAt time.Time
}

func (msg *Message) Save() *Message {
	_, err := msg.db.Exec(
		msg.dbCtx,
		"INSERT INTO messages(id, topic_id, body, created_at, expired_at, deleted_at) VALUES ($1, $2, $3, $4, $5, $6)",
		msg.Id,
		msg.TopicID,
		msg.Body,
		msg.CreatedAT,
		msg.ExpiredAt,
		msg.DeletedAt,
	)

	if err != nil {
		log.Fatalf("message save err: %v", err)
	}

	return msg
}

func CreateMessage(db *pgxpool.Pool, ctx context.Context, topic *Topic, body string, expirationSecondsCount int32, lastId *int, lock *sync.Mutex) *Message {
	expirationDuration := time.Duration(expirationSecondsCount) * time.Second

	lock.Lock()
	*lastId += 1
	message := &Message{
		Id:        *lastId,
		Model:     Model{db: db, dbCtx: ctx},
		TopicID:   topic.Id,
		Body:      body,
		CreatedAT: time.Now(),
		ExpiredAt: time.Now().Add(expirationDuration),
		DeletedAt: time.Time{},
	}
	lock.Unlock()

	return message.Save()
}
