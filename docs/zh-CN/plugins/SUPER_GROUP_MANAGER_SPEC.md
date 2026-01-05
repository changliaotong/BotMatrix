# 超级群管插件 (SuperGroupManager) 设计方案

## 1. 概述
`SuperGroupManager` 是基于 BotMatrix C# SDK 开发的高级群自动化管理插件。它集成了违规监控、自动处罚、成员欢迎及积分联动功能，旨在通过“经济杠杆”和“自动化工具”提升群聊活跃度与合规性。

## 2. 核心功能模块

### 2.1 违规监控与处罚 (Violation Monitor)
- **关键词过滤**：实时扫描消息，支持正则表达式和黑名单。
- **处罚措施**：
    - 撤回消息。
    - 自动禁言（阶梯式禁言：5min -> 1h -> 24h）。
    - **积分扣除**：联动 Economy Service，对违规者扣除 `bot_local` 或 `group` 积分。
- **频率限制 (Anti-Spam)**：检测短时间内的高频刷屏行为。

### 2.2 成员管理与欢迎 (Member Management)
- **智能欢迎语**：支持多模板随机切换，艾特新成员，并发送群规。
- **进群改名**：强制或引导新成员修改群名片。
- **邀请追踪**：记录邀请人，并联动 Economy Service 给予邀请奖励。

### 2.3 互动配置面板 (Interactive Config)
- **免指令配置**：管理员通过 `/config` 进入对话模式，通过按钮或数字选项配置插件。
- **群组差异化**：每个群聊拥有独立的配置存储在 Redis (`Session`)。

### 2.4 积分联动 (Economy Integration)
- **活跃奖励**：每日发言达到一定数量奖励积分。
- **违规罚金**：违规自动从用户积分账户划扣至 `SYSTEM_TAX` 或群组公共账户。
- **邀请佣金**：成功邀请新成员后，系统发放奖励。

## 3. 技术实现细节

### 3.1 消息拦截流程
1.  **Middleware 捕获**：所有消息经过插件中间件。
2.  **敏感词匹配**：调用本地缓存或 Redis 存储的词库。
3.  **执行处罚**：
    - 调用 `delete_message`。
    - 调用 `set_group_ban`。
    - 发送 Skill 请求：`EconomyService.Transfer(from=UserId, to="SYSTEM_TAX", amount=50, reason="违规关键词")`。

### 3.2 存储设计
- **Key 格式**：`table:super_group:group:{group_id}:config`
- **内容**：JSON 序列化的配置对象。

## 4. 接口规范 (Skills)

### 4.1 暴露的 Skill
- `get_group_config`: 获取当前群的配置。
- `update_group_config`: 更新群配置。

### 4.2 调用的外部 Skill
- `EconomyService.transfer`: 执行罚款或奖励。

## 5. 权限控制
- 仅限 `Group Admin` 或 `Bot Admin` 修改配置。
- 插件执行处罚操作需具备 `mute_user` 和 `delete_message` 权限。
