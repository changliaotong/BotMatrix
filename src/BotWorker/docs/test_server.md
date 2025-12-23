# 测试服功能插件文档

## 功能概述

测试服功能插件允许用户选择体验机器人的新功能，在测试服环境中提前尝试尚未正式发布的功能。用户可以自由切换测试服状态，并查看相关说明。

## 核心功能

- **测试服切换**：用户可以自由开启或关闭测试服功能
- **状态查询**：用户可以随时查看自己的测试服状态
- **使用说明**：提供详细的测试服功能说明
- **新功能体验**：在测试服环境中体验机器人的新功能

## 支持的命令

| 命令 | 功能描述 | 权限要求 |
|------|----------|----------|
| 切换测试服 | 切换测试服状态（开启/关闭） | 普通用户 |
| 测试服状态 | 查询当前测试服状态 | 普通用户 |
| 测试服说明 | 查看测试服功能的详细说明 | 普通用户 |

## 数据模型

### 用户测试服状态

```go
type UserTestServerStatus struct {
    UserID       string    `json:"user_id"`
    Enabled      bool      `json:"enabled"`
    CreatedAt    time.Time `json:"created_at"`
    LastUpdatedAt time.Time `json:"last_updated_at"`
}
```

## 使用方法

### 1. 切换测试服状态

发送 `切换测试服` 命令可以开启或关闭测试服功能：

- 如果当前未启用测试服，命令执行后会开启测试服
- 如果当前已启用测试服，命令执行后会关闭测试服

### 2. 查询测试服状态

发送 `测试服状态` 命令可以查看当前的测试服状态：

- 返回当前是否处于测试服状态
- 显示测试服的相关信息

### 3. 查看测试服说明

发送 `测试服说明` 命令可以查看测试服功能的详细说明：

- 包含测试服的功能介绍
- 提供使用指南和注意事项

## 实现细节

### 插件架构

测试服插件实现了 `plugin.Plugin` 接口，包含以下方法：

```go
type TestServerPlugin struct {
    db        *sql.DB
    robot     plugin.Robot
    cmdParser *CommandParser
}

func (p *TestServerPlugin) Name() string { return "TestServerPlugin" }
func (p *TestServerPlugin) Description() string { return "测试服功能插件" }
func (p *TestServerPlugin) Version() string { return "1.0.0" }
func (p *TestServerPlugin) Init(robot plugin.Robot) {
    // 初始化逻辑
}
```

### 数据库集成

测试服功能使用全局数据库连接 `GlobalDB` 进行数据持久化：

- 用户测试服状态存储在 `user_test_server_status` 表中
- 确保用户状态在机器人重启后仍然保留

### 命令处理

测试服插件实现了以下命令处理逻辑：

```go
func (p *TestServerPlugin) handleMessage(robot plugin.Robot, event *onebot.Event) {
    // 处理测试服相关命令
    switch event.Message {
    case "切换测试服":
        p.toggleTestServerStatus(robot, event, event.UserID)
    case "测试服状态":
        p.getTestServerStatus(robot, event, event.UserID)
    case "测试服说明":
        p.showTestServerHelp(robot, event)
    }
}
```

## 测试服功能特点

1. **用户自主选择**：用户可以根据自己的需求选择是否使用测试服
2. **数据隔离**：测试服功能与正式服功能数据隔离，不影响正式使用
3. **提前体验**：用户可以提前体验机器人的新功能
4. **随时切换**：用户可以随时切换测试服状态
5. **详细说明**：提供完整的测试服使用说明和帮助

## 注意事项

1. 测试服功能可能包含尚未完全稳定的新功能
2. 在使用测试服功能时，可能会遇到一些问题或异常
3. 测试服数据可能会在功能更新时被重置
4. 如果遇到问题，可以关闭测试服并使用正式服功能
5. 建议仔细阅读测试服说明后再使用相关功能

## 插件版本信息

- 版本：1.0.0
- 开发者：BotMatrix Team
- 最后更新：2025-12-23
- 更新内容：
  - 实现测试服功能的核心逻辑
  - 支持测试服状态切换
  - 提供测试服状态查询和说明
  - 优化用户体验
