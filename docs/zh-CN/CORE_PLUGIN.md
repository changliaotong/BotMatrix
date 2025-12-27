# Nexus Core Plugin (系统级核心插件)

> [🌐 English](../en-US/CORE_PLUGIN.md) | [简体中文](CORE_PLUGIN.md)
> [⬅️ 返回文档中心](README.md) | [🏠 返回项目主页](../../README.md)

## 概述
`CorePlugin` 是集成在 `BotNexus` 消息路由层的系统级核心插件。它负责在消息转发到 Worker 模块之前，对所有原始消息进行安全性、合规性和状态性的裁决。

**注意**：这是系统级功能，直接运行在 Nexus 核心进程中，不属于任何 Worker 模块。

## 核心功能
- **消息流控制**：支持全局开启/关闭系统，以及维护模式。
- **权限裁决**：多维度的黑白名单（系统级用户、机器人、群组）。
- **内容过滤**：
    - **敏感词库**：支持明文匹配和正则表达式匹配。
    - **URL 过滤器**：防止恶意链接传播。
- **流量统计**：实时统计各类消息的处理量和拦截量。
- **管理员指令**：通过聊天界面直接控制系统行为。

## 管理员指令 (Admin Commands)
指令前缀：`/system` 或 `/nexus`

| 指令 | 参数 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| `status` | 无 | 查看系统运行状态、在线统计及今日流水 | `/system status` |
| `top` | 无 | 查看今日最活跃的用户和群组 (发言统计) | `/system top` |
| `open` | 无 | 开启系统，允许消息转发 | `/system open` |
| `close` | 无 | 关闭系统，拦截除管理员指令外的所有消息 | `/system close` |
| `whitelist` | `add <target> <id>` | 添加白名单 (`target`: `system`, `robot`, `group`) | `/system whitelist add system 123456` |
| `blacklist` | `add <target> <id>` | 添加黑名单 (`target`: `system`, `robot`, `group`) | `/system blacklist add group 789012` |
| `reload` | 无 | 从 Redis 强制重新加载最新配置 | `/system reload` |

## 统计监控
插件会将统计数据异步写入 Redis，键名格式为：
- **今日统计**: `core:stats:yyyy-mm-dd` (Hash 类型)
- **最近拦截记录**: `core:blocked:yyyy-mm-dd` (List 类型，保留最近 100 条)

可以使用 `status` 指令实时查看。

## 测试与验证
1. **状态检查**: 
   发送 `/system status`，系统应返回当前在线的 Bot 数量、Worker 数量以及消息处理统计。
2. **全局拦截验证**:
   - 发送 `/system close`。
   - 尝试发送普通聊天消息，系统应不再回复或 Worker 不再收到该消息。
   - 发送 `/system status`，指令仍应能正常执行。
   - 发送 `/system open` 恢复系统。
3. **黑名单验证**:
   - 发送 `/system blacklist add system <你的UID>`。
   - 尝试发送消息，检查 Nexus 日志，应显示 `Message blocked: ... (reason: user_blacklisted)`。

## 配置与扩展
配置信息在 `src/BotNexus/core_plugin.go` 中定义。支持多实例部署，所有状态同步通过 Redis 实现：
- **系统开关状态**: `core:system_open`
- **核心配置 JSON**: `core:config`

---
*Generated on 2025-12-22*
