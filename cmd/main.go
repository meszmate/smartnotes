package main

import (
	"github.com/joho/godotenv"
	"github.com/meszmate/smartnotes/internal/api"
	"github.com/meszmate/smartnotes/internal/config"
	"github.com/meszmate/smartnotes/internal/pkg/ai"
	"github.com/meszmate/smartnotes/internal/pkg/captcha"
)

func main() {
	_ = godotenv.Load("cmd/.env", ".env")

	cfg := config.New()

	h := &api.Handler{
		Captcha: captcha.NewTurnstile(cfg.TurnstileSecret),
		AI:      ai.NewAIClient(cfg.ApiKey, cfg.RateLimitInterval, cfg.TokenLimit),
	}

	api.Start(h, cfg.Open, cfg.Port)
}
