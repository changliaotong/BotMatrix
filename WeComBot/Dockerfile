FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o wecom-bot main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/wecom-bot .
COPY config.sample.json config.json

EXPOSE 8001

CMD ["./wecom-bot"]
