# MCP and Database/Vector Database Integration Guide

> [简体中文](../../zh-CN/core/MCP_DATABASE_INTEGRATION_GUIDE.md) | [English](MCP_DATABASE_INTEGRATION_GUIDE.md)
> [⬅️ Back to Docs](../README.md) | [🏠 Back to Home](../../../README.md)

This document describes how BotMatrix deeply integrates PostgreSQL and pgvector via MCP (Model Context Protocol) to provide AI with persistent memory and high-performance semantic search capabilities.

## 🏗️ Architectural Background

In BotMatrix's "Agent OS" vision, AI should not directly manipulate the underlying database, but instead interact via standardized tool interfaces. MCP serves as the core bus protocol, encapsulating low-level data capabilities into AI-perceivable tools.

### Why use MCP instead of direct calls?
1. **Decoupling**: AI logic is separated from storage implementation; changing the database doesn't require modifying the AI core.
2. **Security**: Unified PII desensitization and access control at the MCP level.
3. **Interoperability**: A single set of tool definitions can be shared by multiple models (GPT, Claude, Doubao, etc.).

## 🧠 Cognitive Memory Integration (Memory MCP)

The `MemoryMCPHost` is now fully connected to the PostgreSQL-based long-term memory system.

### Key Improvements
- **Persistence**: Memories are no longer stored in-memory, but in the database via [CognitiveMemoryService](file:///d:/projects/BotMatrix/src/BotNexus/internal/app/cognitive_memory.go).
- **Vectorization**: Leveraging the **pgvector** extension, every memory generates an embedding upon storage.
- **Semantic Retrieval**: The `search_memory` tool uses vector similarity (`<=>` operator) to find relevant memories instead of simple keyword matching.

### Tools Provided
- `store_memory`: Store important information with category and importance level.
- `search_memory`: Semantically search through historical memories, returning a list with memory IDs.
- `forget_memory`: Remove a specific long-term memory using its ID, enabling the "forgetting" capability.

## 📚 Knowledge Base Integration (Knowledge MCP)

The `KnowledgeMCPHost` has been upgraded from simple local file search to semantic search based on RAG (Retrieval-Augmented Generation).

### Key Improvements
- **Hybrid Search**: Combines pgvector vector retrieval (semantic relevance) and full-text indexing (keyword matching).
- **RAG 2.0**: Integrated Query Refinement, which automatically optimizes user prompts before retrieval.
- **Multi-dimensional Filtering**: Supports fine-grained isolation of knowledge by BotID, UserID, and group permissions.

### Tools Provided
- `search_knowledge`: Search for technical documentation, architecture details, and project information.

## ⚙️ Integration and Configuration

In [BotNexus](file:///d:/projects/BotMatrix/src/BotNexus/internal/app/main.go), the system automatically handles the following initialization logic:

1. **Vector Service Initialization**: Finds the specified Embedding model (e.g., `doubao-embedding`).
2. **Knowledge Base Preparation**: Initializes `PostgresKnowledgeBase` and performs necessary database migrations.
3. **Dependency Injection**:
   - Injects the vector service into `CognitiveMemoryService`.
   - Injects the knowledge base instance into `KnowledgeMCPHost` within the `MCPManager`.

```go
// Example: Injection logic in main.go
es := rag.NewTaskAIEmbeddingService(m.AIIntegrationService, embedModel.ID, embedModel.ModelID)
kb := rag.NewPostgresKnowledgeBase(m.GORMDB, es, m.AIIntegrationService, chatModel.ID)

if aiSvc, ok := m.AIIntegrationService.(*AIServiceImpl); ok {
    aiSvc.SetKnowledgeBase(kb) // Inject into MCP Manager
}
```

## 🚀 Recommendations for Developers

1. **Prompt Optimization**: When calling `store_memory`, it's recommended that the AI extracts core entities and facts to avoid storing meaningless filler words.
2. **Knowledge Maintenance**: Periodically upload the latest technical documentation to the `knowledge_docs` table via the management console; the system will automatically handle vectorization and chunking.
3. **Monitoring**: Token consumption for vectorization services can be monitored via the `ai_usage_logs` table.
