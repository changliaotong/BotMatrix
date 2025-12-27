# SlackBot 📢

A **Go-based** Slack Robot implementation for [BotMatrix](../README.md), utilizing **Socket Mode** for secure, firewall-friendly enterprise integration.

## ✨ Features

*   **Socket Mode**: No need to expose public HTTP endpoints.
*   **OneBot 11 Compliance**: Maps Slack Channels to Groups and DMs to Private messages.
*   **Burn After Reading**: **New!** Supports message recall.

### 🔥 Burn After Reading (Message Recall)

*   **ID Format**: `ChannelID:Timestamp` (e.g., `C123456:167890000.123456`).
*   **Mechanism**: Uses `chat.delete` API.
*   **Permissions**: Requires `chat:write` scope.

## 🛠 Configuration

Create `config.json`:

```json
{
    "nexus_addr": "ws://bot-manager:3005",
    "app_token": "xapp-...",
    "bot_token": "xoxb-..."
}
```

*   `app_token`: Level-1 Token (starts with `xapp-`), enable "Socket Mode" in Slack App settings.
*   `bot_token`: Bot User OAuth Token (starts with `xoxb-`).

## 🚀 Deployment

```bash
cd SlackBot
go build -o SlackBot.exe main.go
./SlackBot.exe
```
