# 🛠️ BotMatrix Core Components
> [⬅️ Back to Docs](../README.md) | [🏠 Back to Home](../../README.md)

This document provides technical details for the core services and adapters within the BotMatrix ecosystem.

---

## 1. BotNexus (The Brain)
BotNexus is the central hub for routing, scheduling, and management.
- **3D Visualization**: Real-time topology visualization based on Three.js.
- **Intelligent Routing**: RTT-aware routing and failover.
- **Web Dashboard**: Modern interface for monitoring, logs, and configuration.
- **Security**: RBAC permission model and JWT-based authentication.

## 2. BotWorker (The Limbs)
BotWorker is the execution node that handles business logic and plugins.
- **Protocol Support**: Full OneBot v11 compatibility.
- **Plugin Engine**: Multi-language support (Go, Python, C#).
- **Communication**: Reverse WebSocket and HTTP support.
- **State Management**: Distributed session tracking via Redis.

## 3. SystemWorker (The Cortex)
SystemWorker handles complex system-level logic and visualizations.
- **Visual Status**: Generates HD system status images via `#sys status`.
- **Remote Execution**: Secure Python execution for debugging via `#sys exec`.
- **Omni-Channel Broadcast**: Global message pushing across all platforms.

## 4. Overmind (Mobile Center)
Overmind is the mobile control center for remote monitoring.
- **Tech Stack**: Built with Flutter for cross-platform support.
- **Dashboard**: Real-time status and log streaming on mobile devices.
- **Aesthetic**: Sci-fi "Overmind" dark mode interface.

## 5. Miniprogram Admin
A WeChat Miniprogram for lightweight management.
- **Features**: Robot status monitoring, log viewing, and basic settings management.
- **Architecture**: Native WeChat Miniprogram integrated with Overmind APIs.

---
*Last Updated: 2026-01-13*
