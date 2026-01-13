# BotMatrix 开发文档 (Development Notes)
[English](../../en-US/development/BotWorker_DEV.md) | [返回项目主页](../../../README.md) | [返回文档中心](../README.md)

本文档记录了项目中的关键设计决策、架构调整以及针对 QQGuild 平台的特殊处理。

## 1. ID 设计与自动生成策略

为了兼容 QQGuild 平台并支持多平台扩展，项目采用了以下 ID 设计：

- **数据类型**: 数据库中所有的 `user_id` 和 `group_id` 均使用 `BIGINT` (Go 中对应 `int64`)。
- **自动生成范围**:
  - **用户 ID (UserID)**: 从 `980000000000` 开始自增。
  - **群组 ID (GroupID)**: 从 `990000000000` 开始自增。
- **查询逻辑**: 每次生成新 ID 时，会查询当前表中最大的 ID 并加 1。如果表为空或最大 ID 小于起点值，则从起点开始。

## 2. QQGuild 平台集成

QQGuild 平台的 ID (OpenID/TinyID) 通常是长字符串或大数字，处理逻辑如下：

- **SelfId 处理**: 机器人自身的 `SelfId` 被视为一个特殊用户，其 ID 也会被映射到 `users` 表中的 `user_id`。
- **关联字段**:
  - `users` 表增加了 `target_user_id` (BIGINT) 和 `user_openid` (VARCHAR) 列。
  - `groups` 表增加了 `target_group_id` (BIGINT) 和 `group_openid` (VARCHAR) 列。
- **映射逻辑**:
  - `target_id` 用于存储能够转换为数字的平台 ID。
  - `openid` 用于存储字符串格式的平台 ID。
  - `EnsureIDs` 函数负责在事件到达时，根据这些关联字段查找或生成内部的 `int64` ID。

## 3. 数据库架构 (PostgreSQL)

- **Emoji 支持**: PostgreSQL 默认支持 Emoji 存储。只要数据库编码设置为 `UTF8` (默认值)，`VARCHAR` 和 `TEXT` 字段即可直接存储 Emoji，无需像 SQL Server 那样定义为 `NVARCHAR`。
- **字段兼容性**: 
  - 所有涉及 ID 的表列已统一修改为 `BIGINT`。
  - 提供了自动化的 `ALTER TABLE` 语句，确保现有数据库能够平滑迁移。

## 4. 关键技术实现

### FlexibleInt64
为了处理 JSON 中可能同时出现的字符串格式数字和原生数字数字，引入了 `FlexibleInt64` 自定义类型。它在 `UnmarshalJSON` 时会自动尝试将字符串解析为 `int64`，解析失败则回退到 `0`。

### 并发安全
针对 WebSocket 的并发写入，在 `tencentBot/main.go` 中引入了 `sync.Mutex` 保护连接对象，防止在高并发下出现 `concurrent write to websocket connection` 的 panic。

## 5. OneBot 协议约定
为了保持 OneBot 插件的通用性，内部传输时不增加额外字段。所有的平台映射逻辑均在 `BotWorker` 内部完成，插件只需处理标准的 `user_id` 和 `group_id` (int64)。
