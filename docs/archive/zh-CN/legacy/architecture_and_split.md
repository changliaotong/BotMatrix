# BotMatrix 插件架构与拆分最佳实践

本文档定义了 BotMatrix 插件化的标准实现方式，指导如何将业务逻辑从主程序剥离为高性能、低耦合的独立插件。

## 1. 核心接口与元数据 (Standard Interface)

所有插件必须实现 `IPlugin` 接口，并使用 `[BotPlugin]` 特性标注元数据。

### [BotPlugin] 特性说明
- **Id**: 插件唯一标识（建议格式：`category.name.vN`）。
- **Intents**: 声明插件感兴趣的意图和关键词，用于自动生成指令帮助。
- **Category**: 插件分类（如 Admin, Game, Tool）。

### IPlugin 接口要求
```csharp
public interface IPlugin
{
    // 插件初始化，在此处注册技能、创建数据库表
    Task InitAsync(IRobot robot);
    
    // 插件停止时释放资源
    Task StopAsync();
}
```

## 2. 消息处理与路由 (Message Routing)

BotMatrix 采用**中间件 (Middleware) + 意图分发**的模式：

1. **中间件拦截**：
   - 全局逻辑（如黑名单、内置指令过滤）在 `Pipeline` 中处理。
   - `BuiltinCommandMiddleware` 负责拦截已迁移到插件的指令，防止主程序重复处理。
2. **意图路由**：
   - 插件通过 `robot.RegisterSkillAsync` 注册具体指令。
   - 消息匹配到指令后，由 `PluginMiddleware` 分发至对应插件的 `HandleCommandAsync`。

## 3. 跨插件协作：技能系统 (Skill System)

插件之间严禁直接访问对方数据库。协作应通过 `IRobot` 提供的技能系统完成：

- **技能注册**：插件在 `InitAsync` 中注册可供外部调用的技能。
- **技能调用**：使用 `robot.CallSkillAsync(new Intent { Name = "..." })`。

### 示例：超级群管调用通知技能
```csharp
await _robot.CallSkillAsync(new Intent { 
    Name = "NotifyAdmin", 
    Action = "Send", 
    Parameters = new Dictionary<string, object> { { "Content", "发现违规行为" } } 
});
```

## 4. 数据库持久化标准

- **MetaDataGuid 继承**：插件模型应继承 `MetaDataGuid<T>` 以获得自动化的 ORM 支持。
- **表自动创建**：插件应在 `InitAsync` 中通过 `EnsureTablesCreatedAsync()` 确保其私有表结构存在。

## 5. 拆分步骤 (Step-by-Step)

1. **迁移模型**：在 `Modules/Games` (或对应目录) 下创建插件私有模型。
2. **实现逻辑**：创建 `Service` 类实现 `IPlugin`，将原主程序逻辑封装。
3. **注册指令**：在 `InitAsync` 中注册所有相关指令关键词。
4. **清理主程序**：
   - 在 `BuiltinCommandMiddleware` 中添加指令过滤。
   - 注销或删除 `AdminCommandHandler` 等旧的处理器。
5. **完善文档**：在 `docs/zh-CN/plugins/` 下创建独立的插件说明文档。
