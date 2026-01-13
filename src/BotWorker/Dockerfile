# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go.mod and go.sum and download dependencies
COPY src/BotWorker/go.mod src/BotWorker/go.sum ./src/BotWorker/
COPY src/Common/go.mod src/Common/go.sum ./src/Common/
RUN cd src/BotWorker && go mod download

# Copy the source code
COPY src/BotWorker ./src/BotWorker
COPY src/Common ./src/Common

# Build the binary
RUN cd src/BotWorker && \
    CGO_ENABLED=1 GOOS=linux go build -o bot-worker main.go

# Stage 2: Final runtime image
FROM alpine:latest

# Use Aliyun mirror for faster package installation (optional)
# RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# Install runtime dependencies (including Python for Python plugins)
RUN apk add --no-cache \
    tzdata \
    ca-certificates \
    python3 \
    py3-pip \
    bash \
    curl

WORKDIR /app

# Copy the pre-built binary
COPY --from=builder /build/src/BotWorker/bot-worker ./

# Copy default configs and locales
COPY src/BotWorker/configs/config.yaml ./configs/
COPY src/BotWorker/locales ./locales/

# Create a directory for plugins and set as a volume
RUN mkdir -p /app/plugins
VOLUME /app/plugins

# Environment variables with defaults
ENV BM_WORKER_PORT=8080
ENV BM_REDIS_ADDR=redis:6379
ENV BM_LOG_LEVEL=info

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${BM_WORKER_PORT}/health || exit 1

EXPOSE ${BM_WORKER_PORT}

# Run the worker
CMD ["./bot-worker"]
