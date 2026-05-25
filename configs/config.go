package configs

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv               string
	AppPort              string
	DatabaseURL          string
	RedisAddr            string
	RedisPassword        string
	RedisDB              int
	JWTSecret            string
	JWTTTLHours          int
	WebhookURL           string
	WebhookSecret        string
	PaymentExpiryMinutes int
}

func Load() Config {
	return Config{
		AppEnv:               getEnv("APP_ENV", "development"),
		AppPort:              getEnv("APP_PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://flipay:flipay@localhost:5432/flipay?sslmode=disable"),
		RedisAddr:            getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:        getEnv("REDIS_PASSWORD", ""),
		RedisDB:              getEnvInt("REDIS_DB", 0),
		JWTSecret:            getEnv("JWT_SECRET", "change-me"),
		JWTTTLHours:          getEnvInt("JWT_TTL_HOURS", 24),
		WebhookURL:           getEnv("WEBHOOK_URL", "http://localhost:8080/api/v1/callbacks/payment"),
		WebhookSecret:        getEnv("WEBHOOK_SECRET", "webhook-secret"),
		PaymentExpiryMinutes: getEnvInt("PAYMENT_EXPIRY_MINUTES", 30),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value, err := strconv.Atoi(getEnv(key, ""))
	if err != nil {
		return fallback
	}
	return value
}
