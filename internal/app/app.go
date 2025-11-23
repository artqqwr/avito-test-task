package app

import (
	"avito-test-task/internal/config"
	httpcontroller "avito-test-task/internal/controller/http"
	"avito-test-task/internal/database"
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
)

type App struct{}

func (app App) MustRun() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Database
	pool, err := database.NewPool(ctx, cfg.Database)

	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Repositories
	teamRepo := postgres.NewTeamRepo(pool)
	userRepo := postgres.NewUserRepo(pool)
	prRepo := postgres.NewPRRepo(pool)

	// Service & Controller
	svc := service.NewService(teamRepo, userRepo, prRepo)
	ctrl := httpcontroller.NewController(svc)

	// Server
	addr := fmt.Sprintf("0.0.0.0:%s", cfg.Server.Port)
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
