package app

import (
	httpcontroller "avito-test-task/internal/controller/http"
	"avito-test-task/internal/repository/postgres"
	"avito-test-task/internal/service"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type App struct{}

const migrationsDir = "./migrations"

func (app App) MustRun() {
	dbDSN := os.Getenv("DATABASE_URL")
	if dbDSN == "" {
		dbDSN = "postgres://user:password@localhost:5432/pr_reviewer?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := context.Background()

	// Database

	log.Println("Connecting to database...")
	poolConfig, err := pgxpool.ParseConfig(dbDSN)
	if err != nil {
		log.Fatalf("Unable to parse config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Ping failed: %v", err)
	}

	log.Printf("Running migrations from %s\n", migrationsDir)

	stdDB := stdlib.OpenDB(*poolConfig.ConnConfig)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}
	if err := goose.Up(stdDB, migrationsDir); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	if err := stdDB.Close(); err != nil {
		log.Fatalf("Error closing DB: %v", err)
	}

	// Repositories
	teamRepo := postgres.NewTeamRepo(pool)
	userRepo := postgres.NewUserRepo(pool)
	prRepo := postgres.NewPRRepo(pool)

	// Service & Controller
	svc := service.NewService(teamRepo, userRepo, prRepo)
	ctrl := httpcontroller.NewController(svc)

	// Server
	addr := fmt.Sprintf("0.0.0.0:%s", port)

	server := &http.Server{
		Addr:         addr,
		Handler:      ctrl.Handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Server starting
	go func() {
		log.Printf("Server starting on %s", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Grateful shutdown

	<-quit
	log.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	pool.Close()

	log.Println("Server successfully stopped")
}
