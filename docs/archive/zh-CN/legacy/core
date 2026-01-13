# BotNexus 路由规则使用指南

> [🌐 English](../en-US/ROUTING_RULES.md) | [简体中文](ROUTING_RULES.md)
> [⬅️ 返回文档中心](README.md) | [🏠 返回项目主页](../../README.md)

## 📋 概述

BotNexus 提供智能消息路由功能，支持两种路由模式：

1. **API请求路由**：外部API请求使用轮询负载均衡
2. **消息事件路由**：Bot消息使用智能路由规则进行定向分配

## 🎯 路由逻辑

### 消息流向图
```
用户消息 → Bot (via self_id) → BotNexus → 路由规则检查 → 指定Worker
                                           ↓
                                   无匹配规则 → 随机Worker (负载均衡)

Worker处理 → 返回消息 (带self_id) → BotNexus → 根据self_id → 原Bot
```

### 路由优先级
1. **精确匹配 (Exact Match)**：首先检查 `user_123456`, `group_789012` 或 `bot_123` 的直接对应关系。
2. **通配符匹配 (Wildcard Match)**：支持 `*` 通配符（如 `*_test` 或 `123*`）。
3. **智能负载均衡 (RTT-based LB)**：无匹配规则时，根据 Worker 的平均响应时间 (AvgRTT) 和健康状态选择最优节点。
4. **故障回退 (Fallback)**：若指定 Worker 离线，系统将自动回退到智能负载均衡，确保消息不丢失。

### 持久化存储 (Persistence)
所有通过 API 设置的路由规则都会自动持久化到 `botnexus.db` (SQLite 数据库) 中。系统重启后会自动从数据库重新加载所有规则，确保配置不丢失。

## 🔧 路由规则配置

### 规则格式
```json
{
    "key": "user_123456",    // 匹配模式：精确ID或带*的通配符
    "worker_id": "worker1"  // 目标Worker ID
}
```

### 匹配模式示例
- `user_123456` - 精确匹配用户 ID
- `group_789012` - 精确匹配群组 ID
- `bot_123` - 精确匹配机器人 ID
- `*_test` - 匹配任何以 `_test` 结尾的 ID
- `123*` - 匹配任何以 `123` 开头的 ID

### 设置示例
```bash
# 设置用户123456的消息路由到worker1
curl -X POST http://localhost:8080/api/admin/routing \
  -H "Content-Type: application/json" \
  -d '{"key": "user_123456", "worker_id": "worker1"}'

# 设置所有测试群组 (*_test) 路由到worker2
curl -X POST http://localhost:8080/api/admin/routing \
  -H "Content-Type: application/json" \
  -d '{"key": "*_test", "worker_id": "worker2"}'
```

## 💼 使用场景

### 1. VIP用户专属服务
```json
// 高价值客户群组路由到高性能Worker
{"key": "VIP_GROUP_001", "worker_id": "high_performance_worker"}
```

### 2. 测试环境隔离
```json
// 测试消息路由到测试Worker
{"key": "TEST_GROUP", "worker_id": "test_worker"}
```

### 3. 业务模块分离
```json
// 不同业务使用不同Worker处理
{"key": "CUSTOMER_SERVICE", "worker_id": "service_worker"}
{"key": "TECH_SUPPORT", "worker_id": "tech_worker"}
```

### 4. 负载分配优化
```json
// 高负载群组分散到多个Worker
{"key": "HIGH_TRAFFIC_GROUP_1", "worker_id": "worker_1"}
{"key": "HIGH_TRAFFIC_GROUP_2", "worker_id": "worker_2"}
```

## 🧪 测试验证

### 使用测试工具
打开 `test_routing_simple.html` 进行路由功能验证：
1. 检查当前Worker连接状态
2. 设置测试路由规则
3. 发送测试消息验证路由效果

### 日志监控
在BotNexus控制台查看详细的路由调试日志：
```
[ROUTING] Rule Matched: user_123456 -> Target Worker: worker1
[ROUTING] Rule Matched: group_789012 (via pattern group_*) -> Target Worker: worker2
[ROUTING] No target worker worker1 for rule user_123456, falling back to load balancer
[ROUTING] Failed to send to target worker worker1 for rule user_123456: connection closed
```

## ⚠️ 注意事项

1. **Worker可用性**：确保目标Worker处于连接状态
2. **规则冲突**：`group_id`优先级高于`self_id`
3. **性能影响**：大量规则可能略微增加路由延迟
4. **故障转移**：指定Worker不可用时自动回退到随机选择
5. **权限管理**：只有管理员可以配置路由规则

## 🔍 故障排查

### 常见问题

**Q: 路由规则不生效**
- 检查Worker是否连接：`GET /api/workers`
- 确认规则设置成功：`GET /api/admin/routing`
- 验证消息格式是否包含正确的`group_id`或`self_id`

**Q: 消息还是被随机分配**
- 检查路由键是否匹配（区分大小写）
- 确认Worker ID是否正确
- 查看日志确认路由查找过程

**Q: 路由后Worker处理失败**
- 检查Worker连接状态
- 查看Worker端日志
- 确认消息格式符合Worker要求

### 调试建议
1. 使用测试工具验证基本功能
2. 逐步添加规则进行测试
3. 监控BotNexus日志了解路由过程
4. 检查Worker端的接收和处理日志

## 📚 相关文档

- [BotNexus API文档](API.md)
- [Overmind使用指南](Overmind/README.md)
- [Worker开发指南](docs/WORKER_DEVELOPMENT.md)