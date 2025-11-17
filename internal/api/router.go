package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/meszmate/smartnotes/internal/pkg/ai"
	"github.com/meszmate/smartnotes/internal/pkg/captcha"
)

type Handler struct {
	Captcha *captcha.Turnstile
	AI      *ai.AIClient
}

func Start(h *Handler, open bool, port string) {
	r := gin.New()
	r.Use(cors.Default())

	r.POST("/generate", h.GenerateResponse)

	var p string
	if open {
		p = "0.0.0.0"
	}
	r.Run(p + ":" + port)
}
