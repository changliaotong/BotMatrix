# KookBot 🎮

A **Go-based** Kook (formerly Kaiheila) Robot implementation for [BotMatrix](../README.md).

## ✨ Features

*   **WebSocket Connection**: Real-time event handling.
*   **Rich Text Support**: KMarkdown support.
*   **Burn After Reading**: **New!** Supports message recall via `delete_msg`.

### 🔥 Burn After Reading (Message Recall)

*   **Mechanism**: Uses `message/delete` API.
*   **Usage**: Send `delete_msg` with the `message_id` returned from send actions.

## 🛠 Configuration

Create `config.json`:

```json
{
    "nexus_addr": "ws://bot-manager:3005",
    "token": "YOUR_BOT_TOKEN"
}
```

## 🚀 Deployment

```bash
cd KookBot
go build -o KookBot.exe main.go
./KookBot.exe
```
