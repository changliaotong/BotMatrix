# OneBot 11 协议兼容性文档

> [🌐 English](../en-US/ONEBOT_COMPATIBILITY.md) | [简体中文](ONEBOT_COMPATIBILITY.md)
> [⬅️ 返回文档中心](README.md) | [🏠 返回项目主页](../../README.md)

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
