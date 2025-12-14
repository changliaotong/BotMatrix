# BotMatrix Deployment Guide

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
*   **File**: `docker-compose.yml` (environment variables)
*   **Port**: `5000` (Web Dashboard), `3005` (WebSocket Gateway)
*   **Config**:
    *   `REDIS_ADDR`: Redis server address.
    *   `REDIS_PWD`: Redis password.

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

### 🏢 WeComBot (Enterprise WeChat)
*   **Type**: Go / Callback & API
*   **Config**: `WeComBot/config.json`
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

## 5. Troubleshooting

*   **Ports Occupied**: Check `docker-compose.yml` and change mapped ports (e.g., `5000:5000` -> `5050:5000`).
*   **Connection Failed**: Ensure `NEXUS_ADDR` in bot configs points to `ws://bot-manager:3005` (internal Docker network).
*   **Logs**: Use `docker-compose logs -f [service_name]` to debug specific bots.
