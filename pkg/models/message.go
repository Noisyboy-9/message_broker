package models

import (
	"context"
	"log"
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
	_, err := msg.db.Exec(msg.dbCtx, "INSERT INTO messages() VALUES ($1, $2, $3, $4, $5)", msg.TopicID, msg.Body, msg.CreatedAT, msg.ExpiredAt, msg.DeletedAt)
	if err != nil {
		log.Fatalf("message write error: %v", err)
	}

	return msg
}

func CreateMessage(db *pgxpool.Pool, ctx context.Context, topic *Topic, body string, expirationSecondsCount int32) *Message {
	expirationDuration := time.Duration(expirationSecondsCount) * time.Second

	message := &Message{
		Model:     Model{db: db, dbCtx: ctx},
		TopicID:   topic.Id,
		Body:      body,
		CreatedAT: time.Now(),
		ExpiredAt: time.Now().Add(expirationDuration),
		DeletedAt: time.Time{},
	}

	return message.Save()
}
