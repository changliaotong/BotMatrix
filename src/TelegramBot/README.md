# TelegramBot ✈️

A **Go-based** Telegram Robot implementation for [BotMatrix](../README.md).

## ✨ Features

*   **OneBot 11 Compliance**:
    *   **Sending**: `send_group_msg` (to Groups), `send_private_msg` (to Users).
    *   **Receiving**: Supports Text messages.
    *   **Meta**: `get_login_info`.
    *   **Recall**: `delete_msg` support.
*   **Long Polling**: Uses standard Telegram Long Polling for event updates.
*   **Zero-Config Networking**: Connects outbound to BotNexus; no public IP required for the bot itself.
*   **Burn After Reading**: **New!** Supports message recall.

### 🔥 Burn After Reading (Message Recall)

TelegramBot supports message recall via the `delete_msg` action.

*   **ID Format**: `ChatID:MessageID` (e.g., `-100123456789:55`).
*   **Mechanism**: Uses the standard `deleteMessage` API.
*   **Permissions**: The bot must be an Admin in groups to delete messages from others (though for its own messages, standard rights usually suffice).

## 🛠 Configuration

TelegramBot supports two ways to configure:

1.  **Web UI (Recommended)**:
    *   Start the bot.
    *   Access `http://localhost:8085/config-ui` (default port is 8085).
    *   Fill in the fields and click "Save & Restart".

2.  **Manual JSON**:
    Create a `config.json` file in the root directory:

```json
{
    "bot_token": "YOUR_TELEGRAM_BOT_TOKEN",
    "nexus_addr": "ws://bot-nexus:3005",
    "log_port": 8085
}
```

### Configuration Guide

1.  Chat with [@BotFather](https://t.me/BotFather) on Telegram.
2.  Send `/newbot`.
3.  Follow instructions to get your **Bot Token**.
4.  Paste it into `config.json`.

| Field | Description |
| :--- | :--- |
| `bot_token` | Your Telegram Bot Token. |
| `nexus_addr` | Address of the BotNexus WebSocket server. |
| `log_port` | Port for the Web UI and Log viewer. |

## 🚀 Deployment

### Docker (Recommended)

This service is part of the BotMatrix `docker-compose.yml`.

1.  Enable the service in `docker-compose.yml` (uncomment the `telegram-bot` section).
2.  Place your `config.json` in `TelegramBot/config.json`.
3.  Run:
    ```bash
    docker-compose up -d --build telegram-bot
    ```

### Manual Build

```bash
cd TelegramBot
go mod tidy
go build -o TelegramBot main.go
./TelegramBot
```
