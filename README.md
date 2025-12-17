# BotMatrix 🌌

**The Next-Generation Enterprise Bot Management System**
**新一代企业级 OneBot 机器人集群管理系统**

[![Go](https://img.shields.io/badge/Go-1.19%2B-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Python](https://img.shields.io/badge/Python-3.9%2B-blue?style=for-the-badge&logo=python)](https://www.python.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker)](Dockerfile)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge)](LICENSE)

---

## 📢 Recent Updates | 最近更新

### v1.1.67 - Temporary Fixed Routing
*   **🔧 Testing Routing Rules**: Added temporary routing rules to direct specific group or bot messages to fixed workers for testing purposes.
*   **🔐 Admin API**: New `/api/admin/routing` REST API for managing routing rules (admin only).
*   **🔄 Priority Routing**: Messages first check routing rules before falling back to round-robin load balancing.
*   **🛡️ Failure Recovery**: Automatic fallback to round-robin if fixed worker is unavailable.
*   **📊 UI Enhancements**: Overmind routing screen now displays worker handled counts and improved dropdown selection.

### v1.1.66 - Cross-Bot Message Prevention
*   **🔒 Enhanced Message Routing Security**: Completely removed fallback logic that could cause messages to be sent to incorrect bots.
*   **🚫 Strict Target Validation**: Messages with invalid or missing self_id (including "0") are now rejected instead of being routed to random bots.
*   **🔄 Simplified Worker Architecture**: Removed Worker-BotID binding logic as Workers are designed to be shared competing consumers.
*   **📋 Improved Error Logging**: Enhanced error messages to clearly indicate when messages are rejected due to invalid target bot identification.
*   **🏷️ Worker ID Tracking**: Added unique ID assignment for better Worker connection tracking and debugging.

### v1.1.65 - Message Queue System Enhancement
*   **🔄 Separated Retry Queue**: Completely separated message persistence queue from retry queue to prevent successful messages from being re-sent.
*   **💾 Retry Queue Isolation**: Failed messages now go into a dedicated retry queue instead of being mixed with the persistence queue.
*   **📊 Queue Management APIs**: Added new REST API endpoints (`/api/queue/messages` and `/api/queue/retries`) for monitoring both persistence and retry queue status.
*   **🔧 Enhanced Retry Logic**: Improved retry processing to only handle messages in the retry queue, preventing interference with successful message deliveries.

### v1.1.64 - System Reliability & Stability
*   **🔄 Worker Disconnect Detection**: Automatic heartbeat monitoring with 60-second timeout detection and cleanup.
*   **💾 Message Persistence**: In-memory message queue prevents message loss during Worker disconnections.
*   **🔄 Intelligent Message Retry**: Exponential backoff retry mechanism (1min, 2min, 4min) with max 3 attempts.
*   **🤖 Bot Heartbeat Monitoring**: Automatic heartbeat tracking for Bot connections with 5-minute timeout detection and cleanup.
*   **🎯 Enhanced Message Routing**: Improved target bot selection logic with debug logging to prevent incorrect message routing.
*   **🔧 Compile Error Fix**: Resolved WxBotGo build failure by adding missing "os" package import.
*   **🔒 Thread Safety**: Enhanced mutex locking for all shared resource operations.

### v1.1.63 - Security & I18n
*   **🔒 Security**: Implemented optional token authentication for WebSocket connections to prevent unauthorized access.
*   **🌏 I18n**: Enhanced internationalization for Overmind integration and sidebar menus.
*   **🐳 Docker Management**: New sidebar menu for visualizing and managing Docker containers (view status, logs, etc.).
*   **🧠 Overmind Integration**: Seamless access to Overmind system via the sidebar for unified control.
*   **✨ UI Enhancements**: Optimized sidebar structure and added multi-language support for new modules.

> For detailed update history, please refer to [CHANGELOG.md](CHANGELOG.md).
>
> 更多更新记录请查看 [CHANGELOG.md](CHANGELOG.md)。

---

## 📖 Introduction | 简介

**BotMatrix** is a high-performance, distributed robot management platform designed for enterprise scale. It decouples the connection layer from the logic layer, allowing for massive scalability and robust management.

*   **BotNexus (The Core)**: A high-concurrency Gateway written in **Go**. It provides a unified WebSocket interface, REST API, and a powerful **Real-time Dashboard**.
*   **SystemWorker (The Brain)**: A system-level **Python** worker. It handles global commands, visualizes status, and executes remote code.
*   **WxBot (The Bot)**: A flexible Bot Node written in **Python**. It handles protocol adaptation (WeChat/OneBot) and executes business logic.
*   **WeComBot (Enterprise)**: A **Go-based** implementation for WeChat Work (WeCom), supporting internal app integration via callbacks.
*   **NapCat (Personal)**: A **Containerized** implementation for Personal QQ, utilizing NTQQ and OneBot 11.
*   **TencentBot (The Official Worker)**: A high-performance Worker written in **Go**, utilizing the official Tencent Bot SDK (`botgo`) for stable, compliant QQ Guild and Group operations.
*   **DingTalkBot (The Enterprise Worker)**: A **Go-based** implementation supporting DingTalk's Webhook and Stream Mode for enterprise internal integration.
*   **FeishuBot (The Modern Worker)**: A **Go-based** implementation for Feishu/Lark, utilizing official WebSocket SDK for secure, firewall-friendly enterprise operations.
*   **TelegramBot (International)**: A **Go-based** implementation for Telegram, connecting via Long Polling.
*   **DiscordBot (Community)**: A **Go-based** implementation for Discord, supporting channel messages and DMs.
*   **SlackBot (Enterprise)**: A **Go-based** implementation for Slack, utilizing Socket Mode for enterprise integration.
*   **KookBot (Community)**: A **Go-based** implementation for Kook (Kaiheila), utilizing WebSocket for real-time interaction.
*   **EmailBot (Utility)**: A **Go-based** implementation for Email (IMAP/SMTP), bridging emails to OneBot messages.

---

## ✨ Key Features | 核心功能

### 📊 Real-Time Visual Analytics (实时可视化分析)
> Experience the heartbeat of your bot cluster.
*   **Dynamic Charts**: Live visualization of **CPU Usage**, **Memory Trends**, and **Message Throughput (QPS)**.
*   **System Health**: Monitor Goroutines, GC cycles, and server uptime in real-time.
*   **Process Monitor**: Top 10 high-resource processes table to keep server performance in check.

### 🤖 Advanced Bot Fleet Management (集群管理)
*   **Unified List**: View all connected bots with details like **IP Address**, **Connection Duration**, and **Owner**.
*   **Status Tracking**: Instant visibility into bot health and connectivity.
*   **Remote Control**: Manage specific bots directly from the dashboard.

### 👥 User & Group Insights (用户与群组洞察)
*   **Activity Ranking**: "Top 5 Active Groups" and "Top 5 Active Users" (Dragon King) leaderboards.
*   **Member Management**: Search, ban, kick, or modify card names for group members via a unified UI.

### 🔒 Enterprise Security (企业级安全)
*   **Role-Based Access**: Granular permissions for **Admins** and standard **Users**.
*   **Multi-User Auth**: Secure login system with token-based authentication.

### 🔥 Burn After Reading (阅后即焚)
> *Privacy First, Trace Free.*
*   **Secure & Ephemeral**: Auto-recall messages after a configurable delay (0-120s).
*   **High-End Privacy**: Ensure sensitive info vanishes without a trace.
*   **Universal Control**: Seamlessly integrated into the dashboard for supported bots.

### 📜 Centralized Log Management (集中式日志管理)
> *See it all in one place.*
*   **Unified View**: Monitor console logs from all connected bots directly in the dashboard.
*   **Smart Display**: Auto-truncates long logs with click-to-expand; auto-pauses refresh on interaction.
*   **Real-Time Streaming**: Logs are streamed via WebSocket with practically zero latency.
*   **Per-Bot Filtering**: Easily switch between viewing system logs or specific bot logs.

### 🐳 Docker Container Management (容器管理)
> *Control your infrastructure.*
*   **Visual Management**: View and manage Docker containers directly from the sidebar.
*   **Status at a Glance**: Check container ID, names, image, and running status instantly.
*   **One-Click Refresh**: Keep track of container changes in real-time.

### 🧠 Overmind Integration (Overmind 集成)
> *Next-level control.*
*   **Seamless Access**: One-click navigation to the Overmind frontend directly from the sidebar.
*   **Unified Workflow**: Manage your bot cluster and Overmind system without switching contexts.

### 🧠 Intelligent Robot Collaboration (机器人智能协作)
> *One Click, Full Linkage.*
*   **Smart Wake-Up**: Bypasses Tencent Guild Bot restrictions (passive reply only) by coordinating with ordinary bots to "wake up" guild bots via @mentions.
*   **Cross-Protocol Synergy**: Bridges the gap between different bot platforms (e.g., QQ Guild ↔ QQ Group).
*   **Unstoppable Workflow**: Ensures critical messages are delivered even under strict platform limitations. [Read More](docs/QQ_GUILD_SMART_SEND_CN.md).

---

## 🛠 Architecture | 架构

```mermaid
graph TD
    User["Admin / User"] -->|HTTPS / WSS| Nexus["BotNexus (Go Gateway)"]
    Nexus -->|Monitor| Dashboard["Web Dashboard"]
    
    subgraph "Worker Cluster"
        SystemWorker["SystemWorker (Python)"]
        WxBot["WxBot (Python)"]
        TencentBot["TencentBot (Go)"]
        DingTalkBot["DingTalkBot (Go)"]
        FeishuBot["FeishuBot (Go)"]
        TelegramBot["TelegramBot (Go)"]
        DiscordBot["DiscordBot (Go)"]
        SlackBot["SlackBot (Go)"]
        KookBot["KookBot (Go)"]
        EmailBot["EmailBot (Go)"]
        WeComBot["WeComBot (Go)"]
        NapCat["NapCat (Docker)"]
    end
    
    Nexus <-->|WebSocket| SystemWorker
    Nexus <-->|WebSocket| WxBot
    Nexus <-->|WebSocket| TencentBot
    Nexus <-->|WebSocket| DingTalkBot
    Nexus <-->|WebSocket| FeishuBot
    Nexus <-->|WebSocket| TelegramBot
    Nexus <-->|WebSocket| DiscordBot
    Nexus <-->|WebSocket| SlackBot
    Nexus <-->|WebSocket| KookBot
    Nexus <-->|WebSocket| EmailBot
    Nexus <-->|WebSocket| WeComBot
    Nexus <-->|WebSocket| NapCat
    
    WxBot <-->|Protocol| WeChat["WeChat Servers"]
    TencentBot <-->|OpenAPI| QQ["Tencent QQ Platform"]
    NapCat <-->|NTQQ| PersonalQQ["Personal QQ"]
    DingTalkBot <-->|Stream/Hook| DingTalk["DingTalk Cloud"]
    FeishuBot <-->|WebSocket/API| Feishu["Feishu Cloud"]
    TelegramBot <-->|Long Polling| Telegram["Telegram Cloud"]
    DiscordBot <-->|Gateway| Discord["Discord Cloud"]
    SlackBot <-->|Socket Mode| Slack["Slack Cloud"]
    KookBot <-->|WebSocket| Kook["Kook Cloud"]
    EmailBot <-->|IMAP/SMTP| Email["Email Servers"]
    WeComBot <-->|Callback/API| WeCom["WeCom Cloud"]
```

## 📂 Project Structure | 项目结构

```text
BotMatrix/
├── BotNexus/            # [Go] The Brain (Gateway & Dashboard)
│   ├── main.go          # Core Logic
│   ├── index.html       # Modern Responsive UI (Bootstrap 5 + Chart.js)
│   └── Dockerfile       # Deployment config
├── WxBot/               # [Python] The Brawn (WeChat Bot)
│   ├── bots/            # Business Logic
│   └── web_ui.py        # Legacy UI (Deprecated)
├── TencentBot/          # [Go] The Official (QQ Worker)
│   ├── main.go          # BotGo Implementation
│   └── config.json      # Bot Configuration
├── DingTalkBot/         # [Go] The Enterprise (DingTalk Worker)
│   ├── main.go          # Stream/Webhook Implementation
│   └── config.json      # Dual-mode Config
├── FeishuBot/           # [Go] The Modern (Feishu Worker)
│   ├── main.go          # WebSocket Implementation
│   └── config.json      # App Config
├── TelegramBot/         # [Go] The International (Telegram Worker)
│   ├── main.go          # Long Polling Implementation
│   └── config.json      # Bot Token Config
├── DiscordBot/          # [Go] The Community (Discord Worker)
│   ├── main.go          # Gateway Implementation
│   └── config.json      # Bot Token Config
├── SlackBot/            # [Go] The Enterprise (Slack Worker)
│   ├── main.go          # Socket Mode Implementation
│   └── config.json      # App/Bot Token Config
├── KookBot/             # [Go] The Community (Kook Worker)
│   ├── main.go          # WebSocket Implementation
│   └── config.json      # Bot Token Config
├── EmailBot/            # [Go] The Utility (Email Worker)
│   ├── main.go          # IMAP/SMTP Implementation
│   └── config.json      # Server/Auth Config
├── WeComBot/            # [Go] The Enterprise (WeCom Worker)
│   ├── main.go          # Callback/API Implementation
│   └── config.json      # App/Token Config
├── NapCat/              # [Docker] The Personal (QQ Worker)
│   ├── config/          # OneBot 11 Config
│   └── qq/              # QQ Session Data
└── docker-compose.yml   # One-Click Deployment
```

---

## 🏁 Quick Start (Docker) | 快速开始

### Prerequisites
*   Docker & Docker Compose
*   (Optional) Redis for data persistence

### 1. Deploy
```bash
git clone https://github.com/changliaotong/BotMatrix.git
cd BotMatrix

# Configure TencentBot (Optional)
cp TencentBot/config.sample.json TencentBot/config.json
# Edit TencentBot/config.json with your AppID and Secret

docker-compose up -d --build
```

### 2. Access
*   **Dashboard**: `http://localhost:5000` (Default Account: `admin` / `123456`)
*   **WebSocket Gateway**: `ws://localhost:3005`

### 3. Connect a Bot
The `WxBot` container will automatically try to connect to `BotNexus`.
1.  Open the Dashboard (`http://localhost:5000`).
2.  Watch the **Bot List** update in real-time as workers connect.
3.  Scan the QR code in the logs if required.

---

## 📚 Documentation | 文档导航

- **部署与维护 / Deployment & Ops**
  - [docs/DEPLOY.md](docs/DEPLOY.md) - **Docker 部署与远程自动部署脚本使用说明**（推荐从这里开始）
  - [docs/SERVER_MANUAL.md](docs/SERVER_MANUAL.md) - **旧版 Python OneBot 网关服务端手册（Legacy / 兼容文档）**  
    > 当前推荐架构为：`BotNexus (Go 网关)` + 各平台 Bot Worker + `SystemWorker`，Python 网关仅用于兼容老项目。

- **特性与协议说明 / Features & Protocols**
  - [docs/QQ_GUILD_SMART_SEND_CN.md](docs/QQ_GUILD_SMART_SEND_CN.md) - QQ 频道/群组 **智能协作发送（Smart Send / WakeUp）** 机制说明
  - [docs/CSharpReconnectExample.cs](docs/CSharpReconnectExample.cs) - C# 客户端重连示例代码

- **客户端 / Clients**
  - `Overmind/README.md` - Overmind 可视化控制台（Flutter 桌面/移动端）
  - `wx_bot_app/README.md` - 实验性移动端控制应用（Flutter）

- **子模块 / Bots & Workers**
  - `BotNexus/` - Go 网关与 Web 面板
  - `SystemWorker/README.md` - SystemWorker 系统 Worker 说明
  - 各平台 Bot：`WxBot/`, `TencentBot/`, `DingTalkBot/`, `FeishuBot/`, `TelegramBot/`, `DiscordBot/`, `SlackBot/`, `KookBot/`, `WeComBot/`, `EmailBot/` 等目录下均有各自的 `README.md`（使用前建议阅读对应文档）

---

*Made with ❤️ by BotMatrix Team*
