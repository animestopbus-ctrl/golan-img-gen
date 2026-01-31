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

# Install compatible pinned versions to avoid breaking changes
# - torch CPU wheels
# - diffusers/huggingface_hub versions before cached_download removal
# - NumPy 1.x to avoid torch NumPy 2 incompatibility
RUN pip install --no-cache-dir --upgrade pip \
 && pip install --no-cache-dir --prefix=/install \
    "torch==2.1.2" "torchvision==0.16.2" "torchaudio==2.1.2" --index-url https://download.pytorch.org/whl/cpu \
    diffusers==0.24.0 \
    transformers==4.36.2 \
    accelerate==0.25.0 \
    huggingface_hub==0.16.4 \
    numpy==1.26.4 \
 && pip install --no-cache-dir --prefix=/install -r requirements.txt

# Final image: slim Python runtime + supervisor + app
FROM python:3.10-slim
ENV PYTHONUNBUFFERED=1
WORKDIR /app

# Install supervisor and ca-certificates (needed for HTTPS/HF hub)
RUN apt-get update \
 && apt-get install -y --no-install-recommends supervisor ca-certificates \
 && rm -rf /var/lib/apt/lists/*

# Copy Go binary
COPY --from=go-builder /app/bot /app/bot
RUN chmod +x /app/bot

# Copy installed Python packages
COPY --from=py-builder /install /usr/local

# Copy source and supervisor config
COPY py_server /app/py_server
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Make py_server a proper package
RUN touch /app/py_server/__init__.py

# Expose the port your API listens on (change if needed)
EXPOSE 1000

# Run supervisord in foreground
CMD ["supervisord", "-n", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
