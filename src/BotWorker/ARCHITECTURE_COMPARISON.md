# 现有架构 vs. 规划架构对比分析

通过对现有代码 (`BotWorker/internal`) 的分析，以下是现有实现与规划架构的详细对比。

## 1. 现有实现的优缺点

### 优点 (Pros)
*   **热重载机制成熟**: `PluginBridge` 已经实现了基于 `fsnotify` 的文件监听和“影子拷贝 (Shadow Copy)”机制，支持插件的热更新，这对开发体验非常友好。
*   **基础功能完备**: 已经集成了 `BaseBot` (HTTP/WebSocket), `Config` 加载, `LogManager` 等基础设施。
*   **数据库/Redis 集成**: 已经有了 `db` 和 `redis` 的封装，并且在处理事件时（如 `EnsureIDs`）自动处理了用户/群组 ID 的持久化和映射。

### 缺点 (Cons)
*   **与 OneBot/QQGuild 强耦合**:
    *   事件处理逻辑 (`internal/onebot/event.go`, `internal/server/event_helper.go`) 中包含大量 `if e.Platform != "qqguild"` 这样的硬编码判断。
    *   事件结构体直接使用了 `onebot.Event` 的别名，没有抽象出通用的事件层。如果要接入 Discord 或 Telegram，需要修改大量核心代码。
*   **缺乏统一的事件总线**:
    *   目前的事件处理流程比较线性且分散（在 Server 中接收 -> EnsureIDs -> 插件调用）。缺乏一个中心化的 Event Bus 来进行消息的广播、拦截和中间件处理。
*   **插件加载逻辑不够通用**:
    *   `LoadInternalPlugins` 中硬编码了 `AIPlugin` 的加载逻辑。
    *   内外部插件的处理逻辑混合在 `PluginBridge` 中，职责不够单一。
*   **中间件缺失**:
    *   目前没有看到明显的“中间件管道”模式。鉴权、限流、日志等逻辑可能散落在各个处理函数中，难以复用和统一管理。

## 2. 规划架构的改进点

### 改进 1: 解耦平台差异 (Adapters Layer)
*   **现状**: 代码里到处是 `qqguild` 的特判。
*   **规划**: 引入 `Adapter` 模式。OneBot, Discord, Telegram 各自作为一个 Adapter，负责将各自的协议转换为系统统一的 `Event` 结构。核心逻辑不再感知具体平台。

### 改进 2: 统一事件模型 (Unified Event Model)
*   **现状**: 使用 `onebot.Event` 作为核心数据结构。
*   **规划**: 定义一个与平台无关的 `core.Event` 结构。包含 `Sender`, `Channel`, `Content` 等通用字段。
    *   *好处*: 编写一个“天气插件”，可以同时服务于 QQ 群和 Discord Channel，无需修改代码。

### 改进 3: 中间件管道 (Middleware Pipeline)
*   **现状**: 缺乏统一的预处理机制。
*   **规划**: 在事件分发给插件前，经过 `Auth`, `RateLimit`, `Logger`, `Context` 等中间件链。
    *   *好处*: 可以在不修改业务插件的情况下，全局增加“禁止刷屏”或“黑名单”功能。

### 改进 4: AI 能力下沉 (AI Kernel)
*   **现状**: AI 服务作为一个普通 Service 传递。
*   **规划**: 将 AI 能力（LLM 调用、RAG 检索、上下文记忆）作为核心内核的一部分，提供标准接口供所有插件调用。
    *   *好处*: 任何插件（比如“待办事项插件”）都可以轻松获得 AI 的自然语言理解能力，而不需要自己去调 API。

## 3. 总结

现有的代码是一个**“能跑的、针对特定场景优化（OneBot/QQ）的实现”**，而规划的架构是一个**“面向未来、多平台、高扩展的框架”**。

如果目标仅仅是维护现有的 QQ 机器人，现有代码足够。但既然目标是**“方便实现更多机器人的功能”**，那么进行架构重构是必须的，主要工作量在于**抽象层的建立**和**现有逻辑的迁移**。
