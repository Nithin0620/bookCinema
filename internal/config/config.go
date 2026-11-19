package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr         string
	DatabaseURL  string
	RedisAddr    string
	HoldTTL      time.Duration
	SessionTTL   time.Duration
	CookieSecure bool
	StaticDir    string
	AdminEmails  []string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:         env("APP_ADDR", ":8080"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		RedisAddr:    env("REDIS_ADDR", "localhost:6380"),
		HoldTTL:      2 * time.Minute,
		SessionTTL:   7 * 24 * time.Hour,
		CookieSecure: envBool("COOKIE_SECURE", false),
		StaticDir:    env("STATIC_DIR", "web/dist"),
		AdminEmails:  csvEnv("ADMIN_EMAILS"),
	}

	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("DATABASE_URL is required")
	}

	var err error
	if cfg.HoldTTL, err = envDuration("HOLD_TTL", cfg.HoldTTL); err != nil {
		return cfg, fmt.Errorf("HOLD_TTL: %w", err)
	}
	if cfg.SessionTTL, err = envDuration("SESSION_TTL", cfg.SessionTTL); err != nil {
		return cfg, fmt.Errorf("SESSION_TTL: %w", err)
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func csvEnv(key string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	return time.ParseDuration(value)
}
