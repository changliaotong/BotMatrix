# TencentBot 🐧

A **Go-based** Official QQ Guild/Group Robot implementation for [BotMatrix](../README.md), using the official `botgo` SDK.

## ✨ Features

*   **Official API**: Compliant with Tencent's requirements for QQ Guilds and Groups.
*   **OneBot 11 Compliance**: Bridges official events to OneBot standard.
*   **Comprehensive Event Handling**: Supports 8 types of events including messages, guild events, member events, and more.
*   **Message Management**:
    *   Send messages to QQ Guilds, Groups, and Private chats
    *   Support for various media types: text, images, videos, audio, and files
    *   **Burn After Reading**: Supports message recall using `RetractMessage` API
*   **Channel Management**:
    *   Create, update, and delete channels
    *   Get channel lists and channel details
*   **Member Management**:
    *   Get guild member lists and profiles
    *   Manage member roles
*   **WebSocket Communication**: Real-time bidirectional communication with BotNexus.
*   **Keep-Alive Mechanism**: Maintains persistent connection with Tencent's API.
*   **Docker Support**: Easy deployment using Docker containers.

### 🔥 Burn After Reading (Message Recall)

*   **Mechanism**: Uses `RetractMessage` API.
*   **Requirement**:
    *   For Group messages, requires the `group_id` context (handled internally via ID mappings or `channel_id` if available).
    *   Returns valid `message_id` for recall operations.

### 📋 Supported OneBot API

#### Message Operations
- `send_msg`: Send messages to guilds, groups, or private chats
- `send_group_msg`: Send group messages
- `send_private_msg`: Send private messages
- `send_guild_channel_msg`: Send guild channel messages
- `delete_msg`: Recall messages
- `get_message`: Get message details

#### Channel Operations
- `create_guild_channel`: Create a new channel
- `update_guild_channel`: Update channel information
- `delete_guild_channel`: Delete a channel
- `get_guild_channel_list`: Get channel list

#### Guild Operations
- `get_guild_list`: Get guild list
- `get_guild_meta`: Get guild details

#### Member Operations
- `get_guild_member_list`: Get guild member list
- `get_guild_member_profile`: Get guild member profile
- `get_guild_member_info`: Get guild member information

#### Role Operations
- `get_guild_roles`: Get guild roles
- `create_guild_role`: Create a new role
- `update_guild_role`: Update role information
- `delete_guild_role`: Delete a role

#### System Operations
- `get_login_info`: Get bot login information
- `get_group_list`: Get group list (mapped from guilds)
- `get_version_info`: Get version information
- `get_logs`: Get bot logs

### 🎯 Event Types Supported

- Channel @ messages
- Direct messages
- Guild events
- Guild member events
- Channel events
- Message reaction events
- QQ Group @ messages
- QQ C2C (private) messages

## 🛠 Configuration

TencentBot supports two ways to configure:

1.  **Web UI (Recommended)**:
    *   Start the bot.
    *   Access `http://localhost:3133/config-ui` (default port is 3133).
    *   Fill in the fields and click "Save & Restart".

2.  **Manual JSON**:
    Create a `config.json` file in the root directory:

```json
{
    "app_id": 123456789,
    "token": "YOUR_BOT_TOKEN",
    "secret": "YOUR_APP_SECRET",
    "sandbox": true,
    "log_port": 3133,
    "nexus_addr": "ws://bot-nexus:3005"
}
```

### Configuration Options

| Field            | Type      | Description                                                                 |
|------------------|-----------|-----------------------------------------------------------------------------|
| `app_id`         | uint64    | Your Tencent bot's AppID                                                   |
| `token`          | string    | Your Tencent bot's Token                                                   |
| `secret`         | string    | Your Tencent bot's Secret                                                   |
| `sandbox`        | bool      | Whether to use sandbox environment (for testing)                            |
| `log_port`       | int       | Port for the Web UI and Log viewer                                          |
| `nexus_addr`     | string    | Address of the BotNexus WebSocket server                                    |

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
