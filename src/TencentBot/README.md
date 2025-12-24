# TencentBot 🐧

A **Go-based** Official QQ Guild/Group Robot implementation for [BotMatrix](../README.md), using the official `botgo` SDK.

## ✨ Features

*   **Official API**: Compliant with Tencent's requirements for QQ Guilds and Groups.
*   **OneBot 11 Compliance**: Bridges official events to OneBot standard.
*   **Burn After Reading**: **New!** Supports message recall.
*   **WebSocket Communication**: Real-time bidirectional communication with BotNexus.
*   **Keep-Alive Mechanism**: Maintains persistent connection with Tencent's API.
*   **Docker Support**: Easy deployment using Docker containers.

### 🔥 Burn After Reading (Message Recall)

*   **Mechanism**: Uses `RetractMessage` API.
*   **Requirement**:
    *   For Group messages, requires the `group_id` context (handled internally via ID mappings or `channel_id` if available).
    *   Returns valid `message_id` for recall operations.

## 🛠 Configuration

Create `config.json` from the sample:

```json
{
    "app_id": 123456789,
    "token": "YOUR_BOT_TOKEN",
    "secret": "YOUR_APP_SECRET",
    "sandbox": true,
    "self_id": "",
    "log_port": 3133,
    "file_host": "https://www.sz84.com",
    "media_route": "/bot-files/",
    "websocket_addrs": [
        {
            "name": "Default",
            "url": "ws://bot-nexus:3001"
        }
    ]
}
```

### Configuration Options

| Field            | Type      | Description                                                                 |
|------------------|-----------|-----------------------------------------------------------------------------|
| `app_id`         | uint64    | Your Tencent bot's AppID                                                   |
| `token`          | string    | Your Tencent bot's Token                                                   |
| `secret`         | string    | Your Tencent bot's Secret                                                   |
| `sandbox`        | bool      | Whether to use sandbox environment (for testing)                            |
| `self_id`        | string    | Optional: Manually set self ID                                              |
| `log_port`       | int       | Port for HTTP Log Viewer                                                    |
| `file_host`      | string    | Public base URL for serving files                                           |
| `media_route`    | string    | Internal route path for media files                                         |
| `websocket_addrs`| array     | List of WebSocket addresses to connect to                                   |
| `nexus_addr`     | string    | BotNexus address for backward compatibility                                 |

## 🚀 Deployment

### Option 1: Local Development

```bash
cd src/TencentBot
go build -o TencentBot.exe main.go
./TencentBot.exe
```

### Option 2: Docker Deployment

#### Build Docker Image
```bash
docker build -t tencent-bot .
```

#### Run Docker Container
```bash
docker run -d --name tencent-bot \
  -v $(pwd)/config.json:/app/config.json \
  --network botmatrix \
  tencent-bot
```

### Option 3: Automated Deployment

Use the deploy.py script for automated deployment to Ubuntu server:

```bash
python deploy.py --target tencent-bot --ip 192.168.0.167 --user derlin
```

## 📦 Dockerfile

The project includes a Dockerfile for containerized deployment:

```dockerfile
FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o TencentBot main.go

CMD ["./TencentBot"]
```

## 📝 Logging

The bot provides detailed logging:
- Connection status to BotNexus
- Message processing events
- API request/response details
- Error handling and recovery

## 🔧 Troubleshooting

### Common Issues

1. **Compilation Errors**
   - Ensure Go 1.21+ is installed
   - Run `go mod tidy` to resolve dependencies

2. **Connection Failed**
   - Check if BotNexus is running
   - Verify `nexus_addr` in config.json
   - Ensure firewall allows WebSocket connections

3. **API Errors**
   - Verify `app_id`, `token`, and `secret` are correct
   - Check if the bot has been approved by Tencent

## 📄 License

MIT License

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.
