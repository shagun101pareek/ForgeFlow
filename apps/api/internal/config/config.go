package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	Port               string
	DatabaseURL        string
	RedisURL           string
	JWTSecret          string
	OpenAIKey          string
	GenerationProvider string
	NATSURL            string
	S3Bucket           string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	cfg := &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		RedisURL:           getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:          getEnv("JWT_SECRET", "change-me-in-production"),
		OpenAIKey:          os.Getenv("OPENAI_API_KEY"),
		GenerationProvider: strings.ToLower(getEnv("GENERATION_PROVIDER", "mock")),
		NATSURL:            getEnv("NATS_URL", "nats://localhost:4222"),
		S3Bucket:           getEnv("S3_BUCKET", "forgeflow-assets"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if _, err := strconv.Atoi(cfg.Port); err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	switch cfg.GenerationProvider {
	case "mock", "openai":
	default:
		return nil, fmt.Errorf("invalid GENERATION_PROVIDER %q", cfg.GenerationProvider)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
