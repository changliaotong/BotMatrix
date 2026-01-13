# 💓 System Pulse (系统脉动) 审计系统

## 1. 系统概述

**System Pulse** 是 BotMatrix 的“神经监测系统”。它通过监听 `SystemAuditEvent`，实时追踪系统内部的关键逻辑变更，并将这些变更以可视化的方式反馈给管理员或终端用户。

## 2. 核心机制

### 环形日志队列 (Ring Buffer)
在 `EventNexus` 中维护了一个最大容量为 50 的并发安全队列 (`ConcurrentQueue`)。
- **实时性**：一旦事件发布，监控视图立即同步。
- **低开销**：只保留最近 50 条记录，不持久化到数据库，极大地降低了 I/O 压力。

### 审计分级
| 级别 | 图标 | 触发场景 |
| :--- | :--- | :--- |
| **Success** | ✅ | 位面晋升、任务完成、大奖抽取 |
| **Info** | ℹ️ | 常规系统操作、心跳包 |
| **Warning** | ⚠️ | 大额交易、异常操作行为 |
| **Critical** | 🚨 | 数据库异常、核心服务中断 |

---

## 3. 技术落地

### 事件定义
```csharp
public class SystemAuditEvent : BaseEvent {
    public string Level { get; set; }
    public string Source { get; set; }
    public string Message { get; set; }
}
```

### 触发示例 (以进化系统为例)
当用户升级时，`EvolutionService` 会发布一条审计消息：
```csharp
await robot.Events.PublishAsync(new SystemAuditEvent {
    Level = "Success",
    Source = "Evolution",
    Message = "用户 12345 晋升位面: 协议"
});
```

---

## 4. 用户交互

用户通过 **超级菜单 -> 💓 系统脉动** 即可进入实时看板。看板展示最近 15 条关键审计日志，每条日志包含：
- 时间戳 (HH:mm:ss)
- 严重程度图标
- 详细业务描述
