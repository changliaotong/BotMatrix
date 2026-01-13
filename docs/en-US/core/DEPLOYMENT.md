# 🚀 BotMatrix Deployment & Operations Guide
> [⬅️ Back to Docs](../README.md) | [🏠 Back to Home](../../README.md)

This guide describes how to deploy the **BotMatrix** ecosystem and provides best practices for containerized environments.

---

## 1. Quick Start (Docker Compose)
The recommended way to deploy BotMatrix is using Docker Compose.

```bash
# 1. Clone the repository
git clone https://github.com/changliaotong/BotMatrix.git
cd BotMatrix

# 2. Configure your bots (e.g., KookBot)
cp KookBot/config.sample.json KookBot/config.json
# Edit config.json with your credentials

# 3. Start the ecosystem
docker-compose up -d --build
```

---

## 2. Configuration Guide
BotMatrix uses a modular architecture. Configure only the components you need.

### 🧠 BotNexus (Core Manager)
- **Dashboard**: `http://localhost:5000`
- **Gateway**: `ws://localhost:3001` (OneBot V11 Reverse WS)
- **Database**: PostgreSQL (Mandatory for persistence)
- **Environment Variables**:
    - `REDIS_ADDR`: Redis server for task queuing.
    - `DB_HOST`, `DB_USER`, `DB_PASSWORD`: PostgreSQL credentials.
    - `JWT_SECRET`: Secret for auth tokens.

---

## 3. Container Best Practices
To achieve high availability and scalability, follow these principles:

### Stateless Worker + Stateful Plugins
- **Stateless Worker**: BotWorker nodes should be stateless to allow horizontal scaling.
- **State Management**: Use external Redis for session cache and PostgreSQL for persistent data.
- **Shared Storage**: Mount a shared volume (e.g., NFS) for plugin files if running multiple Workers.

### Zero-Downtime Updates
- **Canary Release**: Use the `canary_weight` in `plugin.json` to route a portion of traffic to new plugin versions.
- **Hot Loading**: Use the Plugin Market or `bm-cli` to update plugins without restarting the Worker container.

---

## 4. Troubleshooting & Monitoring
- **Logs**: Access live logs via the Dashboard or `docker logs -f bot-nexus`.
- **Health Checks**: Configure Docker health checks to automatically restart failed instances.
- **Resource Limits**: Set CPU and memory limits in `docker-compose.yml` to prevent resource exhaustion.

---
*Last Updated: 2026-01-13*
