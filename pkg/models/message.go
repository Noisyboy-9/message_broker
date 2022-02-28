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
	err := msg.db.QueryRow(
		msg.dbCtx,
		"INSERT INTO messages(topic_id, body, created_at, expired_at, deleted_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		msg.TopicID,
		msg.Body,
		msg.CreatedAT,
		msg.ExpiredAt,
		msg.DeletedAt,
	).Scan(&msg.Id)

	if err != nil {
		log.Fatalf("message save err: %v", err)
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
