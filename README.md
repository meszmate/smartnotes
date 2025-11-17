# Smart Notes
AI-powered study companion that turns your notes into summaries, flashcards, and quizzes
24 hours programming challenge

## 🎬 Demo Video
https://youtu.be/Q973GvgQz4s

## Features
- Upload notes (PDF, TXT, or copy-paste)
- AI-generated summaries
- AI-generated flashcards
- Mini quizzes for self-testing
- Clean, intuitive UI

## Tech Stack
- Frontend: React + Tailwind CSS + shadcn
- Backend: Golang + Gin
- AI: OpenAI GPT API

## Getting Started
1. Clone the repo
2. `go run cmd/main.go`
3. `cd frontend` → `npm install && npm start`
4. Setup env, upload notes and start learning!

## Frontend Environment Variables
`VITE_API_URL`: e.g. http://localhost:8080<br/>
`VITE_TURNSTILE_KEY`: e.g. 1x00000000000000000000AA<br/>

## Backend Environment Variables
`API_KEY` (OpenAI Api Key): e.g. `sk-proj-...`<br/>
`TURNSTILE_SECRET`: e.g. `1x0000000000000000000000000000000AA`<br/>
`OPEN`: `true` or `false` (listen on 0.0.0.0)<br/>
`PORT`: e.g. `8080`<br/>
`RATELIMIT_INTERVAL` (seconds): e.g. `3600` for 1 hour<br/>
`TOKEN_LIMIT`: e.g. `10000` tokens within each rate limit interval<br/>

## 📝 License
MIT © 2025 Matthew Meszaros
