# Moderation Plugin Documentation

## Overview
The Moderation plugin provides moderation features for the bot, including:
- 撤回 (Message Recall)
- 禁言 (Mute)
- 踢出 (Kick)
- 拉黑 (Blacklist)
- 灰名单 (Graylist)
- 脏话检测 (Profanity Detection)
- 广告检测 (Advertisement Detection)
- 图片检测 (Image Detection)
- 网址检测 (URL Detection)
- 敏感词检测 (Sensitive Word Detection)
- 白名单 (Whitelist)
- 群配置管理 (Group Configuration Management)
- 被踢加黑 (Kick to Blacklist)
- 被踢提示 (Kick Notification)
- 退群加黑 (Leave to Blacklist)
- 退群提示 (Leave Notification)

## Commands

### !撤回 <message_id>
Recall a message.

**Parameters:**
- `<message_id>`: The ID of the message to recall

**Examples:**
```
!撤回 123456
```

### !禁言 <user_id> <duration>
Mute a user for a specified duration.

**Parameters:**
- `<user_id>`: The ID of the user to mute
- `<duration>`: The duration of the mute (e.g., "1h", "1d", "1w")

**Examples:**
```
!禁言 123456 1h
!禁言 789012 1d
```

### !踢出 <user_id>
Kick a user from the group.

**Parameters:**
- `<user_id>`: The ID of the user to kick

**Examples:**
```
!踢出 123456
```

### !拉黑 <user_id>
Add a user to the blacklist. The user will be automatically kicked from the group.

**Parameters:**
- `<user_id>`: The ID of the user to blacklist

**Examples:**
```
!拉黑 123456
```

### !灰名单 <user_id>
Add a user to the graylist.

**Parameters:**
- `<user_id>`: The ID of the user to add to the graylist

**Examples:**
```
!灰名单 123456
```

### !白名单 <user_id>
Add a user to the whitelist. Whitelisted users are exempt from all moderation checks.

**Parameters:**
- `<user_id>`: The ID of the user to add to the whitelist

**Examples:**
```
!白名单 123456
```

### !群配置 <config>
Configure group-specific moderation settings.

**Parameters:**
- `<config>`: The configuration to set. Available options:
  - `被踢加黑开启` - Enable kick to blacklist
  - `被踢加黑关闭` - Disable kick to blacklist
  - `被踢提示开启` - Enable kick notification
  - `被踢提示关闭` - Disable kick notification
  - `退群加黑开启` - Enable leave to blacklist
  - `退群加黑关闭` - Disable leave to blacklist
  - `退群提示开启` - Enable leave notification
  - `退群提示关闭` - Disable leave notification
  - `清空黑名单` - Clear current group's blacklist (requires 3-digit confirmation code)
  - `清空白名单` - Clear current group's whitelist (requires 3-digit confirmation code)
  - `查看` - View current group configuration

**Examples:**
```
!群配置 被踢加黑开启
!群配置 退群提示关闭
!群配置 查看
```

## Configuration

The Moderation plugin requires the following configuration:

```json
{
  "moderation": {
    "sensitive_words": ["脏话1", "脏话2", "脏话3"],
    "whitelist": ["user1", "user2", "user3"],
    "blacklist": ["user4", "user5", "user6"],
    "group_configs": {
      "group1": {
        "kick_to_black": true,
        "kick_notify": true,
        "leave_to_black": true,
        "leave_notify": true
      }
    }
  }
}
```

## Notes
- Moderation commands can only be used by group administrators
- Sensitive words can be customized in the configuration file
- Whitelisted users are not subject to moderation checks and can send any message
- Blacklisted users cannot send messages in the group and will be automatically kicked when added
- Graylisted users are subject to additional moderation checks
- When a user is blacklisted, they will be automatically kicked from the group
- Dangerous operations such as clearing blacklist/whitelist require a 3-digit numeric confirmation code sent by the bot
