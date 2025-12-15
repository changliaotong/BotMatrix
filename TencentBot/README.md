# TencentBot 🐧

A **Go-based** Official QQ Guild/Group Robot implementation for [BotMatrix](../README.md), using the official `botgo` SDK.

## ✨ Features

*   **Official API**: Compliant with Tencent's requirements for QQ Guilds and Groups.
*   **OneBot 11 Compliance**: Bridges official events to OneBot standard.
*   **Burn After Reading**: **New!** Supports message recall.

### 🔥 Burn After Reading (Message Recall)

*   **Mechanism**: Uses `RetractMessage` API.
*   **Requirement**:
    *   For Group messages, requires the `group_id` context (handled internally via ID mappings or `channel_id` if available).
    *   Returns valid `message_id` for recall operations.

## 🛠 Configuration

Create `config.json`:

```json
{
    "nexus_addr": "ws://bot-manager:3005",
    "appid": "YOUR_APPID",
    "token": "YOUR_TOKEN",
    "secret": "YOUR_SECRET"
}
```

## 🚀 Deployment

```bash
cd TencentBot
go build -o TencentBot.exe main.go
./TencentBot.exe
```
