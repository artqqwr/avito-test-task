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
	DSN            string
	MigrationsPath string
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	log.Println("Connecting to database...")
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN)

	if err != nil {
		return nil, fmt.Errorf("unable to parse config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping failed: %v", err)
	}

	log.Printf("Running migrations from %s\n", cfg.MigrationsPath)

	stdDB := stdlib.OpenDB(*poolConfig.ConnConfig)
	defer stdDB.Close()

	if err := migrate(stdDB, cfg.MigrationsPath); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	return pool, nil
}

const dialect = "postgres"

func migrate(db *sql.DB, migrationsPath string) error {
	if err := goose.SetDialect(dialect); err != nil {
		return err
	}

	if err := goose.Up(db, migrationsPath); err != nil {
		return err
	}

	return nil
}
