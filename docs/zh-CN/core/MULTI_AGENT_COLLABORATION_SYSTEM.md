# 🤖 多智能体协作与任务追踪系统 (Multi-Agent Collaboration & Task System)

本文档详细说明了 BotMatrix 系统中多智能体协作与任务追踪的技术实现细节。

## 1. 设计目标
- **跨智能体协作**：允许数字员工相互咨询问题或委派任务。
- **任务闭环追踪**：实现任务从发起、执行到结果反馈的全生命周期监控。
- **全链路追踪**：支持跨 Session、跨 Agent 的调用链追踪。
- **权限与隔离**：确保协作过程符合企业内部及跨企业的安全策略。

## 2. 核心架构

### 2.1 协作网桥 (Collaboration Bridge)
实现在 `src/BotNexus/internal/app/agent_collaboration_bridge.go`。它作为 MCP (Model Context Protocol) 服务运行，为 AI 提供以下工具：

- **`colleague_list`**: 获取企业内可用同事列表及其职责描述。
- **`colleague_consult`**: 同步咨询工具。Agent 会挂起当前推理，等待目标同事的答复。
- **`task_delegate`**: 异步委派工具。返回 `execution_id`，允许 Agent 继续处理其他子任务或等待汇报。
- **`task_status`**: 进度查询工具。Agent 可主动检查之前委派任务的状态。
- **`task_report`**: 结果汇报工具。被委派者完成工作后，通过此工具更新任务结果。

### 2.2 任务追踪 (Execution Tracking)
利用 `tasks.Execution` 模型记录任务状态：
- **ExecutionID**: 唯一任务标识。
- **Status**: `running`, `success`, `failed`。
- **TraceID**: 关联原始请求的 SessionID，实现纵向链路分析。
- **Result**: 存储最终产出或错误详情。

### 2.3 链路关联机制
通过上下文（Context）传递关键元数据：
- **`sessionID`**: 当前对话会话 ID。
- **`parentSessionID`**: 触发当前协作的父级会话 ID。
- **`executionID`**: 关联的具体任务执行 ID。

当 AI 调用 `task_delegate` 时，系统会自动在目标 Agent 的 `InternalMessage.Extras` 中注入这些 ID。目标 Agent 在响应时会根据这些 ID 自动更新任务数据库。

## 3. 技术实现要点

### 3.1 意图引导
在 `AIServiceImpl.ChatWithEmployee` 中，如果检测到消息包含 `executionID`，系统会自动在系统提示词（System Prompt）中追加以下指令：
> "注意：你当前正在执行一个被委派的任务（ID: [ID]）。完成任务后，请务必使用 `task_report` 工具汇报进度或结果。"

这确保了 Agent 意识到自己的职责边界。

### 3.2 AI 驱动的自动学习与冲突解决
在协作过程中产生的知识通过 `AutoLearnFromConversation` 自动提取：
- **JSON 提取**: 使用结构化 Prompt 强制 AI 输出 JSON 格式的知识点。
- **冲突检测**: 
  - 相似度 > 0.95: 视为重复，自动跳过。
  - 相似度 0.6 ~ 0.95: 触发 AI 冲突解决逻辑，由 LLM 合并新旧知识点。
- **异步处理**: 学习过程在后台协程执行，不阻塞用户对话。

### 3.3 稳定性保障
- **高并发支持**: 核心逻辑经过 `TestChatWithEmployeeConcurrency` 压力测试，确保在大量并发协作请求下，数据库事务和 AI 客户端连接池保持稳定。
- **超时策略**: 所有协作调用均受 `60s` 全局超时控制。

## 4. 下一步演进
- **联邦搜索**: 实现跨 Agent 的分布式知识检索，允许 Agent A 搜索 Agent B 拥有权限的知识库。
- **任务编排可视化**: 在 Overmind 管理后台提供任务委派的甘特图/拓扑图展示。
- **优先级调度**: 根据任务优先级（Low/Medium/High）自动调整 AI 推理队列。
