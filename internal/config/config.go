package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
	AppEnv  string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBMaxConns int32
	DBMinConns int32

	RedisURL string

	JWTSecret        string
	JWTRefreshSecret string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration

	SMTPHost    string
	SMTPPort    string
	SMTPUser    string
	SMTPPass    string
	FromEmail   string
	FrontendURL string

	MinIOBucket    string
	MinIORegion    string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOEndpoint  string

	RateLimitRequests int
	RateLimitWindow   time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}

	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}

	rateLimitWindow, err := time.ParseDuration(getEnv("RATE_LIMIT_WINDOW", "1m"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_WINDOW: %w", err)
	}

	cfg := &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "bdgovtjobs"),
		DBUser:     getEnv("DB_USER", "bduser"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBMaxConns: int32(getEnvAsInt("DB_MAX_CONNS", 25)),
		DBMinConns: int32(getEnvAsInt("DB_MIN_CONNS", 5)),

		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379/0"),

		JWTSecret:        getEnv("JWT_SECRET", ""),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
		JWTAccessTTL:     accessTTL,
		JWTRefreshTTL:    refreshTTL,

		SMTPHost:    getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:    getEnv("SMTP_PORT", "587"),
		SMTPUser:    getEnv("SMTP_USER", ""),
		SMTPPass:    getEnv("SMTP_PASS", ""),
		FromEmail:   getEnv("FROM_EMAIL", "noreply@yourdomain.com"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),

		MinIOBucket:    getEnv("MINIO_BUCKET", "bd-govt-jobs-assets"),
		MinIORegion:    getEnv("MINIO_REGION", "us-east-1"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "http://localhost:9000"),

		RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   rateLimitWindow,
	}

	return cfg, nil
}

func AppEnv() string {
	return getEnv("APP_ENV", "development")
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}
