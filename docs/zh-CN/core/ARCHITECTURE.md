# 🏗️ BotMatrix 系统架构概览

> [🌐 English](../en-US/ARCHITECTURE.md) | [简体中文](ARCHITECTURE.md)
> [⬅️ 返回文档中心](../README.md) | [🏠 返回项目主页](../../README.md)

BotMatrix 是一个采用分布式、解耦设计的机器人矩阵管理系统。它通过核心的消息分发中心与多个执行节点协作，实现了高并发和高可扩展性。

## 🏗️ 核心组件

### 1. BotNexus (中心控制节点)
BotNexus 是整个系统的“大脑”和“路由器”。
- **职责**:
    - 维护与客户端（如 WxBot, QQBot）的 WebSocket 连接。
    - 接收原始消息事件 (Events)。
    - **AI 意图识别**: 初步解析用户意图，决定是调用传统插件还是 AI 技能。
    - **Global Agent Mesh (核心)**: 作为 Mesh 网络的枢纽，处理跨域发现、联邦身份验证与 B2B 协作逻辑。
    - **MCP Host**: 深度集成 Model Context Protocol，管理并调度本地与远程的 MCP Server 工具集。
    - **AI 隐私堡垒 (Privacy Bastion)**: 对全链路（包括 Mesh 协作）外发消息进行 PII 脱敏与还原处理。
    - **技能路由**: 根据意图分发至对应的 AI 模型、执行节点或 Mesh 合作伙伴。
    - 管理 Worker 节点的注册与心跳。
    - 提供 Web 管理后台 (WebUI)。
- **技术栈**: Go, Gin, WebSocket, Redis (Pub/Sub).

### 2. BotWorker (任务执行节点)
BotWorker 是实际处理业务逻辑的“四肢”。
- **职责**:
    - 监听 Redis 任务队列。
    - **AI 推理执行**: 调用 LLM 接口，执行提示词模板填充与结果解析。
    - **MCP 工具执行**: 承载具体的 MCP Tool 调用逻辑，支持与外部工具链的深度交互。
    - **RAG 2.0 (Agentic RAG)**: 结合向量数据库进行知识库检索，支持自我反思与意图补全。
    - **GraphRAG**: 利用知识图谱进行跨文档关系推理。
    - 运行插件 (Plugins)。
    - 将处理结果返回给 BotNexus 或直接发送。
- **技术栈**: Go, Python, .NET (多语言支持)。

### 3. MCP Server (能力提供方)
MCP Server 是系统的“功能插件”标准。
- **职责**:
    - 暴露标准化的 Resources, Tools 和 Prompts。
    - 与 BotNexus 通过 SSE 或 STDIO 进行通信。
    - 允许第三方开发者以任何语言编写功能，并无缝接入 BotMatrix 生态。

### 4. Redis (中间件)
Redis 在系统中扮演着至关重要的角色，作为核心通信总线。
- **职责**:
    - **消息分发**: 使用 Pub/Sub 机制实现 Nexus 与 Worker 之间的实时通信。
    - **任务队列**: 存储待处理的异步任务。
    - **状态存储**: 存储机器人在线状态、限流策略和动态配置。
    - **会话缓存**: 维护用户上下文 (Session Context)。

### 5. PostgreSQL/SQLite (持久化数据库)
- **职责**:
    - 存储用户数据、路由规则、持久化配置和操作日志。
    - 存储复杂业务逻辑数据（如宝宝系统、结婚系统数据）。

---

## 🎯 路由与分发逻辑 (Routing Rules)

BotNexus 提供智能消息路由功能，确保消息精准送达对应的执行节点。

### 1. 路由优先级
1. **精确匹配 (Exact Match)**：检查 `user_ID`, `group_ID` 或 `bot_ID` 的直接对应关系。
2. **通配符匹配 (Wildcard Match)**：支持 `*` 通配符（如 `*_test` 或 `123*`）。
3. **智能负载均衡 (RTT-based LB)**：无匹配规则时，根据 Worker 的平均响应时间 (AvgRTT) 选择最优节点。
4. **故障回退 (Fallback)**：若指定 Worker 离线，自动回退到负载均衡模式。

### 2. 消息流向流程
1.  **接收**: 外部机器人客户端 (Client) 通过 WebSocket 将消息发送到 **BotNexus**。
2.  **决策**: BotNexus 根据 `RoutingRules` 匹配目标 Worker。
3.  **分发**: BotNexus 将消息发布到 **Redis** 的指定频道。
4.  **执行**: 订阅了该频道的 **BotWorker** 接收消息并运行相应的插件逻辑。
5.  **反馈**: BotWorker 处理完成后，将响应指令回传给 BotNexus 或直接调用 API 接口。

---

## 🤖 数字员工任务引擎 (Task Engine)

任务引擎负责管理数字员工从任务接收到结果反馈的全生命周期。

### 1. 核心概念
- **Task (任务)**: 工作的最小单元，包含 ExecutionID、Status、Plan 和 Results。
- **Task Plan (计划)**: AI 将复杂描述拆解为多个有序步骤，包含 Tool 调用和审批标志。
- **HITL (人工干预)**: 关键步骤暂停并请求人工审批，确保安全性。

### 2. 任务状态机
任务通过以下状态流转：`Pending` -> `Planning` -> `Executing` -> `Validating` -> `Completed/Failed`。在执行高风险步骤时，会进入 `PendingApproval` 状态。

---

## 📈 扩展性设计

- **水平扩展**: 可以启动多个 BotWorker 节点来分担负载。
- **高可用**: BotNexus 支持集群部署（需配合负载均衡器）。
- **插件化**: 支持动态加载插件，无需停机即可更新业务逻辑。
- **多语言生态**: 支持 Go, Python, .NET 等多种语言编写插件。
- **跨平台适配**: 通过统一的 OneBot 兼容层，适配 QQ、微信、钉钉等多个主流平台。
