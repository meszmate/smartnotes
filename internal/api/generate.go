package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meszmate/smartnotes/internal/errx"
)

func (h *Handler) GenerateResponse(c *gin.Context) {
	var data struct {
		Prompt        string `json:"prompt"`
		Summary       bool   `json:"summary"`
		FlashCards    bool   `json:"flashcards"`
		QuizQuestions bool   `json:"quiz"`
		Turnstile     string `json:"turnstile"`
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		errx.Handle(c, errx.ErrInvalid)
		return
	}

	validCaptcha, err := h.Captcha.Verify(c.Request.Context(), data.Turnstile, c.RemoteIP())
	if err != nil {
		fmt.Printf("[ERROR] Captcha verify: %s\n", err.Error())
		errx.Handle(c, errx.ErrCaptcha)
	}
	if !validCaptcha {
		errx.Handle(c, errx.ErrCaptcha)
		return
	}

	if !data.Summary && !data.FlashCards && !data.QuizQuestions {
		errx.Handle(c, errx.ErrOptions)
		return
	}

	resp, err := h.AI.Generate(c.Request.Context(), data.Prompt, data.Summary, data.FlashCards, data.QuizQuestions)
	if err != nil {
		errx.Handle(c, errx.ErrGeneration(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}
