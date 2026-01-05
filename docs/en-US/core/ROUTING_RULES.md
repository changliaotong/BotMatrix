# BotNexus Routing Rules Guide

> [🌐 English](ROUTING_RULES.md) | [简体中文](../zh-CN/ROUTING_RULES.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

## 📋 Overview

BotNexus provides intelligent message routing, supporting two routing modes:

1.  **API Request Routing**: External API requests use round-robin load balancing.
2.  **Message Event Routing**: Bot messages use intelligent routing rules for directed distribution.

## 🎯 Routing Logic

### Message Flow Diagram
```
User Message → Bot (via self_id) → BotNexus → Routing Rule Check → Targeted Worker
                                           ↓
                                   No Match Found → Random Worker (Load Balancing)

Worker Processing → Return Message (with self_id) → BotNexus → Based on self_id → Original Bot
```

### Routing Priority
1.  **Exact Match**: First checks for direct relationships like `user_123456`, `group_789012`, or `bot_123`.
2.  **Wildcard Match**: Supports `*` wildcards (e.g., `*_test` or `123*`).
3.  **Intelligent Load Balancing (RTT-based LB)**: When no rules match, chooses the optimal node based on average response time (AvgRTT) and health status.
4.  **Fallback**: If a targeted Worker is offline, the system automatically falls back to intelligent load balancing.

### Persistent Storage
All routing rules set via API are automatically persisted to the `botnexus.db` (SQLite database). Rules are reloaded on system restart.

## 🔧 Configuration

### Rule Format
```json
{
    "key": "user_123456",    // Match pattern: exact ID or wildcard with *
    "worker_id": "worker1"  // Target Worker ID
}
```

### Matching Pattern Examples
- `user_123456` - Exact match User ID
- `group_789012` - Exact match Group ID
- `bot_123` - Exact match Bot ID
- `*_test` - Matches any ID ending with `_test`
- `123*` - Matches any ID starting with `123`
