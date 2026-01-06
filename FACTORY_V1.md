# BotMatrix Digital Employee Factory V1

## 概述
BotMatrix 数字员工工厂是一个全自动化的代码生产与修复中心。它通过将 AI 代理（数字员工）与本地开发环境、Git 工作流以及 CI/CD 流水线集成，实现了从问题发现到自动修复的闭环。

## 核心特性

### 1. 数据库驱动的数字劳动力 (Database-Driven Workforce)
- **角色模板 (DigitalRoleTemplate)**：所有的数字员工行为准则（SOP）和 ReAct 提示词都存储在数据库中，支持动态调整而无需重新发布代码。
- **灵活招聘 (Recruitment)**：支持从模板一键生成具有特定技能和性格的数字员工。

### 2. ReAct 自主思考模型 (Autonomous ReAct Pattern)
- 数字员工遵循 **Thought -> Plan -> Action -> Observation -> Reflection** 的循环。
- 通过集成 `local_dev` MCP Host，数字员工可以安全地读取代码、编写代码、运行命令并提交 Git。

### 3. 全链路 CI/CD 闭环 (CI/CD Integration)
- **GitLab Webhook**：实时监听代码仓库的流水线状态。
- **自动修复**：当流水线失败时，系统自动识别问题、匹配最合适的数字专家（如 Code Repair Expert），并立即启动 AI 修复任务。

### 4. 数字化工厂监控看板 (Monitoring Dashboard)
- **战略目标跟踪**：实时展示工厂的整体演进目标及各阶段里程碑进度。
- **活跃指标**：监控在线数字员工数量、待处理任务及今日完成情况。
- **实时任务流**：可视化展示数字员工正在进行的具体工作。

### 5. 多模型兼容与安全 (Model Agnostic & Safety)
- **模型中心**：支持 OpenAI, DeepSeek 等多种大模型的动态切换。
- **安全沙箱**：通过白名单和 MCP 协议限制数字员工的文件访问权限。

## 技术架构
- **后端**: Go (GORM, PostgreSQL, Redis)
- **AI 核心**: ReAct Agent, MCP (Model Context Protocol)
- **前端**: HTML5/JS (实时刷新, 多语言支持)

## 未来路线图 (V2)
- [ ] 接入多模型协同 (Agent Collaboration)
- [ ] 强化认知记忆 (Cognitive Memory)
- [ ] 自动化测试用例生成与验证
