package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

func connect() (*pgxpool.Pool, context.Context, error) {
	username := "postgres"
	password := "postgres"
	host := "postgres"
	port := 5432
	dbName := "broker"
	ctx := context.Background()

	pool, err := pgxpool.Connect(ctx, fmt.Sprintf("postgres://%s:%s@%s:%d/%s", username, password, host, port, dbName))
	if err != nil {
		return nil, nil, err
	}

	return pool, ctx, err
}

func migrate(database *pgxpool.Pool, ctx context.Context) (*pgxpool.Pool, context.Context) {
	// create topics table
	_, err := database.Exec(ctx, "CREATE DATABASE topics("+
		"id INT NOT NULL PRIMARY KEY AUTO INCREMENT,"+
		"name VARCHAR(255) NOT NULL UNIQUE,"+
		"created_at TIMESTAMP NOT NULL,"+
		"deleted_at TIMESTAMP"+
		")")

	if err != nil {
		log.Fatalf("topics migration err: %v", err)
	}

	// create users table
	_, err = database.Exec(ctx, "CREATE DATABASE messages("+
		"id INT NOT NULL PRIMARY KEY AUTO INCREMENT,"+
		"topic_id INT NOT NULL PRIMARY KEY AUTO INCREMENT,"+
		"body TEXT NOT NULL,"+
		"created_at TIMESTAMP NOT NULL,"+
		"expired_at TIMESTAMP NOT NULL,"+
		"deleted_at TIMESTAMP"+
		")")
	if err != nil {
		log.Fatalf("messages migration err: %v", err)
	}

	return database, ctx
}

func Setup() (*pgxpool.Pool, context.Context) {
	db, ctx, err := connect()
	if err != nil {
		log.Fatalf("postgress connection err: %v", err)
	}

	return migrate(db, ctx)
}
