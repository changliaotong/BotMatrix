# Build Stage
FROM golang:alpine AS builder

# Use Aliyun mirror for APK
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o FeishuBot .

# Run Stage
FROM alpine:latest

# Use Aliyun mirror for APK
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

WORKDIR /root/

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/FeishuBot .
COPY --from=builder /app/config.sample.json ./config.json

# Expose port (if needed, though it's mainly websocket client)
# EXPOSE 8080

# Command to run
CMD ["./FeishuBot"]
