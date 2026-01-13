# 🔌 接口规范与协议兼容性 (Interface & Protocols)

> **版本**: 2.0
> **状态**: 核心规范已发布
> [🌐 English](../en-US/INTERFACE_SPECIFICATIONS.md) | [简体中文](INTERFACE_SPECIFICATIONS.md)
> [⬅️ 返回文档中心](README.md) | [🏠 返回项目主页](../../README.md)

BotMatrix 构建在 **OneBot v11** 标准之上，并通过 **Model Context Protocol (MCP)** 和自定义扩展实现了多平台适配、跨企业协作及 AI 技能分发。

---

## 1. 通信基础 (Communication Basics)

- **OneBot 适配层**: WebSocket (正向/反向)，默认端口 `3001` (BotNexus)。
- **AI 扩展层**: MCP SSE (Server-Sent Events)，支持动态能力发现。
- **管理后台**: RESTful API，默认端口 `3002` (WebUI)。
- **数据格式**: 统一使用 `JSON`。

---

## 2. OneBot v11 协议兼容性

BotMatrix 深度兼容 OneBot v11 标准，支持各类主流机器人客户端（如 NapCat, WxBotGo 等）。

### 2.1 平台适配状态
| 客户端 | 兼容状态 | 核心支持功能 | 备注 |
| :--- | :--- | :--- | :--- |
| **NapCat (QQ)** | 完整支持 | 消息、群管、私聊、多媒体 | 标准 OneBot 11 实现 |
| **WxBotGo (WeChat)** | 基本支持 | 私聊、群聊、登录信息 | 受限于 Web 协议，暂不支持撤回/禁言 |
| **DingTalkBot** | 已完成 | 消息转换、Nexus 命令 | 支持核心 Action |
| **DiscordBot** | 已完成 | 消息/频道映射、CQ 码处理 | 映射 Discord ChannelID -> group_id |
| **FeishuBot** | 已完成 | P2MessageReceiveV1 转换 | 支持核心 API 集成 |
| **EmailBot** | 已完成 | 邮件 <-> 私聊消息转换 | 通过 WebSocket 接入 Nexus |

### 2.2 核心事件 (Events)
所有上报遵循 OneBot `post_type` 标准：
- `message`: 消息事件（private/group）。
- `notice`: 通知事件（群成员变动、文件上传等）。
- `request`: 请求事件（加群、加好友申请）。
- `meta_event`: 元事件（心跳 `heartbeat`、生命周期 `lifecycle`）。

---

## 3. 系统扩展指令 (Custom Actions)

除了标准的 `send_msg`, `get_login_info` 外，BotMatrix 扩展了以下指令：

- **`#status`**: 返回 Nexus 服务器运行状态。
- **`#reload`**: 零重启热加载插件配置。
- **`#broadcast`**: 向所有在线机器人发送广播消息。

---

## 4. 技能系统 (Skill System) 与智能分发

为了平滑引入 AI 技能而不错乱旧版逻辑，系统采用了动态开关机制与双层意图调度架构。

### 4.1 技能中心 (Skill Center)
技能中心是机器人能力的“应用商店”，负责技能的发现、分发与挂载。
- **一键挂载**: 用户可以将技能直接绑定到特定的机器人实例或群组。
- **热插拔设计**: 技能的开启与关闭即时生效，无需重启任何服务。
- **隔离机制**: 租户间的技能完全隔离，支持私有技能、组织内共享技能与全局公开技能。

### 4.2 智能意图分发系统 (Intent Dispatch)
系统充当了“分诊台”，在消息进入核心逻辑之前，先通过轻量级 AI 模型识别用户意图。
- **双层调度**: 
    - **系统级**: 由 BotNexus 决定消息转发给哪个 Worker 或 AI 技能。
    - **用户侧**: 在群组中实现多机器人角色分工（如：群管机器人 A 识别技术问题后转发给技术机器人 B）。
- **工作流程**: 
    1. **意图判定**: 调用 `AIIntentGORM` 定义的意图列表，通过模型返回结构化意图 Code。
    2. **路由查找**: 根据 `IntentID` 在 `AIIntentRoutingGORM` 表中查找目标（Skill/Worker/Plugin）。
    3. **任务分发**: 加载对应配置并执行。

### 4.3 技能训练中心 (Skill Training Center)
训练中心通过数据驱动让 AI 表现持续提升。
- **提示词 IDE**: 可视化编辑 System/User Prompt，支持变量动态注入（如 `{{.UserName}}`）。
- **语料标注**: 自动抓取回复效果不佳的对话，转化为 Few-shot 示例进行纠错。
- **RAG 实验室**: 在线上传知识库 (PDF/MD)，调试分段策略与检索准确度。

### 4.4 开关机制 (`ENABLE_SKILL`)
- **默认关闭**: 生产环境默认不启动 GORM 与任务调度器。
- **配置优先级**: 环境变量 > `config.json` > Web UI 配置。

### 4.5 智能路由策略
- **定向投递**: `Dispatcher` 仅将 `skill_call` 分发给在 `botmatrix:worker:register` 频道中报备了对应能力的 Worker。
- **隔离旧版**: 旧版 Worker 因不报备 `capabilities`，会自动被技能分发名单排除，确保其仅处理基础消息。

---

## 5. MCP & Global Agent Mesh API

针对 AI 协作与跨企业场景的专用接口。

### 5.1 MCP SSE 端点 (`GET /api/mcp/sse`)
- **功能**: 实现符合 MCP 标准的工具发现与函数调用通知。
- **认证**: 支持标准 JWT 或 B2B 联邦身份令牌。

### 5.2 跨域工具调用 (`POST /api/mesh/call`)
- **描述**: 代理调用远程企业授权的 MCP 工具。
- **请求示例**:
```json
{
    "target_ent_id": 1001,
    "tool_name": "check_inventory",
    "arguments": { "item_id": "SKU-001" }
}
```

---

## 6. WebUI 管理接口 (RESTful)

- **日志查询**: `GET /api/logs` - 获取系统实时运行日志。
- **机器人管理**: `GET /api/bots` - 获取所有在线机器人状态与平台信息。
- **路由管理**: `POST /api/routing/update` - 动态修改消息路由规则。
- **头像代理**: `GET /api/proxy/avatar?url=...` - 绕过 Referer 限制。

---

## 7. 安全与认证 (Security)

- **JWT 令牌**: 用于 WebUI 与标准 API 访问。
- **B2B 联邦身份**: 跨企业调用时，通过企业私钥签名生成短期令牌，实现双向信任。
- **PII 脱敏**: 接口层内置隐私堡垒，发送至 LLM 前自动屏蔽手机号、姓名等敏感字段。
