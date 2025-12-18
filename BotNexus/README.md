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
1. 前端发送POST请求到 `/api/login`
2. 后端验证用户名密码（从 SQLite 数据库加载）
3. 返回JWT令牌
4. 前端存储令牌到localStorage

### 安全特性
- **SQLite 持久化**: 用户数据永久存储在本地数据库，解决 Redis 重启导致的数据丢失。
- **密码加密**: 使用 bcrypt 强哈希算法存储密码。
- **SSO 登录**: 支持跨系统（BotNexus & Overmind）的令牌透传。

## 📡 API接口

### 认证相关
- `POST /api/login` - 用户登录
- `GET /api/user/info` - 获取当前用户信息
- `GET /api/system/stats` - 系统运行详细统计
- `GET /api/bots` - 机器人列表
- `POST /api/bot/toggle` - 切换机器人状态

### WebSocket接口
- `ws://localhost:3005` - 实时消息推送与系统监控

## 🚀 部署说明

### 环境要求
- Go 1.19+
- SQLite 3 (自动初始化)
- Redis 6.0+ (用于消息缓存，非必须)

### 启动步骤
1. 运行BotNexus主程序
2. 首次启动会自动创建 `bot_nexus.db` 数据库文件
3. 默认管理员账号: `admin`, 默认密码: `admin123` (可在 `config.go` 修改)
4. 访问Web界面: `http://localhost:5000`

## 🎯 未来规划

### 长期目标
- [ ] 插件化架构
- [ ] 分布式部署
- [ ] AI智能管理

## 🐛 已修复问题
- [x] 前端/后端登录端点不匹配
- [x] Redis 依赖导致的密码丢失
- [x] 登录页面无法在移动端输入
- [x] 系统统计数据 undefined 显示问题
- [x] 登录按钮点击无反应问题

## 📚 相关文档

- [Go官方文档](https://golang.org/doc/)
- [Redis文档](https://redis.io/documentation)
- [WebSocket协议](https://developer.mozilla.org/zh-CN/docs/Web/API/WebSockets_API)

## 🤝 贡献指南

欢迎提交Issues和Pull Requests来改进系统。

## 📞 支持

如有问题，请在GitHub Issues中提交。