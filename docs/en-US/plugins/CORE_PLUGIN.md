# Nexus Core Plugin (System-Level)

> [🌐 English](CORE_PLUGIN.md) | [简体中文](../zh-CN/CORE_PLUGIN.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

## Overview
`CorePlugin` is a system-level core plugin integrated into the `BotNexus` message routing layer. It is responsible for making security, compliance, and stateful decisions on all raw messages before they are forwarded to Worker modules.

**Note**: This is a system-level feature that runs directly within the Nexus core process and does not belong to any Worker module.

## Core Features
- **Message Flow Control**: Supports global system opening/closing and maintenance modes.
- **Permission Arbitration**: Multi-dimensional black/white lists (system-level users, robots, groups).
- **Content Filtering**:
    - **Sensitive Word Library**: Supports plaintext and regex matching.
    - **URL Filter**: Prevents the spread of malicious links.
- **Traffic Statistics**: Real-time statistics on processing volume and interception volume for various types of messages.
- **Admin Commands**: Control system behavior directly through the chat interface.

## Admin Commands
Command prefixes: `/system` or `/nexus`

| Command | Parameters | Description | Example |
| :--- | :--- | :--- | :--- |
| `status` | None | View system status, online statistics, and today's throughput | `/system status` |
| `top` | None | View today's most active users and groups (speech statistics) | `/system top` |
| `open` | None | Open the system to allow message forwarding | `/system open` |
| `close` | None | Close the system to intercept all messages except admin commands | `/system close` |
| `whitelist` | `add <target> <id>` | Add to whitelist (`target`: `system`, `robot`, `group`) | `/system whitelist add system 123456` |
| `blacklist` | `add <target> <id>` | Add to blacklist (`target`: `system`, `robot`, `group`) | `/system blacklist add group 789012` |
| `reload` | None | Force reload latest configuration from Redis | `/system reload` |

## Statistics & Monitoring
The plugin asynchronously writes statistics to Redis. Key formats are:
- **Today's Stats**: `core:stats:yyyy-mm-dd` (Hash type)
- **Recent Blocked Records**: `core:blocked:yyyy-mm-dd` (List type, keeps the last 100 entries)

You can use the `status` command to view these in real-time.

## Testing & Verification
1. **Status Check**: 
   Send `/system status`. The system should return the number of online Bots, Workers, and message processing statistics.
2. **Global Interception Verification**:
   - Send `/system close`.
   - Try sending a normal chat message. The system should no longer reply, or the Worker should no longer receive the message.
   - Send `/system status`. The command should still execute normally.
   - Send `/system open` to restore the system.
3. **Blacklist Verification**:
   - Send `/system blacklist add system <YourUID>`.
   - Try sending a message. Check Nexus logs; it should show `Message blocked: ... (reason: user_blacklisted)`.

## Configuration & Extension
Configuration is defined in `src/BotNexus/core_plugin.go`. Supports multi-instance deployment with all state synchronization handled via Redis:
- **System Switch State**: `core:system_open`
- **Core Config JSON**: `core:config`

---
*Generated on 2025-12-22*
