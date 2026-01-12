# 智能化 TODO 系统技术文档

## 1. 概述
本系统旨在通过增强原有的待办事项功能，打造一个支持开发（Dev）、测试（Test）进度跟踪及通用任务管理的智能化系统。支持优先级设定、进度百分比管理、详细描述及分类查询。

## 2. 功能特性
- **分类管理**: 任务可归类为 `Todo` (默认), `Dev` (开发), `Test` (测试)。
- **优先级系统**: 支持 `P1` (High), `P2` (Medium), `P3` (Low) 三级优先级。
- **进度跟踪**: 支持 0-100% 的数值进度更新，并自动同步任务状态。
- **详细描述**: 支持为任务附加长文本描述。
- **智能查询**: 支持按关键词、分类（如 `dev`）、优先级（如 `p1`）进行过滤查询。

## 3. 命令指南
### 3.1 新增任务
语法: `todo + 内容 [分类] [优先级]`
- `todo + 重构 AdminService Dev P1` -> 创建一个开发类的高优先级任务。
- `todo + 准备周报 P2` -> 创建一个普通优先级任务。

### 3.2 更新任务
语法: `todo #编号 [值/done/desc 内容]`
- `todo #1 50` -> 将任务 #1 的进度设为 50% (状态自动变为 InProgress)。
- `todo #1 done` -> 完成任务 #1 (进度设为 100%，状态变为 Completed)。
- `todo #1 P1` -> 将任务优先级改为 High。
- `todo #1 desc 修复了同步死锁问题` -> 更新详细描述。

### 3.3 查询任务
语法: `todo [关键词/分类/优先级]`
- `todo` -> 列出最近的 5 条待办。
- `todo dev` -> 只看开发类任务。
- `todo p1` -> 只看高优先级任务。
- `todo #1` -> 查看任务 #1 的详细信息。

## 4. 技术实现
### 4.1 数据模型
- **C#**: [Todo.cs](file:///d:/projects/BotMatrix/src/BotWorker/Infrastructure/Tools/Todo.cs)
- **Go**: [gorm_models.go](file:///d:/projects/BotMatrix/src/Common/models/gorm_models.go)

### 4.2 数据库变更
需手动执行以下 SQL 脚本：
```sql
ALTER TABLE [dbo].[Todo] ADD [Description] NVARCHAR(MAX) DEFAULT '' NOT NULL;
ALTER TABLE [dbo].[Todo] ADD [Priority] NVARCHAR(50) DEFAULT 'Medium' NOT NULL;
ALTER TABLE [dbo].[Todo] ADD [Progress] INT DEFAULT 0 NOT NULL;
ALTER TABLE [dbo].[Todo] ADD [Status] NVARCHAR(50) DEFAULT 'Pending' NOT NULL;
ALTER TABLE [dbo].[Todo] ADD [Category] NVARCHAR(50) DEFAULT 'Todo' NOT NULL;
ALTER TABLE [dbo].[Todo] ADD [DueDate] DATETIME NULL;
```

## 5. 提交记录
- 修改 `BotWorker/Infrastructure/Tools/Todo.cs` 以实现智能化逻辑。
- 修改 `Common/models/gorm_models.go` 以同步数据模型。
- 新增技术文档 `docs/zh-CN/SMART_TODO_SYSTEM.md`。
