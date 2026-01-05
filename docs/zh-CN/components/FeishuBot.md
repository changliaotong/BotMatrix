# FeishuBot 🐦

A **Go-based** Feishu/Lark (飞书) Robot implementation for [BotMatrix](../README.md), utilizing the official **WS (WebSocket) Mode** for seamless enterprise integration without requiring a public IP.

## ✨ Features

*   **Stream Mode (WebSocket)**: Uses the official `larksuite/oapi-sdk-go` v3 WebSocket support. No need for Public IP or Callback URL configuration.
*   **OneBot 11 Compliance**:
    *   **Sending**: `send_group_msg`, `send_private_msg`, `delete_msg`.
    *   **Receiving**: Supports **Text**, **Image**, **File**, **Audio**, and **Rich Text** (Post) messages.
    *   **Meta**: `get_login_info`, `get_group_list`.
    *   **CQ Codes**: Automatically converts rich media to `[CQ:image]`, `[CQ:file]`, etc.
*   **Burn After Reading**: **New!** Supports message recall.

### 🔥 Burn After Reading (Message Recall)

FeishuBot supports message recall via the `delete_msg` action.

*   **Mechanism**: Uses the `im.v1.messages.delete` API.
*   **Usage**: The `message_id` returned by send actions can be used directly for recall.

## 🛠 Configuration

Create a `config.json` file in the root directory:

```json
{
    "app_id": "cli_...",
    "app_secret": "...",
    "encrypt_key": "...",
    "verification_token": "...",
    "nexus_addr": "ws://bot-manager:3005"
}
```

### Configuration Guide

1.  Go to [Feishu Developer Console](https://open.feishu.cn/app).
2.  Create a "Custom App" (企业自建应用).
3.  Enable **Robot** capability.
4.  In "Permissions" (权限管理), grant:
    *   `im:message` (Receive messages)
    *   `im:message:send_as_bot` (Send messages)
    *   `im:chat:readonly` or `im:chat` (Get group list)
5.  In "Event Subscriptions" (事件订阅):
    *   Set Encrypt Key (Optional, but recommended).
    *   Add Event: `Receive Message` (v2.0).
6.  Copy `App ID` and `App Secret` to `config.json`.

## 🚀 Deployment

### Docker (Recommended)

This service is part of the BotMatrix `docker-compose.yml`.

1.  Enable the service in `docker-compose.yml` (uncomment the `feishu-bot` section).
2.  Place your `config.json` in `FeishuBot/config.json`.
3.  Run:
    ```bash
    docker-compose up -d --build feishu-bot
    ```

### Manual Build

```bash
# Enter directory
cd FeishuBot

# Install dependencies
go mod tidy

# Build
go build -o FeishuBot.exe main.go

# Run
./FeishuBot.exe
```

## 🔗 References

*   [Feishu Open Platform](https://open.feishu.cn/document/home/index)
*   [Lark Go SDK](https://github.com/larksuite/oapi-sdk-go)
