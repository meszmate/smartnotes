package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/invopop/jsonschema"
	"github.com/meszmate/smartnotes/internal/models"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	tiktoken "github.com/pkoukk/tiktoken-go"
)

type Response struct {
	ID            string                `json:"id"`
	Title         string                `json:"title"`
	Summary       string                `json:"summary"`
	Flashcards    []models.Flashcard    `json:"flashcards"`
	QuizQuestions []models.QuizQuestion `json:"quiz_questions"`
}

type AIClient struct {
	client            openai.Client
	rateLimitInterval time.Duration
	tokenLimit        int

	mu         sync.Mutex
	usedTokens int
	lastReset  time.Time
}

var (
	ErrRateLimitReached = errors.New("rate limit reached, please try again later")
)

func GenerateSchema[T any]() *jsonschema.Schema {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	return reflector.Reflect(v)
}

var ResponseSchema = GenerateSchema[models.Response]()

func NewAIClient(apiKey string, rateLimitInterval time.Duration, tokenLimit int) *AIClient {
	return &AIClient{
		client:            openai.NewClient(option.WithAPIKey(apiKey)),
		rateLimitInterval: rateLimitInterval,
		tokenLimit:        tokenLimit,
		lastReset:         time.Now(),
	}
}

func estimateTokens(text string) (int, error) {
	enc, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return 0, err
	}
	tokens := enc.Encode(text, nil, nil)
	return len(tokens), nil
}

func (a *AIClient) Generate(ctx context.Context, text string, includeSummary, includeFlashcards, includeQuiz bool) (*Response, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Reset token usage when interval has passed
	if time.Since(a.lastReset) >= a.rateLimitInterval {
		a.usedTokens = 0
		a.lastReset = time.Now()
	}

	approxTokens, err := estimateTokens(text)
	if err != nil {
		fmt.Printf("[ERROR] Failed to estimate tokens: %v", err)
		return nil, errors.New("failed to estimate tokens")
	}

	// Check token limit within the current interval
	if a.tokenLimit > 0 && a.usedTokens+approxTokens > a.tokenLimit {
		return nil, fmt.Errorf("%w: used %d / %d tokens in current window", ErrRateLimitReached, a.usedTokens, a.tokenLimit)
	}

	a.usedTokens += approxTokens

	taskList := ""
	if includeSummary {
		taskList += "1. Write a short, clear summary of the text.\n"
	}
	if includeFlashcards {
		taskList += "2. Create study flashcards (Q/A pairs).\n"
	}
	if includeQuiz {
		taskList += "3. Generate multiple-choice quiz questions with 4 options each and mark the correct one.\n"
	}
	if taskList == "" {
		return &Response{
			ID: uuid.NewString(),
		}, nil
	}

	instructions := fmt.Sprintf(`Perform the following tasks on this text:
%s
If a section is not requested, leave it empty.

Text:
%s`, taskList, text)

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "smartnotes_output",
		Description: openai.String("Structured AI output containing a summary, study flashcards, and multiple-choice quiz questions for student learning."),
		Schema:      ResponseSchema,
		Strict:      openai.Bool(true),
	}

	req := openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4oMini,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are an educational assistant that outputs structured JSON for study materials."),
			openai.UserMessage(instructions),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{JSONSchema: schemaParam},
		},
	}

	resp, err := a.client.Chat.Completions.New(ctx, req)
	if err != nil {
		fmt.Printf("[ERROR] OpenAI API call failed: %v\n", err)
		return nil, fmt.Errorf("AI service is currently unavailable")
	}

	var parsed models.Response
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &parsed); err != nil {
		fmt.Printf("[ERROR] JSON parse failed: %v\nRaw: %s\n", err, resp.Choices[0].Message.Content)
		return nil, fmt.Errorf("failed to process AI response")
	}

	if !includeSummary {
		parsed.Summary = ""
	}
	if !includeFlashcards {
		parsed.Flashcards = []models.Flashcard{}
	}
	if !includeQuiz {
		parsed.QuizQuestions = []models.QuizQuestion{}
	}

	return &Response{
		ID:            uuid.NewString(),
		Title:         parsed.Title,
		Summary:       parsed.Summary,
		Flashcards:    parsed.Flashcards,
		QuizQuestions: parsed.QuizQuestions,
	}, nil
}
