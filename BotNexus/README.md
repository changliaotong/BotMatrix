# BotNexus 系统文档

## 🏗️ 系统架构

BotNexus是一个多机器人管理系统，支持QQ、微信等平台的机器人统一管理。

### 核心组件
- **BotNexus**: 主服务，提供API接口和Web界面
- **Overmind**: Flutter移动端管理应用
- **小程序**: 微信/QQ小程序版本

### 技术栈
- **后端**: Go语言
- **前端**: HTML/CSS/JavaScript
- **移动端**: Flutter
- **小程序**: 原生小程序框架

## 🔐 认证机制

### 登录流程
1. 前端发送POST请求到 `/login`
2. 后端验证用户名密码
3. 返回JWT令牌
4. 前端存储令牌到localStorage

### 已知问题
- 前端使用 `/api/login`，后端实际为 `/login`
- Redis重启后密码丢失问题

### 解决方案
```javascript
// 临时解决方案 - 直接设置token
localStorage.setItem('token', 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...');

// 或者使用正确的登录端点
fetch('/login', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({username: 'admin', password: 'admin123'})
});
```

## 📡 API接口

### 认证相关
- `POST /login` - 用户登录
- `GET /api/stats` - 系统状态
- `GET /api/bots` - 机器人列表
- `POST /api/bot/toggle` - 切换机器人状态

### WebSocket接口
- `ws://localhost:3005` - 实时消息推送

## 🚀 部署说明

### 环境要求
- Go 1.19+
- Redis 6.0+
- Node.js 16+ (开发环境)

### 启动步骤
1. 启动Redis服务
2. 运行BotNexus主程序
3. 访问Web界面

### 配置说明
- 端口配置: 默认5000 (HTTP), 3005 (WebSocket)
- Redis配置: 默认localhost:6379
- 日志配置: 支持文件和控制台输出

## 🔧 开发指南

### 前端开发
- 使用原生JavaScript
- 支持现代浏览器
- 响应式设计

### API开发
- RESTful API设计
- JSON数据格式
- 统一的错误处理

### WebSocket开发
- 实时状态推送
- 断线重连机制
- 消息队列处理

## 🐛 已知问题

1. **认证问题**: 前端/后端登录端点不匹配
2. **Redis依赖**: 密码存储在Redis，重启后丢失
3. **错误处理**: 部分API错误信息不够友好
4. **性能优化**: 大量机器人时性能待优化

## 🎯 未来规划

### 短期目标
- [ ] 修复认证端点不匹配问题
- [ ] 改进密码存储机制
- [ ] 优化错误处理

### 中期目标
- [ ] 支持更多机器人平台
- [ ] 增加权限管理
- [ ] 完善日志系统

### 长期目标
- [ ] 插件化架构
- [ ] 分布式部署
- [ ] AI智能管理

## 📚 相关文档

- [Go官方文档](https://golang.org/doc/)
- [Redis文档](https://redis.io/documentation)
- [WebSocket协议](https://developer.mozilla.org/zh-CN/docs/Web/API/WebSockets_API)

## 🤝 贡献指南

欢迎提交Issues和Pull Requests来改进系统。

## 📞 支持

如有问题，请在GitHub Issues中提交。