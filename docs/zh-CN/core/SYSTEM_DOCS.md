# BotMatrix 系统说明文档

本文件记录了 BotMatrix 系统的核心设定、数据库架构以及近期进行的逻辑变更。

## 1. 核心系统设定

### 1.1 积分系统 (Points System)
系统支持两种积分管理模式，根据群组设置自动切换：
- **全局模式 (Global Mode)**：用户积分在所有群组间共享，存储在 `"User"` 表的 `Credit` 字段。
- **群模式 (Group Mode)**：用户积分仅在特定群组内有效，存储在 `"GroupMember"` 表的 `Credit` 字段。

**核心逻辑：**
- **自动路由**：所有积分操作（增加、查询、冻结、转账）都会通过 `IsGroupCreditSystemEnabled` 检查群组配置，自动选择操作目标表。
- **日志记录**：所有积分变动都会记录到 `"Credit"` 日志表中，包含变动原因、分类及变动后的余额。

### 1.2 储蓄系统 (Savings System)
用户可以将可用积分存入储蓄账户以赚取利息。
- **利息计算**：系统根据配置的日利率（目前默认为 0.05%）按天计算简单利息。
- **自动结息**：在进行存款、取款或查询余额操作时，系统会自动触发 `applySavingsInterestTx` 结转未结算的利息。
- **元数据管理**：使用 `"UserSavingsMetadata"` 表记录每个用户上次结息的时间戳。

### 1.3 AI 技能引擎 (AI Skill Engine)
系统采用 AI 原生架构，将所有机器人交互视为技能执行：
- **多租户 AI 配置**：支持系统级与用户级 AI 提供商配置（`AIProvider`），用户可配置私有 API Key。
- **隐私保护 (Privacy Bastion)**：集成 PII 脱敏技术，在调用云端模型前自动替换敏感信息，并在返回后还原。
- **RAG 2.0 (Agentic RAG)**：支持意图补全（Query Refinement）与自省（Self-Reflection），提高检索精度。
- **GraphRAG**：基于实体-关系的知识图谱检索，支持跨文档关系挖掘。
- **多模态解析**：支持 PDF 内嵌图片提取、独立图片 OCR 及 Excel 结构化数据索引。
- **调度优先级**：执行 AI 任务时，系统遵循 `用户私有配置 > 系统公共配置` 的调度逻辑。
- **技能生命周期**：包含技能定义（`AISkill`）、提示词管理、语料标注（`AITrainingData`）与 RAG 知识库挂载。

## 2. 数据库架构 (Database Schema)

### 2.1 核心表定义
- **`"User"`**: 全局用户信息，包含全局积分 `Credit` 和储蓄余额 `SaveCredit`。
- **`"Group"`**: 群组信息，包含 `CreditMode`（积分模式开关）。
- **`"GroupMember"`**: 群成员信息，包含群内积分 `Credit` 和冻结积分 `FreezeCredit`。
- **`"Credit"`**: 积分变动日志。
- **`"UserSavingsMetadata"`**: 储蓄系统结息元数据。

### 2.2 AI 相关表 (AI Models)
- **`ai_providers`**: AI 服务提供商配置（OpenAI, DeepSeek 等），包含 `APIKey`（加密存储）与 `UserID`。
- **`ai_models`**: 具体的 AI 模型配置（如 `gpt-4`, `deepseek-chat`）。
- **`ai_skills`**: 机器人技能定义，包含 `Prompt` 模板与分类。
- **`ai_prompt_templates`**: 场景化提示词模板，支持版本管理。
- **`ai_knowledge_bases`**: RAG 知识库配置。
- **`knowledge_docs`**: 知识库文档分段存储。
- **`knowledge_entities`**: GraphRAG 实体定义。
- **`knowledge_relations`**: GraphRAG 关系定义。
- **`ai_training_data`**: 标注数据与训练语料，支持 Few-shot 学习。
- **`ai_usage_logs`**: AI Token 消耗审计与成本控制日志。

### 2.2 扩展表
- **`group_admins`**: 群管理员权限管理。
- **`group_rules`**: 群规及语音配置。
- **`group_features`**: 功能开关覆盖设置。
- **`WhiteList`**: 群白名单用户。

## 3. 近期变更记录 (Recent Changes)

### 3.1 积分系统重构
- **逻辑优化**：重构了 `AddPoints`、`GetPoints`、`FreezePoints` 等核心函数，使其支持群/全局模式的自动路由。
- **转账功能**：新增 `TransferPoints` 函数，支持跨用户转账，并自动处理接收者不存在时的初始化逻辑（全局模式下）。

### 3.2 储蓄系统升级
- **架构变更**：引入 `user_savings_metadata` 表，解决了利息结算时间戳追踪问题。
- **安全性提升**：所有储蓄操作均在事务中执行，并强制执行 `FOR UPDATE` 行级锁以防止并发冲突。

### 3.3 RAG 2.0 与隐私保护升级 (2026-01-03)
- **PII 脱敏**：在 `Common/ai` 引入 `PrivacyFilter`，实现 PII 信息的自动掩码与还原。
- **GraphRAG 落地**：实现了实体与关系的自动提取与向量检索，极大增强了复杂问题的回答能力。
- **多模态增强**：集成了 PDF 图片提取 (pdfcpu) 与图片 OCR (AI Vision)，支持独立图片 (.jpg/.png) 的索引。
- **私人助理模式**：规划并落地了“数字分身”助理模式，支持个人私有知识库与隐私堡垒。

### 3.4 代码质量与修复
- **编译修复**：解决了 `WithdrawPointsFromSavings` 缺少 `botUin` 参数的问题。
- **SQL 优化**：修复了 SQLite 模式下 GraphRAG 表不存在的问题，改用共享缓存模式。
- **清理冗余**：移除了 `db.go` 中多个函数内声明但未使用的 `targetTable` 变量。
- **初始化增强**：在 `InitDatabase` 中增加了自动创建储蓄系统相关表的逻辑。

---
*最后更新日期：2026-01-03*
