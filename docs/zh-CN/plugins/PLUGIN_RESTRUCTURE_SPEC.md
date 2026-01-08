# 插件系统重构技术规范 (2026)

## 1. 重构背景
为了支持多语言开发（Go, Python, JS）并统一进程间通信协议，对 `BotWorker` 的插件底层进行了深度重构。核心目标是实现“一次编写，多端分发”的插件生态，并对齐 Go 版本的核心协议。

## 2. 核心架构变更

### 2.1 统一元数据接口 (`IModuleMetadata`)
所有插件类型（Native, Process, Remote）现在必须实现统一的元数据标准：
- **Intents**: 定义插件响应的指令及关键词。
- **Permissions**: 细粒度的权限控制（如 `send_message`, `call_skill`）。
- **Events**: 声明插件监听的事件类型。

### 2.2 跨插件调用机制 (`Skill Call`)
引入了基于 `CorrelationID` 的异步调用机制：
1. **调用方**：发送 `call_skill` Action，包含目标插件 ID、技能名及参数。
2. **内核**：路由请求至目标插件。
3. **响应方**：执行逻辑并返回结果。
4. **内核**：通过 `skill_result` 事件将结果回传给调用方的 `CorrelationID`。

## 3. 通信协议规范 (JSON)

### 3.1 事件消息 (Core -> Plugin)
```json
{
  "id": "uuid",
  "type": "event",
  "name": "on_command",
  "payload": {
    "command": "fishing",
    "args": ["cast"],
    "from": "user_id"
  }
}
```

### 3.2 响应消息 (Plugin -> Core)
```json
{
  "id": "uuid",
  "ok": true,
  "actions": [
    {
      "type": "reply",
      "text": "已抛竿！"
    },
    {
      "type": "call_skill",
      "payload": { "skill": "economy.add_gold", "amount": 10 }
    }
  ]
}
```

## 4. 自动化运维
### 4.1 数据库同步
原生 C# 插件通过 `SchemaSynchronizer` 实现表结构自动同步：
- 插件启动时调用 `EnsureTablesCreatedAsync`。
- 自动读取 `MetaData<T>` 实体类的属性，生成并执行 `CREATE TABLE` SQL。
- 支持 `[PrimaryKey]` 和 `[IgnoreColumn]` 特性。

## 5. 关键文件索引
- [IPlugin.cs](file:///d:/projects/BotMatrix/src/BotWorker/Domain/Interfaces/IPlugin.cs)：协议与元数据定义基石。
- [ProcessPlugin.cs](file:///d:/projects/BotMatrix/src/BotWorker/Plugins/ProcessPlugin.cs)：Stdin/Stdout 通信实现。
- [PluginManager.cs](file:///d:/projects/BotMatrix/src/BotWorker/Plugins/PluginManager.cs)：插件生命周期与路由中心。

---
**文档状态**：已发布
**维护团队**：BotMatrix Core Team
