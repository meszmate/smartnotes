package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/meszmate/smartnotes/internal/models"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Response struct {
	Summary       string                `json:"summary"`
	Flashcards    []models.Flashcard    `json:"flashcards"`
	QuizQuestions []models.QuizQuestion `json:"quiz_questions"`
}

type AIClient struct {
	client openai.Client
}

// Source: https://github.com/openai/openai-go/blob/main/examples/structured-outputs/main.go

func GenerateSchema[T any]() *jsonschema.Schema {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

var ResponseSchema = GenerateSchema[models.Response]()

func NewAIClient(apiKey string) *AIClient {
	return &AIClient{
		client: openai.NewClient(option.WithAPIKey(apiKey)),
	}
}

func (a *AIClient) Generate(ctx context.Context, text string, includeSummary, includeFlashcards, includeQuiz bool) (*models.Response, error) {
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
		return &models.Response{}, nil
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

	return &parsed, nil
}
