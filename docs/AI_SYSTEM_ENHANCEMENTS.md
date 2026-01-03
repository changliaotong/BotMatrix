# AI 系统增强功能文档

本文档记录了 BotMatrix AI 系统的最新增强功能，包括自主 Agent 循环、上下文管理、可观察性追踪以及基于 MCP 的工具集成。

## 1. 自主 Agent 循环 (ChatAgent)

### 功能描述
`ChatAgent` 是一个更高级的对话接口，它不仅调用 LLM，还能自动处理模型返回的工具调用指令（Tool Calls）。

### 核心特性
- **自动循环**：当模型要求调用工具时，系统会自动执行工具并将结果反馈给模型，直到模型给出最终答复。
- **最大迭代限制**：默认为 10 次，防止无限循环和 Token 浪费。
- **统一工具执行**：无缝支持普通的 Skill 系统和基于 MCP (Model Context Protocol) 的工具。

### 代码参考
- [ai_service.go](file:///D:/projects/BotMatrix/src/BotNexus/internal/app/ai_service.go#L286) 中的 `ChatAgent` 方法。

---

## 2. 上下文管理 (ContextManager)

### 功能描述
为了应对 LLM 上下文窗口限制，系统引入了 `ContextManager` 来自动计算 Token 并修剪过长的对话历史。

### 核心特性
- **Token 估算**：基于字符长度的快速估算模型（约 1 token ≈ 4 字符）。
- **智能修剪**：
    - 始终保留第一条 `system` 消息（包含核心设定）。
    - 采用“滑动窗口”策略，保留最近的消息。
    - 自动丢弃中间过旧的消息以适应 `MaxTokens` 限制。
- **自动摘要（规划中）**：支持将修剪掉的消息生成简短摘要以保留长期记忆。

### 代码参考
- [context_manager.go](file:///D:/projects/BotMatrix/src/Common/ai/context_manager.go)

---

## 3. 可观察性与追踪 (AIAgentTrace)

### 功能描述
为了方便调试和监控 Agent 的行为，系统增加了详细的执行轨迹记录。

### 核心特性
- **全生命周期记录**：记录 `llm_response`（模型响应）、`tool_call`（工具请求）和 `tool_result`（工具执行结果）。
- **异步持久化**：追踪日志通过协程异步写入 PostgreSQL，不影响主对话流程。
- **Session 绑定**：每个 Agent 任务分配唯一的 `SessionID`，方便回溯整个推理链条。

### 代码参考
- [gorm_models.go](file:///D:/projects/BotMatrix/src/Common/models/gorm_models.go) 中的 `AIAgentTraceGORM` 模型。
- [ai_service.go](file:///D:/projects/BotMatrix/src/BotNexus/internal/app/ai_service.go) 中的 `saveTrace` 方法。

---

## 4. 数据库与向量数据库深度集成

### 功能描述
系统现在充分利用了 PostgreSQL 和 pgvector 进行认知记忆和知识库的存储与检索。

### 核心特性
- **向量检索适配器**：`MemoryMCPHost` 现在通过 `CognitiveMemoryService` 直接与 pgvector 交互，实现语义记忆搜索。
- **RAG 增强**：`KnowledgeMCPHost` 使用 `PostgresKnowledgeBase` 进行分片知识检索。
- **统一管理**：所有 MCP 适配器均通过 `MCPManager` 进行初始化和依赖注入。

### 代码参考
- [mcp_memory_adapter.go](file:///D:/projects/BotMatrix/src/BotNexus/internal/app/mcp_memory_adapter.go)
- [mcp_knowledge_adapter.go](file:///D:/projects/BotMatrix/src/BotNexus/internal/app/mcp_knowledge_adapter.go)

---

## 5. 快速开始

### 接口调用
```go
// 使用 AIService 发起自主对话
resp, err := aiService.ChatAgent(ctx, modelID, messages, tools)
```

### 查看追踪日志
```sql
SELECT * FROM ai_agent_traces WHERE session_id = 'your_session_id' ORDER BY step ASC;
```
