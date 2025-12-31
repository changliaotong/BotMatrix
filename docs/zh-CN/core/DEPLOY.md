# BotMatrix 部署指南

> [🌐 English](../en-US/DEPLOY.md) | [简体中文](DEPLOY.md)
> [⬅️ 返回文档中心](README.md) | [🏠 返回项目主页](../../README.md)

本指南介绍如何使用 Docker 部署 **BotMatrix** 生态系统。

## 1. 环境准备

*   安装 **Docker** 和 **Docker Compose**。
*   安装 **Git**。
*   (可选) **Redis** 服务器用于数据持久化 (生产环境推荐)。

## 2. 快速开始

```bash
# 1. 克隆仓库
git clone https://github.com/changliaotong/BotMatrix.git
cd BotMatrix

# 2. 配置所需的机器人 (见第 3 节)
# 示例: 配置 KookBot
cp KookBot/config.sample.json KookBot/config.json
# 编辑 KookBot/config.json 并填入你的 Token

# 3. 启动系统
docker-compose up -d --build
```

## 3. 配置说明

BotMatrix 采用模块化架构。你只需要配置并启用你打算使用的机器人。

### 🧠 BotNexus (核心管理器)
*   **文件**: `docker-compose.yml` (环境变量) 或 `config.json` (持久化配置)
*   **端口**: `5000` (Web 管理后台), `3001` (WebSocket 网关 - 默认)
*   **配置**:
    *   **持久化配置**: 支持同目录下的 `config.json` 文件。该文件可以通过 WebUI (管理员设置) 进行管理。
    *   **环境变量** (覆盖 `config.json`):
        *   `WS_PORT`: WebSocket 网关端口 (例如 `:3001`)。
        *   `WEBUI_PORT`: Web 管理后台端口 (例如 `:5000`)。
        *   `REDIS_ADDR`: Redis 服务器地址 (例如 `127.0.0.1:6379`)。
        *   `REDIS_PWD`: Redis 密码。
        *   `JWT_SECRET`: 用于 JWT Token 生成的密钥。
        *   **数据库配置** (PostgreSQL 必填):
            *   `DB_HOST`: PostgreSQL 主机名
            *   `DB_PORT`: PostgreSQL 端口
            *   `DB_USER`: PostgreSQL 用户名
            *   `DB_PASSWORD`: PostgreSQL 密码
            *   `DB_NAME`: PostgreSQL 数据库名
            *   `DB_SSL_MODE`: PostgreSQL SSL 模式 (例如 `disable`)
    *   **WebUI 配置**: 以管理员身份登录后，你可以直接在 **系统设置** 选项卡中修改这些设置。大多数更改 (如 Redis) 会立即生效，而端口更改则需要重启服务。

### 🟢 WxBot (微信)
*   **类型**: Python / OneBot
*   **登录**: 通过日志或管理后台扫描二维码。
*   **配置**: `docker-compose.yml` (`BOT_SELF_ID`)。

### 🐧 TencentBot (腾讯官方 QQ)
*   **类型**: Go / BotGo SDK
*   **配置**: `TencentBot/config.json`
    ```json
    {
      "app_id": 123456,
      "secret": "YOUR_SECRET",
      "sandbox": false
    }
    ```

### 🐱 NapCat (个人 QQ)
*   **类型**: Docker / OneBot 11 (NTQQ)
*   **配置**: `NapCat/config/onebot11.json` (已为 BotMatrix 预配置)
*   **登录**: 通过 WebUI (`http://localhost:6099/webui`) 或日志扫描二维码。

### 钉 DingTalkBot (钉钉)
*   **类型**: Go / Webhook & Stream
*   **配置**: `DingTalkBot/config.json`
    ```json
    {
      "client_id": "YOUR_CLIENT_ID",
      "client_secret": "YOUR_CLIENT_SECRET"
    }
    ```

### ✈️ FeishuBot (飞书)
*   **类型**: Go / WebSocket
*   **配置**: `FeishuBot/config.json`
    ```json
    {
      "app_id": "cli_xxx",
      "app_secret": "xxx"
    }
    ```

### ✈️ TelegramBot
*   **类型**: Go / Long Polling
*   **配置**: `TelegramBot/config.json`
    ```json
    {
      "bot_token": "123456:ABC-DEF"
    }
    ```
