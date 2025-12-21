FROM golang:alpine AS builder

# Use Aliyun mirror for APK
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

WORKDIR /app
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .
RUN go build -o wxbot-go main.go

FROM alpine:latest

# Use Aliyun mirror for APK
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# Install timezone data
RUN apk add --no-cache tzdata

WORKDIR /app
COPY --from=builder /app/wxbot-go .

CMD ["./wxbot-go"]
