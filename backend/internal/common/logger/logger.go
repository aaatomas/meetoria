package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDKey     contextKey = "request_id"
	CorrelationIDKey contextKey = "correlation_id"
	OrganizationIDKey contextKey = "organization_id"
	UserIDKey        contextKey = "user_id"
)

var defaultLogger *slog.Logger

func Init(service, environment string) {
	defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})).With(
		slog.String("service", service),
		slog.String("environment", environment),
	)
}

func Default() *slog.Logger {
	if defaultLogger == nil {
		Init("meetoria-api", "development")
	}
	return defaultLogger
}

func FromContext(ctx context.Context) *slog.Logger {
	l := Default()
	if v, ok := ctx.Value(RequestIDKey).(string); ok && v != "" {
		l = l.With(slog.String("request_id", v))
	}
	if v, ok := ctx.Value(CorrelationIDKey).(string); ok && v != "" {
		l = l.With(slog.String("correlation_id", v))
	}
	if v, ok := ctx.Value(OrganizationIDKey).(string); ok && v != "" {
		l = l.With(slog.String("organization_id", v))
	}
	if v, ok := ctx.Value(UserIDKey).(string); ok && v != "" {
		l = l.With(slog.String("user_id", v))
	}
	return l
}

func WithContext(ctx context.Context, keys ...contextKey) context.Context {
	for _, key := range keys {
		if ctx.Value(key) == nil {
			ctx = context.WithValue(ctx, key, uuidString())
		}
	}
	return ctx
}

func uuidString() string {
	return uuid.New().String()
}
