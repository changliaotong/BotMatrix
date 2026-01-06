# BotWorker 架构规划建议

为了方便未来扩展更多机器人的功能，建议采用 **模块化微内核架构 (Modular Micro-Kernel Architecture)**。这种架构将核心逻辑与具体业务功能（插件/技能）分离，并统一不同平台的接入方式。

## 1. 总体架构图

```mermaid
graph TD
    subgraph "接入层 (Adapters Layer)"
        A1[OneBot Adapter]
        A2[Discord Adapter]
        A3[Telegram Adapter]
        A4[HTTP/WebSocket API]
    end

    subgraph "核心层 (Core Kernel)"
        B1[事件总线 (Event Bus)]
        B2[上下文管理 (Context Manager)]
        B3[中间件管道 (Middleware Pipeline)]
        B4[插件加载器 (Plugin Loader)]
    end

    subgraph "能力层 (Capabilities/Plugins)"
        C1[通用技能 (Weather, Time)]
        C2[AI 助手 (Chat, RAG)]
        C3[自动化任务 (Playwright)]
        C4[管理工具 (Admin)]
    end

    subgraph "基础设施 (Infrastructure)"
        D1[存储 (DB/Redis)]
        D2[配置中心 (Config)]
        D3[日志/监控 (Observability)]
    end

    A1 -->|标准化事件| B1
    A2 -->|标准化事件| B1
    A3 -->|标准化事件| B1
    
    B1 --> B3
    B3 --> B2
    B2 --> B4
    B4 --> C1 & C2 & C3 & C4
    
    C1 & C2 & C3 & C4 -->|调用| D1
```

## 2. 核心模块详解

### 2.1 统一事件模型 (Unified Event Model)
不同平台（QQ, Discord, Telegram）的消息格式各异。为了让插件“写一次，到处运行”，我们需要定义一套标准化的内部事件结构。

*   **建议结构**:
    ```go
    type Event struct {
        ID        string
        Type      EventType // Message, Notice, Request
        Platform  string    // "qq", "discord", "telegram"
        Sender    User
        Channel   Channel   // 群组或私聊频道
        Content   MessageContent // 统一的消息内容（文本、图片、文件）
        Raw       interface{}    // 原始数据，供特殊需求使用
        Context   context.Context
    }
    ```

### 2.2 插件系统 (Plugin System)
插件应该是功能的最小单元。为了方便开发，建议设计简单的接口。

*   **插件接口 (Interface)**:
    ```go
    type Plugin interface {
        ID() string
        Name() string
        Init(ctx Context) error
        // 匹配器：判断是否处理该事件
        Match(event *Event) bool 
        // 处理器：实际业务逻辑
        Handle(ctx Context, event *Event) error
    }
    ```
*   **生命周期管理**: 支持热加载/卸载（如果使用 Go plugin 或 RPC 机制），或者至少支持配置化的启用/禁用。

### 2.3 中间件管道 (Middleware Pipeline)
在消息到达插件之前，通过管道进行预处理。

*   **典型中间件**:
    1.  **Recovery**: 防止 Panic 导致服务崩溃。
    2.  **Logger**: 记录请求日志。
    3.  **Auth/Permission**: 检查用户是否有权限执行命令。
    4.  **RateLimiter**: 防止刷屏。
    5.  **State**: 加载用户会话状态。

### 2.4 AI 深度集成
鉴于 `BotMatrix` 包含 AI 能力，建议将 AI 作为“第一公民”集成到核心中，而不仅仅是一个插件。

*   **AI Gateway**: 统一封装 OpenAI, Claude, Local LLM 等接口。
*   **Memory Service**: 统一管理短期对话历史（Redis）和长期记忆（Vector DB）。
*   **Intent Recognition (意图识别)**: 在分发给插件前，可以使用轻量级模型判断用户意图，直接路由到对应插件。

## 3. 目录结构建议

建议重构 `BotWorker` 的目录结构以清晰分离关注点：

```
BotWorker/
├── cmd/                # 入口文件
├── configs/            # 配置文件
├── internal/
│   ├── core/           # 核心逻辑
│   │   ├── event/      # 事件定义
│   │   ├── bus/        # 事件总线
│   │   └── kernel/     # 核心启动器
│   ├── adapters/       # 平台适配器 (OneBot, Discord...)
│   ├── plugins/        # 内置插件实现
│   │   ├── system/     # 系统级插件 (Help, Ping)
│   │   ├── ai/         # AI 相关插件
│   │   └── automation/ # 自动化插件 (Playwright)
│   ├── middleware/     # 中间件
│   └── pkg/            # 内部工具库
└── pkg/                # 可导出的公共库 (SDK)
    ├── plugin/         # 插件开发 SDK
    └── types/          # 公共类型定义
```

## 4. 实施步骤建议

1.  **定义标准接口**: 首先确立 `Event` 和 `Plugin` 的 Go 接口定义。
2.  **重构 Adapter**: 将现有的 OneBot 逻辑封装为 Adapter，确保其输出标准 `Event`。
3.  **实现总线与分发**: 完成事件从 Adapter -> Bus -> Middleware -> Plugin 的链路。
4.  **迁移现有功能**: 将目前的业务逻辑拆分为独立的 Plugin。
5.  **增强 AI 能力**: 接入 RAG 和 MCP 到新的架构中。

## 5. 优势

*   **扩展性**: 新增机器人功能只需编写新的 Plugin，无需修改核心代码。
*   **多平台支持**: 新增平台只需编写新的 Adapter。
*   **可维护性**: 职责分离，Bug 更容易定位。
*   **复用性**: 业务逻辑与平台解耦，同一套逻辑可用于所有平台。
