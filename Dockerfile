# Stage 1: Build Go binary
FROM golang:1.22-alpine AS go-builder
WORKDIR /app
COPY go.mod ./
RUN go mod tidy && go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bot main.go

# Stage 2: Python base with dependencies
FROM python:3.10-slim AS py-builder
WORKDIR /app/py_server
COPY py_server/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Final stage: Combine Go and Python, use supervisord
FROM python:3.10-slim
RUN apt-get update && apt-get install -y supervisor && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=go-builder /app/bot /app/bot
COPY --from=py-builder /usr/local/lib/python3.10/site-packages /usr/local/lib/python3.10/site-packages
COPY py_server /app/py_server
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf
COPY . /app  # Copy other files if needed, but main logic is in bot and py_server
ENV PYTHONPATH=/app/py_server
CMD ["supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
