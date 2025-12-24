# WxBotGo 更改日志

## v1.0.1 - 2025-12-25

### 功能增强

#### 自身消息支持
- 新增自身消息上报开关功能
- 支持通过配置文件或 WebUI 界面控制是否上报自身消息
- 默认开启自身消息上报功能
- 实现 OneBot 标准的自身消息格式上报

#### 配置管理
- 在 `FeaturesConfig` 中新增 `ReportSelfMsg` 字段
- 在 WebUI 界面添加 "上报自身消息" 开关选项
- 支持通过命令控制机器人及设置功能

### 代码优化

#### 消息处理逻辑
- 修改 `HandleWeChatMsg` 函数，根据配置决定是否上报自身消息
- 优化自身消息的 OneBot 事件格式转换
- 提高消息处理的灵活性和可配置性

#### 界面集成
- 更新 WebUI 模板，添加自身消息上报开关
- 完善配置保存逻辑，支持新的配置项

### 文档更新

- 更新 `FEATURES.md`，详细说明自身消息支持功能
- 更新 `CHANGELOG.md`，记录版本变更
- 补充 OneBot 事件类型说明

### 兼容性

- 完全兼容 OneBot v11 协议标准
- 与 BotMatrix 管理平台无缝集成
- 支持跨平台部署（Windows、Linux、macOS）

## v1.0.0 - 2025-12-25

### 主要功能实现

#### OneBot 标准协议支持

##### 已实现的动作

1. **消息发送**
   - `send_private_msg`: 发送私聊消息
   - `send_group_msg`: 发送群聊消息
   - `send_msg`: 通用消息发送（支持私聊和群聊）

2. **信息查询**
   - `get_login_info`: 获取登录信息
   - `get_self_info`: 获取机器人自身信息
   - `get_friend_list`: 获取好友列表
   - `get_group_list`: 获取群列表
   - `get_group_member_list`: 获取群成员列表

3. **群管理**
   - `set_group_name`: 修改群名称（支持）
   - `set_group_kick`: 踢人（不支持，返回错误信息）
   - `set_group_ban`: 禁言（不支持，返回错误信息）

##### 数据结构扩展

- **OneBotEvent**: 新增请求和通知类型字段
  - `RequestType`: 请求类型（好友请求、群邀请等）
  - `Comment`: 请求备注信息
  - `Flag`: 请求标识
  - `NoticeType`: 通知类型

- **ActionParams**: 新增管理类和查询类参数
  - `UserIDs`: 批量用户ID
  - `Duration`: 禁言时长
  - `Reason`: 操作原因
  - `NoCache`: 是否禁用缓存

### 代码结构优化

1. **core/bot.go**: 重构 HandleAction 方法，支持更多动作类型
2. **core/models.go**: 扩展数据结构以支持完整的 OneBot 协议
3. **cmd/main.go**: 优化 QR 码生成和处理流程

### 错误处理改进

- 对不支持的操作返回明确的错误信息
- 统一错误码规范：
  - `-1`: 通用错误
  - `100`: 不支持的操作

### 依赖更新

- 升级 openwechat 库到最新版本
- 优化 WebSocket 连接管理

### 兼容性

- 支持 OneBot v11 协议标准
- 与 BotMatrix 管理端完美集成
- 支持跨平台部署（Windows、Linux、macOS）