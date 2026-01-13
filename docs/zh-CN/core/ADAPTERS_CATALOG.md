# BotMatrix 适配器目录 (Adapters Catalog)

> [⬅️ 返回文档中心](../README.md) | [🏠 返回项目主页](../../README.md)

BotMatrix 通过适配器模式支持多种 IM 平台，所有适配器均将平台特定协议转换为标准的 **OneBot v11** 协议。

---

## 1. 核心特性 (Common Features)

所有适配器均支持以下核心功能：
- **OneBot v11 兼容**：支持 `send_msg`, `send_group_msg`, `send_private_msg` 等。
- **消息撤回 (Burn After Reading)**：支持通过 `delete_msg` 撤回已发送的消息。
- **多媒体支持**：自动处理图片、文件、语音等 CQ 码。
- **自动注册**：连接至 Nexus 时自动上报 `self_id` 和平台能力。

---

## 2. 适配器列表 (Adapter List)

### 2.1 飞书 (Feishu/Lark)
- **模式**：官方 WebSocket 模式 (无需公网 IP)。
- **配置**：`app_id`, `app_secret`, `encrypt_key`。
- **优势**：完美支持企业自建应用，消息类型丰富。

### 2.2 钉钉 (DingTalk)
- **模式**：支持 Webhook 模式与 Stream 模式。
- **配置**：`client_id`, `client_secret` (Stream 模式必需)。
- **注意**：消息撤回功能仅在 Stream 模式下可用。

### 2.3 企业微信 (WeCom)
- **模式**：企业自建应用 API。
- **配置**：`corp_id`, `agent_id`, `secret`。
- **特性**：支持 24 小时内消息撤回，深度集成企业办公场景。

### 2.4 腾讯 QQ 机器人 (QQ Guild/Official)
- **模式**：官方机器人 API。
- **优化**：针对 QQ 频道 ID (OpenID) 进行了 `FlexibleInt64` 映射，确保 int64 兼容性。
- **特性**：支持“智能发送”机制，规避腾讯消息频率限制。

### 2.5 Telegram / Discord / Slack
- **模式**：官方 Bot API / WebSocket。
- **全球化**：完美适配海外主流社交平台。

### 2.6 微信个人号 (WxBot / WxBotApp)
- **模式**：基于特定协议的个人号接入。
- **特性**：支持朋友圈监控、群管理等高级功能。

### 2.7 WxBotGo (Android/NDK)
- **概述**：基于 Go 语言开发的 WeChat 机器人框架，深度适配 Android 运行环境。
- **消息支持**：文本、图片、表情、语音、名片及文件消息。
- **局限性**：由于协议限制，不支持 `delete_msg` (撤回) 和部分群管理功能 (如禁言、踢人)。
- **部署**：支持 NDK 编译，可在 Android 设备或模拟器中高效运行。

---

## 3. 部署与配置

### 3.1 统一配置模板 (`config.json`)
```json
{
    "nexus_addr": "ws://your-nexus:5000",
    "access_token": "optional_token",
    "platform_specific_config": "..."
}
```

### 3.2 运行方式
所有适配器均提供 Docker 镜像，推荐使用 `docker-compose` 统一部署。

---

## 4. 开发者指南

如果你想为新平台开发适配器，请参考 `DEVELOPMENT_GUIDE.md` 中的“适配器开发规范”部分。

---
