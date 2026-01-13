# 🏗️ BotMatrix 核心架构全景 (Architecture)

> [⬅️ 返回文档中心](README.md) | [🏠 返回项目主页](../../README.md)

BotMatrix 是一个分布式、AI 原生且高度解耦的机器人矩阵管理系统。它不仅是一个消息转发工具，更是一个由“数字员工”驱动的协同进化生态系统。

---

## 1. 核心设计哲学 (Core Philosophy)

- **分布式解耦**: 采用中心控制 (Nexus) 与边缘执行 (Worker) 分离的架构，通过 Redis 异步总线通信。
- **AI 原生 (AI-Native)**: 将 AI 视为系统的一等公民，所有交互均可由 AI 技能 (Skills) 承载。
- **去插件化 (System Modules)**: 面向用户隐藏“插件”概念，所有功能抽象为“系统模块”，共同构成一个统一的数字世界。
- **服务端无头代理 (Headless Agent)**: 坚持服务端执行逻辑，而非依赖 IDE 插件，实现真正的自动化（Human-on-the-loop）。
- **完全自动化**: 核心目标是减少人工干预，在各个环节增加数字员工介入，实现从“编程助手”到“数字员工”的跨越。

---

## 2. 系统核心组件 (System Components)

### 2.1 BotNexus (中心控制节点)
系统的“大脑”与“消息网关”。
- **职责**: 管理连接、意图识别、路由决策、MCP 调度、B2B 协作 (Mesh)。
- **Identity Manager**: 管理企业级公私钥对，负责跨企业调用的 JWT 签名与验证。
- **技术栈**: Go, Gin, WebSocket, Redis.

### 2.2 BotWorker (任务执行节点)
系统的“四肢”与“逻辑容器”。
- **职责**: 监听任务队列、执行插件逻辑、运行 AI 推理、承载 MCP 工具、管理 RAG 知识库。
- **技术栈**: Go, Python, .NET (多语言支持)。

### 2.3 Event Nexus (事件中枢)
系统的“神经网络”，基于发布/订阅模式实现跨模块联动。
- **零耦合扩展**: 模块间通过事件通信，互不依赖。开发新功能只需订阅相关事件，无需修改核心代码。
- **全自动进化**: 所有的经济行为（签到、转账、游戏输赢）都会产生事件，驱动等级与成就系统。

### 2.4 MCP Server (能力提供方)
遵循 Model Context Protocol 的标准能力单元，为 AI 提供 Tools, Resources 和 Prompts。

---

## 3. 消息流转与智能路由 (Message Flow & Routing)

### 3.1 消息流向流程
1. **接收**: 外部客户端通过 WebSocket 发送至 BotNexus。
2. **裁决**: 经过 `CorePlugin` 进行权限与敏感词过滤。
3. **路由**: 根据 `RoutingRules` 决定目标 Worker。
4. **执行**: Worker 从 Redis 订阅任务，处理并反馈结果。

### 3.2 智能路由规则 (Routing Rules)
- **精确匹配 (Exact Match)**: 检查 `user_id`, `group_id` 或 `bot_id` 的直接对应关系。
- **通配符匹配 (Wildcard Match)**: 支持 `*` 通配符（如 `*_test`）。
- **智能负载均衡 (RTT-based LB)**: 无匹配规则时，根据 Worker 的平均响应时间 (AvgRTT) 和健康状态选择最优节点。
- **故障回退 (Fallback)**: 若指定 Worker 离线，自动回退到智能负载均衡，确保消息不丢失。

---

## 4. 任务引擎与 AI 服务层 (Task Engine & AI)

### 4.1 数字员工系统 (Digital Employee System)
数字员工是 BotMatrix 对 AI 机器人的高级拟人化封装。它不仅是一个机器人实例，而是一个拥有**工号、职位、部门、人设、技能集以及 KPI 考核**的虚拟雇员。

- **核心架构 (五感六觉)**:
    - **身份 (Identity)**: `IdentityGORM` / `BotID` 定义工号、职位与权限。
    - **感知 (Perception)**: `Intent Dispatcher` 接收并解析语义意图。
    - **思维 (Cognition)**: 基于 LLM 的推理、规划与决策中心。
    - **记忆 (Memory)**: **短期**: 会话上下文；**长期**: `Cognitive Memory` 事实片段。
    - **技能 (Skills)**: `MCP Toolset` 调用的具体工具集。
    - **协作 (Social)**: `Agent Mesh` 实现跨员工/跨企业任务委派。
    - **进化 (Evolution)**: 从工作中提取知识，自我纠错与能力提升。

- **绩效与 KPI 考核**: 系统根据任务执行结果（完成率、执行效率、自主度、Token 成本）自动计算绩效，实现 AI 能力的量化管理。

### 4.2 任务引擎 (Task Engine)
管理任务的全生命周期：`Pending` -> `Planning` -> `Executing` -> `Completed`。
- **任务拆解**: AI 将复杂指令转化为步骤化的计划 (Task Plan)。
- **核心模式**:
    - **调度者-执行者 (Orchestrator-Workers)**: 统筹分发与质量审查。
    - **流水线与断路器**: `需求分析 -> 资料检索 -> 初稿 -> 审查 -> 交付`，错误超限自动熔断并请求 **HITL (人工干预)**。
    - **状态机管理**: 持久化中间变量，支持长时任务的崩溃恢复。

### 4.3 增强型 AI 服务层 (AI Capability Integration)
- **RAG 2.0 (Agentic RAG)**:
    - **Nexus RAG**: 支持 PDF, Word, Markdown 等多格式索引。
    - **高级特性**: 查询改写、自我反思、知识图谱漫游。
- **MCP 集成**: 系统的“神经网络接口”，连接 Tools (可执行动作)、Resources (只读数据) 和 Prompts (思维模板)。
- **自举认知 (Bootstrap)**: 机器人通过 Skills 清单和 RAG 知识库，逐步建立对自身功能的认知。
- **隐私堡垒 (Privacy Bastion)**: PII 脱敏技术，外发消息自动脱敏与还原。

---

## 5. 全球智能体网格 (Global Agent Mesh)

实现跨企业的 B2B 协作，构建“智能体即服务 (Agent as a Service)”的网络。
- **联邦身份认证**: 基于公钥基础设施 (PKI) 的组织 ID (OrgID) 与跨域 JWT 授权。
- **动态发现机制**: 支持在联邦网络中实时并发搜索可用的 MCP 服务。
- **任务接力 (Task Relay)**: 跨域调用 (B2B Call)，当本地机器人无法处理意图时，安全地委派给远程节点。
- **典型场景**: 跨国企业库存查询、私人助理集群协作。

---

## 6. 进化系统与数值设计 (Evolution & Numerical)

### 5.1 数值公式 (The Singularity Plan)
采用分段多项式确保长期挑战性：`ExpForLevel(L) = 50 * L^2 + 150 * L - 200`。
- **经验来源**: 积分获取（Amount * 0.8）、积分消费（Amount * 1.2）、社交活跃（10 EXP/msg）。
- **位面梯度**: 
  - [原质] (1-9) -> [构件] (10-29) -> [逻辑] (30-59) -> [协议] (60-89) -> [矩阵] (90-119) -> [奇点] (120+)。

### 5.2 勋章与成就
系统识别里程碑（如首笔积分、达到特定等级），自动授予勋章并实时反馈在界面中。

---

## 7. 系统监控与审计 (Monitoring & Audit)

### 7.1 系统脉动 (System Pulse)
基于环形日志队列 (Ring Buffer) 的实时审计系统。
- **审计分级**: Success ✅, Info ℹ️, Warning ⚠️, Critical 🚨。
- **特性**: 并发安全，只保留最近 50 条记录，低开销，不持久化。

---

## 8. 系统术语表 (Glossary)

为了保持项目的一致性与独特的文化氛围，系统采用了一套具有“赛博朋克”风格的术语体系。

- **矩阵 (Matrix)**: 指代整个 BotMatrix 机器人系统及其运行环境。
- **奇点 (Singularity)**: 系统进化的终极目标，代表 AI 与人类协作的完美状态。
- **位面 (Plane)**: 用户等级的分组体系（如 [原质]、[构件]、[逻辑] 等）。
- **事件中枢 (Event Nexus)**: 进程内异步事件总线，实现插件间的零耦合联动。
- **系统脉动 (System Pulse)**: 实时审计日志监控系统。
- **进化服务 (Evolution Service)**: 等级与成就系统，处理经验值计算。
- **积分 (Points)**: 矩阵内的基础货币，通过复式记账法管理。

---

## 9. 开发者与维护建议

- **禁止直接提及“插件”**: 面向用户统一使用“系统模块”或“逻辑”术语。
- **Skill 导向**: 模块间通讯优先使用 `CallSkillAsync` 而非直接代码引用。
- **强制审计**: 关键业务逻辑必须发布 `SystemAuditEvent` 到 `EventNexus`。
- **数据库迁移**: 所有 GORM 模型需在 `AutoMigrate` 中注册。
- **安全拦截**: 敏感操作必须在 SDK 端和 Core 端进行双重校验。

---
*最后更新日期：2026-01-13*
