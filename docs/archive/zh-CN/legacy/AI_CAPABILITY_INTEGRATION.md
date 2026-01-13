# BotMatrix AI 能力集成技术文档

## 1. 概述 (Overview)

BotMatrix 采用 **AI 原生 (AI-Native)** 架构，将传统的机器人指令系统进化为具备语义理解和自主决策能力的智能体 (Agent) 系统。本项目借鉴了 C# `Semantic Kernel` 的设计思想，在 Go 核心中实现了高度解耦、供应商中立的 AI 调度引擎。

## 2. 核心架构 (Core Architecture)

### 2.1 统一 AI 抽象层 (Common/ai)
为了支持多种模型供应商，我们在 `src/Common/ai` 定义了统一的接口协议：
- **`types.go`**: 定义了 `Message` (对话消息)、`Tool` (函数定义)、`ChatRequest` 等标准结构，完全兼容 OpenAI Chat Completion 协议。
- **`OpenAIAdapter`**: 核心适配器，支持所有兼容 OpenAI 格式的 API（如 DeepSeek, Ollama, Azure OpenAI, 阿里通义千问等）。支持流式响应 (SSE) 和函数调用 (Function Calling)。

### 2.2 枢纽服务 (BotNexus Services)
- **`AIIntegrationService`**: AI 调度中心，负责模型选择、Provider 实例化以及意图分发。
- **`DigitalEmployeeService`**: 数字员工管理，负责关联机器人 ID 与企业职能人设 (Persona)，并记录 KPI 考评数据。

## 3. 核心功能实现 (Key Features)

### 3.1 意图识别进化 (Intent Recognition)
系统支持双轨制意图识别：
1.  **正则匹配 (Regex)**: 适用于高频、固定格式的简单指令（通过 `AIParser.MatchSkillByRegex` 实现）。
2.  **语义理解 (LLM)**: 适用于自然语言描述的复杂任务。
    - **自动工具化**: `BotWorker` 报备的技能 (`Capability`) 会被自动转换为 LLM 的 `Tool` 定义。
    - **函数调用**: AI 通过 Function Calling 协议返回解析后的技能名及结构化参数。

### 3.2 技能中心集成 (Skill Center)
技能清单 (`SystemManifest`) 动态生成：
- **核心动作**: 如 `send_message`, `mute_group`。
- **业务技能**: 由不同 Worker 节点热插拔报备。
- **提示词工程**: 系统会自动生成包含所有可用技能描述的 `System Prompt`，引导模型进行精准识别。

## 4. 数据库模型 (Data Models)

基于 GORM 维护以下核心 AI 配置：
- **`ai_providers`**: 存储 API Key（加密）、BaseURL 和提供商类型。
- **`ai_models`**: 存储具体模型参数（上下文窗口、功能支持、默认标志）。
- **`ai_agents`**: 智能体核心定义（Prompt、模型、工具集、温度等）。
- **`ai_sessions`**: 对话会话管理，支持跨平台上下文关联。
- **`ai_chat_messages`**: 详尽的对话历史，包含 Role、Content、Token 消耗及 Tool 调用记录。
- **`digital_employees`**: 存储智能体人设、职能、归属 BotID 及其 KPI 统计。

## 5. 开发者指南 (Developer Guide)

### 5.1 如何接入新模型
1.  在数据库 `ai_providers` 表中添加配置。
2.  在 `ai_models` 表中关联该提供商的模型。
3.  通过 `AIIntegrationService.Chat` 接口进行调用。

### 5.2 技能报备流程
Worker 节点通过 WebSocket 或 Redis 发送 `Capability` 报备：
```json
{
  "name": "weather_query",
  "description": "查询指定城市的天气",
  "params": {
    "city": "城市名"
  },
  "required": ["city"]
}
```
BotNexus 接收后会自动将其转化为 LLM 可感知的工具。

### 5.3 数据迁移
如果你需要从旧版 `sz84` (SQL Server) 迁移数据，请参考：
- [AI 数据迁移指南](file:///d:/projects/BotMatrix/docs/zh-CN/core/AI_DATA_MIGRATION.md)

## 6. 未来规划 (Roadmap)
- **RAG 增强**: 集成向量数据库，支持基于私有知识库的对话。
- **Agent 记忆**: 引入会话上下文持久化，支持跨平台、长周期的任务跟踪。
- **多模型编排**: 实现自动路由，根据任务复杂度自动选择性价比最高的模型。

---
*文档更新日期: 2025-12-31*
