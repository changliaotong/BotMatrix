# BotMatrix Redis System Upgrade

> [🌐 English](REDIS_UPGRADE.md) | [简体中文](../zh-CN/REDIS_UPGRADE.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

## 1. Overview
This project introduces Redis as core middleware to achieve complete decoupling between the "Message Routing Center" (BotNexus) and the "Robot Execution End" (BotWorker). This upgrade provides high concurrency processing, dynamic scaling, rate limiting protection, idempotency de-duplication, and session management capabilities without introducing complex middleware like RabbitMQ or Kafka.

## 2. Core Architecture
- **Decoupling Mode**: BotNexus is only responsible for receiving messages, making initial decisions (routing), and enqueuing. BotWorker asynchronously listens to the queue and executes time-consuming tasks (e.g., AI calls, external APIs).
- **Scalability**: Supports horizontal scaling of multiple Workers. Workers can achieve load balancing by listening to public or dedicated queues.
- **Dynamic Nature**: All policies (rate limiting, routing, TTL) are stored in Redis, supporting hot updates without restarting services.

## 3. Redis Key Design

| Key Pattern | Type | Description | Lifespan |
| :--- | :--- | :--- | :--- |
| `botmatrix:queue:default` | List | Global public message queue | Permanent (consumed by Workers) |
| `botmatrix:queue:worker:{id}` | List | Dedicated queue for a specific Worker | Permanent |
| `botmatrix:ratelimit:user:{id}` | String | User-level rate limit counter | 60s (Sliding window) |
| `botmatrix:ratelimit:group:{id}` | String | Group-level rate limit counter | 60s |
| `botmatrix:msg:idempotency:{id}` | String | Message idempotency flag (de-duplication) | Dynamic (Default 1h) |
| `botmatrix:session:{platform}:{user}` | String | Session context (JSON) | Dynamic (Default 24h) |
| `botmatrix:session:state:{platform}:{user}` | String | Specific session state (JSON) | Dynamic |
| `botmatrix:config:ratelimit` | Hash | Dynamic rate limit configuration table | Permanent |
| `botmatrix:config:ttl` | Hash | Dynamic TTL configuration table | Permanent |
| `botmatrix:rules:routing` | Hash | Dynamic routing rule table | Permanent |

## 4. Implementation Details

### 4.1 Asynchronous Message Queue
- **Enqueue Policy**: BotNexus uses `RPush` to push standardized messages to Redis. Supports an **exponential backoff retry mechanism** (up to 3 times) to handle transient Redis jitters.
- **Dequeue Policy**: BotWorker uses `BLPop` for blocking queue listening, prioritizing dedicated queues and staying dormant when no messages are present.

### 4.2 Dynamic Rate Limiting
- **Dimensions**: Supports dual dimensions of `user_id` and `group_id`.
- **Dynamic Control**:
    - Global Default: `user_limit_per_min`, `group_limit_per_min`.
    - Individual Override: `user:{id}:limit`, `group:{id}:limit`.
- **Storage**: Configurations are stored in `botmatrix:config:ratelimit` and take effect immediately upon modification.

### 4.3 Idempotency & De-duplication
- **Two-level Cache**:
    1. **Local Hot Cache** (`sync.Map`): Intercepts duplicate requests within a very short time to reduce Redis pressure.
    2. **Redis Remote Cache**: Ensures global uniqueness in a distributed environment.
- **Cleanup Mechanism**: Background goroutines automatically clean up expired local data.
- **Identification Algorithm**: Supports OneBot standard `message_id`, or generates characteristic IDs based on `post_type + time + user_id`.

### 4.4 Session & State Management
- **Context Tracking**: Automatically maintains `last_msg`, `last_time`, and the last 5 messages in history (`history`).
- **State Isolation**: Provides a dedicated `State` interface for storing intermediate states (e.g., waiting for user confirmation, AI task status).
- **Cross-end Sync**: Workers can retrieve or update session information maintained on the Nexus side at any time.

### 4.5 Dynamic Routing
- **Matching Priority**:
    1. Redis dynamic exact rules (User/Group/Bot)
    2. Redis dynamic wildcard rules
    3. Nexus memory-based static rules (Fallback)
- **Hot Update**: Updating `botmatrix:rules:routing` via the management back-end allows real-time switching of the corresponding Worker for a robot.

## 5. Fault Tolerance & High Availability
- **Fail-open**: If Redis is unreachable, BotNexus automatically downgrades to traditional WebSocket direct forwarding to ensure basic functionality.
- **Monitoring Interface**: Added `/api/admin/redis/config` management endpoint for visualizing and operating internal Redis configurations.

## 6. Admin API Reference

### Get Redis Dynamic Configuration
- **URL**: `/api/admin/redis/config`
- **Method**: `GET`
- **Response**: Returns a snapshot of current rate limits, TTL, and routing rules.

### Update Redis Dynamic Configuration
- **URL**: `/api/admin/redis/config`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "type": "ratelimit", // ratelimit, ttl, rules
    "data": { "user_limit_per_min": "30" },
    "clear": false
  }
  ```

---
*Document Version: 1.0*
*Update Date: 2025-12-23*
