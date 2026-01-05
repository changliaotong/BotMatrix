# BotMatrix 📚 文档中心

> [🌐 English](../en-US/README.md) | [简体中文](README.md)
> [🏠 返回项目主页](../../README.md)

欢迎来到 BotMatrix 项目文档中心。这里包含了系统的详细架构、开发指南、部署流程以及各项功能的详细说明。

## 🗺️ 文档导航

### 🤖 AI 智能与技能中心
- **[AI 融入总方案](core/AI_INTEGRATION_PLAN.md)** - 打造 AI 原生系统的核心规划与实现路径。
- **[数字员工系统](core/DIGITAL_EMPLOYEE_SYSTEM.md)** - 将机器人进化为具备职位、部门与 KPI 的虚拟雇员。
- **[智能意图分发系统](core/INTENT_DISPATCH_SYSTEM.md)** - 详解如何通过 AI 实现精准的消息路由与 Agent 分发。
- **[技能中心与训练场](core/SKILL_CENTER.md)** - 技能的发现、分发与模型调优指南。
- **[AI 助手指南](development/AI_GUIDE.md)** - 如何利用内置 AI 能力辅助开发与日常运维。

### 🚀 快速开始 & 部署
- **[更新日志](CHANGELOG.md)** - 项目版本迭代与功能变更记录。
- **[系统架构图](core/ARCHITECTURE.md)** - 了解 BotMatrix 的核心设计思想与组件协作。
- **[部署指南](core/DEPLOY.md)** - 如何在不同环境下安装 and 运行 BotMatrix。
- **[容器化部署最佳实践](deployment/CONTAINER_BEST_PRACTICES.md)** - 在 Docker/K8s 环境下部署 BotWorker 与插件。
- **[API 接口参考](core/API_REFERENCE.md)** - 详细的 JSON 消息格式与通信协议说明。
- **[Redis 升级说明](legacy/REDIS_UPGRADE.md)** - 关于 Redis 依赖的升级和配置说明。
- **[WebUI 升级计划](legacy/WEBUI_UPGRADE_PLAN.md)** - Web 管理界面的重构与升级路线图。

### 📖 使用手册
- **[服务端使用手册](core/SERVER_MANUAL.md)** - 详细的功能说明、配置指南和管理员指令。
- **[插件功能手册](plugins/README.md)** - 所有内置插件的详细使用指南。
- **[插件 SDK 开发指南](plugins/sdk_guide.md)** - 使用 Go/Python/C# SDK 开发高效插件。
- **[插件市场规范 (BMPK)](plugins/market_spec.md)** - 标准化插件分发与分发协议。
- **[路由规则说明](core/ROUTING_RULES.md)** - 消息路由和分发逻辑的详细定义。
- **[技能开关与兼容性](core/SKILL_SWITCH_AND_COMPATIBILITY.md)** - 机器人技能的管理与 OneBot 协议兼容性说明。

### 🛠️ 核心系统说明
- **[文档冲突与改进计划](legacy/CONFLICT_PLAN.md)** - 记录已知冲突与后续文档优化方向。
- **[组件详细文档](components/README.md)** - 各个适配器与核心组件的技术细节。
- **[小程序管理后台](core/MINIPROGRAM.md)** - 移动端小程序管理平台说明。
- **[裂变系统](plugins/FISSION_SYSTEM.md)** - 自动化拉群与用户增长系统的实现。
- **[育儿与婚姻系统](plugins/BABY_AND_MARRIAGE_SYSTEM.md)** - 趣味性互动功能的逻辑说明。
- **[核心插件说明](plugins/CORE_PLUGIN.md)** - 内置核心插件的功能与配置。
- **[QQ 频道智能发送](plugins/QQ_GUILD_SMART_SEND.md)** - 针对 QQ 频道的消息优化发送方案。

### 📋 系统核心设定
- **[系统详细设定](core/SYSTEM_DOCS.md)** - 核心设定、数据库架构及近期逻辑变更。

### 💻 开发者指南
- **[开发与规划中心](development/README.md)** - 插件开发、协议兼容性及系统规划。
- **[开发指南 (编译/调试)](development/DEVELOPMENT.md)** - 快速启动、代码更新与常见问题。
- **[插件开发指南](plugins/PLUGIN_DEVELOPMENT.md)** - 如何为 BotMatrix 编写自定义插件。
- **[OneBot 协议兼容性](core/ONEBOT_COMPATIBILITY.md)** - 协议实现的细节与覆盖范围。
- **[网站规划](legacy/WEBSITE_PLAN.md)** - 官方文档网站的建设规划。

---

## 🏗️ 项目架构
1. **BotNexus**: 系统大脑，负责连接管理、路由分发和插件调度。
2. **BotWorker**: 执行节点，负责具体的任务处理和插件运行。
3. **WebUI**: 基于 Vue 3 的管理后台，提供可视化监控和配置。

---

[🏠 返回项目主页](../../README.md)
