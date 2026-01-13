# 数字员工 AI 能力升级与工具规范 v1 落地文档

## 1. 背景
随着 RAG (Retrieval-Augmented Generation) 技术的演进，传统的 Naive RAG 已难以满足复杂的业务需求。同时，为了让 AI 能够安全、受控地调用系统工具，我们引入了 **Agentic RAG** 架构与 **《数字员工工具接口规范 v1》**。

## 2. Agentic RAG 升级
我们在 `RagService` 中实现了 Agentic RAG 的核心能力，通过 AI 的自我迭代优化检索质量。

### 核心特性
- **查询改写 (Query Refinement)**：AI 会将原始用户提问改写为更适合向量搜索的关键词，提升召回率。
- **自我反思 (Self-Reflection)**：检索出的内容会经过 AI 的二次评估，剔除无关干扰信息，确保上下文的精准度。
- **循环迭代**：支持在检索不到结果时自动调整策略重新尝试。

### 代码实现
参考：[RagService.cs](file:///d:/projects/BotMatrix/src/BotWorker/Modules/AI/Services/RagService.cs)

---

## 3. 数字员工工具接口规范 v1 落地
为了解决 AI “幻觉”导致的高风险操作问题，我们建立了一套完整的工具分级与审计体系。

### 3.1 风险等级定义 (Risk Levels)
我们通过 `ToolRiskAttribute` 对所有可被 AI 调用的函数进行标记：
- **Low (低风险)**：只读、无副作用的操作。如：检索知识库、读取公共文件。自动执行。
- **Medium (中风险)**：有限副作用的操作。如：发送消息、执行普通插件功能。自动执行并审计。
- **High (高风险)**：影响系统状态或关键数据的操作。如：修改数据库、合并代码。**必须经过人工审批**。

### 3.2 审计与追溯 (Audit)
所有工具调用都会实时记录到 `ToolAuditLogs` 表中，包含：
- 任务 ID 与执行员工 ID
- 工具名称与入参 (JSON)
- 执行结果或错误信息
- 风险等级与审批人

### 3.3 拦截器机制 (Interceptor)
通过 `DigitalEmployeeToolFilter` 实现了基于 Semantic Kernel 的全局拦截：
- **自动记录日志**：调用前后无感记录。
- **高风险拦截**：发现 `High` 风险调用时，立即挂起任务，状态置为 `PendingApproval`，并返回友好提示。

---

## 4. 目录结构说明
- `src/BotWorker/Modules/AI/Tools/`：工具元数据与审计服务。
- `src/BotWorker/Modules/AI/Filters/`：工具调用拦截器与安全策略。
- `src/BotWorker/Modules/AI/Services/`：核心 Agent 执行引擎。

## 5. 后续规划
1. **扩展工具库**：增加文件操作、数据库查询、代码分析等更多标准工具。
2. **可视化审批面板**：在管理后台增加审批流处理页面。
3. **多智能体协作 (Multi-Agent)**：基于规范实现 Planner 与 Executor 的职责分离。
