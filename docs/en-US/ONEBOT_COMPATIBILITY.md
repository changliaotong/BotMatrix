# OneBot 11 Protocol Compatibility

> [🌐 English](ONEBOT_COMPATIBILITY.md) | [简体中文](../zh-CN/ONEBOT_COMPATIBILITY.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

This document records the compatibility of various protocol clients in the BotMatrix project with the OneBot 11 standard protocol.

## Protocol Client Compatibility Status

### 1. DingTalkBot
- **Status**: Completed
- **Feature Implementation**:
  - Basic OneBot 11 compatibility.
  - Handles Nexus commands (`send_group_msg`, `send_private_msg`, `delete_msg`, `get_login_info`).
  - DingTalk events converted to OneBot format.
- **Notes**: Core actions supported; advanced events (notice/request) to be improved.

### 2. DiscordBot
- **Status**: Completed
- **Feature Implementation**:
  - OneBot 11 compatibility.
  - Message conversion (Discord Message → OneBot Event).
  - Fixed template import errors.
  - Basic CQ code processing.
- **Notes**: Message type mapping correctly implemented (Discord ChannelID → group_id).

### 3. FeishuBot
- **Status**: Completed
- **Feature Implementation**:
  - OneBot 11 compatibility.
  - Message conversion (Feishu P2MessageReceiveV1 → OneBot Event).
  - Multiple command support (`send_group_msg`, `send_private_msg`, `delete_msg`, `get_login_info`, `get_group_list`, `get_group_member_list`).
  - Feishu API integration.
- **Notes**: Core features implemented.

### 4. KookBot
- **Status**: Completed
- **Feature Implementation**:
  - OneBot 11 compatibility.
  - Message conversion (Kook Text/Image/Kmarkdown Message → OneBot Event).
  - Multiple command support (`send_group_msg`, `send_private_msg`, `delete_msg`, `get_login_info`).
  - WebSocket communication with BotNexus.
- **Notes**: Full message type support implemented.

### 5. WxBotGo
- **Status**: Completed
- **Feature Implementation**:
  - OneBot 11 compatibility.
  - Message conversion (WeChat Message → OneBot Event).
  - Multiple command support (`send_private_msg`, `send_group_msg`, `get_login_info`, `get_group_list`, `get_group_member_info`).
  - WebSocket communication with BotNexus.
- **Limitations**:
  - Due to `openwechat` library restrictions, some operations are not supported:
    - `set_group_kick`
    - `delete_msg`
    - `set_group_ban`
    - `set_friend_add_request`
    - `set_group_add_request`
- **Notes**: Basic chat functionality implemented.

### 6. EmailBot
- **Status**: Completed
- **Feature Implementation**:
  - OneBot 11 compatibility.
  - Email reception converted to OneBot message events.
  - Sending emails via OneBot protocol.
  - WebSocket connection to BotNexus.
  - Config UI and log viewing features.
- **Notes**: All emails are treated as private chat messages.

### 7. NapCat
- **Status**: Completed
- **Feature Implementation**:
  - Full OneBot 11 standard implementation.
  - Supports both forward and reverse WebSocket connections.
  - Supports all OneBot 11 standard features.
  - Configuration set to use reverse WebSocket to BotNexus.
- **Notes**: NapCat itself is fully compatible with the OneBot 11 standard.

## OneBot 11 Standard Implementation Notes

### Core Event Types
- `message` - Message events
- `notice` - Notice events
- `request` - Request events
- `meta_event` - Meta events

### Core Fields
- `post_type` - Event type
- `message_type` - Message type (group/private)
- `time` - Event timestamp
- `self_id` - Bot's own ID
- `user_id` - User ID
- `group_id` - Group ID (if applicable)
- `message_id` - Message ID
- `message` - Message content
- `raw_message` - Raw message content

### Core API Actions
- `send_msg` - Send message
- `send_private_msg` - Send private message
- `send_group_msg` - Send group message
- `delete_msg` - Recall message
