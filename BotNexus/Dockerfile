# Build Stage
FROM golang:alpine AS builder

WORKDIR /build

# Copy dependency files
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

# Copy source code
COPY main.go ./

# Build the application
RUN go build -o BotNexus main.go

# Runtime Stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/BotNexus ./

# Copy static files (HTML)
COPY index.html ./
# Copy stats.json if it exists (optional, usually volume mounted)
# COPY stats.json ./ 

# Expose ports
EXPOSE 3001 5000

# Set default environment variables
ENV WS_PORT=:3001
ENV WEBUI_PORT=:5000
ENV REDIS_ADDR=192.168.0.126:6379
ENV REDIS_PWD=redis_zsYik8

# Run the application
CMD ["./BotNexus"]
