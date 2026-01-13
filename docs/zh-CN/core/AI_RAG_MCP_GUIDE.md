# 🧠 BotMatrix AI, RAG & MCP 核心指南 (AI System Core)

> **版本**: 2.5
> **状态**: 核心架构已实现，应用层扩展中
> [🌐 English](../en-US/AI_RAG_MCP_GUIDE.md) | [简体中文](AI__RAG_MCP_GUIDE.md)
> [⬅️ 返回文档中心](README.md) | [🏠 返回项目主页](../../README.md)

本指南详细介绍了 BotMatrix 的 AI 能力、RAG (检索增强生成) 架构、MCP (Model Context Protocol) 接入层，以及基于这些技术构建的**数字员工 (Digital Employee)** 系统。

---

## 🔌 1. MCP 接入层 (Model Context Protocol)

MCP 是 BotMatrix 的“驱动程序接口”，将能力提供方与模型使用方解耦。

### 1.1 核心支柱
- **Resources (数据)**: 允许模型读取静态或动态数据（如历史记录、数据库报表）。
- **Tools (函数)**: 允许模型执行动作（如发送消息、调用 API、执行脚本）。
- **Prompts (模板)**: 提供预定义的提示词模板（如专家人格、任务拆解流）。

### 1.2 数据库与向量集成 (Storage Integration)
系统通过 MCP 深度集成 PostgreSQL 和 pgvector，为 AI 提供持久化记忆。
- **持久化**: 记忆通过 `CognitiveMemoryService` 存储在数据库中，不再依赖内存。
- **向量化**: 利用 **pgvector** 插件，每条记忆/知识片段在存储时生成 Embedding（默认使用 BGE-M3 或 豆包-embedding）。
- **语义检索**: 使用向量相似度计算 (`<=>` 操作符) 实现毫秒级语义检索。

### 1.3 全球智能体网络 (Global Agent Mesh)
每一个 BotMatrix 节点都是一个标准的 MCP Host/Server：
- **联邦身份认证**: 基于 PKI 的 OrgID 和 JWT 鉴权。
- **跨域调用**: 当本地无法处理意图时，安全地代理调用远程企业的 MCP 工具。

---

## 📚 2. RAG 2.0 (检索增强生成)

RAG 机制使机器人具备“自举”能力，能够基于系统文档和外部知识库进行精准回答。

### 2.1 技术选型
- **向量数据库**: PostgreSQL + **pgvector**。
- **混合搜索 (Hybrid Search)**: 结合向量检索（语义）与全文索引（关键词）。
- **RAG 2.0 优化**: 引入查询重写 (Query Refinement)，在检索前自动优化用户提问。

### 2.2 机器人自举 (Bootstrap) 机制
机器人通过内置的身份清单和能力描述建立自我认知：
- **BotIdentity**: 定义名称、角色、性格。
- **SystemManifest**: 聚合所有已注册的技能和动作。
- **RAG Enhancement**: 挂载深度知识库，提供“如何使用功能”的指导。

---

## � 3. 数字员工系统 (Digital Employee System)

**数字员工** 是对上述 AI 能力的高级拟人化封装。它拥有工号、职位、人设、技能集以及 KPI 考核。

### 3.1 核心架构：“五感六觉”
| 维度 | 对应组件 | 功能描述 |
| :--- | :--- | :--- |
| **身份 (Identity)** | `IdentityGORM` | 工号、职位、部门、权限范围。 |
| **感知 (Perception)** | `Intent Dispatcher` | 识别来自 IM 或 API 的用户意图。 |
| **思维 (Cognition)** | `AI Service Layer` | 基于 LLM 的推理、规划与决策中心。 |
| **记忆 (Memory)** | `Cognitive Memory` | **短期**: 会话上下文；**长期**: 事实片段、用户偏好。 |
| **技能 (Skills)** | `MCP Toolset` | 能够调用的工具（数据库、API、跨企业服务）。 |
| **进化 (Evolution)** | `Auto-Learning` | 自我纠错与能力提升。 |

### 3.2 协作机制
- **同步咨询**: 实时问答协作。
- **异步委派**: 任务指派与结果汇报（支持 A2A 协作协议）。
- **跨域授权**: 使用双向 JWT 握手，跨企业身份受控。

---

## � 4. KPI 考核与进化体系

系统根据任务执行结果自动计算绩效，驱动 AI 自主优化。

### 4.1 核心考核维度
| 维度 | 计算逻辑 | 权重 |
| :--- | :--- | :--- |
| **完成率 (Success Rate)** | `成功任务数 / 总任务数` | 40% |
| **执行效率 (Efficiency)** | `平均步骤耗时 vs 模板基准` | 30% |
| **自主度 (Autonomy)** | `无人工干预执行数 / 总任务数` | 20% |
| **成本 (Cost)** | `消耗 Token 数 vs 任务价值` | 10% |

### 4.2 自动调优
KPI 分数低会触发“再培训”：系统分析最近失败的任务，基于 AI 自动生成更精准的 `Bio` (人设) 或提示词补丁。

---

## 🛡️ 5. 安全与隐私 (Privacy Bastion)

- **PII 脱敏**: 发送至 LLM 前自动屏蔽手机号、姓名等敏感字段。
- **沙箱运行**: 远程工具调用在受限环境中执行。
- **操作审计**: `AIAgentTrace` 记录每一次工具调用的全量参数与返回。

---

## ⚙️ 管理接口 (Admin API)

- `GET /api/admin/employees/tasks`: 获取员工任务队列及状态。
- `GET /api/admin/employees/kpi`: 获取绩效统计数据。
- `POST /api/admin/employees/optimize`: 触发 AI 驱动的自动优化。
- `POST /api/knowledge/upload`: 上传并向量化文档。
