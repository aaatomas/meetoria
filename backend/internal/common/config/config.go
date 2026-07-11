package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env         string
	Port        string
	DatabaseURL string
	RedisURL    string
	RabbitMQURL string
	UploadDir   string
	Keycloak    KeycloakConfig
	RateLimit   RateLimitConfig
}

type KeycloakConfig struct {
	URL          string
	Realm        string
	ClientID     string
	ClientSecret string
	JWTIssuer    string
}

type RateLimitConfig struct {
	RequestsPerMinute int
}

func Load() *Config {
	return &Config{
		Env:         getEnv("APP_ENV", "development"),
		Port:        getEnv("APP_PORT", "8081"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://meetoria:meetoria@localhost:5432/meetoria?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://meetoria:meetoria@localhost:5672/"),
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
		Keycloak: KeycloakConfig{
			URL:          getEnv("KEYCLOAK_URL", "http://localhost:8080"),
			Realm:        getEnv("KEYCLOAK_REALM", "meetoria"),
			ClientID:     getEnv("KEYCLOAK_CLIENT_ID", "meetoria-api"),
			ClientSecret: getEnv("KEYCLOAK_CLIENT_SECRET", "meetoria-api-secret"),
			JWTIssuer:    getEnv("JWT_ISSUER", "http://localhost:8080/realms/meetoria"),
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: getEnvInt("RATE_LIMIT_RPM", 120),
		},
	}
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func (c *Config) ShutdownTimeout() time.Duration {
	return 10 * time.Second
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
