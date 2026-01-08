# 多轮对话与危险操作确认开发指南 (C#)

为了支持分布式多 Worker 架构，BotWorker 采用基于 Redis 的无状态会话管理。这意味着任何一个 Worker 实例都能通过用户的 `(GroupId, UserId)` 恢复对话上下文。

## 1. 核心概念

### 会话状态 (UserSession)
存储在 Redis 中的上下文信息，包含：
- `PluginId`: 当前锁定该会话的插件。
- `Action`: 当前正在执行的操作标识。
- `Step`: 当前对话所处的步骤（可选）。
- `ConfirmationCode`: 待验证的确认码（仅限确认流程）。
- `DataJson`: 插件自定义的持久化数据。

## 2. 危险操作确认流程

当插件执行删除、清空等高危操作时，应强制进入确认流程。

### 实现步骤：
1. **发起确认**：在插件 Handler 中调用 `StartConfirmationAsync`。
2. **拦截验证**：系统自动拦截用户下一条消息。如果匹配验证码，设置 `ctx.IsConfirmed = true` 并再次回调插件。
3. **执行操作**：插件检查 `ctx.IsConfirmed`，若为 `true` 则执行逻辑。

### 示例代码：
```csharp
public async Task<string> HandleClearBlacklist(IPluginContext ctx, string[] args)
{
    // 情况 A：用户刚刚输入了正确的验证码
    if (ctx.IsConfirmed && ctx.SessionAction == "ClearBlacklist")
    {
        await _db.ClearBlacklistAsync();
        return "✅ 黑名单已成功清空。";
    }

    // 情况 B：发起确认请求
    var code = await _robot.Sessions.StartConfirmationAsync(
        ctx.UserId, ctx.GroupId, "my.plugin.id", "ClearBlacklist");
        
    return $"⚠️ 正在尝试清空黑名单，请输入验证码【{code}】确认，或发送“取消”退出。";
}
```

## 3. 多轮对话 / 多级菜单

用于处理需要用户多次输入信息的复杂指令。

### 实现步骤：
1. **启动对话**：调用 `StartDialogAsync` 并指定 `step`。
2. **状态流转**：插件通过 `ctx.SessionStep` 判断当前进度。
3. **完成或继续**：更新 `step` 进入下一步，或调用 `ClearSessionAsync` 结束对话。

### 示例代码：
```csharp
public async Task<string> HandleOrderMusic(IPluginContext ctx, string[] args)
{
    // 步骤 2：处理收到的歌名
    if (ctx.SessionAction == "OrderMusic" && ctx.SessionStep == "WaitSongName")
    {
        var songName = ctx.Message;
        await _robot.Sessions.ClearSessionAsync(ctx.UserId, ctx.GroupId);
        return $"🎵 已为您点播歌曲：{songName}";
    }

    // 步骤 1：发起点歌请求
    await _robot.Sessions.StartDialogAsync(
        ctx.UserId, ctx.GroupId, "music.plugin", "OrderMusic", "WaitSongName");
        
    return "🎸 请输入您想听的歌名：";
}
```

## 4. 用户交互指令

在任何会话状态下，用户发送以下指令具有特殊含义：
- **“取消”**：系统将自动清除 Redis 中的会话状态，并回复“✅ 已取消当前操作”。插件无需手动处理。

## 5. 最佳实践

- **超时控制**：默认会话有效期较短（确认码 60秒，对话 300秒），过期后 Redis 自动删除，用户需重新发起。
- **无状态设计**：不要在插件类中使用成员变量存储用户信息，必须使用 `ctx.SessionData` 或 `SetSessionAsync` 中的 `data` 参数。
- **插件隔离**：系统通过 `PluginId` 确保消息只会路由回发起会话的插件。

## 6. 意图识别 (Intent Recognition)

意图识别允许插件处理非固定格式的自然语言输入。它在直接指令匹配（Commands）失败后触发。

### 配置方式
在注册 Skill 时，填充 `Intents` 列表：

```csharp
await robot.RegisterSkillAsync(new SkillCapability
{
    Name = "Weather",
    Intents = new List<Intent>
    {
        new Intent { 
            Name = "QueryWeather", 
            Regex = ".*天气.*", 
            Keywords = new[] { "气温", "下雨" } 
        }
    }
}, HandleWeather);
```

### 优先级
1. **活跃会话** (Session Interception)
2. **直接指令** (Command Prefix)
3. **意图匹配** (Regex & Keywords)
4. **AI 语义识别** (LLM Fallback - 规划中)
