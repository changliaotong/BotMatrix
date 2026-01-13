# 🚀 部署与运维指南 (Deployment & Operations)

> **版本**: 2.0
> **状态**: 生产就绪
> [🌐 English](../en-US/DEPLOYMENT_AND_OPERATIONS.md) | [简体中文](DEPLOYMENT_AND_OPERATIONS.md)
> [⬅️ 返回文档中心](README.md) | [🏠 返回项目主页](../../README.md)

BotMatrix 采用容器化、模块化架构，核心理念为 **“无状态 Worker + 有状态插件与会话”**。

---

## 1. 核心架构与部署模式

为了实现高可用与弹性扩缩容，BotWorker 保持无状态，所有状态通过外部系统管理。

```text
[ 用户/平台 ] -> [ 负载均衡 (Nginx) ] 
                 |
        [ BotWorker 集群 (副本: N) ]
          /            |            \
 [ Redis 会话 ] [ 共享存储 (插件) ] [ BotNexus 控制中心 ]
    (状态)          (代码/模型)          (指令/鉴权)
```

---

## 2. 环境准备与快速开始

### 2.1 环境要求
- **Docker** 20.10+ & **Docker Compose** 2.0+
- **PostgreSQL** (核心数据库)
- **Redis** (缓存与异步任务)

### 2.2 快速启动
```bash
# 1. 克隆仓库
git clone https://github.com/changliaotong/BotMatrix.git
cd BotMatrix

# 2. 初始化配置 (以 KookBot 为例)
cp KookBot/config.sample.json KookBot/config.json

# 3. 启动全栈
docker-compose up -d --build
```

---

## 3. 核心配置说明 (BotNexus)

配置支持 `config.json` 与 **环境变量** (优先级更高)。

### 3.1 关键环境变量
- `WS_PORT`: WebSocket 网关端口 (默认 `:3001`)。
- `WEBUI_PORT`: Web 管理后台端口 (默认 `:5000`)。
- `REDIS_ADDR`: Redis 地址 (例如 `redis:6379`)。
- `DB_HOST / DB_NAME / DB_USER / DB_PASSWORD`: PostgreSQL 连接信息。
- `JWT_SECRET`: 用于管理后台登录的安全密钥。

---

## 4. 容器化最佳实践

### 4.1 插件版本化与热更新
推荐在容器中挂载共享卷 (`/app/plugins`)，并采用版本化目录：
- **热更新**: 将新版本解压至插件目录，通过 WebUI 或 `#reload` 指令触发热加载，无需重启容器。
- **灰度发布 (Canary)**: 在 `plugin.json` 中设置 `canary_weight`。Nexus 会根据 Session 粘滞性将指定比例的流量引导至新版本。

### 4.2 弹性扩容
在 `docker-compose.yml` 中利用 `deploy.replicas` 实现水平扩展：
- 使用 `update_config.order: start-first` 确保零停机滚动更新。
- 确保所有 Worker 挂载同一个分布式文件系统卷 (如 NFS/Ceph) 或通过 `SyncFromMarket` API 同步。

---

## 5. 平台部署速查

| 平台 | 类型 | 部署关键点 |
| :--- | :--- | :--- |
| **NapCat (QQ)** | OneBot 11 | 使用预配置镜像，通过 `:6099` 扫码登录。 |
| **WxBot (微信)** | Python/OneBot | 扫描容器日志或 WebUI 二维码。 |
| **DingTalk / Feishu** | Go/Stream | 配置 `AppID` 与 `Secret`，确保公网回调可达。 |
| **Telegram** | Go/Polling | 仅需 `BotToken`，国内环境需配置代理。 |

---

## 6. 系统运维与指令

### 6.1 Web 控制台
访问 `http://IP:5000` (或配置端口)：
- **Dashboard**: 查看 CPU/内存、连接数、消息吞吐量。
- **实时日志**: 滚动查看系统全局日志。
- **设置中心**: 动态修改 Redis、数据库及机器人路由规则。

### 6.2 管理员指令 (仅限超级管理员)
在聊天框输入以 `#` 开头的指令：
- `#status`: 查看服务器负载、运行时间。
- `#reload`: 强制热重载所有插件。
- `#broadcast <msg>`: 全局广播通知。
- `#db_clean <days>`: 清理指定天数前的聊天记录。

---

## 7. 安全与合规
- **PII 脱敏**: 开启 `ENABLE_PRIVACY_GUARD=true`，系统将自动识别并屏蔽日志与外发数据中的手机号、姓名。
- **健康检查**: 配置 Docker `HEALTHCHECK` 确保故障实例自动剔除。
- **审计跟踪**: 每一项关键操作 (如 `send_msg`) 均通过 `AIAgentTrace` 记录 `execution_id` 供回溯。
