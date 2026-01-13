# Member & Group 配置重构指南 (Member & Group Setup Refactor Guide)

## 1. 背景与目标 (Background & Objectives)
为了实现管理权上移至 BotNexus (Go)，需要重构机器人（Member）和群组（Group）的配置接口。
目标是确保 Go 端接口能正确处理自有机器人与官方机器人的差异，并保持与 C# (BotWorker) 逻辑的权限一致性。

## 2. 数据模型映射 (Data Model Mapping)

### 2.1 机器人配置 (Member / BotInfo)
- **数据库表**: `Member`
- **Go 模型**: `models.BotInfo` ([sz84_models.go](file:///c:/Users/彭光辉/projects/BotMatrix/src/Common/models/sz84_models.go))
- **C# 模型**: `BotWorker.Domain.Entities.BotInfo` ([BotInfo.cs](file:///c:/Users/彭光辉/projects/BotMatrix/src/BotWorker/Domain/Entities/BotInfo.cs))
- **关键字段**:
    - `BotUin`: 唯一标识。
    - `BotType`: 机器人类型（0: Mirai, 1: QQ, 2: Weixin, 3: Public, 等）。
    - `AdminId`: 所属管理员 ID。
    - `IsCredit`: 是否开启积分系统。

### 2.2 群组配置 (Group / GroupInfo)
- **数据库表**: `Group`
- **Go 模型**: `models.GroupInfo` ([sz84_models.go](file:///c:/Users/彭光辉/projects/BotMatrix/src/Common/models/sz84_models.go))
- **C# 模型**: `BotWorker.Domain.Entities.GroupInfo` ([GroupInfo.cs](file:///c:/Users/彭光辉/projects/BotMatrix/src/BotWorker/Domain/Entities/GroupInfo.cs))
- **权限字段说明**:
    - `RobotOwnerName` (Go) / `RobotOwner` (C#): 机器人主人的标识。
    - `GroupOwnerName` (Go) / `GroupOwner` (C#): 群主人的标识。
    - **注意**: Go 端使用 `Name` 后缀的字段通常是 C# 端对应 ID 的字符串表示或关联名称，需确保逻辑一致。

## 3. 权限体系 (Permission Model)
### 3.1 核心权限校验逻辑
在 `handlers_setup.go` 中，所有接口均已接入基于 JWT Claims 的权限校验：
- **RobotOwner (机器人主人)**: `claims.UserID` (Int64) 匹配 `Member.AdminId`。
- **GroupOwner (群组主人)**: `claims.Username` (String) 匹配 `GroupInfo.GroupOwnerName` 或 `GroupInfo.RobotOwnerName`。
- **Admin (管理员)**: `claims.IsAdmin` 为 `true`，可跨过所有所有权校验。

## 4. 机器人类型差异逻辑 (Bot Type Logic)
- **自有机器人 (Member Bot)**: 用户自己接入的机器人，通常具有完整的配置权限。
- **官方机器人 (Proxy/Official Bot)**: 由系统统一提供的机器人，用户仅能配置部分群组插件功能，不能更改核心协议参数。
- **判断依据**: 通过 `BotType` 和 `AdminId` 进行区分。

## 5. API 接口设计 (API Design)

### 5.1 接口变动与保护 (位于 `handlers_setup.go`)
- `GET /admin/member/setup`: 现由 `JWTMiddleware` 保护。普通用户仅能查询自己作为 `AdminId` 的机器人列表。
- `PUT /admin/member/setup`: 现由 `JWTMiddleware` 保护。普通用户仅能更新自己作为 `AdminId` 的机器人配置。
- `GET /admin/group/setup`: 现由 `JWTMiddleware` 保护。普通用户仅能查询自己作为 `RobotOwnerName` 或 `GroupOwnerName` 的群组。
- `PUT /admin/group/setup`: 现由 `JWTMiddleware` 保护。普通用户仅能更新自己作为 `RobotOwnerName` 或 `GroupOwnerName` 的群组配置。

## 6. 开发进度与计划 (Roadmap)
- [x] 完成 `handlers_setup.go` 基础框架实现。
- [x] 文档化重构方案与逻辑映射。
- [x] 在 `handlers_setup.go` 中加入基于 JWT 的 `RobotOwner` 权限过滤逻辑。
- [ ] **待办**: 验证 `RobotOwnerName` 在 Go 和 C# 之间的数据同步正确性。
- [ ] **待办**: 更新 WebUI 前端调用，对接新接口。
- [ ] **待办**: 分析并整理自定义机器人与官方机器人编号配置差异的逻辑。
