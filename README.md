# BotMatrix

[🌐 English](#english) | [简体中文](#简体中文)

---

## 简体中文

BotMatrix 是一个 **AI 原生 (AI-Native)** 的分布式跨平台机器人技能平台。它不仅整合了不同平台的机器人（如微信、QQ、钉钉、Discord 等），更通过深度集成大语言模型 (LLM)，将传统机器人进化为具备感知、思考与执行能力的智能体 (Agent)。

### 🌟 核心特性

- **AI 原生架构**: 全面接入 DeepSeek, OpenAI, Ollama 等主流模型，支持用户配置私有 Token。
- **技能中心 (Skill Center)**: 像应用商店一样管理机器人能力，支持一键订阅、热插拔与权限隔离。
- **训练中心 (Training Center)**: 内置语料标注、提示词 (Prompt) IDE 与自动化微调 (Fine-tuning) 工作流。
- **分布式执行**: 采用 **BotNexus** (枢纽) 与 **BotWorker** (执行节点) 分离的设计，支持多机部署与大规模并发。
- **多平台适配**: 统一适配微信、QQ、钉钉、飞书、公众号、抖音、Discord、Telegram、Slack 等主流平台。
- **数字员工 (Digital Employee)**: 引入工号、职位、KPI 考评体系，将机器人升级为具备企业属性的虚拟雇员。
- **B2B 协作**: 支持不同企业数字员工之间的自然语言业务对接，建立安全可靠的跨企业通信协议。
- **Global Agent Mesh**: 去中心化的智能体协作网络，支持跨域发现、联邦搜索与任务共识。
- **MCP 标准支持**: 深度集成 Model Context Protocol，通过 SSE 与 B2B 协议连接全球 AI 工具生态。
- **隐私堡垒 (Privacy Bastion)**: 内置全链路隐私保护，自动识别并脱敏敏感信息，确保 AI 交互安全。
- **管理 App**: 提供专属移动 App，随时随地管理数字员工，支持实时人工干预与技能分发。
- **实时控制台 (WebUI)**: 提供全局弹出式聊天窗口，支持实时消息监控、技能结果反馈与未读提醒。
- **Online 机器人仿真**: 支持模拟机器人连接，无需外部 Docker 容器即可测试技能逻辑与消息路由。

### 🏗️ 系统组件

1. **[BotNexus](docs/zh-CN/components/BotNexus.md)**: 系统“大脑”，负责连接管理、消息路由、AI 意图识别及技能分发。
2. **[BotWorker](docs/zh-CN/components/BotWorker.md)**: 执行节点，承载具体的插件运行环境与 AI 推理/任务处理逻辑。
3. **[AI 引擎](docs/zh-CN/core/AI_INTEGRATION_PLAN.md)**: 统一的模型调度中心，支持多模型负载均衡、RAG 知识库与长期记忆。
4. **[Overmind](docs/zh-CN/components/Overmind.md)**: 监控与管理后台，提供可视化技能训练、配置管理及统计分析。
4. **[SystemWorker](docs/zh-CN/components/SystemWorker.md)**: 专门负责系统级维护任务与高优先级作业。

### 📚 文档中心

- **[简体中文文档中心](docs/zh-CN/README.md)**
  - [系统架构图](docs/zh-CN/core/ARCHITECTURE.md)
  - [API 接口参考](docs/zh-CN/core/API_REFERENCE.md)
  - [部署指南](docs/zh-CN/core/DEPLOY.md)
  - [AI 助手指南](docs/zh-CN/development/AI_GUIDE.md)
  - [插件开发手册](docs/zh-CN/plugins/PLUGIN_DEVELOPMENT.md)
  - [国际化 (I18N) 开发指南](docs/zh-CN/development/I18N_GUIDE.md)

### 🚀 快速开始

请参考 **[部署指南](docs/zh-CN/core/DEPLOY.md)** 完成环境配置与系统启动。

---

## English

BotMatrix is an **AI-Native** distributed cross-platform bot skill platform. It not only integrates bots from various platforms (such as WeChat, QQ, DingTalk, Discord, etc.) but also evolves traditional bots into intelligent Agents with perception, reasoning, and execution capabilities through deep integration with Large Language Models (LLMs).

### 🌟 Key Features

- **AI-Native Architecture**: Full support for DeepSeek, OpenAI, Ollama, and other major models, allowing users to configure private tokens.
- **Skill Center**: Manage bot capabilities like an app store, supporting one-click subscription, hot-plugging, and permission isolation.
- **Training Center**: Built-in corpus annotation, Prompt IDE, and automated fine-tuning workflows.
- **Distributed Execution**: Separation of **BotNexus** (Hub) and **BotWorker** (Execution Node), supporting multi-machine deployment and large-scale concurrency.
- **Multi-Platform Support**: Unified adapters for WeChat, QQ, DingTalk, Feishu, Discord, Telegram, Slack, and more.
- **Multi-Language Plugin System**: Support for writing AI-capable plugins in Go, Python, C#, and other languages.
- **Enterprise Control**: Includes RBAC permission management, audit logs, sensitive word filtering, and unified cross-platform identity mapping (NexusUID).
- **Global Agent Mesh**: Decentralized smart agent collaboration network supporting cross-domain discovery and federated search.
- **MCP Protocol**: Deep integration with Model Context Protocol via SSE and B2B protocols to connect with the global AI tool ecosystem.
- **Privacy Bastion**: Built-in full-link privacy protection with automatic de-identification and restoration of sensitive data.
- **Real-time Console (WebUI)**: Features a global pop-up chat window for real-time message monitoring, skill results, and unread notifications.
- **Online Bot Simulation**: Supports simulated bot connections for testing skill logic and message routing without external Docker containers.

### 🏗️ System Components

1. **[BotNexus](docs/en-US/components/BotNexus.md)**: The "Brain" of the system, responsible for connection management, message routing, AI intent recognition, and skill distribution.
2. **[BotWorker](docs/en-US/components/BotWorker.md)**: Execution nodes that host specific plugin environments and AI inference/task processing logic.
3. **[AI Engine](docs/zh-CN/core/AI_INTEGRATION_PLAN.md)**: Unified model scheduling center, supporting multi-model load balancing, RAG knowledge bases, and long-term memory.
4. **[Overmind](docs/en-US/components/Overmind.md)**: Monitoring and management backend, providing visual skill training, configuration management, and statistical analysis.
4. **[SystemWorker](docs/en-US/components/SystemWorker.md)**: Dedicated to system-level maintenance tasks and high-priority jobs.

### 📚 Documentation Hub

- **[English Documentation Hub](docs/en-US/README.md)**
  - [System Architecture](docs/en-US/core/ARCHITECTURE.md)
  - [API Reference](docs/en-US/core/API_REFERENCE.md)
  - [Deployment Guide](docs/en-US/core/DEPLOY.md)
  - [AI Assistant Guide](docs/en-US/development/AI_GUIDE.md)
  - [Plugin Development](docs/en-US/plugins/PLUGIN_DEVELOPMENT.md)
  - [I18N Development Guide](docs/en-US/development/I18N_GUIDE.md)

### 🚀 Quick Start

Please refer to the **[Deployment Guide](docs/en-US/core/DEPLOY.md)** for environment configuration and system startup.

---

**BotMatrix Team** | 2025
