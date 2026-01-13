# 🧠 BotMatrix AI, RAG & MCP 核心指南

> [🌐 English](../en-US/AI_RAG_MCP_GUIDE.md) | [简体中文](AI_RAG_MCP_GUIDE.md)
> [⬅️ 返回文档中心](README.md) | [🏠 返回项目主页](../../README.md)

本指南详细介绍了 BotMatrix 的 AI 能力、RAG (检索增强生成) 架构以及 MCP (Model Context Protocol) 接入层。

---

## 🔌 1. MCP 接入层 (Model Context Protocol)

MCP 是 BotMatrix 的“驱动程序接口”，将能力提供方与模型使用方解耦。

### 1.1 核心支柱
- **Resources (数据)**: 允许模型读取静态或动态数据（如历史记录、数据库报表）。
- **Tools (函数)**: 允许模型执行动作（如发送消息、调用 API、执行脚本）。
- **Prompts (模板)**: 提供预定义的提示词模板（如专家人格、任务拆解流）。

### 1.2 双栈 (Dual-Stack) 架构
BotMatrix 采用“适配器模式 + MCP 模式”并行架构：
- **适配器模式**: 处理 OneBot, Discord 等原生协议。
- **MCP 模式**: 将适配器能力封装为 MCP 工具，供 AI 按需调用。

### 1.3 全球智能体网络 (Global Agent Mesh)
每一个 BotMatrix 节点都是一个标准的 MCP Host/Server，实现跨企业、跨平台的智能体协作。
- **联邦身份认证**: 基于 PKI 的 OrgID 和 JWT 鉴权。
- **任务接力**: 当本地无法处理意图时，安全地调用远程 MCP 工具。

---

## 📚 2. RAG 2.0 (检索增强生成)

RAG 机制使机器人具备“自举”能力，能够基于系统文档和外部知识库进行精准回答。

### 2.1 技术选型
- **向量数据库**: PostgreSQL + **pgvector**。
- **向量化模型**: **BGE-M3** (1024 维)，通过 Ollama 本地部署。
- **知识流**: Indexer (文档分片) -> Storage (向量存储) -> Retriever (语义检索)。

### 2.2 机器人自举 (Bootstrap) 机制
机器人通过内置的身份清单和能力描述建立自我认知：
- **BotIdentity**: 定义名称、角色、性格。
- **SystemManifest**: 聚合所有已注册的技能和动作。
- **RAG Enhancement**: 挂载深度知识库，提供“如何使用功能”的指导。

---

## 🛡️ 3. 安全与隐私 (Privacy Bastion)

在 AI 协作和 RAG 检索过程中，安全是首要考虑：
- **PII 脱敏**: 自动识别并脱敏姓名、手机号等敏感信息。
- **权限隔离**: 远程 MCP 工具调用在受限沙箱中运行。
- **操作审计**: 所有 AI 执行步骤及其原始结果永久留存。

---

## 🗺️ 4. 演进路线图

1. **已完成**: MCP 基础架构、动态注册、权限隔离、SSE 实时协议。
2. **已完成**: 基于 B2B 联邦身份的跨域调用、联邦搜索、隐私堡垒。
3. **未来**: 语义向量检索增强的认知记忆、BotMatrix Hub 动态组网。
