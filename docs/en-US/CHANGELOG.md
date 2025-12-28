# Changelog
[简体中文](../zh-CN/CHANGELOG.md) | [Back to Home](../../README.md) | [Back to Docs Center](../README.md)

All notable changes to this project will be documented in this file.

## v1.7.1 (2025-12-28)
*   **🖼️ Avatar Standardization**:
    - **Real Avatar Integration**: Fully integrated official QQ group and user avatars, utilizing a backend proxy to bypass CORS and Referer restrictions.
    - **Smart Platform Logo Fallback**: Automatically displays platform logos for non-QQ protocols when the user ID exceeds 980000000000.
    - **Global Consistency**: Standardized avatar display across all modules, including Bot List, Contacts, Group Members, Message Logs, and Identity Mapping.
    - **Backend Proxy Support**: Introduced a new `/api/proxy/avatar` endpoint to ensure reliable loading of external avatar images within the WebUI.

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
    *   **Metadata Caching**: BotNexus now automatically caches group names and user nicknames from Napcat events.
    *   **Message Enrichment**: Forwarded messages are now enriched with cached metadata before reaching Workers.
*   **⚙️ System Reliability & Auto-Reply**:
    *   **Settings Persistence**: Fixed a critical bug where `LogLevel` and `AutoReply` settings were not saved correctly to the configuration file.
    *   **Intelligent Auto-Reply**: Implemented global `AutoReply` logic that automatically notifies users with a maintenance message when the system is closed.
*   **🧹 Dynamic Worker Management**: Added real-time state filtering to the 3D visualizer to ensure only online Worker nodes are displayed.

## v1.5.0 (2025-12-24)
*   **🐘 PostgreSQL-First Architecture**: Completely removed legacy SQLite support to streamline database operations and focus on enterprise-grade performance with PostgreSQL.
*   **🧹 Dependency Modernization**: Eliminated `modernc.org/sqlite` and related CGO-free SQLite drivers, reducing binary size and complexity.
*   **⚙️ Unified Configuration WebUI**: Redesigned the backend configuration interface in the WebUI to align with the new PostgreSQL-only architecture.
*   **🛡️ Type-Safe API Communication**: Enforced strict numeric validation for port fields in the WebUI and Go backend, resolving data type mismatches.
*   **🌐 Enhanced Localization**: Added comprehensive translations for all new PostgreSQL and system configuration fields in `locales.js`.

## v1.3.2 (2025-12-22)
*   **🧠 BotWorker Confirmation & Dialog Engine**: Added a shared confirmation/session utility layer in `BotWorker` using Redis (when available) or the database as fallback.
*   **🔊 Voice Reply & Burn-After-Reading**: Introduced per-group feature switches for AI voice replies and burn-after-reading auto-recall in `BotWorker`.
*   **🗄️ Shared Persistence Layer**: Added a GORM-based persistence layer and `dblib` helper to the `Common` and `BotWorker` modules.
*   **📚 Documentation Refresh**: Updated root and BotWorker documentation (including plugin docs and development notes).

## v1.3.1 (2025-12-22)
*   **🐾 BotWorker Feature Expansion**: Introduced a new **Pet System** plugin (`pets.go`), allowing users to adopt, feed, and interact with virtual pets.
*   **⌨️ Command Parsing Engine**: Added a robust `CommandParser` utility to handle structured bot commands with optional prefixes and parameters.
*   **🛡️ Group Management Testing**: Implemented a comprehensive testing suite for group management features, including a new `GROUP_MANAGER_TEST_PLAN.md`.
*   **🛠️ Developer Tooling**: Added several new testing scripts and CLI tools (`test_cli.go`, `test_pets.go`, etc.).
*   **📦 Integration Fix**: Formally integrated the `BotWorker` source code into the main repository by resolving nested Git repository conflicts.

## v1.3.0 (2025-12-22)
*   **🏗️ Architectural Decoupling**: Successfully separated `BotNexus` (Core Gateway) and `BotAdmin` (Management Backend).
*   **🚀 Gateway Optimization**: Removed redundant Docker management dependencies and legacy Admin HTTP handlers from `BotNexus`.
*   **📏 Code Standardization**: Enforced PascalCase across all shared types and methods in the `Common` package.
*   **🧹 Logic Consolidation**: Migrated duplicate statistics and monitoring functions to the `Common` package.
*   **🔧 Routing Refinement**: Simplified wildcard routing logic in the gateway for better performance and easier maintenance.

## v1.2.5 (2025-12-21)
*   **🌐 Comprehensive i18n Fixes**: Resolved internationalization issues for "More" button popups in statistics, cloud matrix sorting labels, and implemented real-time language refresh.
*   **🖼️ Visual Identification Restored**: Re-integrated backend avatar proxy in the 3D routing visualizer and list views to fix CORS issues.
*   **👥 Robust Group Management**: Fixed group member list loading for guild-type channels using fallback API logic.
*   **🎨 UI/UX Polishing**: Optimized dashboard layout by reducing card spacing.
*   **⚡ Real-time Sync**: Exposed render functions to the global window object to ensure UI components react instantly to language changes.

## v1.2.4 (2025-12-20)
*   **📡 Overmind Data Sync**: Fixed critical WebSocket port and API path issues in the Overmind Flutter service.
*   **🔐 Seamless Auth Integration**: Implemented automatic JWT token and language preference passing for the embedded Overmind iframe.
*   **📐 Cloud Matrix UI Refinement**: Optimized the Cloud Matrix page with full-height layouts and independent scrollable lists for Workers and Bots.
*   **🚀 Automated Deployment**: Synchronized Overmind Flutter Web builds with the BotNexus static file server.

## v1.2.3 (2025-12-20)
*   **🌐 Real-time i18n Sync**: Fixed an issue where the Routing Visualizer wouldn't update its language without a page refresh.
*   **🎨 Dynamic 3D Textures**: Added automatic re-generation of node textures on language change.

## v1.2.2 (2025-12-20)
*   **🛡️ Self-Deletion Protection**: Prevented administrators from deleting their own accounts to avoid accidental lockout.
*   **⚡ Robust User Operations**: Fixed a critical issue where user creation/deletion might prompt "failure" despite succeeding.
*   **📝 Improved Audit Logs**: Added detailed server-side logging for all administrative user management actions.

## v1.2.1 (2025-12-20)
*   **🛠️ Compilation Fixes**: Resolved duplicate method declaration for `sendWorkerHeartbeat` in `handlers.go`.
*   **🔢 Type Safety**: Fixed map type mismatch in `NewManager` initialization within `main.go`.

## v1.2.0 (2025-12-20)
*   **🛡️ Docker Reliability**: Removed unstable `python:3.9-slim` image fallback in container management.
*   **👥 Contact Data Consistency**: Fixed an issue where group and friend interfaces showed cached data without name/id.
*   **🧠 Seamless Overmind Integration**: Embedded Overmind system now supports in-app loading with automatic language parameter passing.
*   **🌐 Localization Perfection**: Resolved duplicate i18n labels for home page time metrics.
*   **📐 Adaptive Matrix Layout**: Refactored Cloud Matrix page background logic to use adaptive heights.

## v1.1.99 (2025-12-20)
*   **⚙️ Live Config Panel**: Added a real-time settings panel to adjust 3D visualization parameters.
*   **💾 Persistent Settings**: Visualization preferences are now saved to local storage.
*   **🎨 Dynamic Layout Engine**: Refactored the positioning logic to use configurable parameters.
*   **🛠️ Reset & Save**: Included quick reset to defaults and manual save options for the 3D universe.

## v1.1.98 (2025-12-20)
*   **🌌 Galactic Scale Expansion**: Increased all 3D node distances and distribution radii by a full order of magnitude.
*   **🔭 Enhanced Camera Optics**: Adjusted camera far-plane and zoom limits to accommodate the massive new cosmic scale.
*   **✨ Massive Visual Overhaul**: Scaled up node sizes, particle effects, message hints, and star fields.
*   **🚀 High-Speed Interstellar Travel**: Optimized message particle speeds and dashed line dash sizes for the larger distances.

## v1.1.97 (2025-12-20)
*   **🌌 Dynamic Bot-Nexus Distance**: Bots now automatically adjust their distance from the Nexus core based on their group count.
*   **👥 Adaptive Group Clustering**: Group-to-Bot distance now scales dynamically with the total number of groups.
*   **🧭 Navigation Persistence**: Fixed a critical issue where refreshing the "Visualization" page would redirect users back to the dashboard.
*   **✨ Smooth Real-time Re-positioning**: Implemented smooth lerp transitions for nodes.

## v1.1.96 (2025-12-20)
*   **🖥️ Full-Screen Mode**: Added native full-screen support for the 3D routing map.
*   **💾 Docker Persistence**: Refactored database storage to a dedicated `data/` directory with persistent volume mapping.
*   **🛡️ Permission Hardening**: Fixed database initialization errors in Docker by properly configuring directory ownership.
*   **🌐 Localization**: Integrated full-screen UI controls with multi-language support.

## v1.1.95 (2025-12-19)
*   **🌌 3D Group Clustering & Persistence Sync**:
    *   **👥 Group Member Clustering**: Implemented automatic spatial clustering.
    *   **🌲 Tree-like Link Optimization**: Optimized the 3D visualization to use a hierarchical link structure.
    *   **💾 Robust Persistence Sync**: Added SQLite persistence for global message counts and contact caches.
    *   **📡 WebSocket Initial Sync**: Introduced a `SyncState` protocol for WebSocket connections.
    *   **⏱️ Advanced Visual Lifecycle**: Added fade-out animations for message trails and holographic hints.

## v1.1.94 (2025-12-19)
*   **🖼️ Avatar Proxy**:
    *   **解决跨域问题**: Added backend avatar proxy interface `/api/proxy/avatar`.
    *   **智能切换**: Frontend automatically identifies external URLs and loads them via proxy.

## v1.1.93 (2025-12-19)
*   **⚡ High-Speed Communication**:
    *   **🚀 Warp Particles**: Particle balls now travel at 5x speed.
    *   **🕸️ Virtual Links**: Added persistent "Link Lines" between nodes during message exchange.
    *   **📍 Midpoint Anchoring**: Message text hints are now anchored at the midpoint of the link.
    *   **⏱️ TTL-based Persistence**: Communication visuals stay visible for 3 seconds before fading out.

## v1.1.92 (2025-12-19)
*   **👤 User Node Enhancements**:
    *   **🖼️ Avatar Integration**: User avatars are now correctly fetched and displayed.
    *   **🧲 Proximity Attraction**: Users who send messages move closer to the Nexus hub.
    *   **🌌 Dynamic Drifting**: Inactive users smoothly drift back to the outer rim of the galaxy.
    *   **💎 Visual Identity**: Added unique cyan glow effects for user nodes.

## v1.1.91 (2025-12-19)
*   **🌌 Visualization Optimization**:
    *   **🧹 Aggressive Node Cleanup**: Offline Workers and Bots are removed almost immediately.
    *   **🛡️ Fail-safe Retention**: Added a short grace period if the API fails to return any active nodes.

## v1.1.90 (2025-12-19)
*   **🌌 Visualization & Visualizer Enhancements**:
    *   **✨ Complete Flow Visualization**: Added missing `bot_to_user` routing event.
    *   **🔮 High-Clarity Holographic Hints**: Significant improvement in text clarity for holographic message hints.

## v1.1.89 (2025-12-19)
*   **🌌 3D Visualization Refinement & Protocol Robustness**:
    *   **✨ Shaking Nexus Core**: Added a high-frequency shaking effect to the central Nexus node.
    *   **🌌 Dynamic Cosmic Scaling**: Inactive user nodes smoothly drift to a further "Outer Rim".
    *   **🌠 Full Trajectory Content Sync**: Message holographic hints follow the exact 3D arc trajectory.
    *   **♾️ Permanent User Nodes**: Keep user nodes permanently in the scene.
    *   **🔧 OneBot 11 Echo Normalization**: Fixed non-string `echo` field type assertion failures.
    *   **📡 Consistent Platform Forwarding**: Added missing `platform` field to `RoutingEvent`.
    *   **🛡️ Undefined Variable Fix**: Resolved a compile error in `handlers.go`.
    *   **👥 WxBot Group Member Fix**: Fixed `get_group_member_info` in `WxBot`.

## v1.1.88 (2025-12-19)
*   **🌌 3D Cosmic Visualization & Load Balancing Optimization**:
    *   **✨ 3D Routing Visualizer**: Refactored the forwarding path page into a high-performance Three.js 3D environment.
    *   **🔮 Holographic Message Hints**: Introduced animated holographic message hints.
    *   **⚖️ Intelligent Load Balancing**: Upgraded worker selection algorithm to prioritize `AvgProcessTime`.
    *   **🖼️ High-Tech Avatar System**: Implemented matrix-inspired 3D node sprites.
    *   **👤 Dynamic User Clustering**: Added user nodes to the 3D space.
