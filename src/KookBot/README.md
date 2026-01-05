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

KookBot supports two ways to configure:

1.  **Web UI (Recommended)**:
    *   Start the bot.
    *   Access `http://localhost:8088/config-ui` (default port is 8088).
    *   Fill in the fields and click "Save & Restart".

2.  **Manual JSON**:
    Create a `config.json` file in the root directory:

```json
{
    "nexus_addr": "ws://bot-nexus:3005",
    "token": "YOUR_BOT_TOKEN",
    "log_port": 8088
}
```

| Field | Description |
| :--- | :--- |
| `nexus_addr` | Address of the BotNexus WebSocket server. |
| `token` | Your Kook Bot Token. |
| `log_port` | Port for the Web UI and Log viewer. |

## 🚀 Deployment

```bash
cd KookBot
go build -o KookBot.exe main.go
./KookBot.exe
```
