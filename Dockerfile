# Stage 1: build Go binary (Debian-based builder for stability)
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
# install into a transient /install directory so we can copy into final
RUN pip install --no-cache-dir --upgrade pip \
 && pip install --no-cache-dir --prefix=/install -r requirements.txt

# Final image: Python runtime + supervisor + our app
FROM python:3.10-slim
ENV PYTHONUNBUFFERED=1
WORKDIR /app

# install supervisor and any system deps your python packages need
RUN apt-get update \
 && apt-get install -y --no-install-recommends supervisor ca-certificates \
 && rm -rf /var/lib/apt/lists/*

# Copy go binary from builder and make it executable
COPY --from=go-builder /app/bot /app/bot
RUN chmod +x /app/bot

# Copy python libraries installed into /install by py-builder into system site-packages
# NOTE: target path may vary by python minor version — this uses the typical path for 3.10-slim
COPY --from=py-builder /install /usr/local

# Copy python server sources and supervisor config
COPY py_server /app/py_server
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Ensure PYTHONPATH contains app python server
ENV PYTHONPATH=/app/py_server

# Expose ports used by python/gRPC etc (if any)
# EXPOSE 8000

# Run supervisord in foreground so container stays alive on Render
CMD ["supervisord", "-n", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
