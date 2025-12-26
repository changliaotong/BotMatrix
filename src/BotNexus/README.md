# BotNexus 系统文档

BotNexus 是一个统一的机器人矩阵管理系统，采用 Go 语言后端与现代化 Web 前端，支持大规模机器人的拓扑可视化、实时监控、Docker 容器管理及智能路由分发。

## 🏗️ 系统架构

BotNexus 作为一个中心枢纽，连接并管理多个机器人实例（Bots）与处理节点（Workers）。

### 核心特性
- **3D 拓扑可视化 (Matrix 3D)**: 基于 Three.js 的实时宇宙拓扑。支持节点聚类、实时消息粒子特效、自动连线优化、头像全景代理及**全路径路由追踪（User-Group-Nexus-Worker）**。
- **智能路由分发**: 具备 RTT 感知的动态路由算法，支持精确 ID 及通配符（*）匹配，规则持久化，支持节点离线自动回退。具备**智能缓存增强**功能，自动补全消息元数据。
- **多语言支持 (i18n)**: 完整支持 中文/英文 界面切换，适配全球化管理需求。
- **系统日志管理**: 实时流式日志展示，支持关键词过滤、日志一键清空及日志历史导出。
- **系统核心插件 (Core Plugin)**: 集成在消息路由层的安全拦截器。支持全局开关、黑白名单、敏感词过滤、URL 过滤及管理员指令控制。支持**全局自动回复**通知。
- **用户管理体系**: 完善的 RBAC 权限模型。管理员可创建用户、重置密码、切换用户状态（启用/禁用）。支持 `session_version` 强制 Token 失效。
- **数据持久化**: 核心缓存（联系人/统计/配置）均支持 **PostgreSQL** 持久化，确保服务重启后数据秒级同步。

### 技术栈
- **后端**: Go 1.20+, PostgreSQL, Redis (Cluster), JWT (Auth), Docker SDK
- **前端**: Vue 3 (Composition API), Three.js (3D), Tailwind CSS, Lucide Icons
- **移动端**: Flutter (Overmind)

## 🚀 快速开始

### 环境要求
- **Docker**: 必须安装并运行（若需容器管理功能）
- **Go**: 1.19+（本地编译）
- **PostgreSQL**: 核心业务数据库
- **Redis**: 用于消息队列与状态同步

### 启动步骤
1. **获取代码**:
   ```bash
   git clone <repository_url>
   cd BotNexus
   ```
2. **运行服务**:
   ```bash
   go run .
   ```
3. **访问后台**:
   - URL: `http://localhost:5000`
   - 默认账号: `admin`
   - 默认密码: `admin123`

## 📡 API 概览

### 认证
- `POST /api/login` - 获取 JWT Token
- `GET /api/me` - 获取个人信息

### Docker 管理
- `GET /api/docker/list` - 获取容器列表
- `POST /api/docker/action` - 执行容器操作 (start/stop/restart/delete)
- `POST /api/docker/add-bot` - 部署机器人
- `POST /api/docker/add-worker` - 部署处理节点
- `GET /api/admin/docker/logs` - 获取容器日志

### 用户管理
- `GET /api/admin/users` - 获取所有用户
- `POST /api/admin/users` - 管理用户 (create/delete/reset_password/toggle_active)

### 系统日志
- `GET /api/logs` - 获取流式日志
- `POST /api/logs/clear` - 清空日志

## 🎯 核心逻辑

### 3D 优化 (Performance)
- **材质缓存**: 复用 GPU 纹理与材质，减少内存占用。
- **光源限制**: 动态限制实时点光源数量，确保在消息量激增时维持 60 FPS。
- **自动同步**: WebSocket `sync_state` 确保前端节点状态与后端严格一致。

### 安全模型
- **JWT Middleware**: 全局接口权限校验。
- **Admin Middleware**: 核心管理操作二次验证。
- **Password Hashing**: 使用 bcrypt 进行高强度加密。

## 🤝 贡献与反馈
欢迎通过 GitHub Issues 提交建议或报告 Bug。

---
*BotNexus - Powering your bot matrix with elegance.*
