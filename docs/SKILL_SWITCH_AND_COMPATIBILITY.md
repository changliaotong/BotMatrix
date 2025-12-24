# BotMatrix 技能系统兼容性与功能开关文档

本文档记录了 BotMatrix 技能系统的兼容性设计、功能开关机制以及在测试阶段的隔离策略。

## 1. 设计目标

为了在引入 Redis 异步任务队列和技能系统（Skill System）的同时，确保与旧版 BotWorker 客户端的完全兼容，系统采用了“功能开关控制”与“能力动态发现”相结合的策略。

- **默认关闭**：所有技能相关功能在生产环境默认关闭。
- **平滑降级**：旧版 Worker 依然可以通过 WebSocket 正常处理基础消息。
- **环境隔离**：技能系统仅在 `ENABLE_SKILL=true` 的测试环境中激活。

## 2. 功能开关机制

### 2.1 全局配置
可以通过以下三种方式控制开关：

- **Web UI 后台**：在 BotNexus 配置中心的“核心配置”页签中，可以直接勾选或取消“启用技能系统”并保存重启。
- **配置文件** (`config.json`):
  ```json
  {
    "enable_skill": false
  }
  ```
- **环境变量**: `ENABLE_SKILL=true` 或 `ENABLE_SKILL=1` 可强制开启。

### 2.2 BotNexus (服务端) 行为
当 `ENABLE_SKILL` 为 `false` 时：
1. **组件不初始化**：不启动 GORM 数据库连接、`TaskManager` 调度器及 Redis 订阅监听。
2. **结果上报拦截**：即使收到 Worker 上报的 `skill_result`，也会被 `handleWorkerMessage` 丢弃并记录日志。
3. **路由降级**：系统仅执行传统的 OneBot 消息转发逻辑。

### 2.3 BotWorker (客户端) 行为
当 `EnableSkill` 为 `false` 时：
1. **隐藏能力报备**：启动时不向 `botmatrix:worker:register` 频道发送 `capabilities` 列表，Nexus 将其视为基础转发节点。
2. **拒绝执行指令**：在 Redis 队列监听中，如果收到 `skill_call` 类型的消息，将直接忽略而不进入执行流。

## 3. 兼容性路由策略

为了支持混合版本（新旧 Worker 同时在线）的环境，BotNexus 实现了以下智能路由逻辑：

### 3.1 技能感知分发
- **定向投递**：`Dispatcher` 在分发技能任务前，会通过 `FindWorkerBySkill` 检索显式报备了该技能的 Worker。
- **负载均衡**：在支持该技能的 Worker 集合中进行随机分发。
- **隔离旧版**：未报备技能能力的旧版 Worker 不会进入候选名单，从而避免收到无法解析的 `skill_call` 消息。

### 3.2 结果回传兼容
- **双通道支持**：支持通过 Redis Pub/Sub 或 WebSocket 回传技能结果。
- **ID 回退机制**：
  - 优先使用 `execution_id` 精确匹配任务执行记录。
  - 对于不支持 `execution_id` 的旧版上报，回退到根据 `task_id` 更新该任务最近一次的执行状态。

## 4. 测试与上线建议

1. **测试阶段**：
   - 部署独立的测试 Nexus 和 Worker 实例。
   - 配置文件中设置 `"enable_skill": true`。
   - 验证 `skill_call` 到 `skill_result` 的闭环流程。

2. **灰度上线**：
   - 先升级部分 Worker 并开启技能开关。
   - Nexus 开启开关，观察任务是否准确路由到新版 Worker。

3. **正式上线**：
   - 全量更新 Worker 并在配置文件中统一开启开关。

---
*最后更新日期：2025-12-24*
