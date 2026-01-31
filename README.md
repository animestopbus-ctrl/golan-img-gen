# AI Image Generator Telegram Bot

## Setup
1. Copy `.env.example` to `.env` and fill in BOT_TOKEN and MONGO_URI.
2. Build Docker image: `docker build -t image-bot .`
3. Run: `docker run -d -p 8000:8000 --env-file .env image-bot`
4. For local dev without Docker:
   - Go: `go run main.go`
   - Python: `cd py_server && uvicorn main:app --host 0.0.0.0 --port 8000`

## Features
- /start: Welcome message.
- /generate [prompt]: AI image if prompt, else random from Picsum.
- /history: Last 5-10 prompts.
- Rate limit: 5/hour/user.
- Local AI generation on CPU.

## Deployment
Deploy on Render/Heroku with Docker support. Set env vars.