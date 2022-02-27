package models

import (
	"context"
	"log"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Topic struct {
	Model
	Id        int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func (topic *Topic) Save(db *pgxpool.Pool, ctx context.Context) *Topic {
	err := topic.db.QueryRow(
		topic.dbCtx,
		"INSERT INTO topics (name, created_at, updated_at, deleted_at) VALUES ($1, $2, $3, $4) RETURNING id",
		topic.Name,
		topic.CreatedAt,
		topic.UpdatedAt,
		topic.DeletedAt,
	).Scan(&topic.Id)

	if err != nil {
		log.Fatalf("topic save err: %v", err)
	}

	return topic
}

func (topic *Topic) Messages() (messages []*Message) {
	if err := pgxscan.Select(
		topic.dbCtx,
		topic.db,
		&messages,
		"SELECT * FROM messages WHERE topic_id = $1",
		topic.Id,
	); err != nil {
		log.Fatalf("something went wrong for reading topic: %v", err)
	}

	return
}

func GetOrCreateTopicByName(dbConnection *pgxpool.Pool, dbCtx context.Context, name string) *Topic {
	return &Topic{
		Model: Model{
			db:    dbConnection,
			dbCtx: dbCtx,
		},
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// if TopicExist(dbConnection, dbCtx, name) {
	// 	return GetTopicByName(dbConnection, dbCtx, name)
	// }
	//
	// // 	topic doesn't exist create and persist it
	// topic := Topic{
	// 	Model: Model{
	// 		db:    dbConnection,
	// 		dbCtx: dbCtx,
	// 	},
	// 	Name:      name,
	// 	CreatedAt: time.Now(),
	// 	UpdatedAt: time.Now(),
	// }
	//
	// return topic.Save(dbConnection, dbCtx)
}

func GetTopicByName(db *pgxpool.Pool, ctx context.Context, name string) *Topic {
	topic := &Topic{}
	err := db.QueryRow(ctx, "SELECT * FROM topics WHERE name = $1", name).Scan(&topic.Id, &topic.Name, &topic.CreatedAt, &topic.UpdatedAt, &topic.DeletedAt)

	if err != nil {
		log.Fatalf("get topic by name read err: %v", err)
	}

	topic.db = db
	topic.dbCtx = ctx

	return topic
}

func TopicExist(db *pgxpool.Pool, ctx context.Context, name string) (exist bool) {
	rows, err := db.Query(ctx, "SELECT id FROM topics WHERE name = $1", name)
	if err == pgx.ErrNoRows {
		return false
	} else if err != nil {
		log.Fatalf("something went wrong for topic exist: %v", err)
	}

	rowsCount := 0
	for rows.Next() {
		rowsCount++
	}
	return rowsCount != 0
}
