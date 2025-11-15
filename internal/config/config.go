package config

import (
	"os"
	"strconv"
)

type Config struct {
	TurnstileSecret string
	ApiKey          string
	Open            bool
	Port            string
}

func New() *Config {
	return &Config{
		TurnstileSecret: os.Getenv("TURNSTILE_SECRET"),
		ApiKey:          os.Getenv("API_KEY"),
		Open:            envBool("OPEN"),
		Port:            os.Getenv("PORT"),
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
