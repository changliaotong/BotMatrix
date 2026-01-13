# BotWorker 重构计划文档

## 1. 日志系统重构 (Logging System) - [进行中]
- **目标**: 废弃自定义混合逻辑，全面转向 Serilog 标准化集成。
- **状态**: 
    - [x] 配置 `builder.Host.UseSerilog()`。
    - [x] 移除 `AppConfig` 日志方法。
    - [x] 重构 `Logger.cs` 封装 Serilog。
    - [ ] 全局搜索并替换旧的日志调用 (ShowMessage -> Logger.Show, ErrorMessage -> Logger.Error)。

## 2. 配置管理优化 (Configuration) - [已完成]
- **目标**: 从静态常量转向 `appsettings.json` 驱动。
- **状态**:
    - [x] `AppConfig` 改为属性读取 `IConfiguration`。
    - [x] `Program.cs` 初始化 `AppConfig`。

## 3. 依赖注入 (DI) 深度应用 - [待启动]
- **目标**: 减少 `static` 类和 `new` 关键字的使用，提高可测试性。
- **执行步骤**:
    - [ ] 将 `SqlService`, `HttpHelper` 等工具类重构为非静态服务并注册到 DI。
    - [ ] 插件处理类通过构造函数注入所需服务。

## 4. 数据库访问层 (DAL) 标准化 - [待启动]
- **目标**: 统一 EF Core 与 Dapper 的使用规范。
- **执行步骤**:
    - [ ] 移除 `SQLConn.cs` 等原始 ADO.NET 操作，尽可能迁移至 EF Core 或 Dapper。
    - [ ] 统一数据库连接字符串管理。

## 5. 项目结构与代码规范清理 - [进行中]
- **目标**: 清理历史残留，统一命名空间。
- **执行步骤**:
    - [x] 建立 `GlobalUsings.cs`。
    - [ ] 彻底移除 `sz84` 命名空间残留。
    - [ ] 将文件按功能移动到 `Core`, `Infrastructure`, `Features` 等标准目录下。
