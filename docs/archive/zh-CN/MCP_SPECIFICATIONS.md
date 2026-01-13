# 🔌 MCP 核心協議與系統集成 (MCP Specification & Integration)

> [⬅️ 返回文档中心](../README.md) | [🏠 返回项目主页](../../README.md)

## 1. 概述与设计哲学 (Overview & Philosophy)

Model Context Protocol (MCP) 是 BotMatrix 的核心总线协议。它类似于硬件世界的 **USB 标准**，将“能力提供方”与“模型使用方”彻底解耦。

### 1.1 核心痛点
在传统的 AI 开发中，集成新工具通常是“烟囱式”的，导致适配成本随工具和模型增加呈指数级增长。

### 1.2 我们的哲学
我们不认为机器人只是一个对话框。我们认为：
> **机器人 = 实时通信 (IM) + 标准化能力 (MCP) + 全球协作网络 (Mesh)**

MCP 赋予了机器人“手脚”与“感知”，而 Mesh 赋予了它们“社交能力”。

---

## 2. MCP 的三大支柱 (Core Pillars)

### 2.1 Resources (只读上下文)
- **原理**：允许模型按需读取静态或动态数据。
- **场景**：历史对话记录、用户笔记、数据库实时报表。

### 2.2 Tools (可执行动作)
- **原理**：允许模型执行代码或调用外部 API。
- **场景**：发送消息、修改配置、执行 Shell 脚本、操作桌面应用。

### 2.3 Prompts (预定义思维)
- **原理**：提供复杂的提示词模板。
- **场景**：专家人格切换、任务拆解流程模板。

---

## 3. 系统架构与双栈模式 (Architecture)

BotMatrix 作为一个 **Universal MCP Host**，实现了高度解耦的架构。

### 3.1 架构层次
1.  **接入层 (Access Layer)**：支持 **STDIO**, **HTTP** 和 **SSE** 通信协议。
2.  **核心层 (Kernel Layer)**：BotNexus 负责权限调度、`Scope` 管理与 **Privacy Bastion (隐私堡垒)** 集成。
3.  **桥接层 (Bridge Layer)**：将现有插件系统自动打包为 **Internal MCP Server**。

### 3.2 双栈 (Dual-Stack) 架构
- **适配器模式 (Legacy)**：负责高效的消息收发与基础路由。
- **MCP 模式 (Modern)**：将适配器封装为 MCP 工具，允许 AI “按需调用”通信能力。

---

## 4. 数据库与向量集成 (Database & Vector Integration)

BotMatrix 通过 MCP 深度集成 PostgreSQL 和 pgvector，为 AI 提供持久化记忆。

### 4.1 认知记忆集成 (Memory MCP)
`MemoryMCPHost` 接入了基于 PostgreSQL 的长期记忆系统：
- **持久化**: 记忆通过 [CognitiveMemoryService](../../../src/BotNexus/internal/app/cognitive_memory.go) 存储。
- **向量化**: 利用 **pgvector** 插件生成 Embedding。
- **安全性**: 在 MCP 层级统一进行 PII 脱敏和权限控制。

---

## 5. Mesh、MCP 与适配器的演进 (Evolution)

### 5.1 类比理解
- **IM 适配器** = 耳朵与嘴巴（负责收发消息）。
- **MCP** = 通用神经网络接口（负责接入工具与能力）。
- **Global Agent Mesh** = 神经网络/互联网（负责跨地域、跨企业协作）。

### 5.2 协作流程
1. **跨域发现**：数字员工可以发现合作伙伴公开的 MCP 工具。
2. **联邦认证**：通过 B2B 联邦认证实现安全的能力输出。
3. **隐私隔离**：所有跨域数据经过自动脱敏，确保数据安全。

---

## 6. 国内生态适配建议 (Domestic Ecosystem)

针对境内访问限制，推荐以下本土化适配方案：

### 6.1 核心推荐服务
- **搜索**: Bing Search MCP, Bocha (博查), 知乎/公众号爬虫。
- **办公**: 飞书 (Feishu/Lark) MCP, 钉钉 (DingTalk) MCP, 企业微信 MCP。
- **基础**: 本地文件系统, 数据库 (MySQL/PostgreSQL), Python 执行器。

---

## 7. 开发者指南 (Developer Guide)

### 如何接入新功能
1. **编写标准 MCP Server**：使用任何语言（Python, JS, Go）实现。
2. **连接至 BotMatrix**：在后台输入 MCP Server 的 URL 或命令。
3. **能力发现**：BotNexus 自动将其包装为全局可用的 Skill。

---
*Last Updated: 2026-01-13*
