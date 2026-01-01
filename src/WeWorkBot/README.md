# WeWorkBot 🏢

A **Go-based** WeWork (Enterprise WeChat/企业微信) Robot implementation for [BotMatrix](../README.md).

## ✨ Features

*   **Enterprise Integration**: Supports WeWork Enterprise Internal Apps.
*   **OneBot 11 Compliance**:
    *   `send_group_msg` -> Sends message to Chat (requires ChatID).
    *   `send_private_msg` -> Sends message to User (requires UserID).
    *   `delete_msg` -> **New!** Supports recalling messages (within 24 hours).
*   **Message Types**: Supports Text, Image, Markdown, and more.

### 🔥 Burn After Reading (Message Recall)

WeWorkBot fully supports the **Burn After Reading** feature.

*   **Mechanism**:
    *   The bot uses the Enterprise WeChat API `message/recall` to retract messages.
    *   Messages can be recalled within 24 hours of sending.
    *   The `message_id` returned by send actions is compatible with the `delete_msg` action.

## 🛠 Configuration

Create a `config.json` file in the root directory:

```json
{
    "nexus_addr": "ws://bot-manager:3005",
    "corp_id": "YOUR_CORP_ID",
    "agent_id": 1000001,
    "secret": "YOUR_APP_SECRET"
}
```

| Field | Description |
| :--- | :--- |
| `nexus_addr` | Address of the BotNexus WebSocket server. |
| `corp_id` | Your Enterprise CorpID. |
| `agent_id` | The AgentID of your Internal App. |
| `secret` | The Secret of your Internal App. |

## 🚀 Deployment

### Docker

Add to your `docker-compose.yml` or use the existing service definition.

### Manual Build

```bash
cd WeWorkBot
go mod tidy
go build -o WeWorkBot.exe main.go
./WeWorkBot.exe
```
