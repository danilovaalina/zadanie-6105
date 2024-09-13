package postgres

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultDB = "postgres://postgres:postgres@localhost:5432/tender"

func Pool() (*pgxpool.Pool, error) {
	db := os.Getenv("POSTGRES_CONN")
	if db == "" {
		db = defaultDB
	}

	pool, err := pgxpool.New(context.Background(), db)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
