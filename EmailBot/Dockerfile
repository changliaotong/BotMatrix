FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o email-bot main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/email-bot .
COPY config.sample.json config.json

CMD ["./email-bot"]
