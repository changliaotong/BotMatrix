# DiscordBot 🎮

A **Go-based** Discord Robot implementation for [BotMatrix](../README.md).

## ✨ Features

*   **OneBot 11 Compliance**:
    *   **Sending**: `send_group_msg` (to Channels), `send_private_msg` (Direct Message).
    *   **Receiving**: Supports Text messages.
    *   **Meta**: `get_login_info`.
    *   **Recall**: `delete_msg` support.
*   **WebSocket Gateway**: Uses official Discord Gateway via `discordgo`.
*   **Channel Mapping**: Maps Discord Channels to OneBot `group_id`.
*   **Burn After Reading**: **New!** Supports message recall.

### 🔥 Burn After Reading (Message Recall)

DiscordBot supports message recall via the `delete_msg` action.

*   **ID Format**: `ChannelID:MessageID` (e.g., `123456789:987654321`).
*   **Mechanism**: Uses the standard `ChannelMessageDelete` API.
*   **Permissions**: The bot needs the `MANAGE_MESSAGES` permission to delete messages from others, but can always delete its own.

## 🛠 Configuration

DiscordBot supports two ways to configure:

1.  **Web UI (Recommended)**:
    *   Start the bot.
    *   Access `http://localhost:8084/config-ui` (default port is 8084).
    *   Fill in the fields and click "Save & Restart".

2.  **Manual JSON**:
    Create a `config.json` file in the root directory:

```json
{
    "bot_token": "YOUR_DISCORD_BOT_TOKEN",
    "nexus_addr": "ws://bot-nexus:3005",
    "log_port": 8084
}
```

### Configuration Guide

1.  Go to [Discord Developer Portal](https://discord.com/developers/applications).
2.  Create a New Application.
3.  Go to **Bot** tab -> **Add Bot**.
4.  **Important**: Enable **Message Content Intent** under "Privileged Gateway Intents".
5.  Copy **Token** to `config.json`.
6.  Invite Bot to server: `OAuth2` -> `URL Generator` -> `bot` scope -> Copy URL.

| Field | Description |
| :--- | :--- |
| `bot_token` | Your Discord Bot Token. |
| `nexus_addr` | Address of the BotNexus WebSocket server. |
| `log_port` | Port for the Web UI and Log viewer. |

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
