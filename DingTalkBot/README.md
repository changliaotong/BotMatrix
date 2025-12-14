# DingTalkBot 🤖

A **Go-based** DingTalk (钉钉) Robot implementation for [BotMatrix](../README.md), supporting both **Webhook Mode** (Outgoing) and **Stream Mode** (Incoming/Interactive).

> **Note**: This bot acts as a bridge between DingTalk and BotMatrix, converting DingTalk events into OneBot 11 standard events.

## ✨ Features

*   **Dual Mode Support**:
    *   **Webhook Mode**: Standard outgoing messages via HTTP Webhook.
    *   **Stream Mode**: Real-time event receiving via WebSocket (Official SDK), suitable for Enterprise Internal Robots.
*   **OneBot 11 Compliance**:
    *   `send_group_msg` / `send_msg` -> DingTalk Group Message.
    *   `send_private_msg` -> Simulated via @Mention in Group (Webhook limitation).
*   **Security**: Implements DingTalk's HMAC-SHA256 signature algorithm for safe message delivery.
*   **Auto-Configuration**: Automatically generates a unique `self_id` based on your token if not provided.

## 🛠 Configuration

Create a `config.json` file in the root directory:

```json
{
    "nexus_addr": "ws://bot-manager:3005",
    "access_token": "YOUR_WEBHOOK_ACCESS_TOKEN",
    "secret": "YOUR_WEBHOOK_SECRET",
    "client_id": "YOUR_APP_KEY",
    "client_secret": "YOUR_APP_SECRET"
}
```

### Configuration Guide

| Field | Description | Mode |
| :--- | :--- | :--- |
| `nexus_addr` | Address of the BotNexus WebSocket server. | **Required** |
| `access_token` | Token from the Webhook URL (e.g., `.../send?access_token=THIS_PART`). | **Webhook** |
| `secret` | "Secure Settings" -> "Sign" (SEC...) in DingTalk Robot settings. | **Webhook** |
| `client_id` | AppKey of your Enterprise Internal Robot. | **Stream** |
| `client_secret` | AppSecret of your Enterprise Internal Robot. | **Stream** |

> **Tip**: 
> *   For **Send Only**, just configure `access_token` (and optional `secret`).
> *   For **Receive & Send**, configure `client_id` + `client_secret` (for receiving) AND `access_token` (for sending).

## 🚀 Deployment

### Docker (Recommended)

This service is part of the BotMatrix `docker-compose.yml`.

1.  Enable the service in `docker-compose.yml` (uncomment the `dingtalk-bot` section).
2.  Place your `config.json` in `DingTalkBot/config.json`.
3.  Run:
    ```bash
    docker-compose up -d --build dingtalk-bot
    ```

### Manual Build

```bash
# Enter directory
cd DingTalkBot

# Install dependencies
go mod tidy

# Build
go build -o DingTalkBot.exe main.go

# Run
./DingTalkBot.exe
```

## 🔗 References

*   [DingTalk Custom Robot Guide](https://open.dingtalk.com/document/orgapp/custom-robot-access)
*   [DingTalk Stream Mode SDK](https://github.com/open-dingtalk/dingtalk-stream-sdk-go)
