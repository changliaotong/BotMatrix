# BotMatrix

[🌐 English](#english) | [简体中文](#简体中文)

---

## 简体中文

BotMatrix 是一个高度可扩展、分布式的跨平台机器人矩阵系统。它旨在通过统一的 OneBot v11 协议标准，将不同平台的机器人（如微信、QQ、钉钉、Discord 等）整合进一个智能化的调度网络中。

### 🌟 核心特性

- **分布式架构**: 采用 **BotNexus** (枢纽) 与 **BotWorker** (执行节点) 分离的设计，支持多机部署与负载均衡。
- **AI 智能驱动**: 深度集成大语言模型，支持自然语言创建任务、智能语义路由及自动化策略管理。
- **多平台适配**: 统一适配微信 (WxBotGo/WxBot)、QQ (TencentBot/QQGuild)、钉钉、飞书、Discord、Telegram、Slack 等主流平台。
- **多语言插件系统**: 支持使用 Go、Python、C# 等多种语言编写插件，基于标准 JSON 通信，降低开发门槛。
- **企业级管控**: 包含 RBAC 权限管理、全局拦截器、影子执行模式及跨平台身份统一映射 (NexusUID)。

### 🏗️ 系统组件

1. **[BotNexus](docs/zh-CN/components/BotNexus.md)**: 系统“大脑”，负责连接管理、消息路由、任务调度及插件分发。
2. **[BotWorker](docs/zh-CN/components/BotWorker.md)**: 执行节点，承载具体的插件运行环境与消息处理逻辑。
3. **[Overmind](docs/zh-CN/components/Overmind.md)**: 监控与管理后台，提供可视化监控、配置管理及统计分析。
4. **[SystemWorker](docs/zh-CN/components/SystemWorker.md)**: 专门负责系统级维护任务与高优先级作业。

### 📚 文档中心

- **[简体中文文档中心](docs/zh-CN/README.md)**
  - [系统架构图](docs/zh-CN/ARCHITECTURE.md)
  - [API 接口参考](docs/zh-CN/API_REFERENCE.md)
  - [部署指南](docs/zh-CN/DEPLOY.md)
  - [AI 助手指南](docs/zh-CN/development/AI_GUIDE.md)
  - [插件开发手册](docs/zh-CN/PLUGIN_DEVELOPMENT.md)
  - [国际化 (I18N) 开发指南](docs/zh-CN/development/I18N_GUIDE.md)

### 🚀 快速开始

请参考 **[部署指南](docs/zh-CN/DEPLOY.md)** 完成环境配置与系统启动。

---

## English

BotMatrix is a highly scalable, distributed, and cross-platform bot matrix system. It aims to integrate bots from various platforms (such as WeChat, QQ, DingTalk, Discord, etc.) into an intelligent scheduling network based on the unified OneBot v11 protocol standard.

### 🌟 Key Features

- **Distributed Architecture**: Separation of **BotNexus** (Hub) and **BotWorker** (Execution Node), supporting multi-machine deployment and load balancing.
- **AI-Driven Intelligence**: Deep integration with Large Language Models (LLMs), supporting natural language task creation, intelligent semantic routing, and automated policy management.
- **Multi-Platform Support**: Unified adapters for WeChat (WxBotGo/WxBot), QQ (TencentBot/QQGuild), DingTalk, Feishu, Discord, Telegram, Slack, and more.
- **Multi-Language Plugin System**: Support for writing plugins in Go, Python, C#, and other languages, based on standard JSON communication for low-barrier development.
- **Enterprise Control**: Includes RBAC permission management, global interceptors, shadow execution mode, and unified cross-platform identity mapping (NexusUID).

### 🏗️ System Components

1. **[BotNexus](docs/en-US/components/BotNexus.md)**: The "Brain" of the system, responsible for connection management, message routing, task scheduling, and plugin distribution.
2. **[BotWorker](docs/en-US/components/BotWorker.md)**: Execution nodes that host specific plugin environments and message processing logic.
3. **[Overmind](docs/en-US/components/Overmind.md)**: Monitoring and management backend, providing visual monitoring, configuration management, and statistical analysis.
4. **[SystemWorker](docs/en-US/components/SystemWorker.md)**: Dedicated to system-level maintenance tasks and high-priority jobs.

### 📚 Documentation Hub

- **[English Documentation Hub](docs/en-US/README.md)**
  - [System Architecture](docs/en-US/ARCHITECTURE.md)
  - [API Reference](docs/en-US/API_REFERENCE.md)
  - [Deployment Guide](docs/en-US/DEPLOY.md)
  - [AI Assistant Guide](docs/en-US/development/AI_GUIDE.md)
  - [Plugin Development](docs/en-US/PLUGIN_DEVELOPMENT.md)
  - [I18N Development Guide](docs/zh-CN/development/I18N_GUIDE.md)

### � Quick Start

Please refer to the **[Deployment Guide](docs/en-US/DEPLOY.md)** for environment configuration and system startup.

---

**BotMatrix Team** | 2025
