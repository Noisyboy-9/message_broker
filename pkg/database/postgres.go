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
	host := "localhost"
	port := "5432"
	dbName := "broker"
	ctx := context.Background()

	connectionString := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable", username, password, host, port, dbName)
	pool, err := pgxpool.Connect(ctx, connectionString)

	if err != nil {
		return nil, nil, err
	}

	return pool, ctx, err
}

func migrate(database *pgxpool.Pool, ctx context.Context) (*pgxpool.Pool, context.Context) {
	// create topics table
	_, err := database.Exec(ctx, "DROP TABLE IF EXISTS topics")
	if err != nil {
		log.Fatalf("drop table topics err: %v", err)
	}

	_, err = database.Exec(ctx, "CREATE TABLE topics("+
		"id SERIAL NOT NULL PRIMARY KEY,"+
		"name VARCHAR(255) NOT NULL UNIQUE,"+
		"created_at TIMESTAMP NOT NULL,"+
		"updated_at TIMESTAMP NOT NULL,"+
		"deleted_at TIMESTAMP"+
		")")

	if err != nil {
		log.Fatalf("topics migration err: %v", err)
	}

	// create users table
	_, err = database.Exec(ctx, "DROP TABLE IF EXISTS messages")
	if err != nil {
		log.Fatalf("drop table topics err: %v", err)
	}
	_, err = database.Exec(ctx, "CREATE TABLE messages("+
		"id SERIAL NOT NULL PRIMARY KEY,"+
		"topic_id INT NOT NULL,"+
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
