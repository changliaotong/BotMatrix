# MCP 与数据库/向量数据库集成指南

> [🌐 English](../../en-US/core/MCP_DATABASE_INTEGRATION_GUIDE.md) | [简体中文](MCP_DATABASE_INTEGRATION_GUIDE.md)
> [⬅️ 返回文档中心](../README.md) | [🏠 返回项目主页](../../../README.md)

本文档介绍了 BotMatrix 如何通过 MCP (Model Context Protocol) 深度集成 PostgreSQL 和 pgvector，为 AI 提供持久化记忆和高性能语义搜索能力。

## 🏗️ 架构背景

在 BotMatrix 的 "Agent OS" 愿景中，AI 不应直接操作底层数据库，而是通过标准化的工具接口进行交互。MCP 作为核心的总线协议，将底层的数据能力封装为 AI 可感知的工具。

### 为什么使用 MCP 而不是直接调用？
1. **解耦**: AI 逻辑与存储实现分离，更换数据库无需修改 AI 核心。
2. **安全**: 在 MCP 层级统一进行 PII 脱敏和权限控制。
3. **互操作性**: 一套工具定义可供多个模型（GPT, Claude, 豆包等）共同使用。

## 🧠 认知记忆集成 (Memory MCP)

`MemoryMCPHost` 现在已经完全接入了基于 PostgreSQL 的长期记忆系统。

### 核心改进
- **持久化**: 记忆不再存储在内存中，而是通过 [CognitiveMemoryService](file:///d:/projects/BotMatrix/src/BotNexus/internal/app/cognitive_memory.go) 存储在数据库。
- **向量化**: 利用 **pgvector** 插件，每条记忆在存储时都会生成 Embedding。
- **语义检索**: `search_memory` 工具使用向量相似度计算 (`<=>` 操作符) 查找相关记忆，而非简单的关键词匹配。

### 提供的工具
- `store_memory`: 存储重要信息，支持设置分类和重要程度。
- `search_memory`: 基于语义搜索历史记忆。

## 📚 知识库集成 (Knowledge MCP)

`KnowledgeMCPHost` 已经从简单的本地文件搜索升级为基于 RAG (Retrieval-Augmented Generation) 的语义搜索。

### 核心改进
- **混合搜索 (Hybrid Search)**: 结合了 pgvector 向量检索（语义相关）和全文索引（关键词匹配）。
- **RAG 2.0**: 集成了查询重写 (Query Refinement) 功能，在检索前自动优化用户的提问。
- **多维度过滤**: 支持按 BotID、UserID 和群组权限对知识进行精细化隔离。

### 提供的工具
- `search_knowledge`: 搜索技术文档、架构详情和项目信息。

## ⚙️ 集成与配置

在 [BotNexus](file:///d:/projects/BotMatrix/src/BotNexus/internal/app/main.go) 中，系统会自动完成以下初始化逻辑：

1. **向量服务初始化**: 查找配置中指定的 Embedding 模型（如 `doubao-embedding`）。
2. **知识库准备**: 初始化 `PostgresKnowledgeBase` 并执行必要的数据库迁移。
3. **依赖注入**:
   - 将向量服务注入 `CognitiveMemoryService`。
   - 将知识库实例注入 `MCPManager` 中的 `KnowledgeMCPHost`。

```go
// 示例：在 main.go 中的注入逻辑
es := rag.NewTaskAIEmbeddingService(m.AIIntegrationService, embedModel.ID, embedModel.ModelID)
kb := rag.NewPostgresKnowledgeBase(m.GORMDB, es, m.AIIntegrationService, chatModel.ID)

if aiSvc, ok := m.AIIntegrationService.(*AIServiceImpl); ok {
    aiSvc.SetKnowledgeBase(kb) // 注入到 MCP 管理器
}
```

## 🚀 开发者建议

1. **提示词优化**: 在调用 `store_memory` 时，建议 AI 提取核心实体和事实，避免存储无意义的语气词。
2. **知识库维护**: 建议通过管理后台定期上传最新的技术文档到 `knowledge_docs` 表，系统会自动完成向量化切片。
3. **监控**: 可以通过 `ai_usage_logs` 表监控向量化服务的 Token 消耗情况。
