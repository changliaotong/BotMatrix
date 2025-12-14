FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o slack-bot main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/slack-bot .
COPY config.sample.json config.json

CMD ["./slack-bot"]