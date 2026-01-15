# EarlyMeow Worker 開發者指南 (Developer Guide)

歡迎來到 EarlyMeow Worker 開發者指南。本文檔旨在幫助開發者了解 EarlyMeow Worker 的架構、核心介面以及如何擴展其功能。

## 🏛️ 核心架構

EarlyMeow Worker 是基於 .NET 10 構建的現代化機器人處理程序，遵循 OneBot v11 協議。

### 1. 核心接口：`IRobot`
`IRobot` 是插件与系统交互的唯一入口点。它提供了以下核心功能：
- **技能注册**：`RegisterSkillAsync` 允许插件注册指令处理程序。
- **事件处理**：`RegisterEventAsync` 用于监听平台事件（如入群、退群）。
- **消息发送**：`SendMessageAsync` 支持跨平台、跨账户的消息主动发送。
- **AI 能力集成**：通过 `AI`, `Agent`, `Rag` 属性直接访问智能服务。

### 2. 插件系统：`IPlugin`
所有功能模块必须实现 `IPlugin` 接口：
- `InitAsync(IRobot robot)`: 插件初始化入口。
- `StopAsync()`: 插件卸载时的清理逻辑。
- `Intents`: 定义插件支持的意图和关键词（用于 NLP/RAG 匹配）。

### 3. 会话管理：`SessionManager`
系统内置了基于 Redis 的分布式会话管理，支持：
- 用户状态持久化。
- 跨节点的上下文同步。

## 🛠️ 如何开发一个新插件

### 第一步：创建插件类
```csharp
[BotPlugin(Id = "hello_world", Name = "示例插件")]
public class MyPlugin : IPlugin
{
    public async Task InitAsync(IRobot robot)
    {
        await robot.RegisterSkillAsync(new SkillCapability 
        {
            Name = "你好",
            Commands = ["hello", "hi"]
        }, async (ctx, args) => 
        {
            return $"你好, {ctx.SenderName}!";
        });
    }

    public Task StopAsync() => Task.CompletedTask;
}
```

### 2. 利用依赖注入 (DI)
EarlyMeow Worker 全面支持 DI。你可以在构造函数中注入 `ILogger`, `IConfiguration` 或其他系统服务。

## 🧠 AI 与 RAG 集成
EarlyMeow Worker 内置了 RAG（检索增强生成）能力：
- 插件可以通过 `robot.Rag` 检索本地知识库。
- 系统会自动将标记为 `BotPlugin` 的描述和意图向量化，供 AI 自动分发指令。

## 📂 源码布局
- `/Domain`: 核心实体、接口和消息模型。
- `/Infrastructure`: 协议实现（OneBot）、工具类和基础组件。
- `/Modules`: 核心业务模块（游戏、金融、AI）。
- `/Plugins`: 插件加载引擎与 SDK 存根。
