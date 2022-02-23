package models

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Model struct {
	db    *pgxpool.Pool
	dbCtx context.Context
}
