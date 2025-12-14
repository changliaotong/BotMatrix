# TelegramBot ✈️

A **Go-based** Telegram Robot implementation for [BotMatrix](../README.md).

## ✨ Features

*   **OneBot 11 Compliance**:
    *   **Sending**: `send_group_msg` (to Groups), `send_private_msg` (to Users).
    *   **Receiving**: Supports Text messages.
    *   **Meta**: `get_login_info`.
*   **Long Polling**: Uses standard Telegram Long Polling for event updates.
*   **Zero-Config Networking**: Connects outbound to BotNexus; no public IP required for the bot itself.

## 🛠 Configuration

Create a `config.json` file in the root directory:

```json
{
    "bot_token": "YOUR_TELEGRAM_BOT_TOKEN",
    "nexus_addr": "ws://bot-manager:3005"
}
```

### Configuration Guide

1.  Chat with [@BotFather](https://t.me/BotFather) on Telegram.
2.  Send `/newbot`.
3.  Follow instructions to get your **Bot Token**.
4.  Paste it into `config.json`.

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
