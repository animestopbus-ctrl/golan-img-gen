# Stage 1: build Go binary
FROM golang:1.22 as go-builder
WORKDIR /src
# cache modules first
COPY go.mod go.sum ./
RUN go env -w GO111MODULE=on \
 && go mod download
# copy rest and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bot main.go

# Stage 2: build Python packages into /install
FROM python:3.10-slim AS py-builder
WORKDIR /py
COPY py_server/requirements.txt .

# Install from your updated requirements.txt (assumes it has compatible pins)
# Pre-install torch CPU wheels if not in requirements.txt (to ensure CPU versions)
RUN pip install --no-cache-dir --upgrade pip \
 && pip install --no-cache-dir --prefix=/install --index-url https://download.pytorch.org/whl/cpu \
    torch torchvision torchaudio \
 && pip install --no-cache-dir --prefix=/install -r requirements.txt

# Final image
FROM python:3.10-slim
ENV PYTHONUNBUFFERED=1
WORKDIR /app

RUN apt-get update \
 && apt-get install -y --no-install-recommends supervisor ca-certificates \
 && rm -rf /var/lib/apt/lists/*

COPY --from=go-builder /app/bot /app/bot
RUN chmod +x /app/bot

COPY --from=py-builder /install /usr/local

COPY py_server /app/py_server
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

RUN touch /app/py_server/__init__.py

EXPOSE 1000

CMD ["supervisord", "-n", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
