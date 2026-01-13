# BotMatrix Deployment Guide

> [🌐 English](DEPLOY.md) | [简体中文](../zh-CN/DEPLOY.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

This guide describes how to deploy the **BotMatrix** ecosystem using Docker.

## 1. Prerequisites

*   **Docker** & **Docker Compose** installed.
*   **Git** installed.
*   (Optional) **Redis** server for data persistence (recommended for production).

## 2. Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/changliaotong/BotMatrix.git
cd BotMatrix

# 2. Configure desired bots (see Section 3)
# Example: Configure KookBot
cp KookBot/config.sample.json KookBot/config.json
# Edit KookBot/config.json with your token

# 3. Start the ecosystem
docker-compose up -d --build
```

## 3. Configuration Guide

BotMatrix uses a modular architecture. You only need to configure and enable the bots you intend to use.

### 🧠 BotNexus (Core Manager)
*   **File**: `docker-compose.yml` (environment variables) or `config.json` (persistent config)
*   **Port**: `5000` (Web Dashboard), `3001` (WebSocket Gateway - Default)
*   **Config**:
    *   **Persistent Config**: Support for `config.json` file in the same directory. This file can be managed via the WebUI (Admin settings).
    *   **Environment Variables** (Overrides `config.json`):
        *   `WS_PORT`: WebSocket gateway port (e.g., `:3001`).
        *   `WEBUI_PORT`: Web Dashboard port (e.g., `:5000`).
        *   `REDIS_ADDR`: Redis server address (e.g., `127.0.0.1:6379`).
        *   `REDIS_PWD`: Redis password.
        *   `JWT_SECRET`: Secret for JWT token generation.
        *   **Database Configuration** (PostgreSQL mandatory):
            *   `DB_HOST`: PostgreSQL host (e.g., `localhost`)
            *   `DB_PORT`: PostgreSQL port (e.g., `5432`)
            *   `DB_USER`: PostgreSQL username
            *   `DB_PASSWORD`: PostgreSQL password
            *   `DB_NAME`: PostgreSQL database name
            *   `DB_SSL_MODE`: PostgreSQL SSL mode (e.g., `disable`)
    *   **WebUI Config**: Once logged in as an administrator, you can modify these settings directly in the **System Settings** tab. Most changes (like Redis) take effect immediately, while port changes require a service restart.

### 🟢 WxBot (WeChat)
*   **Type**: Python / OneBot
*   **Login**: Scan QR code via logs or Dashboard.
*   **Config**: `docker-compose.yml` (`BOT_SELF_ID`).

### 🐧 TencentBot (Official QQ)
*   **Type**: Go / BotGo SDK
*   **Config**: `TencentBot/config.json`
    ```json
    {
      "app_id": 123456,
      "secret": "YOUR_SECRET",
      "sandbox": false
    }
    ```

### 🐱 NapCat (Personal QQ)
*   **Type**: Docker / OneBot 11 (NTQQ)
*   **Config**: `NapCat/config/onebot11.json` (Pre-configured for BotMatrix)
*   **Login**: Scan QR code via WebUI (`http://localhost:6099/webui`) or logs.

### 钉 DingTalkBot (DingTalk)
*   **Type**: Go / Webhook & Stream
*   **Config**: `DingTalkBot/config.json`
    ```json
    {
      "client_id": "YOUR_CLIENT_ID",
      "client_secret": "YOUR_CLIENT_SECRET"
    }
    ```

### ✈️ FeishuBot (Lark/Feishu)
*   **Type**: Go / WebSocket
*   **Config**: `FeishuBot/config.json`
    ```json
    {
      "app_id": "cli_xxx",
      "app_secret": "xxx"
    }
    ```

### ✈️ TelegramBot
*   **Type**: Go / Long Polling
*   **Config**: `TelegramBot/config.json`
    ```json
    {
      "bot_token": "123456:ABC-DEF"
    }
    ```

### 🎮 DiscordBot
*   **Type**: Go / Gateway
*   **Config**: `DiscordBot/config.json`
    ```json
    {
      "bot_token": "YOUR_BOT_TOKEN"
    }
    ```

### 💬 SlackBot
*   **Type**: Go / Socket Mode
*   **Config**: `SlackBot/config.json`
    ```json
    {
      "bot_token": "xoxb-...",
      "app_token": "xapp-..."
    }
    ```

### 🦜 KookBot (Kaiheila)
*   **Type**: Go / WebSocket
*   **Config**: `KookBot/config.json`
    ```json
    {
      "bot_token": "YOUR_KOOK_TOKEN"
    }
    ```

### 📧 EmailBot
*   **Type**: Go / IMAP & SMTP
*   **Config**: `EmailBot/config.json`
    ```json
    {
      "imap_server": "imap.gmail.com",
      "username": "user@example.com",
      "password": "app_password"
    }
    ```

### 🏢 WeWorkBot (Enterprise WeChat)
*   **Type**: Go / Callback & API
*   **Config**: `WeWorkBot/config.json`
    ```json
    {
      "corp_id": "wx...",
      "agent_id": 10001,
      "secret": "...",
      "token": "...",
      "encoding_aes_key": "..."
    }
    ```
*   **Callback URL**: `http://<YOUR_IP>:5002/callback`

## 4. Dashboard & Management

Access the BotMatrix Dashboard at:
**http://localhost:5000**
*   **Default User**: `admin`
*   **Default Pass**: `123456`

## 4. Automated Deployment Script (For Developers)

We provide a powerful python script `scripts/deploy.py` for automated deployment to a remote server.

### Features
- **Interactive Menu**: Choose exactly what to deploy.
- **Auto Config**: Automatically generates `config.json` from samples if missing.
- **Smart Cleanup**: Handles remote directory conflicts and old containers.
- **Version Bump**: Automatically increments patch version.

### Usage

```bash
# Run the deployment script
python scripts/deploy.py
```

You will be presented with a menu:
```
Select Deployment Target:
  1. [All] Deploy Everything (Default)
  2. [NoWx] Deploy All EXCEPT WxBot (Preserves Login)
  3. [Mgr] Bot Manager Only
  4. [Wx] WxBot Only
  5. [Tencent] TencentBot Only
  6. [Sys] System Worker Only
```

### Modes
- **Full Mode** (Default): Rebuilds docker images and recreates containers.
- **Fast Mode** (`--fast`): Only updates files and restarts containers (no rebuild).
- **Target Selection**:
  - `[NoWx]`: Essential for updates without killing the WeChat bot process (preserves login session).
  - `[All]`: Full system reset/update.

### Configuration
Edit `scripts/deploy.py` to set your server details:
```python
DEFAULT_SERVER_IP = "192.168.x.x"
DEFAULT_USERNAME = "user"
```

## 5. Database Setup

### PostgreSQL Setup (Recommended)
BotMatrix now supports PostgreSQL as the primary database with SQLite as a fallback option.

1. **Install PostgreSQL** (if not already installed):
```bash
# Ubuntu/Debian
sudo apt-get install postgresql postgresql-contrib

# CentOS/RHEL
sudo yum install postgresql-server postgresql-contrib
sudo postgresql-setup initdb
```

2. **Create Database and User**:
```bash
# Connect to PostgreSQL
sudo -u postgres psql

# Create database and user
CREATE DATABASE botmatrix_db;
CREATE USER botmatrix WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE botmatrix_db TO botmatrix;
\q
```

3. **Configure Environment Variables**:
Update your `.env` file with PostgreSQL configuration:
```bash
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=botmatrix
DB_PASSWORD=your_password
DB_NAME=botmatrix_db
DB_SSL_MODE=disable
```

4. **Database Migration**:
BotMatrix will automatically create the necessary tables on startup. No manual migration is required.

### SQLite Setup (Fallback)
For development or small deployments, SQLite is still supported:
```bash
DB_TYPE=sqlite
DB_PATH=./botmatrix.db
```

## 6. Troubleshooting

*   **Ports Occupied**: Check `docker-compose.yml` and change mapped ports (e.g., `5000:5000` -> `5050:5000`).
*   **Connection Failed**: Ensure `NEXUS_ADDR` in bot configs points to `ws://bot-manager:3005` (internal Docker network).
*   **Logs**: Use `docker-compose logs -f [service_name]` to debug specific bots.
*   **Connection Refused**: Ensure `bot-manager` is running and port `3005` is accessible.
*   **Docker Permission Denied**: Run with `sudo` or add user to `docker` group.
