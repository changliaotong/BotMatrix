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

WeComBot supports two ways to configure:

1.  **Web UI (Recommended)**:
    *   Start the bot.
    *   Access `http://localhost:8083/config-ui` (default port is 8083).
    *   Fill in the fields and click "Save & Restart".

2.  **Manual JSON**:
    Create a `config.json` file in the root directory:

```json
{
    "nexus_addr": "ws://bot-nexus:3005",
    "log_port": 8083,
    "corp_id": "YOUR_CORP_ID",
    "agent_id": "1000001",
    "secret": "YOUR_APP_SECRET",
    "token": "YOUR_TOKEN",
    "encoding_aes_key": "YOUR_AES_KEY",
    "listen_port": 8084
}
```

| Field | Description |
| :--- | :--- |
| `nexus_addr` | Address of the BotNexus WebSocket server. |
| `log_port` | Port for the Web UI and Log viewer. |
| `corp_id` | Your Enterprise CorpID. |
| `agent_id` | The AgentID of your Internal App. |
| `secret` | The Secret of your Internal App. |
| `token` | Token for callback verification. |
| `encoding_aes_key` | EncodingAESKey for callback encryption. |
| `listen_port` | Local port to listen for WeCom callbacks. |

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
