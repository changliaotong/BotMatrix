# 组件与适配器详解 (Adapters & Components)

> [⬅️ 返回文档中心](README.md) | [🏠 返回项目主页](../../README.md)

BotMatrix 采用解耦的微服务架构，由核心管理枢纽、分布式工作节点以及多平台机器人适配器组成。本手册详细介绍了各组件的功能、架构与配置。

---

## 1. 核心基础设施 (Core Infrastructure)

### 1.1 BotNexus (系统核心枢纽)
BotNexus 是整个矩阵的“大脑”，负责全局路由、调度与监控。
- **3D 拓扑可视化**: 基于 Three.js 的实时矩阵拓扑，支持全路径路由追踪（User-Group-Nexus-Worker）。
- **智能路由分发**: 具备 RTT 感知的动态路由算法，支持精确 ID 及通配符匹配。
- **RBAC 权限模型**: 完善的用户管理体系，支持 Token 强制失效与管理员二次验证。
- **核心插件**: 集成在消息层的安全拦截器，支持敏感词过滤与黑白名单。

### 1.2 BotWorker (分布式工作节点)
BotWorker 是执行具体任务的“肢体”，兼容 OneBot v11 协议。
- **任务执行**: 承载各种功能插件（游戏、社交、工具、管理）。
- **多语言支持**: 核心采用 Go 编写，插件支持 Go, Python, JS, C#。
- **通信方式**: 同时支持 WebSocket (正向/反向) 和 HTTP 通信。

### 1.3 Overmind & SystemWorker
- **Overmind**: 提供系统监控、日志聚合与管理后台 API 服务。
- **SystemWorker**: 负责系统级后台任务（如数据库清理、定时广播、健康检查）。

---

## 2. 机器人适配器 (Bot Adapters)

适配器负责将各社交平台的原始协议转化为 BotMatrix 统一的 OneBot 消息格式。

### 2.1 微信生态 (WeChat)
- **WxBot**: 基于 Python 的 Web 协议适配器，支持“阅后即焚”撤回功能。
- **WxBotGo**: 基于 Go 的高性能微信实现，提供更稳定的 OneBot v11 接口。
- **WeComBot**: 企业微信适配器，适用于办公自动化与私域管理。

### 2.2 QQ 生态 (Tencent/NapCat)
- **TencentBot**: 基于官方 BotGo SDK 的腾讯官方 QQ 机器人。
- **NapCat**: 兼容 NTQQ 的个人 QQ 适配器，支持全功能 OneBot 11 协议。

### 2.3 办公与协作平台
- **DingTalkBot (钉钉)**: 支持 Webhook 与 Stream 模式，适合企业内部通知。
- **FeishuBot (飞书)**: 基于 WebSocket 的实时机器人，支持富媒体卡片。
- **SlackBot**: 全球流行的协作平台适配器。

### 2.4 全球社交平台
- **DiscordBot**: 针对游戏社区优化的机器人，支持频道管理。
- **TelegramBot**: 基于长轮询的高安全性机器人。
- **KookBot**: 针对国内游戏语音社区的适配器。

---

## 3. 移动端与管理界面

### 3.1 WxBotApp
为微信机器人提供的移动端辅助应用，支持远程状态查看与简单的消息处理。

### 3.2 小程序管理后台 (MiniProgram)
基于微信小程序的移动管理平台，支持：
- 实时系统状态监控 (CPU/内存/磁盘)。
- 机器人列表管理与批量控制。
- 实时流式日志查看。

---

## 4. 技术规格

| 组件 | 核心语言 | 主要通信协议 | 角色 |
| :--- | :--- | :--- | :--- |
| **BotNexus** | Go / Vue 3 | WebSocket / HTTP / gRPC | 中心管理 |
| **BotWorker** | Go / .NET | WebSocket / Redis | 任务执行 |
| **Adapters** | Go / Python | OneBot v11 (JSON) | 协议转换 |
| **Overmind** | Go | REST API | 监控中心 |

---
*最后更新日期：2026-01-13*
