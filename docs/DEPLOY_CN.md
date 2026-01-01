# BotMatrix 部署指南

本指南介绍如何使用 Docker 部署 **BotMatrix** 生态系统。

## 1. 前提条件

*   已安装 **Docker** 和 **Docker Compose**。
*   已安装 **Git**。
*   （可选）用于数据持久化的 **Redis** 服务器（生产环境推荐）。

## 2. 快速开始

```bash
# 1. 克隆仓库
git clone https://github.com/changliaotong/BotMatrix.git
cd BotMatrix

# 2. 配置所需的机器人（见第 3 节）
# 示例：配置 KookBot
cp KookBot/config.sample.json KookBot/config.json
# 编辑 KookBot/config.json 文件，添加你的 token

# 3. 启动生态系统
docker-compose up -d --build
```

## 3. 配置指南

BotMatrix 采用模块化架构。你只需要配置和启用你打算使用的机器人。

### 🧠 BotNexus（核心管理器）
*   **文件**：`docker-compose.yml`（环境变量）或 `config.json`（持久化配置）
*   **端口**：`5000`（Web 仪表盘），`3001`（WebSocket 网关 - 默认）
*   **配置**：
    *   **持久化配置**：支持在同一目录下使用 `config.json` 文件。该文件可通过 WebUI（管理员设置）管理。
    *   **环境变量**（覆盖 `config.json`）：
        *   `WS_PORT`：WebSocket 网关端口（例如：`:3001`）。
        *   `WEBUI_PORT`：Web 仪表盘端口（例如：`:5000`）。
        *   `REDIS_ADDR`：Redis 服务器地址（例如：`127.0.0.1:6379`）。
        *   `REDIS_PWD`：Redis 密码。
        *   `JWT_SECRET`：用于生成 JWT 令牌的密钥。
        *   **数据库配置**（PostgreSQL 为必填项）：
            *   `DB_HOST`：PostgreSQL 主机（例如：`localhost`）
            *   `DB_PORT`：PostgreSQL 端口（例如：`5432`）
            *   `DB_USER`：PostgreSQL 用户名
            *   `DB_PASSWORD`：PostgreSQL 密码
            *   `DB_NAME`：PostgreSQL 数据库名称
            *   `DB_SSL_MODE`：PostgreSQL SSL 模式（例如：`disable`）
    *   **WebUI 配置**：以管理员身份登录后，你可以在 **系统设置** 标签页中直接修改这些设置。大多数更改（如 Redis）会立即生效，而端口更改需要重启服务。

### 🟢 WxBot（微信）
*   **类型**：Python / OneBot
*   **登录**：通过日志或仪表盘扫描二维码。
*   **配置**：`docker-compose.yml`（`BOT_SELF_ID`）。

### 🐧 TencentBot（官方 QQ）
*   **类型**：Go / BotGo SDK
*   **配置**：`TencentBot/config.json`
    ```json
    {
      "app_id": 123456,
      "secret": "YOUR_SECRET",
      "sandbox": false
    }
    ```

### 🐱 NapCat（个人 QQ）
*   **类型**：Docker / OneBot 11（NTQQ）
*   **配置**：`NapCat/config/onebot11.json`（已为 BotMatrix 预配置）
*   **登录**：通过 WebUI（`http://localhost:6099/webui`）或日志扫描二维码。

### 钉 DingTalkBot（钉钉）
*   **类型**：Go / Webhook & Stream
*   **配置**：`DingTalkBot/config.json`
    ```json
    {
      "client_id": "YOUR_CLIENT_ID",
      "client_secret": "YOUR_CLIENT_SECRET"
    }
    ```

### ✈️ FeishuBot（飞书）
*   **类型**：Go / WebSocket
*   **配置**：`FeishuBot/config.json`
    ```json
    {
      "app_id": "cli_xxx",
      "app_secret": "xxx"
    }
    ```

### ✈️ TelegramBot
*   **类型**：Go / Long Polling
*   **配置**：`TelegramBot/config.json`
    ```json
    {
      "bot_token": "123456:ABC-DEF"
    }
    ```

### 🎮 DiscordBot
*   **类型**：Go / Gateway
*   **配置**：`DiscordBot/config.json`
    ```json
    {
      "bot_token": "YOUR_BOT_TOKEN"
    }
    ```

### 💬 SlackBot
*   **类型**：Go / Socket Mode
*   **配置**：`SlackBot/config.json`
    ```json
    {
      "bot_token": "xoxb-",
      "app_token": "xapp-"
    }
    ```

### 🦜 KookBot（开黑啦）
*   **类型**：Go / WebSocket
*   **配置**：`KookBot/config.json`
    ```json
    {
      "bot_token": "YOUR_KOOK_TOKEN"
    }
    ```

### 📧 EmailBot
*   **类型**：Go / IMAP & SMTP
*   **配置**：`EmailBot/config.json`
    ```json
    {
      "imap_server": "imap.gmail.com",
      "username": "user@example.com",
      "password": "app_password"
    }
    ```

### 🏢 WeWorkBot（企业微信）
*   **类型**：Go / Callback & API
*   **配置**：`WeWorkBot/config.json`
    ```json
    {
      "corp_id": "wx",
      "agent_id": 10001,
      "secret": "",
      "token": "",
      "encoding_aes_key": ""
    }
    ```
*   **回调 URL**：`http://<YOUR_IP>:5002/callback`

## 4. 仪表盘和管理

访问 BotMatrix 仪表盘：
**http://localhost:5000**
*   **默认用户**：`admin`
*   **默认密码**：`123456`

## 4. 自动化部署脚本（面向开发者）

我们提供了一个功能强大的 Python 脚本 `scripts/deploy.py`，用于自动部署到远程服务器。

### 特性
- **交互式菜单**：精确选择要部署的内容。
- **自动配置**：如果缺少配置文件，自动从样本生成 `config.json`。
- **智能清理**：处理远程目录冲突和旧容器。
- **版本更新**：自动增加补丁版本号。

### 使用方法

```bash
# 运行部署脚本
python scripts/deploy.py
```

你将看到一个菜单：
```
Select Deployment Target:
  1. [All] Deploy Everything (Default)
  2. [NoWx] Deploy All EXCEPT WxBot (Preserves Login)
  3. [Mgr] Bot Manager Only
  4. [Wx] WxBot Only
  5. [Tencent] TencentBot Only
  6. [Sys] System Worker Only
```

### 模式
- **完整模式**（默认）：重新构建 Docker 镜像并重新创建容器。
- **快速模式**（`--fast`）：仅更新文件并重启容器（不重新构建）。
- **目标选择**：
  - `[NoWx]`：在不终止微信机器人进程的情况下进行更新（保留登录会话）。
  - `[All]`：完整的系统重置/更新。

### 配置
编辑 `scripts/deploy.py` 文件，设置你的服务器详细信息：
```python
DEFAULT_SERVER_IP = "192.168.x.x"
DEFAULT_USERNAME = "user"
```

## 5. 数据库设置

### PostgreSQL 设置（推荐）
BotMatrix 现在支持 PostgreSQL 作为主数据库，SQLite 作为备选。

1. **安装 PostgreSQL**（如果尚未安装）：
```bash
# Ubuntu/Debian
sudo apt-get install postgresql postgresql-contrib

# CentOS/RHEL
sudo yum install postgresql-server postgresql-contrib
sudo postgresql-setup initdb
```

2. **创建数据库和用户**：
```bash
# 连接到 PostgreSQL
sudo -u postgres psql

# 创建数据库和用户
CREATE DATABASE botmatrix_db;
CREATE USER botmatrix WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE botmatrix_db TO botmatrix;
\q
```

3. **配置环境变量**：
更新你的 `.env` 文件，添加 PostgreSQL 配置：
```bash
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=botmatrix
DB_PASSWORD=your_password
DB_NAME=botmatrix_db
DB_SSL_MODE=disable
```

4. **数据库迁移**：
BotMatrix 会在启动时自动创建必要的表。无需手动迁移。

### SQLite 设置（备选）
对于开发或小型部署，仍支持 SQLite：
```bash
DB_TYPE=sqlite
DB_PATH=./botmatrix.db
```

## 6. 故障排除

*   **端口被占用**：检查 `docker-compose.yml` 文件并更改映射端口（例如：`5000:5000` → `5050:5000`）。
*   **连接失败**：确保机器人配置中的 `NEXUS_ADDR` 指向 `ws://bot-manager:3005`（Docker 内部网络）。
*   **日志**：使用 `docker-compose logs -f [service_name]` 调试特定机器人。
*   **连接被拒绝**：确保 `bot-manager` 正在运行且端口 `3005` 可访问。
*   **Docker 权限被拒绝**：使用 `sudo` 运行或将用户添加到 `docker` 用户组。