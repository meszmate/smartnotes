package models

type Response struct {
	Title         string         `json:"title" json_schema:"2-20 character length title"`
	Summary       string         `json:"summary" json_schema:"A short, clear summary of the text in plain language that helps the student understand the main ideas."`
	Flashcards    []Flashcard    `json:"flashcards" json_schema:"A list of study flashcards, each with a question and answer to help review key points."`
	QuizQuestions []QuizQuestion `json:"quiz_questions" json_schema:"A list of multiple-choice quiz questions with options and the correct answer to test comprehension."`
}
