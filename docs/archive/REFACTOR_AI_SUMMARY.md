# AI 系统重构技术总结文档

## 1. 重构目标
- **消除冗余**：废除独立开发的 `HelpAgentService`，整合进现有的 `MatrixOracleService`。
- **系统自举**：实现机器人功能说明的自动提取与 RAG 索引，支持 AI 实时解答用户提问。
- **架构统一**：将 AI 服务（AIService）、智能体执行（AgentExecutor）与 RAG 服务（RagService）与系统底座（LLMApp, KnowledgeBaseService, MCPManager）全面对接。

## 2. 核心改动说明

### 2.1 AI 服务层 (AIService)
- **改动文件**：[AIService.cs](file:///d:/projects/BotMatrix/src/BotWorker/Modules/AI/AIService.cs)
- **内容**：重构 `AIService`，移除占位逻辑，直接调用 `LLMApp` 管理的 `ModelProviderManager`。
- **收益**：支持 DeepSeek, Azure OpenAI, OpenAI 等多模型动态切换，复用系统全局配置。

### 2.2 智能体执行引擎 (AgentExecutor)
- **改动文件**：[AgentExecutor.cs](file:///d:/projects/BotMatrix/src/BotWorker/Modules/AI/AgentExecutor.cs)
- **内容**：实现 `AgentExecutor`，集成 Semantic Kernel。
- **功能**：
    - 动态获取 MCP 工具并转换为 SK 插件。
    - 支持函数调用 (Tool Call) 机制，使 AI 能够调用系统指令。
    - 统一处理对话历史与上下文。

### 2.3 RAG 服务增强 (RagService)
- **改动文件**：[RagService.cs](file:///d:/projects/BotMatrix/src/BotWorker/Modules/AI/RagService.cs)
- **内容**：将 `RagService` 与 `KnowledgeBaseService` 整合。
- **逻辑**：优先检索外部向量知识库，结合本地内存索引（Chunk 存储），实现多级知识检索。

### 2.4 帮助系统自举 (MatrixOracleService)
- **改动文件**：[MatrixOracleService.cs](file:///d:/projects/BotMatrix/src/BotWorker/Modules/Games/MatrixOracleService.cs)
- **内容**：
    - 增加 `IndexSystemManualAsync` 方法，自动遍历 `IRobot.Skills` 提取所有功能描述。
    - 在启动 10 秒后自动生成“系统说明书”并写入 RAG 索引。
    - 结合 RAG 检索与 AI 生成，实现“先知咨询”功能。

## 3. 基础设施调整
- **[IRobot.cs](file:///d:/projects/BotMatrix/src/BotWorker/Domain/Interfaces/IRobot.cs)**：新增 `Agent` 属性，暴露智能体执行引擎。
- **[PluginManager.cs](file:///d:/projects/BotMatrix/src/BotWorker/Plugins/PluginManager.cs)**：实现依赖注入，完善 AI/Agent/Rag 的全量接入。
- **[Program.cs](file:///d:/projects/BotMatrix/src/BotWorker/Program.cs)**：完成 `LLMApp`, `IKnowledgeBaseService`, `IAgentExecutor` 的全局单例注册。

## 4. 验证情况
- **自动化测试**：执行 `dotnet test`，所有 49 项功能测试（含 2048、宠物、求婚、钓鱼等插件）全部通过。
- **编译检查**：修复了 `OpenAIAzureApiHelper` 接口实现不完整及 `MatrixOracleService` 字段引用错误。

## 5. 后续计划
- 接入更多 MCP Server 以扩展 AI 的操作能力。
- 优化 `KnowledgeBaseService` 的检索权重算法。
- 增加多轮对话的上下文持久化支持。
