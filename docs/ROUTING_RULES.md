# BotNexus 路由规则使用指南

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
1. **检查路由规则**：根据`group_id`或`self_id`查找匹配的Worker
2. **默认负载均衡**：无匹配规则时随机选择Worker
3. **失败回退**：指定Worker不可用时回退到随机选择

## 🔧 路由规则配置

### API端点
- **获取规则**：`GET /api/admin/routing`
- **设置规则**：`POST /api/admin/routing`
- **管理权限**：需要管理员权限

### 规则格式
```json
{
    "key": "123456",        // group_id 或 bot_id
    "worker_id": "worker1"  // 目标Worker ID
}
```

### 设置示例
```bash
# 设置群123456的消息路由到worker1
curl -X POST http://localhost:8080/api/admin/routing \
  -H "Content-Type: application/json" \
  -d '{"key": "123456", "worker_id": "worker1"}'

# 设置机器人bot_789的消息路由到worker2
curl -X POST http://localhost:8080/api/admin/routing \
  -H "Content-Type: application/json" \
  -d '{"key": "bot_789", "worker_id": "worker2"}'

# 删除路由规则（worker_id为空）
curl -X POST http://localhost:8080/api/admin/routing \
  -H "Content-Type: application/json" \
  -d '{"key": "123456", "worker_id": ""}'
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
在BotNexus控制台查看路由日志：
```
[SUCCESS] Successfully routed message to worker1 via routing rule
[WARN] No routing rule found for group 123456, using random worker
[ERROR] Target worker worker1 unavailable, falling back to random selection
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