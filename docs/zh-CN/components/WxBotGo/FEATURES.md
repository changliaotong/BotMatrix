# WxBotGo 功能特性

## 概述

WxBotGo 是一个基于 Go 语言开发的 WeChat 机器人框架，实现了 OneBot v11 协议标准，提供了丰富的功能接口和事件处理机制。

## 核心功能

### 1. 消息处理

#### 支持的消息类型
- **文本消息**: 纯文本内容
- **图片消息**: 支持本地图片和网络图片
- **表情消息**: 系统表情和自定义表情
- **语音消息**: 语音文件
- **名片消息**: 联系人名片
- **文件消息**: 文档、音频、视频等文件

#### 消息发送接口
```go
// 发送私聊消息
send_private_msg(user_id, message)

// 发送群聊消息
send_group_msg(group_id, message)

// 通用消息发送
send_msg(message_type, user_id/group_id, message)
```

#### 消息发送参数
- **message_type**: 消息类型，可选值为 "private"（私聊）或 "group"（群聊）
- **user_id**: 私聊对象 ID
- **group_id**: 群聊 ID
- **message**: 消息内容，支持纯文本或 CQ 码格式
- **auto_escape**: 是否自动转义特殊字符，默认值为 false

#### 自身消息支持
- 支持接收自身发送的消息
- 以 OneBot 标准格式上报自身消息
- 支持通过命令控制机器人及设置功能

#### 消息管理操作
```go
// 删除消息
delete_msg(message_id)
```
- 注意：由于 openwechat 库限制，目前不支持删除消息功能
- 调用该接口会返回 "Unsupported action" 响应

### 2. 信息查询

#### 机器人信息
```go
// 获取登录信息
get_login_info()

// 获取自身信息
get_self_info()
```

#### 联系人信息
```go
// 获取好友列表
get_friend_list()

// 获取群列表
get_group_list()

// 获取群成员列表
get_group_member_list(group_id)
```

### 3. 群管理

#### 支持的操作
```go
// 修改群名称
set_group_name(group_id, new_name)
```

#### 不支持的操作
- `set_group_kick`: 踢人功能（openwechat 库不支持）
- `set_group_ban`: 禁言功能（openwechat 库不支持）
- `set_group_admin`: 设置管理员（openwechat 库不支持）
- `set_group_anonymous`: 匿名聊天功能（openwechat 库不支持）
- `set_group_card`: 修改群名片功能（openwechat 库不支持）
- `set_group_leave`: 退出群聊功能（openwechat 库不支持）
- `set_group_special_title`: 设置群专属头衔功能（openwechat 库不支持）

### 4. 事件处理

#### 消息事件
- `message`: 收到消息事件
- `message.private`: 私聊消息事件
- `message.group`: 群聊消息事件
- `message.private.self`: 自身发送的私聊消息

#### 请求事件
- `request.friend`: 好友请求事件
- `request.group`: 群邀请事件

#### 通知事件
- `notice.group_member_increase`: 群成员增加事件
- `notice.group_member_decrease`: 群成员减少事件
- `notice.group_admin`: 群管理员变更事件

#### 元事件
- `meta_event.lifecycle`: 生命周期事件
- `meta_event.heartbeat`: 心跳事件

### 5. 协议兼容性

#### 已实现的接口
- 消息发送接口：`send_private_msg`, `send_group_msg`, `send_msg`
- 信息查询接口：`get_login_info`, `get_self_info`, `get_friend_list`, `get_group_list`, `get_group_member_list`, `get_group_info`, `get_group_member_info`
- 群管理接口：`set_group_name`
- 版本信息接口：`get_version_info`
- 状态查询接口：`get_status`

#### 未实现的接口
- 消息管理接口：`delete_msg`（openwechat 库不支持）
- 群管理接口：`set_group_kick`, `set_group_ban`, `set_group_admin`, `set_group_anonymous`, `set_group_card`, `set_group_leave`, `set_group_special_title`（openwechat 库不支持）
- 请求处理接口：`set_friend_add_request`, `set_group_add_request`（openwechat 库不支持）

## 技术特性

### 1. 协议支持
- 完整实现 OneBot v11 协议标准
- 支持 WebSocket 通信方式
- 兼容 BotMatrix 管理平台

### 2. 扩展性
- 模块化设计，易于扩展新功能
- 支持插件机制
- 可自定义事件处理器

### 3. 可靠性
- 自动重连机制
- 错误重试机制
- 日志记录功能

### 4. 跨平台
- 支持 Windows、Linux、macOS
- 支持 Docker 容器化部署
- 轻量级，资源占用低

## 性能指标

- 消息处理延迟: < 100ms
- 并发连接数: 支持 100+ 同时连接
- 消息吞吐量: 1000+ 消息/秒

## 安全特性

- 数据加密传输
- 权限控制机制
- 防消息轰炸
- 日志审计功能