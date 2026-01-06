# 2025 Agent 元年：架构演进与路线图 (Manus 对标计划)

## 1. 背景与愿景

2025 年被视为 Agent 元年。Manus 等产品的成功证明了通用 Agent 在用户体验和能力闭环上的巨大潜力。作为独立开发者或小团队，要在一周或一月内构建出 Manus 级别的产品，核心在于对 Agent 底层架构的精准拆解与重构。

通过深入研究，我们将通用 Agent 的核心架构拆解为“三大件”：
1.  **Framework (大脑)**: 负责规划 (Planning)、反思 (Reflection)、记忆 (Memory) 和上下文管理。
2.  **Tools (手脚)**: 基于 MCP (Model Context Protocol) 的标准化工具生态，连接数字世界。
3.  **Sandbox (环境)**: 安全、隔离的执行环境（如 Docker），用于代码执行、文件处理和复杂任务闭环。

本项目（BotMatrix）致力于将这“三大件”从外挂插件下沉为**核心基础服务**，实现原生 Agent 能力。

## 2. 核心架构升级进展 (Current Progress)

截至目前，我们已完成 Agent 核心能力的初步整合，将 AI 从“聊天插件”升级为“自主智能体”。

### 2.1 Framework: ReAct 自主循环
*   **实现**: `Common/ai/agent_executor.go`
*   **功能**: 实现了标准的 ReAct (Reasoning + Acting) 循环。
    *   **Thought**: 分析用户意图。
    *   **Plan**: 规划后续步骤。
    *   **Action**: 调用 MCP 工具。
    *   **Observation**: 接收工具反馈并进行下一步推理。
*   **集成**: 已深度集成至 `AIServiceImpl` 和 `WorkerAIService`。调用 `ChatAgent` 接口即可自动触发多步推理，无需人工干预。

### 2.2 Sandbox: Docker 隔离沙箱
*   **实现**: `Common/sandbox/manager.go` & `Common/ai/mcp/sandbox_host.go`
*   **功能**:
    *   **容器管理**: 基于 Docker SDK，支持毫秒级启动隔离容器（默认 Python 环境）。
    *   **资源限制**: 512MB 内存 / 0.5 CPU，保障宿主机安全。
    *   **文件系统**: 支持宿主机与沙箱间的文件写入与读取。
*   **MCP 暴露**: 提供了标准 MCP 工具：
    *   `sandbox_create`: 创建环境。
    *   `sandbox_exec`: 执行 Shell/Python 命令。
    *   `sandbox_write_file` / `sandbox_read_file`: 文件操作。
*   **自动注入**: 在 AI 服务初始化时，自动检测 Docker 环境并注入沙箱工具，Bot 无需配置即可使用。

### 2.3 Tools: MCP 统一生态
*   **实现**: `Common/ai/mcp/mcp_manager.go`
*   **功能**:
    *   **统一入口**: `MCPManager` 统一管理所有工具（沙箱、知识库、搜索、外部 API）。
    *   **Worker 赋能**: 即使是边缘节点的 BotWorker，现在也能通过 `WorkerAIService` 直接调用核心 Agent 能力和宿主机沙箱。
    *   **隐私与安全**: 集成了隐私过滤器，确保敏感数据不泄露给 LLM。

## 3. 差距分析与路线图 (Gap Analysis & Roadmap)

要达到 Manus 级别的用户体验，我们仍需在以下三个维度进行深度迭代：

### 3.1 记忆系统的主动性 (Active Memory)
*   **现状**: 被动注入。系统根据相似度检索历史记录，塞入 Context。
*   **差距**: Agent 无法自主决定“何时回忆”或“回忆什么”，缺乏对长期记忆的主动检索能力。
*   **改进计划**:
    *   将 `MemoryService` 封装为 MCP 工具（如 `memory_recall`, `memory_save`）。
    *   在 System Prompt 中引导 Agent 在信息不足时主动查阅记忆。

### 3.2 规划能力的宏观性 (Global Planning)
*   **现状**: ReAct (边想边做)。适合中短程任务（10-20 步）。
*   **差距**: 缺乏跨越数天、数十个步骤的复杂任务（如“开发一个完整网站”）的宏观规划能力。
*   **改进计划**:
    *   引入 `TaskService` 作为 Agent 的“笔记本”。
    *   实现 **Planner Pattern**: 先生成 Task List，持久化到数据库，再逐个 Execute，防止上下文丢失或任务中断。

### 3.3 感知力的完备性 (Perception / Browser)
*   **现状**: 只有执行力 (Sandbox)，缺乏感知力（看不见网页）。
*   **差距**: Manus 可以像人一样访问网页、阅读文档、点击按钮。
*   **改进计划**:
    *   集成 **Headless Browser (Playwright)**。
    *   开发 `BrowserMCPHost`，提供 `browser_navigate`, `browser_click`, `browser_screenshot`, `browser_extract_text` 等工具。
    *   实现“看网页 -> 写代码 -> 存文件”的完整闭环。

## 4. 下一步行动 (Next Steps)

优先级最高的是补全**感知力**，打通“联网获取信息”这一环。

1.  **集成 Playwright**: 引入 `playwright-go` 或类似库。
2.  **实现 Browser MCP**: 封装浏览器操作为标准工具。
3.  **端到端测试**: 验证 "访问 GitHub -> 抓取数据 -> 存入沙箱 -> 分析生成报告" 的完整流程。
