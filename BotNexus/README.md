# BotNexus 系统文档

## 🏗️ 系统架构

BotNexus是一个多机器人管理系统，支持QQ、微信等平台的机器人统一管理。

### 核心特性
- **智能路由**: 动态 RTT 感知路由，支持基于用户/群组/机器人 ID 的精确及通配符匹配，规则支持数据库持久化，具备离线 Worker 自动回退机制。
- **3D 拓扑可视化**: 基于 Three.js 的 3D 宇宙拓扑图，支持群组聚类（Clustering）、层次化连线优化、实时消息粒子特效及全景头像代理。
- **高可靠性与持久化**: 消息转发失败自动重试，所有缓存（群组/好友/成员/统计）均支持 SQLite 持久化，支持 WebSocket 初始同步（SyncState）确保刷新不掉数。
- **用户管理**: 完善的用户管理体系，支持管理员创建用户、重置密码及用户自助修改密码。
- **现代化 UI**: 响应式设计，实时系统资源监控，独立的运行时间与系统时间显示，仪表盘集成轮播统计块。

### 技术栈
- **后端**: Go 1.19+, SQLite 3 (持久化), JWT (身份认证), bcrypt (密码加密)
- **前端**: HTML5/CSS3/JS, Bootstrap 5, Chart.js, BI Icons
- **移动端**: Flutter (Overmind)
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
- **会话失效**: 密码修改或重置后，通过递增 `session_version` 强制旧 Token 失效。
- **SSO 登录**: 支持跨系统（BotNexus & Overmind）的令牌透传。

## 📡 API接口

### 认证与用户
- `POST /api/login` - 用户登录
- `GET /api/user/info` - 获取当前用户信息
- `POST /api/user/password` - 用户修改密码
- `GET /api/admin/users` - (Admin) 获取用户列表
- `POST /api/admin/user/reset-password` - (Admin) 重置用户密码

### 监控与管理
- `GET /api/system/stats` - 系统运行详细统计
- `GET /api/stats` - 业务统计数据 (群/用户/消息)
- `GET /api/bots` - 机器人列表
- `GET /api/workers` - 处理端 (Workers) 列表
- `POST /api/bot/toggle` - 切换机器人状态

### WebSocket接口
- `/ws/subscriber` - 实时消息推送与系统监控 (需JWT认证)

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
- [x] 3D 群组聚类与成员围绕分布 (Clustering)
- [x] 3D 连线树状优化与性能提升 (Tree-like Links)
- [x] WebSocket 初始状态同步与刷新数据恢复 (Initial Sync)
- [x] SQLite 全局统计与联系人缓存持久化 (Persistence)
- [x] 后台用户管理 (添加、修改、重置密码)
- [x] 智能路由未处理节点轮询问题
- [x] 消息转发确认、重发、自动更换节点与离线缓存
- [x] 运行时间及现在时间实时刷新
- [x] UI 布局优化 (OS 信息独立, 群/用户数量合并)
- [x] Overmind 链接跳转当前页面的 Bug
- [x] 处理端 (Workers) 数量在仪表盘显示为 undefined
- [x] 机器人 (Bots) 统计数据在某些情况下显示为 0
- [x] 机器人选择下拉菜单无法显示头像或昵称 (缺失 self_id)
- [x] 前端/后端登录端点不匹配
- [x] Redis 依赖导致的密码丢失
- [x] 登录页面无法在移动端输入
- [x] 系统统计数据 undefined 显示问题
- [x] 登录按钮点击无反应问题
- [x] WebSocket 持续重连与连接错误问题

## 📚 相关文档

- [Go官方文档](https://golang.org/doc/)
- [Redis文档](https://redis.io/documentation)
- [WebSocket协议](https://developer.mozilla.org/zh-CN/docs/Web/API/WebSockets_API)

## 🤝 贡献指南

欢迎提交Issues和Pull Requests来改进系统。

## 📞 支持

如有问题，请在GitHub Issues中提交。