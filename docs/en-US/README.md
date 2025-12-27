# BotMatrix 📚 Documentation Hub

> [🌐 English](README.md) | [简体中文](../zh-CN/README.md)
> [🏠 Back to Home](../../README.md)

Welcome to the BotMatrix Documentation Hub. This includes detailed system architecture, development guides, deployment processes, and feature descriptions.

## 🗺️ Navigation

### 🚀 Quick Start & Deployment
- **[Changelog](CHANGELOG.md)** - Project version iterations and feature changes.
- **[System Architecture](ARCHITECTURE.md)** - Core design principles and component collaboration.
- **[API Reference](API_REFERENCE.md)** - Detailed JSON message formats and communication protocols.
- **[Deployment Guide](DEPLOY.md)** - How to install and run BotMatrix in different environments.
- **[Redis Upgrade](REDIS_UPGRADE.md)** - Instructions on Redis dependency and configuration.
- **[WebUI Upgrade Plan](WEBUI_UPGRADE_PLAN.md)** - Roadmap for WebUI refactoring and upgrades.

### 📖 Manuals
- **[Server Manual](SERVER_MANUAL.md)** - Detailed feature descriptions, configuration guides, and admin commands.
- **[Plugin Manual](plugins/README.md)** - Detailed guide for all built-in plugins.
- **[Routing Rules](ROUTING_RULES.md)** - Definition of message routing and distribution logic.
- **[Skills & Compatibility](SKILL_SWITCH_AND_COMPATIBILITY.md)** - Bot skill management and OneBot compatibility.

### 🛠️ Core Systems
- **[Conflict & Improvement Plan](CONFLICT_PLAN.md)** - Records known conflicts and future optimization directions.
- **[Component Documentation](components/README.md)** - Technical details for adapters and core components.
- **[Miniprogram Admin](MINIPROGRAM.md)** - Documentation for the mobile miniprogram management platform.
- **[Fission System](FISSION_SYSTEM.md)** - Implementation of automated group invitations and user growth.
- **[Marriage System](BABY_AND_MARRIAGE_SYSTEM.md)** - Logic for interactive marriage and baby systems.
- **[Core Plugins](CORE_PLUGIN.md)** - Features and configuration of built-in core plugins.
- **[QQ Guild Smart Send](QQ_GUILD_SMART_SEND.md)** - Optimized message sending for QQ Guilds.

### 💻 Developer Guide
- **[Development & Planning Hub](development/README.md)** - Plugin development, protocol compatibility, and system planning.
- **[Plugin Development](PLUGIN_DEVELOPMENT.md)** - How to write custom plugins for BotMatrix.
- **[OneBot Compatibility](ONEBOT_COMPATIBILITY.md)** - Details on protocol implementation coverage.
- **[Website Plan](WEBSITE_PLAN.md)** - Planning for the official documentation website.

---

## 🏗️ System Components
1. **BotNexus**: The system brain, responsible for connection management, routing, and plugin scheduling.
2. **BotWorker**: Execution nodes, responsible for task processing and plugin execution.
3. **WebUI**: Management interface based on Vue 3 for monitoring and configuration.

---

[🏠 Back to Project Home](../../README.md)
