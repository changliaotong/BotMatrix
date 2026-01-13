# BotMatrix 项目研发进度追踪表

## 1. AI 核心能力 (Intelligence Core)

| 功能模块 | 子任务 | 状态 | 说明 | 核心代码/文档 |
| :--- | :--- | :--- | :--- | :--- |
| **Agentic RAG** | 查询改写 (Refinement) | ✅ 已完成 | 提升向量搜索召回率 | [RagService.cs](file:///d:/projects/BotMatrix/src/BotWorker/Modules/AI/Services/RagService.cs) |
| | 自我反思 (Reflection) | ✅ 已完成 | 过滤无关干扰信息 | [RagService.cs](file:///d:/projects/BotMatrix/src/BotWorker/Modules/AI/Services/RagService.cs) |
| | 并发优化 | ⏳ 进行中 | 减少多步 RAG 的延迟 | - |
| **工具规范 v1** | 风险分级模型 | ✅ 已完成 | 定义 Low/Medium/High 风险 | [ToolRiskAttribute.cs](file:///d:/projects/BotMatrix/src/BotWorker/Modules/AI/Tools/ToolRiskAttribute.cs) |
| | 自动拦截器 | ✅ 已完成 | 高风险操作强制挂起 | [DigitalEmployeeToolFilter.cs](file:///d:/projects/BotMatrix/src/BotWorker/Modules/AI/Filters/DigitalEmployeeToolFilter.cs) |
| | 审计追踪系统 | ✅ 已完成 | 记录所有工具调用流水 | [ToolAuditService.cs](file:///d:/projects/BotMatrix/src/BotWorker/Modules/AI/Tools/ToolAuditService.cs) |
| **智能体框架** | 基础执行器 | ✅ 已完成 | 支持工具调用的 AgentExecutor | [AgentExecutor.cs](file:///d:/projects/BotMatrix/src/BotWorker/Modules/AI/Services/AgentExecutor.cs) |
| | 多智能体协作 | 📅 待开始 | Planner 与 Executor 职责分离 | - |

## 2. 基础设施与安全 (Infrastructure & Security)

| 功能模块 | 子任务 | 状态 | 说明 | 核心代码/文档 |
| :--- | :--- | :--- | :--- | :--- |
| **会话管理** | Redis Session 迁移 | ✅ 已完成 | 废弃 DB 存储，改用 Redis | [SessionManager.cs](file:///d:/projects/BotMatrix/src/BotWorker/Plugins/SessionManager.cs) |
| | 验证码确认逻辑 | ✅ 已完成 | 关键操作需要验证码确认 | [ConfirmMessage.cs](file:///d:/projects/BotMatrix/src/BotWorker/Domain/Models/Messages/BotMessages/ConfirmMessage.cs) |
| **缓存优化** | 高频字段属性 | ✅ 已完成 | 减少数据库 IO | [UserInfo.cs](file:///d:/projects/BotMatrix/src/BotWorker/Domain/Entities/UserInfo.cs) |
| **工具库扩展** | 系统文件工具 | ✅ 已完成 | AI 可读取项目文档 | [SystemToolPlugin.cs](file:///d:/projects/BotMatrix/src/BotWorker/Modules/AI/Plugins/SystemToolPlugin.cs) |
| | 只读数据库工具 | ✅ 已完成 | AI 可执行 SELECT 查询 | [SystemToolPlugin.cs](file:///d:/projects/BotMatrix/src/BotWorker/Modules/AI/Plugins/SystemToolPlugin.cs) |

## 3. 交互与管理 (UX & Management)

| 功能模块 | 子任务 | 状态 | 说明 | 核心代码/文档 |
| :--- | :--- | :--- | :--- | :--- |
| **文档建设** | 工具规范落地文档 | ✅ 已完成 | 指导后续工具开发 | [Agentic_RAG_v1.md](file:///d:/projects/BotMatrix/docs/zh-CN/ai/Agentic_RAG_and_Digital_Staff_v1.md) |
| **审批中心** | 后端逻辑封装 | ✅ 已完成 | 审批、拒绝、待审查询接口 | [ToolAuditService.cs](file:///d:/projects/BotMatrix/src/BotWorker/Modules/AI/Tools/ToolAuditService.cs) |
| | Web 审批界面 | 📅 待开始 | 管理员可视化操作面板 | - |

---

## 🚀 数字员工全流程上线路径 (Operational Roadmap)

为实现数字员工从“智能对话”到“自主执行任务”的跨越，需按以下阶段逐步推进：

### 第一阶段：智能基座与安全约束 (已基本完成)
- [x] **Agentic RAG**: 确保 AI 能获取准确的业务背景知识。
- [x] **工具规范 v1**: 建立风险分级与拦截机制，防止 AI “乱搞”。
- [x] **审计日志**: 所有操作可追溯、可还原。

### 第二阶段：自主任务拆解与规划 (当前重点 🎯)
- [ ] **Planner 实现**: 引入能够将复杂指令（如“分析本月报表并提交汇总”）拆解为多个子任务的规划器。
- [ ] **上下文记忆增强**: 实现长短期记忆（Long-term Memory），让数字员工记得之前的操作结果，避免重复报错。
- [ ] **多工具链调用**: 优化 AgentExecutor，使其支持在一个会话中连续调用 3 个以上工具。

### 第三阶段：人工审批与闭环管理 (开发中)
- [ ] **审批 API**: 提供前端调用的待审任务查询、批准、拒绝接口。
- [ ] **管理后台 (Web)**: 实现可视化审批中心，管理员可实时查看 AI 的“作案动机”和入参。
- [ ] **反馈循环**: 允许管理员在拒绝时给出理由，AI 根据理由调整行动方案。

### 第四阶段：多角色模板与业务集成 (待启动)
- [ ] **角色定义 (Personas)**: 预设“程序员”、“行政助手”、“数据分析师”等角色模板，配置专属工具集。
- [ ] **异步任务流**: 数字员工执行耗时任务时（如爬虫、大文件分析），能够通过 IM 平台主动推送进度。
- [ ] **组织架构集成**: 让数字员工感知群组权限，仅执行其权限范围内的任务。

---

## 📅 近期详细任务清单 (Next Actions)

| 优先级 | 任务描述 | 目标 | 负责人 |
| :--- | :--- | :--- | :--- |
| **P0** | **审批中心 API 开发** | 实现 `GET /api/admin/ai/approvals` 等接口 | Agent |
| **P0** | **Planner 逻辑引入** | 在 `AgentExecutor` 中加入 Task Decomposition 逻辑 | Agent |
| **P1** | **并发 RAG 优化** | 解决改写+检索+反思带来的延迟问题 | Agent |
| **P2** | **长短期记忆存储** | 接入 Redis 存储最近 10 轮的工具调用结果 | Agent |

---

## 历史进度记录 (Past Milestones)
*(此处保留之前的进度表内容)*
