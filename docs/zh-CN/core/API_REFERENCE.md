# BotMatrix API 接口参考

> [🌐 English](../en-US/API_REFERENCE.md) | [简体中文](API_REFERENCE.md)
> [⬅️ 返回文档中心](README.md) | [🏠 返回项目主页](../../README.md)

BotMatrix 兼容 **OneBot v11** 协议，并在此基础上扩展了多平台支持和系统管理接口。

## 🔌 协议基础

- **通信协议**: WebSocket (Positive/Reverse)
- **数据格式**: JSON
- **默认端口**: 3001 (BotNexus), 3002 (WebUI API)

---

## 📥 机器人上报 (Events)

机器人端上报的消息格式遵循 OneBot 标准。

### 1. 消息事件 (Message Events)
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

### 2. 元事件 (Meta Events)
- **心跳 (heartbeat)**: 周期性上报，确保连接存活。
- **生命周期 (lifecycle)**: 机器人上线/离线通知。

---

## 📤 系统指令 (Actions)

BotNexus 发送给机器人或由 Worker 发出的指令。

### 1. 发送消息 (send_msg)
```json
{
    "action": "send_msg",
    "params": {
        "message_type": "group",
        "group_id": 87654321,
        "message": "这是一条自动回复"
    }
}
```

### 2. 获取登录信息 (get_login_info)
用于获取当前机器人的昵称和 ID。

### 3. 系统管理扩展 (Custom Actions)
- **`#status`**: 获取服务器运行状态。
- **`#reload`**: 重新加载插件。
- **`#broadcast`**: 全局广播。

---

## 🌐 WebUI API

Web 管理后台使用的 RESTful API。

### 1. 获取日志 (GET /api/logs)
- **描述**: 获取最新的系统日志。
- **返回**: 字符串数组。

### 2. 机器人列表 (GET /api/bots)
- **描述**: 获取当前在线的所有机器人信息。
- **返回**:
```json
[
    {
        "self_id": "12345678",
        "platform": "qq",
        "connected_at": "2023-10-01T12:00:00Z",
        "status": "online",
        "avatar": "https://q.qlogo.cn/headimg_dl?dst_uin=12345678&spec=640"
    }
]
```

### 3. 头像代理 (GET /api/proxy/avatar?url=...)
- **描述**: 代理外部头像图片，解决跨域 (CORS) 和 Referer 限制问题。
- **参数**: `url` - 原始头像图片的编码 URL。

### 4. 更新路由规则 (POST /api/routing/update)
- **描述**: 动态修改消息路由规则。

---

## 🧪 调试工具

推荐使用 `wscat` 或 Postman 进行 WebSocket 调试：
```bash
wscat -c ws://localhost:3001/ws/bot -H "X-Self-ID: 123456" -H "X-Platform: wechat"
```
