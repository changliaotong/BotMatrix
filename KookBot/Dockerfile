FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o kook-bot main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/kook-bot .
COPY config.sample.json config.json

CMD ["./kook-bot"]
