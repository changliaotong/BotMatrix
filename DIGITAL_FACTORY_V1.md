# BotMatrix Digital Employee Manufacturing Factory V1

## 1. 概述 (Overview)
BotMatrix 数字员工工厂是一个高度自动化的、无头（Headless）运行的智能体生产环境。它能够根据预定义的角色模板批量生产数字员工，并通过监控看板和 CI/CD 集成实现全自动的任务分发、代码编写、Bug 修复和闭环测试。

## 2. 核心架构 (Core Architecture)

### 2.1 角色模板驱动 (Role Template Driven)
所有的数字员工行为都由数据库中的 `DigitalRoleTemplate` 定义。
- **SOP (Standard Operating Procedure)**: 在 `BasePrompt` 中定义，基于 ReAct (Thought -> Action -> Observation -> Reflection) 模式。
- **技能集 (Skills)**: 定义员工可以访问的 MCP 工具集（如 `local_dev`, `git_ops`）。

### 2.2 无头智能体 (Headless Agents)
不同于传统的 IDE 插件，我们的数字员工直接运行在服务端，通过 MCP (Model Context Protocol) 协议操作文件系统和运行环境，实现真正的无人值守自动化。

## 3. 已实现功能 (Implemented Features)

### 3.1 Web 监控看板 (Monitoring Dashboard)
- **实时状态**: 监控活跃员工数、待处理任务、今日完成任务及失败任务。
- **任务流**: 实时展示正在进行的任务详情（正在写什么代码、修复什么 Bug）。
- **API 驱动**: 后端通过 `DashboardService` 提供标准 RESTful API。
- **访问地址**: `http://localhost:8080`

### 3.2 CI/CD Webhook 集成
- **自动修复**: 接入 GitLab Webhook。当 CI 流水线报错时，WebhookService 会自动提取错误上下文。
- **任务闭环**: 自动创建高优先级的“AI 修复任务”并指派给具备相关技能的数字员工。
- **自动提交**: 数字员工修复代码后会自动运行测试并重新提交。
- **访问地址**: `http://localhost:8081/webhook/gitlab`

### 3.3 真实大模型集成 (Real LLM Integration)
- 支持 **OpenAI** 和 **DeepSeek** (兼容 OpenAI 接口)。
- 通过 `config.json` 配置 API Key 和 BaseURL，实现从 Mock 到生产环境的平滑切换。

## 4. 快速启动 (Quick Start)

1. **配置数据库**: 修改 `config.json` 中的 PostgreSQL 连接。
2. **配置 AI**: 在 `config.json` 中设置 `ai_provider_type` (mock/openai/deepseek) 及 `ai_api_key`。
3. **启动工厂**:
   ```bash
   go run src/Common/cmd/setup_experiment/main.go
   ```

## 5. 工厂扩容路线图 (Roadmap)
1. **多智能体协作 (Swarm)**: 增加架构师、开发者、测试员等不同角色，实现流水线式协作。
2. **知识库 RAG**: 集成公司内部代码规范和历史修复经验。
3. **性能看板**: 统计每个数字员工的 ROI (投入产出比) 和代码质量评分 (KPI)。
