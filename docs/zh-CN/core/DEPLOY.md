# 部署与系统运维指南 (Deployment & Operations)

> [⬅️ 返回文档中心](README.md) | [🏠 返回项目主页](../../README.md)

本指南详细介绍了 BotMatrix 生态系统的环境准备、容器化部署、系统配置、性能优化以及日常运维操作。

---

## 1. 环境准备与快速开始

### 1.1 环境要求
- **Docker** & **Docker Compose**: 核心部署工具。
- **Redis**: 用于数据持久化与异步任务队列 (生产环境推荐)。
- **PostgreSQL**: 核心数据库存储。

### 1.2 快速启动
```bash
# 1. 克隆仓库
git clone https://github.com/changliaotong/BotMatrix.git
cd BotMatrix

# 2. 配置机器人 (以 KookBot 为例)
cp KookBot/config.sample.json KookBot/config.json
# 编辑 config.json 并填入 Token

# 3. 启动系统
docker-compose up -d --build
```

---

## 2. 核心配置说明 (BotNexus)

BotNexus 采用模块化架构，支持通过配置文件、环境变量或 WebUI 进行管理。

- **Web 管理后台**: 默认端口 `5000`。
- **WebSocket 网关**: 默认端口 `3001`。
- **关键环境变量**:
    - `REDIS_ADDR`: Redis 地址 (例: `127.0.0.1:6379`)。
    - `DB_HOST`/`DB_USER`/`DB_PASSWORD`: 数据库连接信息。
    - `JWT_SECRET`: 用于安全认证的密钥。

---

## 3. 服务端管理与运维 (Server Manual)

系统内置了轻量级 Web 控制台与管理员指令系统。

### 3.1 Web 控制台功能
- **仪表盘 (Dashboard)**: 实时监控 CPU/内存、连接数及消息吞吐量。
- **登录管理**: 提供二维码登录页面 (`/login`) 及状态检测。

### 3.2 管理员指令 (仅限管理员使用)
指令需以 `#` 开头：
- `#status`: 查看服务器运行状态及网关连接数。
- `#reload`: 热重载所有插件代码（无需重启服务）。
- `#broadcast <msg>`: 向活跃群组发送系统通知。
- `#db_clean <days>`: 清理指定天数前的历史聊天记录。

---

## 4. 性能优化措施

为了支撑高并发消息处理，系统实现了多级优化：

### 4.1 AI 解析器优化
- **正则预编译**: 在插件报备能力时自动预编译正则表达式，避免运行时 CPU 密集计算。
- **正则缓存**: 使用线程安全的 Map 缓存已编译的正则对象。

### 4.2 Redis 交互策略
- **ConfigCache (二级缓存)**: 每 30 秒从 Redis 同步一次配置，消息主流程仅读取本地内存，延迟近乎零。
- **SessionCache (热点缓存)**: 活跃会话上下文存储在本地 `sync.Map` 中，采用“写穿式”同步，减少同步等待。

---

## 5. 容器化部署最佳实践

### 5.1 核心理念：无状态 Worker + 有状态插件
- **Stateless Worker**: BotWorker 实例不存储状态，支持水平扩容。
- **Stateful Plugins**: 插件状态与用户会话存储在 Redis 或共享存储中。

### 5.2 插件管理
- **热更新**: 支持无需重启容器的插件热加载。
- **灰度发布**: 利用 `canary_weight` 配置，根据 Session 粘滞性将部分流量导向新版本插件。

---

## 6. 移动端管理 (小程序)

BotMatrix 提供配套的微信小程序，方便随时随地管理系统。
- **功能**: 系统状态概览、机器人实时监控、远程指令执行、实时日志查看。
- **集成**: 通过 Overmind REST API 与 WebSocket 服务实现数据同步。

---
*最后更新日期：2026-01-13*
