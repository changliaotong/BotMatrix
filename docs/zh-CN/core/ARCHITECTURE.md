# 🏗️ BotMatrix 系统架构概览

> [🌐 English](../en-US/ARCHITECTURE.md) | [简体中文](ARCHITECTURE.md)
> [⬅️ 返回文档中心](../README.md) | [🏠 返回项目主页](../../README.md)

BotMatrix 是一个采用分布式、解耦设计的机器人矩阵管理系统。它通过核心的消息分发中心与多个执行节点协作，实现了高并发和高可扩展性。

## 🏗️ 核心组件详解

### 1. BotNexus (中心控制节点)
BotNexus 是整个系统的“大脑”和“路由器”。
- **职责**:
    - 维护与客户端（如 WxBot, QQBot）的 WebSocket 连接。
    - **3D 拓扑可视化**: 基于 Three.js 的实时宇宙拓扑，支持节点聚类与全路径追踪。
    - **智能路由分发**: RTT 感知的动态路由算法，支持精确 ID 及通配符匹配。
    - **MCP Host**: 调度本地与远程的 MCP Server 工具集。
    - **AI 隐私堡垒**: 对全链路外发消息进行 PII 脱敏处理。
- **技术栈**: Go 1.20+, Gin, Redis Cluster, JWT, Docker SDK。

### 2. BotWorker (任务执行节点)
BotWorker 是实际处理业务逻辑的“四肢”。
- **职责**:
    - **插件系统**: 支持私聊、群聊消息处理，内置实用工具、成就、游戏、宠物、积分、社交及群管系统。
    - **AI 推理执行**: 调用 LLM 接口，执行提示词模板填充。
    - **RAG 2.0 / GraphRAG**: 结合向量数据库与知识图谱进行深度检索。
- **内置功能**:
    - **实用工具**: 天气、翻译、点歌、计算。
    - **社交娱乐**: 签到、抽奖、三公、宠物系统、坐骑系统。
    - **群管**: 撤回、禁言、敏感词过滤、黑白名单。

### 3. SystemWorker (集中智能单元)
SystemWorker 是机器人的“皮层”，负责全局编排与高阶监控。
- **实时看板**: `#sys status` 命令生成高清系统活力波形图与矩阵状态。
- **上帝模式**: `#sys exec` 允许管理员远程执行 Python 代码（需严格白名单）。
- **全频道广播**: `#sys broadcast` 一键推送公告至所有平台。

### 4. Overmind (移动端控制中心)
- **跨平台支持**: 基于 Flutter 开发，适配 Android/iOS/Web。
- **特性**: 具有“主脑”美学的科幻风 UI，提供 Nexus 仪表盘与实时日志控制台。

### 5. 中间件与数据库

- **Redis (核心通信总线)**:
    - **消息路由**: 使用 Pub/Sub 机制实现 Nexus 与 Worker 之间的实时通信。
    - **任务队列**: 使用 `List` (`BLPop/RPush`) 实现异步任务处理与削峰填谷。
    - **动态限流**: 支持 `user_id` 和 `group_id` 双重维度的滑动窗口限流。
    - **幂等去重**: 结合本地热缓存与 Redis 远程缓存，确保分布式环境下的消息唯一性。
    - **会话管理**: 维护用户上下文 (Context) 与中间状态 (State)，支持跨端同步。
    - **动态配置**: 路由规则、限流策略及 TTL 配置均存储于 Redis，支持热更新。
- **PostgreSQL**: 核心业务数据持久化，存储路由规则、用户数据及复杂业务状态。

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
