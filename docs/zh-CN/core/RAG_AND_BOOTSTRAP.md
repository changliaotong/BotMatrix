# RAG 与机器人自举 (Bootstrap) 架构文档

## 1. 概述
本项目引入了基于 RAG (Retrieval-Augmented Generation) 的机器人“自举”机制。通过将机器人的身份认知、功能指南与外部知识库结合，使 AI 能够具备自我意识，并能根据详细的系统文档回答复杂问题。

## 2. RAG 架构选型
- **开发语言**: Go (与 BotNexus 核心保持一致)
- **向量数据库**: PostgreSQL + [pgvector](https://github.com/pgvector/pgvector)
  - 理由：利用现有 PostgreSQL 基础设施，减少运维复杂度。
  - 支持 `L2 distance (<->)` 和 `Cosine distance (<=>)` 搜索。
- **向量化模型**: 推荐使用开源模型 [BGE-M3](https://huggingface.co/BAAI/bge-m3) (1024 维)
  - 部署方式：通过 [Ollama](https://ollama.com/) 本地运行 `ollama run bge-m3`。
  - 理由：BGE-M3 在中英文语义对齐和长文本处理上表现卓越，且支持本地部署，无需担心隐私和成本。
- **数据流**:
  1. **Indexer**: 扫描 `DOCS.md` 及代码注释，进行 Markdown 分片。
  2. **Storage**: 存储分片内容及其向量。
  3. **Retriever**: 根据用户输入，执行向量相似度检索。

## 3. 机器人自举 (Bootstrap) 机制
自举是指机器人通过内置的身份清单和能力描述，逐步建立起对自身功能的认知过程。

### 3.1 核心组件
- **BotIdentity**: 定义机器人的基本属性（名称、角色、性格、内置知识）。
- **SystemManifest**: 动态聚合所有已注册的技能（Skills）和核心动作（Actions）。
- **RAG Enhancement**: 当内置知识不足以回答时，触发深度知识库检索。

### 3.2 身份自举层次
1. **静态层**: 配置文件中定义的 `Name`, `Role`, `Personality`。
2. **动态层**: 实时汇报的 `Skills` 清单，告知 AI 当前可调用的工具。
3. **知识层**: 挂载的 RAG 知识库，提供“如何使用 [功能]”的深度文档。

## 4. 实现详情
- **数据模型**: [model.go](../../BotMatrix/common/rag/model.go) 定义了 `KnowledgeDoc` 和 `KnowledgeChunk`。
- **知识库实现**: [pg_knowledge.go](../../BotNexus/internal/rag/pg_knowledge.go) 实现了基于 pgvector 的存储与搜索。
- **能力注入**: [capabilities.go](../../BotNexus/tasks/capabilities.go) 负责将身份与 RAG 提示词注入 System Prompt。

## 5. 维护与更新
- **文档同步**: 运行 `Indexer` 工具可自动将最新的开发文档灌入向量库。
- **身份调整**: 修改 `GetDefaultManifest()` 中的 `Identity` 结构即可更新机器人的自我认知。

---
*Last Updated: 2026-01-03*
