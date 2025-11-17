package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	TurnstileSecret   string
	ApiKey            string
	Open              bool
	Port              string
	RateLimitInterval time.Duration
	TokenLimit        int
}

func New() *Config {
	ratelimitInterval, _ := strconv.Atoi(os.Getenv("RATELIMIT_INTERVAL"))
	tokenLimit, _ := strconv.Atoi(os.Getenv("TOKEN_LIMIT"))

	return &Config{
		TurnstileSecret:   os.Getenv("TURNSTILE_SECRET"),
		ApiKey:            os.Getenv("API_KEY"),
		Open:              envBool("OPEN"),
		Port:              os.Getenv("PORT"),
		RateLimitInterval: time.Duration(ratelimitInterval) * time.Second,
		TokenLimit:        tokenLimit,
	}
}

func envBool(key string) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}

	return false
}
