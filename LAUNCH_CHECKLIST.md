# BotMatrix Launch Checklist (系统上线检查清单)

## 1. 核心连通性 (Core Connectivity)
- [ ] **Redis 连接**: 验证 BotNexus 和 BotWorker 是否能正常连接到 Redis 集群。
- [ ] **数据库 (GORM)**: 验证 BotNexus 能否正常读写数据库，且所有模型 (BotEntity, RoutingRule, User, Task, Execution, MessageLog) 已完成 AutoMigrate。
- [ ] **WebSocket 通信**: 验证 BotNexus 能够接收来自 Bot 和 Worker 的 WebSocket 连接。
- [ ] **Redis Stream**: 验证消息队列 `botmatrix:queue:worker:<worker_id>` 是否正常工作，Worker 能够消费消息。

## 2. 技能与路由 (Skills & Routing)
- [ ] **技能发现**: 验证 Worker 启动后通过 `update_metadata` 自动上报插件技能，且 Nexus 能够识别。
- [ ] **正则路由**: 验证 Nexus 能够根据消息内容匹配正则规则，并将请求正确转发给对应的 Worker。
- [ ] **AI 转发**: 验证当消息未命中正则时，Nexus 能够将其转发给 AI 服务处理（如果开启）。
- [ ] **技能执行**: 验证 Worker 执行技能后，能够通过 `skill_result` 回传结果。

## 3. 性能与稳定性 (Performance & Stability)
- [ ] **心跳检测**: 验证 Worker 掉线后 Nexus 能够及时检测（默认 30s 阈值）。
- [ ] **日志监控**: 验证 `clog` 日志是否正常输出到控制台/文件，且没有明显的 Nil Pointer 报错。
- [ ] **并发测试**: 验证在高并发消息下，Redis 队列和 WebSocket 转发的稳定性。

## 4. 环境配置 (Environment Configuration)
- [ ] **环境变量**: 检查 `REDIS_ADDR`, `DB_URL`, `ENABLE_SKILL` 等核心配置是否在生产环境已正确设置。
- [ ] **鉴权秘钥**: 检查 JWT Secret 和各平台 API Key 是否安全存储。

## 5. 测试覆盖 (Testing Coverage)
- [ ] **集成测试**: 确保 `src/BotNexus/internal/app/flow_integration_test.go` 中的所有测试用例在 CI/CD 中通过。
- [ ] **插件测试**: 确保核心插件 (AI, RAG, MCP) 的单元测试全部通过。
