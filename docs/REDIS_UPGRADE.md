# BotMatrix Redis 系统升级文档 (Redis-Based System Upgrade)

## 1. 概述 (Overview)
本项目通过引入 Redis 作为核心中间件，实现了「消息路由中心」(BotNexus) 与「机器人执行端」(BotWorker) 的彻底解耦。该升级在不引入复杂中间件（如 RabbitMQ/Kafka）的前提下，提供了高并发处理、动态扩展、限流保护、幂等去重及会话管理能力。

## 2. 核心架构 (Architecture)
- **解耦模式**：BotNexus 仅负责消息的接收、初步决策（路由）和入队。BotWorker 异步监听队列并执行耗时任务（如 AI 调用、外部 API）。
- **扩展性**：支持多 Worker 横向扩展，Worker 可以通过监听公共队列或专用队列实现负载均衡。
- **动态性**：所有策略（限流、路由、TTL）均存储于 Redis，支持热更新，无需重启服务。

## 3. Redis Key 设计 (Redis Key Design)

| Key 模式 | 类型 | 说明 | 生命周期 |
| :--- | :--- | :--- | :--- |
| `botmatrix:queue:default` | List | 全局公共消息队列 | 永久 (由 Worker 消费) |
| `botmatrix:queue:worker:{id}` | List | 特定 Worker 的专用队列 | 永久 |
| `botmatrix:ratelimit:user:{id}` | String | 用户级限流计数器 | 60秒 (滑动窗口) |
| `botmatrix:ratelimit:group:{id}` | String | 群组级限流计数器 | 60秒 |
| `botmatrix:msg:idempotency:{id}` | String | 消息幂等标识 (去重) | 动态 (默认 1小时) |
| `botmatrix:session:{platform}:{user}` | String | 会话上下文 (JSON) | 动态 (默认 24小时) |
| `botmatrix:session:state:{platform}:{user}` | String | 特定会话状态 (JSON) | 动态 |
| `botmatrix:config:ratelimit` | Hash | 动态限流配置表 | 永久 |
| `botmatrix:config:ttl` | Hash | 动态过期时间配置表 | 永久 |
| `botmatrix:rules:routing` | Hash | 动态路由规则表 | 永久 |

## 4. 功能实现细节 (Implementation Details)

### 4.1 异步消息队列 (Message Queue)
- **入队策略**：BotNexus 使用 `RPush` 将标准化消息推送到 Redis。支持**指数退避重试机制**（最多 3 次），应对 Redis 瞬时抖动。
- **出队策略**：BotWorker 使用 `BLPop` 阻塞式监听队列，支持优先处理专用队列，无消息时保持休眠。

### 4.2 动态限流 (Rate Limiting)
- **维度**：支持 `user_id` 和 `group_id` 双重维度。
- **动态控制**：
    - 全局默认：`user_limit_per_min`, `group_limit_per_min`。
    - 个体覆盖：`user:{id}:limit`, `group:{id}:limit`。
- **存储**：配置存储在 `botmatrix:config:ratelimit` 中，修改即生效。

### 4.3 消息幂等与去重 (Idempotency)
- **二级缓存**：
    1. **本地热缓存** (`sync.Map`)：拦截极短时间内的重复请求，减少 Redis 压力。
    2. **Redis 远程缓存**：保证分布式环境下的全局唯一性。
- **清理机制**：后台协程自动清理本地过期数据。
- **标识算法**：支持 OneBot 标准 `message_id`，或根据 `post_type + time + user_id` 生成特征 ID。

### 4.4 会话与状态管理 (Session & State)
- **上下文追踪**：自动维护 `last_msg`, `last_time` 及最近 5 条消息历史 (`history`)。
- **状态隔离**：提供专用的 `State` 接口用于存储中间状态（如：等待用户确认、AI 任务状态）。
- **跨端同步**：Worker 可随时通过接口获取或更新 Nexus 侧维护的会话信息。

### 4.5 动态路由 (Dynamic Routing)
- **匹配优先级**：
    1. Redis 动态精确规则 (User/Group/Bot)
    2. Redis 动态通配符规则
    3. Nexus 内存静态规则 (Fallback)
- **热更新**：通过管理后台更新 `botmatrix:rules:routing` 即可实时切换机器人对应的 Worker。

## 5. 容错与高可用 (Fault Tolerance)
- **Fail-open (故障开路)**：若 Redis 无法连接，BotNexus 会自动降级为传统的 WebSocket 直接转发模式，确保基础功能不断连。
- **监控接口**：新增 `/api/admin/redis/config` 管理端点，支持可视化查看和操作 Redis 内部配置。

## 6. 管理员 API 参考 (Admin API Reference)

### 获取 Redis 动态配置
- **URL**: `/api/admin/redis/config`
- **Method**: `GET`
- **Response**: 返回限流、TTL、路由规则的当前快照。

### 更新 Redis 动态配置
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
*文档版本: 1.0*
*更新日期: 2025-12-23*
