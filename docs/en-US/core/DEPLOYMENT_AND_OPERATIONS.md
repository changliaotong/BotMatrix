# 🚀 Deployment & Operations

> **Version**: 2.0
> **Status**: Production Ready
> [🌐 English](DEPLOYMENT_AND_OPERATIONS.md) | [简体中文](../zh-CN/core/DEPLOYMENT_AND_OPERATIONS.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

BotMatrix uses a containerized, modular architecture based on the principle of **"Stateless Workers + Stateful Plugins & Sessions"**.

---

## 1. Environment Setup

### 1.1 Requirements
- **Docker** 20.10+ & **Docker Compose** 2.0+
- **PostgreSQL** (Core DB)
- **Redis** (Cache & Async Tasks)

### 1.2 Quick Start
```bash
git clone https://github.com/changliaotong/BotMatrix.git
cd BotMatrix
cp KookBot/config.sample.json KookBot/config.json
docker-compose up -d --build
```

---

## 2. Configuration (BotNexus)

Configuration is managed via `config.json` and **Environment Variables** (higher priority).
- `WS_PORT`: WebSocket gateway port (default `:3001`).
- `REDIS_ADDR`: Redis address (e.g., `redis:6379`).
- `DB_HOST / DB_NAME`: Database connection details.
- `JWT_SECRET`: Security key for WebUI login.

---

## 3. Container Best Practices

### 3.1 Hot Reloading
Mount shared volumes (`/app/plugins`) for plugins. Use the `#reload` command or WebUI to trigger hot-reloading without container restarts.

### 3.2 Elastic Scaling
Use `deploy.replicas` in `docker-compose.yml` to scale Workers. Ensure all workers share a common file system or sync via the Market API.

---

## 4. Platform Deployment Quick-check

| Platform | Type | Key Points |
| :--- | :--- | :--- |
| **NapCat (QQ)** | OneBot 11 | Use pre-configured image, scan QR code at `:6099`. |
| **WxBot (WeChat)** | Python/OneBot | Check container logs or WebUI for QR code. |
| **DingTalk / Feishu** | Go/Stream | Config `AppID` & `Secret`, ensure callback accessibility. |

---

## 5. Performance Optimization

### 5.1 AI Parser Optimization
- **Pre-compilation**: Automatically pre-compiles regex patterns when workers report skills.
- **Regex Cache**: Thread-safe `regexCache` for O(1) lookups.

### 5.2 Redis Strategy
- **ConfigCache**: Workers sync rate limits and TTL from Redis every 30s to local memory.
- **SessionCache**: "Write-through" caching for active sessions (Local-first for reads, Async-to-Redis for writes).

---

## 6. Mobile Management (Miniprogram)

A WeChat miniprogram is available for mobile monitoring and management.
- **Features**: Real-time CPU/Memory alerts, Bot management, and Log filtering.
- **Tech Stack**: Native Miniprogram (Frontend) + Overmind REST API (Backend).
