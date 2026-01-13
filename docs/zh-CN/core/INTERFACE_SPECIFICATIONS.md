# 🔌 接口规范与协议兼容性 (Interface Specifications)

> [⬅️ 返回文档中心](../README.md) | [🏠 返回项目主页](../../README.md)

BotMatrix 采用标准化的协议体系，确保不同平台与模块间的高效通信。系统核心兼容 **OneBot v11** 协议，并扩展了多平台适配与系统管理接口。

---

## 1. 通信协议基础

- **核心协议**: OneBot v11 (兼容性扩展)
- **通信方式**: WebSocket (正向/反向), HTTP RESTful API
- **数据格式**: 标准 JSON
- **默认端口**: 
    - `3001`: BotNexus (消息网关与 WebSocket)
    - `3002`: WebUI 后端 API

---

## 2. 机器人上报 (Events - OneBot 标准)

机器人端向 BotNexus 上报的消息格式遵循 OneBot 标准。

### 2.1 消息事件 (Message Events)
支持私聊 (`private`) 与群聊 (`group`) 消息。包含 `time`, `self_id`, `post_type`, `message_type`, `user_id`, `message` 等核心字段。

### 2.2 元事件 (Meta Events)
- **心跳 (heartbeat)**: 周期性上报，确保连接存活。
- **生命周期 (lifecycle)**: 机器人上线/离线通知。

---

## 3. 系统指令 (Actions - OneBot 扩展)

### 3.1 核心动作
- `send_msg`: 发送通用消息。
- `send_private_msg`: 发送私聊消息。
- `send_group_msg`: 发送群消息。
- `delete_msg`: 撤回消息。
- `get_login_info`: 获取当前机器人账号信息。

### 3.2 系统管理扩展指令
- `#status`: 获取服务器运行状态。
- `#reload`: 重新加载插件。
- `#broadcast`: 全局广播。

---

## 4. OneBot v11 协议适配状态 (Compatibility)

BotMatrix 对不同平台适配器的 OneBot 11 兼容性情况如下：

| 适配器 | 状态 | 核心功能支持 | 备注 |
| :--- | :--- | :--- | :--- |
| **NapCat** | ✅ 已完成 | 完整 OneBot 11 标准支持 | 推荐的 QQ 适配器 |
| **WxBotGo** | ✅ 已完成 | 消息收发、登录信息、群列表 | 受限于底层库，不支持撤回/禁言 |
| **FeishuBot** | ✅ 已完成 | 消息收发、撤回、群/成员列表 | 完整 Feishu API 集成 |
| **DiscordBot**| ✅ 已完成 | 消息收发、ChannelID 映射 | 支持基本 CQ 码 |
| **KookBot** | ✅ 已完成 | 文本/图片/Kmarkdown 支持 | 完整 WebSocket 链路 |
| **DingTalkBot**| ✅ 已完成 | 核心消息收发、登录信息 | 高级事件待完善 |
| **EmailBot** | ✅ 已完成 | 邮件转私聊、邮件发送 | 作为私聊消息处理 |

---

## 5. WebUI 与 扩展 API

### 5.1 管理 API
- `GET /api/logs`: 获取实时系统日志。
- `GET /api/bots`: 获取所有在线机器人状态。
- `POST /api/routing/update`: 动态修改消息路由规则。

### 5.2 Mesh & MCP 核心接口
- `GET /api/mesh/discover`: 联邦网络服务发现。
- `POST /api/mesh/connect`: 建立企业级 B2B 信任连接。
- `POST /api/mcp/v1/tools/call`: 标准 MCP 工具执行接口。
- `GET /api/mcp/sse`: 符合 MCP 规范的 SSE 通信端点。

---

## 6. 调试工具建议

推荐使用 `wscat` 进行 WebSocket 链路测试：
```bash
wscat -c ws://localhost:3001/ws/bot -H "X-Self-ID: 123456" -H "X-Platform: wechat"
```

---
*最后更新日期：2026-01-13*
