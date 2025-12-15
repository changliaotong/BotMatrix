# Build stage
FROM golang:alpine AS builder

# Set Go Proxy
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache gcc musl-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

# Copy source code
COPY main.go ./

# Build
RUN go build -o DingTalkBot main.go

# Runtime stage
FROM alpine:latest

# Set Mirror
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache tzdata ca-certificates

WORKDIR /app

COPY --from=builder /build/DingTalkBot ./
COPY config.sample.json ./config.sample.json

# Use Shanghai Timezone
ENV TZ=Asia/Shanghai

CMD ["./DingTalkBot"]
