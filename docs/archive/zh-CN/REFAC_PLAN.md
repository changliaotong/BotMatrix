# 机器人插件与中间件重构方案

## 1. 目标
- **解耦核心逻辑**：将 `BotMessage` 从繁杂的业务判断中解放出来。
- **增强可扩展性**：通过中间件管道处理过滤逻辑，通过插件处理业务逻辑。
- **职责清晰**：
    - **Middleware**：负责“拦截”、“过滤”、“预处理”。
    - **Plugin**：负责“功能”、“交互”、“业务”。

## 2. 核心架构：管道 (Pipeline) 模式

我们将实现一个事件处理管道，每个事件都会按顺序经过一系列中间件。

### 2.1 IMiddleware 接口
```csharp
public interface IMiddleware
{
    // 返回 true 继续执行下一个中间件，返回 false 中断处理
    Task<bool> InvokeAsync(BotMessage context, Func<Task<bool>> next);
}
```

### 2.2 管道顺序 (建议)
1. **LoggingMiddleware** (日志记录)
2. **SystemBlacklistMiddleware** (全局黑名单拦截)
3. **PowerStatusMiddleware** (开关机/权限检查)
4. **AntiSpamMiddleware** (频率限制)
5. **ContentCleaningMiddleware** (繁简转换/去广告)
6. **SensitiveWordMiddleware** (敏感词拦截)
7. **PluginMiddleware** (插件系统接入点 - 原 PluginManager)
8. **CoreBusinessMiddleware** (现有的遗留核心业务逻辑)

## 3. 功能迁移计划

### 3.1 迁移至 Middleware (中间件)
| 功能 | 现有代码位置 | 目标中间件 |
| :--- | :--- | :--- |
| 系统/群黑名单 | `HandleEventAsync` / `HandleBlackWarnAsync` | `BlacklistMiddleware` |
| 敏感词过滤 | `GroupWarn.GetEditKeyword` | `SensitiveWordMiddleware` |
| 刷屏扣分 | `HandleRefresh` | `AntiSpamMiddleware` |
| 开关机状态 | `Group.IsPowerOn` | `PowerStatusMiddleware` |
| 文本清洗 | `AsJianti` / `RemoveQqAds` | `ContentCleaningMiddleware` |

### 3.2 迁移至独立 Plugin (插件)
| 功能 | 现有代码位置 | 目标插件 |
| :--- | :--- | :--- |
| 踢人/禁言/改名 | `GetKickOutAsync` / `GetMuteResAsync` | `GroupAdminPlugin` |
| 签到系统 | `TrySignIn` | `SignInPlugin` |
| 积分系统 | `MinusCreditRes` | `PointSystemPlugin` |
| AI 智能体 | `GetAgentResAsync` | `AIPlugin` |

## 4. 重构步骤
1. **定义基础架构**：创建 `IMiddleware` 和 `MessagePipeline`。
2. **实现 PluginManager 中间件**：将当前的插件调用封装成一个中间件，放在管道中。
3. **分步迁移逻辑**：每次提取一个逻辑块（如黑名单），封装成中间件并插入管道。
4. **瘦身 BotMessage**：清理 `HandleEventAsync` 中已被迁移的代码。

## 5. 优势
- **即插即用**：可以随时通过配置文件增加或移除中间件（例如暂时关闭敏感词检测）。
- **独立开发**：插件开发者只需关注业务，无需了解底层的过滤逻辑。
- **性能优化**：在管道早期拦截黑名单用户，避免后续昂贵的数据库查询和 AI 调用。
