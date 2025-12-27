# WxBot 💬

A **Python-based** WeChat (Personal/Web Protocol) Robot implementation for [BotMatrix](../README.md).

> **Disclaimer**: This bot uses the Web WeChat protocol. Use with caution as it may carry risks of account restrictions.

## ✨ Features

*   **OneBot 11 Interface**: Provides a standard OneBot 11 implementation over WebSocket.
*   **Rich Message Support**: Text, Image, and basic system events.
*   **Group Management**: Kick, Ban (Simulated), Group List, Member List.
*   **Burn After Reading**: **New!** Supports message recall.

### 🔥 Burn After Reading (Message Recall)

WxBot supports message recall via the `delete_msg` action.

*   **ID Format**: `ToUserName:MsgID:LocalID` (Composite ID).
*   **Mechanism**:
    *   The bot tracks `LocalID` and `MsgID` (Server ID) upon sending.
    *   The `delete_msg` action uses the `webwxrevokemsg` API to retract the message.
    *   Supports both Group and Private messages.
    *   Recall time limit is subject to WeChat's standard rules (usually 2 minutes).

## 🛠 Configuration

Configuration is handled via `config.json` in the root or passed via environment variables in Docker.

```json
{
    "network": {
        "ws_server": {
            "host": "0.0.0.0",
            "port": 3001
        }
    },
    "driver": {
        "nexus_addr": "ws://bot-manager:3005"
    }
}
```

## 🚀 Deployment

### Docker

The `wxbot` service is the default bot in BotMatrix.

```bash
docker-compose up -d wxbot
```

### Manual Run

```bash
cd WxBot
pip install -r requirements.txt
python main.py
```
