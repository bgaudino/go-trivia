package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func GetPool() (*pgxpool.Pool, error) {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "postgres:///trivia"
	}
	return pgxpool.New(context.Background(), dbUrl)
}
