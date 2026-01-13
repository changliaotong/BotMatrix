# 🔌 IM Adapters Catalog

> **Version**: 1.0
> **Status**: Mainstream adapters documented
> [🌐 English](ADAPTERS_CATALOG.md) | [简体中文](../zh-CN/core/ADAPTERS_CATALOG.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

BotMatrix supports a wide range of IM platforms through its adapter system, primarily leveraging the OneBot v11 protocol.

---

## 1. Supported Platforms

### 1.1 NapCat (QQ)
- **Protocol**: OneBot v11.
- **Features**: Group management, private messaging, multi-media support.
- **Deployment**: Best run in Docker with QR code login at port `:6099`.

### 1.2 WxBotGo (WeChat)
- **Protocol**: Custom Go implementation (Android/NDK based).
- **Features**: High stability, supports private/group messaging and media.
- **Limits**: No message recall or ban support due to protocol restrictions.

### 1.3 DingTalk & Feishu
- **Protocol**: Stream-based API.
- **Features**: Integrated with enterprise workflows, support for cards and interactive messages.

### 1.4 Discord
- **Protocol**: WebSocket/Gateway.
- **Features**: Mapping Discord channels to OneBot `group_id`, supports markdown-to-CQCode conversion.

---

## 2. Configuration Tips

- **X-Self-ID**: Always provide the bot's ID in the connection header for multi-bot environments.
- **Reverse WebSocket**: Recommended for production environments to allow BotNexus to manage connections proactively.
