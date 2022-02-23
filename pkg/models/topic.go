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
	Id        int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func (topic *Topic) Save(db *pgxpool.Pool, ctx context.Context) *Topic {
	_, err := db.Exec(
		ctx,
		"INSERT INTO topics (name, created_at, updated_at, deleted_at) VALUES ($1, $2, $3, $4)",
		topic.Name,
		topic.CreatedAt,
		topic.UpdatedAt,
		topic.DeletedAt,
	)

	if err != nil {
		log.Fatalf("topic save err: %v", err)
	}

	return topic
}

func GetOrCreateTopicByName(dbConnection *pgxpool.Pool, dbCtx context.Context, name string) *Topic {
	if TopicExist(dbConnection, dbCtx, name) {
		return GetTopicByName(dbConnection, dbCtx, name)
	}

	// 	topic doesn't exist create and persist it
	topic := Topic{
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return topic.Save(dbConnection, dbCtx)
}

func GetTopicByName(db *pgxpool.Pool, ctx context.Context, name string) (topic *Topic) {
	if err := pgxscan.Select(ctx, db, &topic, "SELECT * FROM topics WHERE name = $1", name); err != nil {
		log.Fatalf("get topic by name err %v", err)
	}

	return
}

func TopicExist(db *pgxpool.Pool, ctx context.Context, name string) (exist bool) {
	if err := db.QueryRow(ctx, "SELECT EXIST(SELECT * FROM topics WHERE name = $1)", name).Scan(&exist); err == pgx.ErrNoRows {
		// topic doesn't exist
		exist = false
	} else if err != nil {
		// 	have some other error other than topic not existing
		log.Fatalf("something went wrong: %v", err)
	}

	// topic exist
	exist = true
	return
}
