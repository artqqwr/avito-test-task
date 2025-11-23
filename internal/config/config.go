package config

import (
	"avito-test-task/internal/database"
	"errors"
	"os"
)

type Config struct {
	Server struct {
		Port string
	}

	Database database.Config
}

const databaseDSNEnvKey = "DATABASE_URL"
const serverPortEnvKey = "PORT"

func Load() (Config, error) {
	var cfg Config

	cfg.Database.DSN = os.Getenv(databaseDSNEnvKey)
	if cfg.Database.DSN == "" {
		return Config{}, errors.New("DATABASE_URL env variable not set")
	}

	cfg.Server.Port = getEnv(serverPortEnvKey, "8080")

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
