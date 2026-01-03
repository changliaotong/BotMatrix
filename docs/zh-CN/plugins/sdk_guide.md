# BotMatrix 插件 SDK 指南 (v3.0)

> [简体中文](sdk_guide.md) | [🌐 English](../../en-US/plugins/sdk_guide.md)
> [⬅️ 返回文档中心](../README.md) | [🏠 返回项目主页](../../../README.md)

欢迎使用 BotMatrix 插件 SDK。本 SDK 支持 **Go**, **Python**, 和 **C#** 三种主流语言，采用统一的架构设计，旨在提供高性能、高交互、分布式友好的插件开发体验。

---

## 核心特性

- **交互式 Ask**: 像写同步代码一样编写异步对话流。
- **智能路由**: 支持字符串前缀、正则表达式匹配指令。
- **基于意图的路由 (Intent-based Routing)**: 支持基于自然语言意图的路由（如：不需要指令前缀即可触发）。
- **跨插件调用 (Plugin Interop)**: 允许插件间导出 Skill 并进行安全调用。
- **Correlation ID**: 原生支持分布式环境下的消息追踪，解决无状态 Worker 状态丢失问题。
- **中间件系统**: 支持中间件扩展，方便实现日志、权限校验等横切关注点。
- **逃生舱接口 (Escape Hatch)**: 预留底层 Action 调用接口，支持 Core 协议的无缝升级。
- **开发工具链**: 提供 `bm-cli` 命令行工具，一键生成多语言插件模板。

---

## 插件开发 CLI 工具 (bm-cli)

为了加速开发，我们提供了 `bm-cli` 工具。

### 安装 (Go)
```bash
go build -o bm-cli src/tools/bm-cli/main.go
```

### 快速创建插件
```bash
# 创建一个 Go 插件
./bm-cli init MyWeatherPlugin --lang go

# 创建一个 Python 插件
./bm-cli init MyAIAgent --lang python
```

### 打包插件 (.bmpk)
使用 `pack` 命令将插件目录打包成标准分发格式：
```bash
./bm-cli pack ./MyWeatherPlugin
```
这将生成 `com.botmatrix.myweatherplugin_1.0.0.bmpk` 文件，可用于在 BotNexus 中进行热安装。

---

## 插件配置与声明 (plugin.json)

每个插件必须包含一个 `plugin.json` 文件，用于声明插件的元数据、入口点、意图以及所需的权限。

```json
{
  "id": "com.example.weather",
  "name": "天气助手",
  "version": "1.0.0",
  "author": "BotMatrix Team",
  "description": "提供全球天气查询服务",
  "entry": "python main.py",
  "permissions": [
    "send_message",
    "send_image"
  ],
  "events": [
    "on_message"
  ],
  "intents": [
    {
      "name": "get_weather",
      "keywords": ["天气", "气温", "下雨吗"],
      "priority": 10
    }
  ]
}
```

### 意图识别 (Intent Routing)
通过在 `intents` 中配置关键字，插件可以实现自然语言触发：
- **Python**: `@app.on_intent("get_weather")`
- **Go**: `p.OnIntent("get_weather", handler)`

---

## 跨插件能力调用 (Skill Call)

插件 A 可以调用插件 B 导出的 Skill。这是通过 Core 层的 IPC 机制实现的。

### 导出能力 (Plugin B)
- **C# (强类型)**:
  ```csharp
  p.ExportSkill<TransferRequest, bool>("transfer", async (ctx, req) => {
      // 业务逻辑...
      return true; 
  });
  ```
- **Go**: `p.ExportSkill("check_stock", handler)`
- **Python**: `@app.export_skill("check_stock")`

### 调用能力 (Plugin A)
- **C# (异步强类型)**:
  ```csharp
  var result = await ctx.CallSkillAsync<bool>("plugin_b_id", "transfer", new { amount = 100 });
  ```
- **Go**: `ctx.CallSkill("plugin_b_id", "check_stock", payload)`
- **Python**: `await ctx.call_skill("plugin_b_id", "check_stock", payload)`

---

## UI 扩展 (UI Components)

插件可以通过声明 `ui` 字段向 BotMatrix WebUI 注入自定义界面组件。

### 配置示例
```json
{
  "ui": [
    {
      "type": "panel",
      "position": "sidebar",
      "title": "天气面板",
      "icon": "cloud",
      "entry": "./web/index.html"
    }
  ]
}
```

### 支持的类型与位置
- **Type**: `panel` (侧边栏面板), `button` (按钮), `tab` (标签页)
- **Position**: `sidebar` (全局侧边栏), `dashboard` (仪表盘), `chat_action` (聊天窗口工具栏)

---

## 权限安全机制

1. **白名单模式**：只有在 `permissions` 中声明的动作（Action），SDK 才会允许执行。
2. **SDK 端拦截**：如果你在代码中尝试调用未声明的动作（如 `ctx.KickUser()`），SDK 会直接拦截并报错，不会发送给核心系统。
3. **Core 端校验**：核心系统也会根据此配置进行二次校验，确保插件安全。

---

## 分布式状态存储 (SessionStore)

在分布式（无状态）环境下，插件可以使用 SDK 提供的 `SessionStore` 来存取持久化数据。

### 使用场景
- 用户登录态、对话上下文缓存。
- 跨 BotWorker 实例的数据同步。
- 复杂工作流的中间状态记录。

### 示例 (Python)
```python
# 存储数据
await ctx.session.set("user_level", 10, expire=3600)

# 获取数据
level = await ctx.session.get("user_level")
```

### 示例 (Go)
```go
// 存储数据
ctx.Session.Set("last_query", time.Now(), 24 * time.Hour)

// 获取数据
var lastTime time.Time
ctx.Session.Get("last_query", &lastTime)
```

---

## 快速开始代码示例

请参考 [插件开发指南](../PLUGIN_DEVELOPMENT.md) 获取更详细的代码示例。

---

## 关键 API 说明

### Context (上下文)

每个 Handler 都会接收到一个 `Context` 对象，它包含了当前事件的所有信息及回复方法。

| 方法/属性 | 描述 |
| :--- | :--- |
| `ctx.Reply(text)` | 快速回复文本消息给发送者/群组。 |
| `ctx.Ask(prompt, timeout)` | **核心**：发送提示语并阻塞等待该用户的下一条回复。 |
| `ctx.UserId` / `ctx.GroupId` | **C# 新增**: 快速获取发送者和群组 ID。 |
| `ctx.CallSkillAsync<T>` | **C# 新增**: 异步调用其他插件的 Skill 并获取强类型结果。 |
| `ctx.Args` | 指令后面的参数数组（按空格拆分）。 |
| `ctx.CorrelationId` | 当前会话的唯一标识，用于分布式追踪。 |

---
*最后更新: 2026-01-01*
