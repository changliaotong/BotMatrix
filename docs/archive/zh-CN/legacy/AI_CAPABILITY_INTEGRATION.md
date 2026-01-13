# 🧠 BotMatrix AI 能力集成与核心架构 (AI Capability & Integration)

> [⬅️ 返回文档中心](../README.md) | [🏠 返回项目主页](../../README.md)

## 1. 概述与核心目标 (Overview & Goals)

BotMatrix 采用 **AI 原生 (AI-Native)** 架构，将传统的机器人指令系统进化为具备语义理解和自主决策能力的智能体 (Agent) 系统。通过集成 **RAG** (检索增强生成) 与 **MCP** (模型上下文协议)，构建起智能体的“大脑”、“记忆”与“手脚”。

### 核心目标
- **供应商中立**: 支持 OpenAI, DeepSeek, Claude, Google Gemini, 阿里通义千问及本地 Ollama 等多种模型调度。
- **自举认知 (Bootstrap)**: 机器人具备自我意识，了解自身功能与边界。
- **能力标准化 (MCP)**: 类似于 USB 标准，将工具能力与模型逻辑彻底解耦。
- **智能记忆 (RAG)**: 基于向量数据库实现长期记忆与海量知识检索。
- **高可靠性**: 具备自动降级、故障切换及全链路任务追踪能力。

---

## 2. 核心技术架构 (Technical Architecture)

系统借鉴了 C# `Semantic Kernel` 的设计思想，在 Go 核心中实现了高度解耦的 AI 调度引擎，并集成了 RAG 与 MCP 两大核心组件。

### 2.1 统一 AI 抽象层 (`Common/ai`)
为了屏蔽不同供应商的 API 差异，我们定义了标准接口：
- **AIService 接口**: 涵盖了 `Chat` (对话)、`Embed` (向量化)、`GetAvailableModels` (模型发现) 等核心方法。
- **通用适配器**: `OpenAIAdapter` 支持所有兼容 OpenAI 格式的 API，包括流式响应 (SSE) 和函数调用 (Function Calling)。

### 2.2 核心总线与存储
- **MCP 总线**: 作为系统的“神经网络接口”，负责连接 Tools (可执行动作)、Resources (只读数据) 和 Prompts (思维模板)。
- **向量存储 (pgvector)**: 存储长期记忆与 RAG 知识库，支持 `L2 distance` 和 `Cosine distance` 搜索。

---

## 3. 智能意图分发系统 (Intelligent Intent Dispatch)

系统充当“分诊台”角色，在消息进入核心逻辑前进行语义识别与精确路由。

### 3.1 双层调度架构
1.  **系统级调度 (System-Level)**: 由 BotNexus 决定由哪个 Worker 或哪种 AI 技能处理。
2.  **用户级调度 (User-Level)**: 在群组协作中，由“调度者 Agent”识别意图并指派给特定“执行者 Agent”。

### 3.2 工作流程
- **语义判定**: 构造特定的 `System Prompt`，引导模型返回结构化的意图 Code。
- **动态路由**: 根据 `IntentID` 查找映射关系，转发至对应的 `skill` (AI 技能)、`worker` (后端节点) 或 `plugin` (传统插件)。

---

## 4. RAG: 检索增强生成与自举 (RAG & Bootstrap)

RAG 为 AI 提供了“第二大脑”，使其能够基于实时更新的文档库回答问题，并实现机器人的“自举”认知。

### 4.1 核心功能 (Nexus RAG)
- **多格式支持**: 自动解析并索引 PDF (含 OCR)、Word、Markdown、Excel、代码及图片。
- **空间隔离**: 区分“个人私有空间”与“群组共享空间”，确保数据安全。
- **Agentic RAG (RAG 2.0)**:
    - **查询改写**: AI 自动优化搜索关键词。
    - **自我反思**: 二次评估检索质量，剔除无关干扰。
    - **知识图谱漫游**: 建立跨文档的实体关联。

### 4.2 机器人自举 (Bootstrap) 机制
自举是指机器人通过内置的身份清单和能力描述，逐步建立起对自身功能的认知过程：
1. **静态层**: 配置文件定义的角色与性格。
2. **动态层**: 实时汇报的 Skills 清单，告知 AI 当前可调用的工具。
3. **知识层**: 挂载的 RAG 知识库，提供“如何使用 [功能]”的深度文档。

---

## 5. MCP: 模型上下文协议与集成 (MCP Integration)

MCP 是 BotMatrix 的核心总线协议，实现了“能力提供方”与“模型使用方”的彻底解耦。

### 5.1 MCP 的三大支柱
- **Resources (只读上下文)**: 按需读取历史对话、用户笔记、实时报表。
- **Tools (可执行动作)**: 执行代码、调用 API、操作桌面应用。
- **Prompts (预定义思维)**: 专家人格切换、任务拆解模板。

### 5.2 接入与双栈模式
- **双栈架构**: 传统“适配器模式”负责基础通信，现代“MCP 模式”将能力封装为工具供 AI 调用。
- **Universal MCP Host**: 支持 **STDIO**, **HTTP** 和 **SSE** 协议，可接入外部任何语言编写的 MCP Server。

---

## 6. 多智能体协作与任务追踪 (Multi-Agent Collaboration)

### 6.1 协作网桥 (Collaboration Bridge)
作为 MCP 服务运行，为 AI 提供以下工具：
- `colleague_list`: 获取可用同事列表及职责。
- `colleague_consult`: 同步咨询，挂载推理等待答复。
- `task_delegate`: 异步委派任务，返回执行 ID。

### 6.2 任务追踪与全链路监控
- **TraceID**: 关联原始请求，支持跨 Session、跨 Agent 的纵向链路分析。
- **状态管理**: 利用 `tasks.Execution` 模型实时记录 `running`, `success`, `failed` 状态。

---

## 7. 数字员工与工具规范 (Digital Employee & Tooling)

为了解决 AI “幻觉”导致的高风险操作问题，建立了完整的工具分级体系。

### 7.1 风险等级与审计
- **Low (低风险)**: 只读操作，自动执行。
- **Medium (中风险)**: 发送消息等有限副作用操作，自动执行并审计。
- **High (高风险)**: 修改数据库等关键操作，**必须经过人工审批 (HITL)**。

### 7.2 性能与 KPI
- 系统自动生成数字员工的 KPI 考评（响应速度、准确率、Token 成本），实现 AI 能力的量化管理。

---

## 8. 进阶核心能力 (Advanced Capabilities)

### 8.1 长期记忆与隐私堡垒
- **长期记忆**: 利用 pgvector 实现海量消息的语义检索，并自动生成会话摘要。
- **隐私堡垒 (PII Masking)**: 云端交互前自动脱敏敏感信息，本地接收后自动还原。

### 8.2 智能交互能力
- **AI Tasking**: 自然语言生成 Cron 任务。
- **AI Policy**: 自然语言调整系统维护策略。
- **AI Tagging**: 批量处理与打标。

---

## 9. 开发者指南 (Developer Guide)

### 9.1 如何接入新能力
1. **编写标准 MCP Server**: 使用 Python/JS/Go 等任何语言实现。
2. **连接至 BotMatrix**: 在后台配置 MCP Server 节点。
3. **能力发现**: 系统自动将其包装为全局可用的 Skill。

### 9.2 数据模型定义
- `ai_providers` & `ai_models`: 提供商与模型参数。
- `ai_agents`: 智能体定义（Prompt、工具集）。
- `ai_intents` & `ai_intent_routings`: 意图识别与路由规则。
- `KnowledgeDoc` & `KnowledgeChunk`: RAG 知识库分片模型。

---

## 10. 未来规划 (Roadmap)
- [ ] **多模型自动编排**: 根据任务复杂度自动选择性价比最高的模型（如 GPT-4 vs DeepSeek）。
- [ ] **自愈式运维**: AI 自动诊断系统日志并给出修复建议。
- [ ] **边缘 AI 策略**: 在端侧运行小参数模型处理敏感任务。
- [ ] **任务编排可视化**: 提供任务委派的拓撲圖與甘特圖展示。

---
*Last Updated: 2026-01-13*
