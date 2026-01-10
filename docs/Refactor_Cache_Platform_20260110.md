# 代码重构与缓存优化文档 (2026-01-10)

## 1. BotWorker 缓存机制重构 (MetaData.cs)

### 1.1 核心逻辑优化
- **ID 提取增强**：重构了 `ExtractIdsFromWhere` 方法，增加了对 `IN` 子句（如 `id IN ({0}, {1})`）的支持。通过正则表达式更精准地从 SQL `WHERE` 子句中提取主键 ID，用于精准失效缓存。
- **标准化缓存同步**：统一了 `SyncCacheField` 方法，并新增了 `SyncCacheFieldAsync` 异步版本。现在所有字段级更新都会同步失效对应的行级缓存，确保数据的一致性。
- **关键字段缓存绕过**：在 `GetAsync<T>` 中针对 `IsPowerOn`、`UseRight`、`IsOpen`、`IsGroup`、`IsPrivate` 等权限相关字段增加了缓存绕过逻辑。这些字段的读取将直接访问数据库，以保证权限校验的绝对实时性。

### 1.2 调试与维护性提升
- **详细日志注入**：在 `MetaData.cs` 的核心方法（`GetCached`、`GetAsync`、`GetValueAsync`、`InvalidateCache` 等）中增加了标准化的调试日志，格式如 `[Cache] Hit/Miss`、`[DB] Loading`、`[CacheSync] Updated`，极大地方便了线上问题的排查。
- **修复编译缺陷**：修正了静态方法中错误访问实例属性（`KeyField`）的问题，统一改用静态属性 `Key` 和 `Key2`。

---

## 2. 平台标识标准化与代码清理

### 2.1 平台标识统一
- **大小写对齐**：根据 BotNexus 的映射要求，将 `Platforms.QQ` 的常量值从 `"QQ"` 修改为小写的 `"qq"`。
- **废弃 NapCat 引用**：移除了代码中所有 `IsNapCat` 相关的逻辑，并统一替换为 `IsQQ`。受影响的文件包括 `BotMessageExt.cs`、`HandleEventMessag.cs`、`IUserService.cs` 以及多个消息处理类。

### 2.2 修复 UserTokens 构建错误
- **重构 GetTokensListAsync**：修复了 `UserTokens.cs` 中导致死循环的递归调用问题。
- **类型转换修复**：解决了枚举类型 `BotData.Platform` 无法隐式转换为 `int` 的编译错误，通过显式转换和重构 SQL 查询逻辑确保了代码的健壮性。

---

## 3. BotNexus 兼容性修复

### 3.1 机器人查找逻辑
- **Key 格式统一**：修改了 BotNexus `handlers.go` 中的机器人注册逻辑，将 Key 格式从单纯的 `self_id` 变更为 `platform:self_id`。
- **默认值处理**：在 WebSocket 握手阶段，若未指定 `X-Platform`，则默认标识为 `qq`，确保了与 BotWorker 下发 Action 时的查找逻辑完全匹配。

---

## 4. 消息管线优化

### 4.1 VipMiddleware 行为调整
- **静默拦截**：对于未授权（非 VIP 或超出限制）的请求，现在通过设置 `botMsg.IsSend = false` 实现静默拦截，不再向用户发送干扰性的报错消息，提升了用户体验。

### 4.2 发送日志增强
- **SendMessage 追踪**：在 `SendMessage.cs` 中增加了对 `ReplyMessageAsync` 委托状态的检查和日志记录，用于定位消息发送环节中的空引用或逻辑中断问题。
