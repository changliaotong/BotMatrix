# BotMatrix Development Notes
[中文](../development/BotWorker_DEV.md) | [Back to Home](../../README.md) | [Back to Docs](../README.md)

This document records key design decisions, architectural adjustments, and special handling for the QQGuild platform within the project.

## 1. ID Design and Auto-Generation Strategy

To be compatible with the QQGuild platform and support multi-platform expansion, the project adopts the following ID design:

- **Data Type**: All `user_id` and `group_id` in the database use `BIGINT` (corresponding to `int64` in Go).
- **Auto-Generation Range**:
  - **User ID (UserID)**: Increments starting from `980000000000`.
  - **Group ID (GroupID)**: Increments starting from `990000000000`.
- **Query Logic**: Each time a new ID is generated, the system queries the maximum ID in the current table and adds 1. If the table is empty or the maximum ID is less than the starting value, it starts from the starting point.

## 2. QQGuild Platform Integration

IDs on the QQGuild platform (OpenID/TinyID) are usually long strings or large numbers. The processing logic is as follows:

- **SelfId Handling**: The bot's own `SelfId` is treated as a special user, and its ID is also mapped to a `user_id` in the `users` table.
- **Associated Fields**:
  - The `users` table has added `target_user_id` (BIGINT) and `user_openid` (VARCHAR) columns.
  - The `groups` table has added `target_group_id` (BIGINT) and `group_openid` (VARCHAR) columns.
- **Mapping Logic**:
  - `target_id` is used to store platform IDs that can be converted to numbers.
  - `openid` is used to store platform IDs in string format.
  - The `EnsureIDs` function is responsible for looking up or generating internal `int64` IDs based on these associated fields when an event arrives.

## 3. Database Architecture (PostgreSQL)

- **Emoji Support**: PostgreSQL supports Emoji storage by default. As long as the database encoding is set to `UTF8` (the default), `VARCHAR` and `TEXT` fields can directly store Emojis, without needing to be defined as `NVARCHAR` like in SQL Server.
- **Field Compatibility**: 
  - All table columns involving IDs have been uniformly modified to `BIGINT`.
  - Automated `ALTER TABLE` statements are provided to ensure a smooth migration for existing databases.

## 4. Key Technical Implementations

### FlexibleInt64
To handle numbers that may appear as both string format and native number format in JSON, a custom `FlexibleInt64` type was introduced. During `UnmarshalJSON`, it automatically attempts to parse strings as `int64`, falling back to `0` if parsing fails.

### Concurrent Safety
For concurrent writes to WebSocket, `sync.Mutex` was introduced in `tencentBot/main.go` to protect the connection object, preventing `concurrent write to websocket connection` panics under high concurrency.

## 5. OneBot Protocol Convention
To maintain the generality of OneBot plugins, no extra fields are added during internal transmission. All platform mapping logic is completed within `BotWorker`; plugins only need to handle standard `user_id` and `group_id` (int64).
