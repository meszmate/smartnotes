package models

type QuizQuestion struct {
	Question string   `json:"question" json_schema:"The text of the quiz question."`
	Options  []string `json:"options" json_schema:"A list of 3–5 possible answers for the question."`
	Correct  string   `json:"correct" json_schema:"The text of the correct answer, exactly matching one of the options."`
}
