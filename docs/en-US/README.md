# BotMatrix 📚 Documentation Hub

> [🌐 English](README.md) | [简体中文](../zh-CN/README.md)
> [🏠 Back to Home](../../README.md)

Welcome to the BotMatrix Documentation Hub. This includes detailed system architecture, development guides, deployment processes, and feature descriptions.

## 🗺️ Documentation Navigation

### 🚀 Getting Started & Deployment
- **[Changelog](CHANGELOG.md)** - Project version iterations and change logs.
- **[Architecture](core/ARCHITECTURE.md)** - Understand the core design and component collaboration.
- **[Deployment Guide](core/DEPLOY.md)** - How to install and run BotMatrix in different environments.
- **[Container Best Practices](deployment/CONTAINER_BEST_PRACTICES.md)** - Deploying BotWorker and plugins with Docker/K8s.
- **[API Reference](core/API_REFERENCE.md)** - Detailed JSON message formats and communication protocols.
- **[Redis Upgrade](legacy/REDIS_UPGRADE.md)** - Notes on Redis dependency upgrades and configuration.
- **[WebUI Upgrade Plan](legacy/WEBUI_UPGRADE_PLAN.md)** - Roadmap for Web interface refactoring.

### 📖 Manuals
- **[Server Manual](core/SERVER_MANUAL.md)** - Detailed features, configuration, and admin commands.
- **[Plugin Manual](plugins/README.md)** - Usage guides for all built-in plugins.
- **[Plugin SDK Guide](plugins/sdk_guide.md)** - Develop efficient plugins with Go/Python/C# SDK.
- **[Market Specification (BMPK)](plugins/market_spec.md)** - Standardized plugin distribution protocol.
- **[Routing Rules](core/ROUTING_RULES.md)** - Detailed definition of message routing and distribution.
- **[Skill Switch & Compatibility](core/SKILL_SWITCH_AND_COMPATIBILITY.md)** - Bot skill management and OneBot compatibility.

### 🛠️ Core System Details
- **[Conflict & Improvement Plan](legacy/CONFLICT_PLAN.md)** - Known conflicts and future document optimizations.
- **[Component Documentation](components/README.md)** - Technical details of adapters and core components.
- **[MiniProgram Backend](core/MINIPROGRAM.md)** - Description of the mobile management platform.
- **[Fission System](plugins/FISSION_SYSTEM.md)** - Implementation of automated group growth.
- **[Baby & Marriage System](plugins/BABY_AND_MARRIAGE_SYSTEM.md)** - Logic for interactive fun features.
- **[Core Plugin](plugins/CORE_PLUGIN.md)** - Features and configuration of built-in core plugins.
- **[QQ Guild Smart Send](plugins/QQ_GUILD_SMART_SEND.md)** - Optimized message sending for QQ Guilds.

### 💻 Developer Guide
- **[Dev & Planning Hub](development/README.md)** - Plugin development, protocol compatibility, and planning.
- **[Plugin Development](plugins/PLUGIN_DEVELOPMENT.md)** - How to write custom plugins for BotMatrix.
- **[OneBot Compatibility](core/ONEBOT_COMPATIBILITY.md)** - Protocol implementation details and coverage.
- **[Website Plan](legacy/WEBSITE_PLAN.md)** - Construction plan for the official documentation site.

---

## 🏗️ System Components
1. **BotNexus**: The system brain, responsible for connection management, routing, and plugin scheduling.
2. **BotWorker**: Execution nodes, responsible for task processing and plugin execution.
3. **WebUI**: Management interface based on Vue 3 for monitoring and configuration.

---

[🏠 Back to Project Home](../../README.md)
