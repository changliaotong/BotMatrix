# BotMatrix 数据库结构文档

本文档详细描述了 BotMatrix 系统中所有数据库表的结构，所有表名和列名均已统一为 `snake_case` 标准。

---

## 1. 核心业务表 (BotNexus 维护)

这些表由 BotNexus 通过 GORM 进行管理和自动迁移。

### 1.1 `bot_entities` (机器人实体)
存储系统中注册的所有机器人实例信息。

| 列名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | SERIAL (PK) | 自增主键 |
| `self_id` | VARCHAR(64) | 机器人自身 ID (如 QQ 号) |
| `nickname` | VARCHAR(128) | 机器人昵称 |
| `platform` | VARCHAR(32) | 平台类型 (onebot, gowebq, etc.) |
| `status` | VARCHAR(32) | 当前状态 (online, offline) |
| `connected` | BOOLEAN | 是否已连接 |
| `last_seen` | TIMESTAMP | 最后在线时间 |
| `created_at` | TIMESTAMP | 创建时间 |
| `updated_at` | TIMESTAMP | 更新时间 |
| `deleted_at` | TIMESTAMP | 软删除时间 |

### 1.2 `message_logs` (消息日志)
记录机器人收发的所有原始消息。

| 列名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | SERIAL (PK) | 自增主键 |
| `bot_id` | VARCHAR(64) | 关联机器人 ID |
| `user_id` | VARCHAR(64) | 发送者/接收者用户 ID |
| `group_id` | VARCHAR(64) | 关联群组 ID (私聊为空) |
| `content` | TEXT | 消息内容 |
| `raw_data` | TEXT | 原始 JSON 数据 |
| `platform` | VARCHAR(32) | 平台类型 |
| `direction` | VARCHAR(16) | 消息方向 (incoming, outgoing) |
| `created_at` | TIMESTAMP | 记录时间 |

### 1.3 `users` (系统用户)
管理 BotMatrix 管理后台的登录用户。

| 列名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | SERIAL (PK) | 自增主键 |
| `username` | VARCHAR(255) | 用户名 (唯一) |
| `password_hash` | VARCHAR(255) | 密码哈希值 |
| `is_admin` | BOOLEAN | 是否为超级管理员 |
| `active` | BOOLEAN | 账号是否启用 |
| `session_version` | INTEGER | 会话版本号 (用于强制退出) |
| `created_at` | TIMESTAMP | 创建时间 |
| `updated_at` | TIMESTAMP | 更新时间 |

---

## 2. AI 核心表 (AI 引擎相关)

### 2.1 `ai_providers` (AI 提供商)
配置不同的 AI 模型供应商（如 OpenAI, DeepSeek）。

| 列名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | SERIAL (PK) | 自增主键 |
| `name` | VARCHAR(100) | 提供商名称 |
| `type` | VARCHAR(50) | 提供商类型 (openai, ollama, etc.) |
| `base_url` | VARCHAR(255) | 接口基础地址 |
| `api_key` | VARCHAR(500) | API 密钥 (加密存储) |
| `is_enabled` | BOOLEAN | 是否启用 |
| `priority` | INTEGER | 优先级 (用于负载均衡) |
| `user_id` | INTEGER | 关联用户 (私有配置) |

### 2.2 `ai_models` (AI 模型)
定义供应商下具体的模型 ID。

| 列名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | SERIAL (PK) | 自增主键 |
| `provider_id` | INTEGER | 关联提供商 ID |
| `model_id` | VARCHAR(100) | API 调用模型 ID (如 gpt-4) |
| `model_name` | VARCHAR(100) | 展示名称 (如 GPT-4 Turbo) |
| `capabilities` | VARCHAR(255) | 能力列表 (chat, vision 等) |
| `context_size` | INTEGER | 上下文窗口大小 |
| `is_default` | BOOLEAN | 是否为默认模型 |

### 2.3 `ai_usage_logs` (AI 使用日志)
详细记录每一次 AI 调用产生的 Token 消耗和费用。

| 列名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | SERIAL (PK) | 自增主键 |
| `user_id` | INTEGER | 关联用户 |
| `agent_id` | INTEGER | 关联智能体 ID |
| `model_name` | VARCHAR(100) | 实际使用的模型名 |
| `input_tokens` | INTEGER | 输入 Token 数 |
| `output_tokens` | INTEGER | 输出 Token 数 |
| `duration_ms` | INTEGER | 响应时长 (毫秒) |
| `status` | VARCHAR(20) | 状态 (success, failed) |
| `question` | TEXT | 用户提问内容 |
| `answer` | TEXT | AI 回答内容 |
| `revenue_deducted` | INTEGER | 扣除的算力额度 |

---

## 3. 插件业务表 (BotWorker 维护)

这些表主要服务于具体的插件业务逻辑，由 BotWorker 在初始化时维护。

### 3.1 `vips` (VIP 会员)
| 列名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `group_id` | BIGINT (PK) | 群组 ID |
| `group_name` | TEXT | 群组名称 |
| `first_pay` | DECIMAL | 首次付费金额 |
| `start_date` | TIMESTAMP | 开始日期 |
| `end_date` | TIMESTAMP | 结束日期 |
| `user_id` | BIGINT | 推荐人/购买人 ID |
| `is_year_vip` | BOOLEAN | 是否为年费 VIP |

### 3.2 `agents` (智能体/小助手)
| 列名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | SERIAL (PK) | 自增主键 |
| `guid` | UUID | 唯一标识符 |
| `name` | TEXT | 智能体名称 |
| `prompt` | TEXT | 提示词设定 |
| `model` | TEXT | 指定使用的模型 |
| `temperature` | DOUBLE | 温度参数 |
| `owner_id` | BIGINT | 所有者 ID |

### 3.3 `black_list` (黑名单)
| 列名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `bot_uin` | BIGINT (PK) | 机器人 ID |
| `group_id` | BIGINT (PK) | 群组 ID |
| `black_id` | BIGINT (PK) | 黑名单用户 ID |
| `black_info` | TEXT | 拉黑原因/备注 |
| `insert_date` | TIMESTAMP | 拉黑时间 |

---

## 4. 缓存与同步表 (Cache)

用于提高查询效率，定期与平台数据同步。

### 4.1 `member_cache` (群成员缓存)
| 列名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `group_id` | VARCHAR(255) | 群号 |
| `user_id` | VARCHAR(255) | 用户号 |
| `nickname` | VARCHAR(255) | 昵称 |
| `card` | VARCHAR(255) | 群名片 |
| `role` | VARCHAR(50) | 角色 (owner, admin, member) |
| `last_seen` | TIMESTAMP | 最后同步时间 |

---

## 5. 命名规范与设计原则

1. **统一蛇形命名**：所有表名和列名必须使用 `snake_case`（如 `user_id` 而非 `UserId`）。
2. **PostgreSQL 兼容**：不再使用双引号包裹标识符，确保跨数据库驱动的兼容性。
3. **字段一致性**：
   - 时间字段统一使用 `created_at` / `updated_at` / `deleted_at`。
   - 外部 ID (如 QQ 号) 统一使用 `VARCHAR(64)` 或 `BIGINT` 以兼容不同平台的长 ID。
   - 状态字段使用 `VARCHAR(20)` 并配合常量枚举。
