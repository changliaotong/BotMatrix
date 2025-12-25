# OneBot 11 协议兼容性文档

本文档记录了 BotMatrix 项目中各种协议客户端与 OneBot 11 标准协议的兼容性情况。

## 协议客户端兼容性状态

### 1. DingTalkBot
- **状态**: 已完成
- **功能实现**:
  - 基本的 OneBot 11 兼容性
  - 处理 Nexus 命令（send_group_msg, send_private_msg, delete_msg, get_login_info）
  - DingTalk 事件转换为 OneBot 格式
- **备注**: 核心动作已支持，高级事件（notice/request）待完善

### 2. DiscordBot
- **状态**: 已完成
- **功能实现**:
  - OneBot 11 兼容性
  - 消息转换（Discord 消息 → OneBot 事件）
  - 修复模板导入错误
  - 基本 CQ 码处理
- **备注**: 消息类型映射已正确实现（Discord ChannelID → group_id）

### 3. FeishuBot
- **状态**: 已完成
- **功能实现**:
  - OneBot 11 兼容性
  - 消息转换（Feishu P2MessageReceiveV1 → OneBot 事件）
  - 多种命令支持（send_group_msg, send_private_msg, delete_msg, get_login_info, get_group_list, get_group_member_list）
  - Feishu API 集成
- **备注**: 核心功能已实现

### 4. KookBot
- **状态**: 已完成
- **功能实现**:
  - OneBot 11 兼容性
  - 消息转换（Kook Text/Image/Kmarkdown 消息 → OneBot 事件）
  - 多种命令支持（send_group_msg, send_private_msg, delete_msg, get_login_info）
  - WebSocket 通信与 BotNexus
- **备注**: 已实现完整消息类型支持

### 5. WxBotGo
- **状态**: 已完成
- **功能实现**:
  - OneBot 11 兼容性
  - 消息转换（WeChat 消息 → OneBot 事件）
  - 多种命令支持（send_private_msg, send_group_msg, get_login_info, get_group_list, get_group_member_info）
  - WebSocket 通信与 BotNexus
- **限制**:
  - 由于 openwechat 库限制，部分操作不支持：
    - set_group_kick
    - delete_msg
    - set_group_ban
    - set_friend_add_request
    - set_group_add_request
- **备注**: 基础聊天功能已实现

### 6. EmailBot
- **状态**: 已完成
- **功能实现**:
  - OneBot 11 兼容性
  - 邮件接收转换为 OneBot 消息事件
  - 通过 OneBot 协议发送邮件
  - WebSocket 连接到 BotNexus
  - 配置 UI 和日志查看功能
- **备注**: 将所有邮件作为私聊消息处理

### 7. NapCat
- **状态**: 已完成
- **功能实现**:
  - 完整的 OneBot 11 标准实现
  - 支持正向和反向 WebSocket 连接
  - 支持所有 OneBot 11 标准功能
  - 配置已设置为使用反向 WebSocket 连接到 BotNexus
- **备注**: NapCat 本身已完全兼容 OneBot 11 标准

## OneBot 11 标准实现说明

### 核心事件类型
- `message` - 消息事件
- `notice` - 通知事件
- `request` - 请求事件
- `meta_event` - 元事件

### 核心字段
- `post_type` - 事件类型
- `message_type` - 消息类型（group/private）
- `time` - 事件时间戳
- `self_id` - 机器人自身 ID
- `user_id` - 用户 ID
- `group_id` - 群组 ID（如适用）
- `message_id` - 消息 ID
- `message` - 消息内容
- `raw_message` - 原始消息内容

### 核心 API 动作
- `send_msg` - 发送消息
- `send_private_msg` - 发送私聊消息
- `send_group_msg` - 发送群消息
- `delete_msg` - 撤回消息
- `get_login_info` - 获取登录信息
- `get_group_list` - 获取群列表
- `get_group_member_list` - 获取群成员列表
- `get_group_member_info` - 获取群成员信息
- `get_self_info` - 获取机器人信息
- `get_friend_list` - 获取好友列表
- `get_version_info` - 获取版本信息
- `get_status` - 获取状态

## WebSocket 通信协议

所有客户端均支持 WebSocket 与 BotNexus 通信，包括:
- 连接建立时的识别包
- 心跳机制（每 10-30 秒）
- 事件上报
- 命令接收与响应
- 配置更新

## 特殊处理

### CQ 码支持
- 基本 CQ 码处理（图片、语音等）
- 消息元素转换

### 消息类型映射
- 不同平台的消息类型正确映射到 OneBot 11 标准
- 文本、图片、语音、表情等消息类型支持

## 已知限制

1. **WxBotGo**: 由于第三方库限制，部分群管理功能不支持
2. **EmailBot**: 将所有邮件作为私聊消息处理，不支持群聊概念
3. **各平台特有功能**: 某些平台特有功能无法完全映射到 OneBot 11 标准

## 开发建议

1. 对于新添加的协议客户端，应遵循以上 OneBot 11 标准实现
2. 优先实现核心消息和命令功能
3. 根据平台特性进行合理的字段映射
4. 实现 WebSocket 连接管理机制
5. 考虑异常处理和重连机制