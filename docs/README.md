# BotMatrix 📚 文档中心 (Documentation Hub)

欢迎来到 BotMatrix 项目文档中心。这里包含了系统的详细架构、开发指南、部署流程以及各项功能的详细说明。

## 🗺️ 文档导航 (Navigation)

### 🚀 快速开始 & 部署 (Quick Start & Deployment)
- **[系统架构图 (Architecture)](ARCHITECTURE.md)** - 了解 BotMatrix 的核心设计思想与组件协作。
- **[API 接口参考 (API Reference)](API_REFERENCE.md)** - 详细的 JSON 消息格式与通信协议说明。
- **[部署指南 (Deployment Guide)](DEPLOY.md)** - 如何在不同环境下安装和运行 BotMatrix。
- **[Redis 升级说明 (Redis Upgrade)](REDIS_UPGRADE.md)** - 关于 Redis 依赖的升级和配置说明。
- **[WebUI 升级计划 (WebUI Upgrade Plan)](WEBUI_UPGRADE_PLAN.md)** - Web 管理界面的重构与升级路线图。

### 📖 使用手册 (Manuals)
- **[服务端使用手册 (Server Manual)](SERVER_MANUAL.md)** - 详细的功能说明、配置指南和管理员指令。
- **[路由规则说明 (Routing Rules)](ROUTING_RULES.md)** - 消息路由和分发逻辑的详细定义。
- **[技能开关与兼容性 (Skills & Compatibility)](SKILL_SWITCH_AND_COMPATIBILITY.md)** - 机器人技能的管理与 OneBot 协议兼容性说明。

### 🛠️ 核心系统说明 (Core Systems)
- **[文档冲突与改进计划](CONFLICT_PLAN.md)** - 记录已知冲突与后续文档优化方向。
- **[裂变系统 (Fission System)](FISSION_SYSTEM.md)** - 自动化拉群与用户增长系统的实现。
- **[育儿与婚姻系统 (Marriage System)](BABY_AND_MARRIAGE_SYSTEM.md)** - 趣味性互动功能的逻辑说明。
- **[核心插件说明 (Core Plugins)](CORE_PLUGIN.md)** - 内置核心插件的功能与配置。
- **[QQ 频道智能发送 (QQ Guild Smart Send)](QQ_GUILD_SMART_SEND_CN.md)** - 针对 QQ 频道的消息优化发送方案。

### 💻 开发者指南 (Developer Guide)
- **[插件开发指南 (Plugin Development)](../PLUGIN_DEVELOPMENT.md)** - 如何为 BotMatrix 编写自定义插件。
- **[OneBot 协议兼容性 (OneBot Compatibility)](../ONEBOT_COMPATIBILITY.md)** - 协议实现的细节与覆盖范围。
- **[网站规划 (Website Plan)](website_plan.md)** - 官方文档网站的建设规划。

---

## 🏗️ 项目架构 (Architecture)
*(此处建议后续补充架构图)*

BotMatrix 由以下核心组件组成：
1. **BotNexus**: 系统大脑，负责连接管理、路由分发和插件调度。
2. **BotWorker**: 执行节点，负责具体的任务处理和插件运行。
3. **WebUI**: 基于 Vue 3 的管理后台，提供可视化监控和配置。

---

[⬅️ 返回项目主页](../README.md)
