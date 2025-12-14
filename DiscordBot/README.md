# DiscordBot 🎮

A **Go-based** Discord Robot implementation for [BotMatrix](../README.md).

## ✨ Features

*   **OneBot 11 Compliance**:
    *   **Sending**: `send_group_msg` (to Channels), `send_private_msg` (Direct Message).
    *   **Receiving**: Supports Text messages.
    *   **Meta**: `get_login_info`.
*   **WebSocket Gateway**: Uses official Discord Gateway via `discordgo`.
*   **Channel Mapping**: Maps Discord Channels to OneBot `group_id`.

## 🛠 Configuration

Create a `config.json` file in the root directory:

```json
{
    "bot_token": "YOUR_DISCORD_BOT_TOKEN",
    "nexus_addr": "ws://bot-manager:3005"
}
```

### Configuration Guide

1.  Go to [Discord Developer Portal](https://discord.com/developers/applications).
2.  Create a New Application.
3.  Go to **Bot** tab -> **Add Bot**.
4.  **Important**: Enable **Message Content Intent** under "Privileged Gateway Intents".
5.  Copy **Token** to `config.json`.
6.  Invite Bot to server: `OAuth2` -> `URL Generator` -> `bot` scope -> Copy URL.

## 🚀 Deployment

### Docker (Recommended)

This service is part of the BotMatrix `docker-compose.yml`.

1.  Enable the service in `docker-compose.yml` (uncomment the `discord-bot` section).
2.  Place your `config.json` in `DiscordBot/config.json`.
3.  Run:
    ```bash
    docker-compose up -d --build discord-bot
    ```

### Manual Build

```bash
cd DiscordBot
go mod tidy
go build -o DiscordBot main.go
./DiscordBot
```
