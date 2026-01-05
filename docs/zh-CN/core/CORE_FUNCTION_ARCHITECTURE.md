# 🏗️ BotMatrix 核心功能架构文档

> [🌐 English](../en-US/CORE_FUNCTION_ARCHITECTURE.md) | [简体中文](CORE_FUNCTION_ARCHITECTURE.md)
> [⬅️ 返回文档中心](../README.md) | [🏠 返回项目主页](../../README.md)

本文档详细介绍了 BotMatrix 系统中四大核心功能的架构设计与实现逻辑：**B2B 协作服务**、**增强型 AI 服务层**、**多智能体协作网桥**以及**自主学习系统**。

---

## 1. 🏢 B2B 协作服务 (Global Agent Mesh)

B2B 协作服务旨在打破企业间的“数据孤岛”，实现跨企业的数字员工身份委派、技能共享与知识检索。

### 1.1 核心组件
- **Identity Manager**: 管理企业级公私钥对，负责跨企业调用的 JWT 签名与验证。
- **Circuit Breaker (熔断器)**: 监控远程企业端点的可用性，防止因第三方服务故障导致的级联雪崩。
- **Mesh Knowledge Bridge**: 并发向所有已建立连接的合作伙伴发起联邦知识检索。

### 1.2 关键流程：跨企业握手 (Handshake)
1. **发起方**: 生成 `challenge`，使用私钥签名，发送至接收方的 `/api/b2b/handshake`。
2. **接收方**: 使用发起方的公钥验证签名。
3. **双向信任**: 接收方在数据库中记录反向连接，并返回带签名的 `acceptance`。

### 1.3 核心代码参考
- [b2b_service.go](file:///d:/projects/BotMatrix/src/BotNexus/internal/app/b2b_service.go)
- [handlers_mesh.go](file:///d:/projects/BotMatrix/src/BotNexus/internal/app/handlers_mesh.go)

---

## 2. 🧠 增强型 AI 服务层 (AI Service Layer)

AI 服务层不仅是 LLM 的简单封装，更是一个具备可观测性、稳定性与上下文感知能力的复杂引擎。

### 2.1 稳定性保障
- **超时控制**: 全局 60s 超时控制，确保高并发下的资源及时释放。
- **结构化日志**: 使用 `zap` 记录每次 AI 调用的 `AIAgentTrace`，涵盖记忆提取、知识检索、工具注入等全生命周期。

### 2.2 上下文增强
- **Session 绑定**: 自动在 AI 推理过程中注入 `sessionID` 和 `executionID`。
- **Tool Injection**: 根据当前上下文（如 B2B 环境、特定技能集）动态注入 MCP 工具。

---

## 3. 🤝 多智能体协作网桥 (Collaboration Bridge)

协作网桥允许 AI 智能体像人类员工一样进行团队协作。

### 3.1 协作工具集 (MCP Tools)
- `colleague_list`: 发现同事。
- `colleague_consult`: 同步咨询。
- `task_delegate`: 异步任务委派（返回 `execution_id`）。
- `task_report`: 进度与结果汇报。

### 3.2 任务链路追踪
系统通过 `parentSessionID` 将一系列协作对话串联成一条完整的链路。当 Agent A 委派任务给 Agent B 时，B 的所有后续行为都将关联至 A 的原始请求。

---

## 4. 📚 自主学习与认知记忆 (Autonomous Learning)

系统具备从日常对话中提取知识、解决冲突并不断进化的能力。

### 4.1 自动学习流程
1. **知识提取**: 对话结束后，后台触发 `ExtractAndSaveMemories`。
2. **冲突解决**: 使用向量相似度检测现有知识库。若发现相似知识点，调用 LLM 进行合并（Merging）而非简单覆盖。
3. **KPI 评价**: 自动评估每次对话的质量，记录 KPI 数据。

### 4.2 认知记忆
- **Short-term**: 对话上下文（Redis 缓存）。
- **Long-term**: `CognitiveMemoryGORM` 表，支持向量检索 (pgvector)。

---

## 5. 🛠️ 开发与维护建议

- **数据库迁移**: 任何新的 GORM 模型必须在 `db.go` 的 `AutoMigrate` 中注册。
- **接口扩展**: 推荐通过增加新的 MCP Host 来扩展 AI 的能力，而非修改核心 AI 逻辑。
- **测试验证**: 核心变更应运行 `concurrency_test.go` 进行稳定性验证。
