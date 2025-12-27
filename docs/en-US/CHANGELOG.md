# Changelog
[简体中文](../zh-CN/CHANGELOG.md) | [Back to Home](../../README.md) | [Back to Docs Center](../README.md)

All notable changes to this project will be documented in this file.

## v1.7.0 (2025-12-27)
*   **🧠 AI Task Orchestrator**:
    - **Natural Language Parsing**: Users can now describe automation tasks in plain Chinese/English (e.g., "Mute group every night at 11 PM").
    - **Structured Task Generation**: AI automatically generates cron-based tasks with correct actions and parameters.
    - **Draft & Confirmation Flow**: Introduced a multi-step workflow where AI proposes a task and the user confirms before deployment.
*   **🛡️ Advanced Routing & Security**:
    - **Strategy Management**: Added a new "Strategies" module for defining complex routing policies and reusable logic blocks.
    - **Shadow Rules**: Implemented non-intrusive message monitoring and redirection rules for auditing and testing.
*   **👤 Identity Mapping System**:
    - **Cross-Platform Identity**: Users from different platforms (QQ, WeChat, Discord) can now be mapped to a single "System Identity".
    - **Centralized Permission Management**: Permissions and stats are now tracked against the unified identity rather than fragmented platform IDs.
*   **🛠️ WebUI Modernization**:
    - **Vue 3 & Pinia Architecture**: Refactored the core dashboard to use Vue 3 Composition API and Pinia for ultra-responsive state management.
    - **System Capabilities Discovery**: Frontend now dynamically discovers backend actions, interceptors, and skills via a new capabilities API.
*   **📊 Real-time System Monitoring**:
    - **Enhanced Docker Integration**: Added live CPU/Memory usage tracking for bot containers directly in the dashboard.
    - **Nexus Status Dashboard**: Real-time visualization of global connection stats, including total bots, workers, and system uptime.

## v1.6.0 (2025-12-25)
*   **🌌 3D Visualization Overhaul**:
    *   **Enhanced Labels**: Increased node label font size for better visibility at a distance.
    *   **Sprite Orientation**: Fixed avatar sprite orientation to always face the camera, regardless of node rotation.
    *   **Visual Identification**: Improved user/bot identification and avatar handling for diverse platforms.
    *   **Dynamic Links**: Introduced a new blue-themed link palette with dynamic transmission flash.
    *   **User-Group Connectivity**: Added persistent, visible links between users and their respective groups in the 3D space.
*   **🛣️ Full Path Routing Visualization**:
    *   Implemented full-path event visualization: `User -> Group -> Nexus -> Worker`.
*   **📡 Smart Cache Enrichment**:
    *   **Metadata Caching**: BotNexus now automatically caches group names and user nicknames.
    *   **Message Enrichment**: Forwarded messages are enriched with cached metadata before reaching Workers.
*   **⚙️ System Reliability & Auto-Reply**:
    *   **Settings Persistence**: Fixed bugs in saving `LogLevel` and `AutoReply` settings.
    *   **Intelligent Auto-Reply**: Global maintenance notifications when the system is closed.

## v1.5.0 (2025-12-24)
*   **🐘 PostgreSQL-First Architecture**: Completely removed legacy SQLite support.
*   **🧹 Dependency Modernization**: Eliminated `modernc.org/sqlite` and related CGO-free drivers.
*   **⚙️ Unified Configuration WebUI**: Redesigned backend configuration for PostgreSQL and Redis Cluster.
*   **🛡️ Type-Safe API Communication**: Enforced strict numeric validation for port fields.
*   **🌐 Enhanced Localization**: Added translations for all new system configuration fields.

---
*(Check root CHANGELOG.md for full history)*
