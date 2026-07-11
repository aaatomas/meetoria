// Package main Meetoria API
//
// @title Meetoria API
// @version 1.0
// @description Multi-tenant appointment scheduling SaaS platform
// @host localhost:8081
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/meetoria/meetoria/backend/internal/common/config"
	"github.com/meetoria/meetoria/backend/internal/common/database"
	"github.com/meetoria/meetoria/backend/internal/common/logger"
	"github.com/meetoria/meetoria/backend/internal/common/rabbitmq"
	redisclient "github.com/meetoria/meetoria/backend/internal/common/redis"
	"github.com/meetoria/meetoria/backend/internal/common/router"
)

func main() {
	cfg := config.Load()
	logger.Init("meetoria-api", cfg.Env)

	db, err := database.New(cfg.DatabaseURL, cfg.IsDevelopment())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	redis, err := redisclient.New(cfg.RedisURL)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer redis.Close()

	publisher, err := rabbitmq.NewPublisher(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer publisher.Close()

	r := router.Setup(router.Dependencies{
		Config:    cfg,
		DB:        db,
		Redis:     redis,
		Publisher: publisher,
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	go func() {
		logger.Default().Info("starting server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Default().Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout())
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
}
