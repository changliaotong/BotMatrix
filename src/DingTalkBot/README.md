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
    *   `delete_msg` -> **New!** Supports recalling messages (Enterprise Mode required).
    *   `message` Event -> Forwards received DingTalk messages (Text/Content) to Nexus.
*   **Intelligent Parsing**:
    *   Automatically parses `im.message.receive_v1` events.
    *   Extracts Text, Sender ID, and Group ID.
    *   Forwards unknown events as `notice` type for debugging.
*   **Security**: Implements DingTalk's HMAC-SHA256 signature algorithm for safe message delivery.
*   **Auto-Configuration**: Automatically generates a unique `self_id` based on your token if not provided.

### 🔥 Burn After Reading (Message Recall)

DingTalkBot supports the **Burn After Reading** feature orchestrated by BotNexus.

*   **Requirement**: You must configure **Stream Mode** (Enterprise Robot with `client_id` and `client_secret`) to use this feature. Webhook-only bots cannot recall messages.
*   **Mechanism**:
    *   When sending a message, the bot returns a special `message_id` format: `conversationID|processQueryKey`.
    *   The `delete_msg` action uses this ID to call the DingTalk `groupMessages/recall` API.
*   **Usage**: Simply enable "Burn After Reading" in the BotNexus dashboard or send a `delete_msg` command with the ID returned by a previous send action.

## 🛠 Configuration

DingTalkBot supports two ways to configure:

1.  **Web UI (Recommended)**:
    *   Start the bot.
    *   Access `http://localhost:8089/config-ui` (default port is 8089).
    *   Fill in the fields and click "Save & Restart".

2.  **Manual JSON**:
    Create a `config.json` file in the root directory:

```json
{
    "nexus_addr": "ws://bot-nexus:3005",
    "access_token": "YOUR_WEBHOOK_ACCESS_TOKEN",
    "secret": "YOUR_WEBHOOK_SECRET",
    "client_id": "YOUR_APP_KEY",
    "client_secret": "YOUR_APP_SECRET",
    "log_port": 8089
}
```

### Configuration Guide

| Field | Description | Mode |
| :--- | :--- | :--- |
| `nexus_addr` | Address of the BotNexus WebSocket server. | **Required** |
| `access_token` | Token from the Webhook URL. | **Webhook** |
| `secret` | "Sign" (SEC...) in Robot settings. | **Webhook** |
| `client_id` | AppKey of your Enterprise Internal Robot. | **Stream (Required for Recall)** |
| `client_secret` | AppSecret of your Enterprise Internal Robot. | **Stream (Required for Recall)** |
| `log_port` | Port for the Web UI and Log viewer. | **Required** |

> **Tip**: 
> *   For **Send Only**, just configure `access_token` (and optional `secret`).
> *   For **Receive, Send & Recall**, configure `client_id` + `client_secret` (Stream Mode). The `access_token` is optional if Stream Mode sending is fully supported, but recommended for fallback or specific webhook features.

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
