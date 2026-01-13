# 🔌 Interface & Protocols

> **Version**: 2.0
> **Status**: Core specification released
> [🌐 English](INTERFACE_SPECIFICATIONS.md) | [简体中文](../zh-CN/core/INTERFACE_SPECIFICATIONS.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

BotMatrix is built on the **OneBot v11** standard and extends it with **Model Context Protocol (MCP)** and custom actions for multi-platform adaptation, cross-enterprise collaboration, and AI skill distribution.

---

## 1. Communication Basics

- **OneBot Adapter Layer**: WebSocket (Forward/Reverse), default port `3001` (BotNexus).
- **AI Extension Layer**: MCP SSE (Server-Sent Events), supporting dynamic capability discovery.
- **Management API**: RESTful API, default port `3002` (WebUI).
- **Data Format**: `JSON`.

---

## 2. OneBot v11 Compatibility

BotMatrix is deeply compatible with the OneBot v11 standard, supporting various clients (NapCat, WxBotGo, etc.).

### 2.1 Adapter Status
| Client | Status | Core Features | Notes |
| :--- | :--- | :--- | :--- |
| **NapCat (QQ)** | Full | Messages, Group Admin, Private, Media | Standard implementation |
| **WxBotGo (WeChat)** | Basic | Private, Group, Login Info | No message recall/ban due to protocol limits |
| **DingTalkBot** | Complete | Message Conversion, Nexus Commands | Supports core actions |
| **DiscordBot** | Complete | Channel Mapping, CQ Code handling | Maps Discord ChannelID -> group_id |

---

## 3. Custom System Actions

Beyond standard `send_msg` and `get_login_info`, BotMatrix extends:
- **`#status`**: Returns Nexus server running status.
- **`#reload`**: Hot-reloads plugin configurations.
- **`#broadcast`**: Sends messages to all online bots.

---

## 4. Skill System & Intent Dispatch

### 4.1 Skill Center
The "App Store" for bot capabilities.
- **One-click Mount**: Bind skills to specific bots or groups.
- **Hot-swappable**: Enable/disable skills instantly without restart.
- **Isolation**: Tenant-level isolation for private and shared skills.

### 4.2 Intent Dispatch System
Acts as a "triage desk" using lightweight AI to identify user intent before routing.
- **Dual-Layer Dispatch**: 
    - **System Level**: Nexus decides which Worker or Skill handles the message.
    - **User Side**: Multi-role bot distribution within a group.
- **Workflow**:
    1. **Intent Identification**: Matches features or calls models to return an intent code.
    2. **Route Lookup**: Finds target destination in `AIIntentRoutingGORM`.
    3. **Dispatch**: Loads configuration and executes.

---

## 5. MCP & Global Agent Mesh API

Dedicated interfaces for AI collaboration and B2B scenarios.

### 5.1 MCP SSE Endpoint (`GET /api/mcp/sse`)
- **Function**: Standard MCP tool discovery and function call notifications.
- **Auth**: JWT or B2B Federation Identity Tokens.

### 5.2 Cross-Domain Tool Call (`POST /api/mesh/call`)
- **Description**: Proxy calls to authorized MCP tools in a remote enterprise.
- **Example**:
```json
{
    "target_ent_id": 1001,
    "tool_name": "check_inventory",
    "arguments": { "item_id": "SKU-001" }
}
```

---

## 6. Security

- **JWT Tokens**: Used for WebUI and API access.
- **B2B Federation**: PKI-based enterprise signatures for cross-domain trust.
- **PII Guard**: Built-in data masking for sensitive information (phones, names) before sending to LLMs.
