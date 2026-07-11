package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/meetoria/meetoria/workers/email-worker/internal/consumer"
	"github.com/meetoria/meetoria/workers/email-worker/internal/provider"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).
		With("service", "meetoria-email-worker"))

	rabbitURL := getEnv("RABBITMQ_URL", "amqp://meetoria:meetoria@localhost:5672/")
	dbURL := getEnv("DATABASE_URL", "postgres://meetoria:meetoria@localhost:5432/meetoria_email?sslmode=disable")
	providerType := getEnv("EMAIL_PROVIDER", "mock")

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	emailProvider := provider.NewProvider(providerType)
	c, err := consumer.NewConsumer(rabbitURL, db, emailProvider)
	if err != nil {
		log.Fatalf("consumer init failed: %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := c.Start(ctx); err != nil && ctx.Err() == nil {
			log.Fatalf("consumer error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down email worker")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
