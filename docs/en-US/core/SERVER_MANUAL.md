# Server Manual

> [🌐 English](SERVER_MANUAL.md) | [简体中文](../zh-CN/SERVER_MANUAL.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

---

This document describes the features, configuration, plugin commands, and operations for the BotMatrix Server.

---

## 1. System Overview

The server core is a **OneBot Protocol Gateway** responsible for:
1.  **Connection Management**: Maintaining WebSocket connections with clients (WeChat/QQ).
2.  **Message Routing**: Distributing received messages to the plugin system and business ends.
3.  **Service Plugins**: Providing system monitoring, logging, broadcast notifications, etc.
4.  **Web Interface**: Providing QR code login, real-time logs, and status monitoring dashboard.

---

## 2. Web Console

The server includes a lightweight built-in web server.

- **Address**: `http://SERVER_IP:3001` (Default port 3001, configurable in `config.json`)
- **Main Features**:
    - **Dashboard**:
        - Real-time CPU / Memory usage charts.
        - Gateway connection counts, system uptime, and message throughput.
        - Real-time scrolling log window.
    - **Login (/login)**:
        - Access this page to get a QR code when the bot needs to re-login.
        - Auto-refreshes to detect login status.

---

## 3. Server Commands

Built-in management commands, **restricted to administrators**. Commands must start with `#`.

### 3.1 Permission Verification
Admin permissions are verified in two ways:
1.  **Config File**: WXID included in the `admins` list in `config.json`.
2.  **Database**: WXID registered in the `User` table and bound as an `AdminId` in the `Member` table.

### 3.2 Command List

| Module | Command | Example | Description |
| :--- | :--- | :--- | :--- |
| **Monitoring** | `#status` | None | View CPU, memory usage, uptime, and connection counts. |
| **System** | `#reload` | None | Hot-reload plugins and configuration. |
| **Messaging** | `#broadcast` | `Message` | Send a global broadcast message. |

---

*For detailed plugin development, see the [Plugin Development Guide](../../PLUGIN_DEVELOPMENT.md).*
