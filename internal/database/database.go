package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type Config struct {
	DSN string
}

const postgresDialect = "postgres"
const migrationsPath = "./migrations"

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	log.Println("Connecting to database...")
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN)

	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping failed: %v", err)
	}

	stdDB := stdlib.OpenDB(*poolConfig.ConnConfig)
	defer stdDB.Close()

	if err := migrate(stdDB, postgresDialect, migrationsPath); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	return pool, nil
}

func migrate(db *sql.DB, dialect string, migrationsPath string) error {
	log.Printf("Running migrations from %s\n", migrationsPath)

	if err := goose.SetDialect(dialect); err != nil {
		return err
	}

	if err := goose.Up(db, migrationsPath); err != nil {
		return err
	}

	return nil
}
