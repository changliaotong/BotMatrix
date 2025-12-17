# main.go 拆分计划

## 当前问题
原始main.go文件太大（4000+行），需要拆分成多个文件便于维护。

## 建议拆分方案

### Phase 1: 提取配置和常量
- [x] config.go - 环境变量和配置常量

### Phase 2: 提取数据结构
- [x] types.go - 所有结构体定义
- [x] BotClient, WorkerClient, Subscriber, User, LogEntry等

### Phase 3: 提取工具函数  
- [x] utils.go - JWT、随机令牌、类型转换等通用函数

### Phase 4: 提取核心业务逻辑
- [ ] worker_handlers.go - Worker连接处理
- [ ] bot_handlers.go - Bot连接处理  
- [ ] api_handlers.go - API接口处理
- [ ] websocket_handlers.go - WebSocket通用处理

### Phase 5: 提取统计和监控
- [x] stats.go - 统计相关功能
- [ ] logger.go - 日志系统

### Phase 6: 提取后台任务
- [ ] timeout_monitor.go - 心跳超时检测
- [ ] periodic_tasks.go - 定时任务

## 下一步建议

由于直接完整拆分会导致大量编译错误，建议：

1. **先保持main.go不变**，确保当前功能正常
2. **逐步提取小功能模块**，每次提取后都测试编译
3. **优先提取日志增强功能**，解决你关心的worker/bot断开问题

## 立即行动

先给main.go添加详细日志，定位worker断开影响bot的问题，然后再考虑文件拆分。