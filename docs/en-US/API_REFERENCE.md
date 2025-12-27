# BotMatrix API Reference

> [🌐 English](API_REFERENCE.md) | [简体中文](../zh-CN/API_REFERENCE.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

BotMatrix is compatible with the **OneBot v11** protocol and extends it with multi-platform support and system management interfaces.

## 🔌 Protocol Basics

- **Communication Protocol**: WebSocket (Positive/Reverse)
- **Data Format**: JSON
- **Default Ports**: 3001 (BotNexus), 3002 (WebUI API)

---

## 📥 Bot Events

The message format reported by the bot client follows the OneBot standard.

### 1. Message Events
```json
{
    "time": 1632832800,
    "self_id": "12345678",
    "post_type": "message",
    "message_type": "group",
    "sub_type": "normal",
    "message_id": 1,
    "group_id": "87654321",
    "user_id": "10001",
    "message": "hello",
    "raw_message": "hello",
    "font": 0,
    "sender": {
        "user_id": "10001",
        "nickname": "User",
        "role": "member"
    }
}
```

### 2. Meta Events
- **heartbeat**: Reported periodically to ensure the connection is alive.
- **lifecycle**: Notification of bot online/offline status.

---

## 📤 System Actions

Commands sent by BotNexus to bots or issued by Workers.

### 1. Send Message (send_msg)
```json
{
    "action": "send_msg",
    "params": {
        "message_type": "group",
        "group_id": 87654321,
        "message": "This is an automatic reply"
    }
}
```

### 2. Get Login Info (get_login_info)
Used to retrieve the nickname and ID of the current bot.

### 3. Custom System Actions
- **`#status`**: Get server running status.
- **`#reload`**: Reload plugins.
- **`#broadcast`**: Global broadcast.

---

## 🌐 WebUI API

RESTful API used by the Web Management Interface.

### 1. Get Logs (GET /api/logs)
- **Description**: Fetch the latest system logs.
- **Returns**: Array of strings.

### 2. Bot List (GET /api/bots)
- **Description**: Fetch information about all currently online bots.
- **Returns**:
```json
[
    {
        "self_id": "12345678",
        "platform": "qq",
        "connected_at": "2023-10-01T12:00:00Z",
        "status": "online"
    }
]
```

### 3. Update Routing Rules (POST /api/routing/update)
- **Description**: Dynamically modify message routing rules.

---

## 🧪 Debugging Tools

Recommended to use `wscat` or Postman for WebSocket debugging:
```bash
wscat -c ws://localhost:3001/ws/bot -H "X-Self-ID: 123456" -H "X-Platform: wechat"
```
