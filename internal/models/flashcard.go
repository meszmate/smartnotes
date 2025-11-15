package models

type Flashcard struct {
	Question string `json:"question" json_schema:"A concise question that tests a key concept from the text."`
	Answer   string `json:"answer" json_schema:"A clear, correct answer to the flashcard question."`
}
