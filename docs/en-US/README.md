# BotMatrix 📚 Documentation Hub
> [🌐 English](README.md) | [简体中文](../zh-CN/README.md)
> [🏠 Back to Home](../../README.md)

Welcome to the BotMatrix Documentation Hub. This includes detailed system architecture, development guides, deployment processes, and feature descriptions.

---

## 🗺️ Documentation Navigation

### 🚀 Core System
- **[System Architecture](core/ARCHITECTURE.md)** - Distributed design, routing logic, and skill system.
- **[Deployment Guide](core/DEPLOYMENT.md)** - Docker deployment, container best practices, and operations.
- **[Core Components](core/COMPONENTS.md)** - Technical details of Nexus, Worker, Overmind, and more.
- **[API Reference](core/API_REFERENCE.md)** - OneBot v11 compatibility and system-level API actions.

### 🛠️ Development & Plugins
- **[Plugin System & Dev](core/PLUGIN_SYSTEM.md)** - How to develop, package, and distribute plugins (SDKs, BMPK).
- **[Plugins Catalog](core/PLUGINS_CATALOG.md)** - Advanced features like Fission, Social Systems, and Smart Send.
- **[OneBot Compatibility](core/ONEBOT_COMPATIBILITY.md)** - Protocol implementation status across different adapters.

### 📖 History & Logs
- **[Changelog](CHANGELOG.md)** - Project version iterations and change logs.

---

## 🏗️ System Overview
1. **BotNexus**: The system brain, responsible for connection management, routing, and plugin scheduling.
2. **BotWorker**: Execution nodes, responsible for task processing and plugin execution.
3. **WebUI**: Management interface based on Vue 3 for monitoring and configuration.

---
[🏠 Back to Project Home](../../README.md)
