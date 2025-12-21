# ======================
# Build Stage
# ======================
FROM golang:alpine AS builder

# Use Aliyun mirror for APK
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# Install build dependencies
RUN apk add --no-cache gcc musl-dev git ca-certificates

WORKDIR /build

# Copy dependency files
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on
ENV GOSUMDB=off
RUN go mod download

# Copy all source code
COPY . .

# Build the binary
RUN go build -o BotNexus

# ======================
# Runtime Stage
# ======================
FROM alpine:latest

# Use Aliyun mirror for APK
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# Install tzdata for timezone support
RUN apk add --no-cache tzdata ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/BotNexus ./
COPY web ./web
COPY overmind ./overmind

# Expose ports
EXPOSE 3001 5000

# Environment variables
ENV WS_PORT=:3001
ENV WEBUI_PORT=:5000
ENV REDIS_ADDR=192.168.0.126:6379
ENV REDIS_PWD=redis_zsYik8

# Optional: run as non-root (but default to root for Docker socket access)
RUN addgroup -S appgroup && adduser -S -D -H -G appgroup appuser

# Create data directory and ensure appuser owns the app directory
RUN mkdir -p /app/data && chown -R appuser:appgroup /app

# USER appuser # Commented out to allow Docker socket access by default
# If you want to run as non-root, ensure /var/run/docker.sock has correct permissions

# Run the application
CMD ["./BotNexus"]
