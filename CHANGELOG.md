# Changelog | 更新日志

All notable changes to this project will be documented in this file.

## v1.2.5 (2025-12-21)
*   **🌐 Comprehensive i18n Fixes**: Resolved internationalization issues for "More" button popups in statistics, cloud matrix sorting labels, and implemented real-time language refresh for all dynamic lists (Bots, Workers, Groups, Friends).
*   **🖼️ Visual Identification Restored**: Re-integrated backend avatar proxy in the 3D routing visualizer and list views to fix CORS issues with QQ and group avatars.
*   **👥 Robust Group Management**: Fixed group member list loading for guild-type channels using fallback API logic; resolved mass message delivery failures for guild channels.
*   **🎨 UI/UX Polishing**: Optimized dashboard layout by reducing card spacing; repositioned access terminal status/protocol badges to the name suffix for better clarity.
*   **⚡ Real-time Sync**: Exposed render functions to the global window object to ensure UI components react instantly to language changes.

## v1.2.4 (2025-12-20)
*   **📡 Overmind Data Sync**: Fixed critical WebSocket port and API path issues in the Overmind Flutter service, ensuring real-time data flow in production environments.
*   **🔐 Seamless Auth Integration**: Implemented automatic JWT token and language preference passing for the embedded Overmind iframe, removing the need for double login.
*   **📐 Cloud Matrix UI Refinement**: Optimized the Cloud Matrix page with full-height layouts and independent scrollable lists for Workers and Bots; enhanced visual contrast with distinct background colors.
*   **🚀 Automated Deployment**: Synchronized Overmind Flutter Web builds with the BotNexus static file server for a unified "one-click" deployment experience.

## v1.2.3 (2025-12-20)
*   **🌐 Real-time i18n Sync**: Fixed an issue where the Routing Visualizer wouldn't update its language without a page refresh; implemented a dynamic update engine for 3D textures and UI components.
*   **🎨 Dynamic 3D Textures**: Added automatic re-generation of node textures on language change to ensure all labels (Nexus, Bot, Group, etc.) are correctly translated.

## v1.2.2 (2025-12-20)
*   **🛡️ Self-Deletion Protection**: Prevented administrators from deleting their own accounts to avoid accidental lockout.
*   **⚡ Robust User Operations**: Fixed a critical issue where user creation/deletion might prompt "failure" despite succeeding; added dual-layer checks (memory + database) and enhanced logging.
*   **📝 Improved Audit Logs**: Added detailed server-side logging for all administrative user management actions.

## v1.2.1 (2025-12-20)
*   **🛠️ Compilation Fixes**: Resolved duplicate method declaration for `sendWorkerHeartbeat` in `handlers.go`.
*   **🔢 Type Safety**: Fixed map type mismatch in `NewManager` initialization within `main.go`, ensuring consistent use of `map[string]int64` for user and group statistics.

## v1.2.0 (2025-12-20)
*   **🛡️ Docker Reliability**: Removed unstable `python:3.9-slim` image fallback in container management, replaced with clear error messaging for `botmatrix-wxbot` image availability.
*   **👥 Contact Data Consistency**: Fixed an issue where group and friend interfaces showed cached data without name/id; implemented explicit field mapping for consistent frontend display.
*   **🧠 Seamless Overmind Integration**: Embedded Overmind system now supports in-app loading with automatic language parameter passing and real-time data refreshing.
*   **🌐 Localization Perfection**: Resolved duplicate i18n labels for home page time metrics; optimized layout for "Uptime" and "Current Time" displays.
*   **📐 Adaptive Matrix Layout**: Refactored Cloud Matrix page background logic to use adaptive heights, ensuring the UI perfectly fits varied content lists.

## v1.1.99 (2025-12-20)
*   **⚙️ Live Config Panel**: Added a real-time settings panel to adjust 3D visualization parameters (distances, multipliers, spread, etc.).
*   **💾 Persistent Settings**: Visualization preferences are now saved to local storage and persist across refreshes.
*   **🎨 Dynamic Layout Engine**: Refactored the positioning logic to use configurable parameters, allowing for custom cosmic layouts.
*   **🛠️ Reset & Save**: Included quick reset to defaults and manual save options for the 3D universe.

## v1.1.98 (2025-12-20)
*   **🌌 Galactic Scale Expansion**: Increased all 3D node distances and distribution radii by a full order of magnitude (10x - 100x larger scene).
*   **🔭 Enhanced Camera Optics**: Adjusted camera far-plane and zoom limits (up to 50,000 units) to accommodate the massive new cosmic scale.
*   **✨ Massive Visual Overhaul**: Scaled up node sizes, particle effects, message hints, and star fields to match the expanded universe.
*   **🚀 High-Speed Interstellar Travel**: Optimized message particle speeds and dashed line dash sizes for the larger distances.

## v1.1.97 (2025-12-20)
*   **🌌 Dynamic Bot-Nexus Distance**: Bots now automatically adjust their distance from the Nexus core based on their group count, preventing central overcrowding.
*   **👥 Adaptive Group Clustering**: Group-to-Bot distance now scales dynamically with the total number of groups, ensuring a clear and spacious layout even for high-load bots.
*   **🧭 Navigation Persistence**: Fixed a critical issue where refreshing the "Visualization" page would redirect users back to the dashboard.
*   **✨ Smooth Real-time Re-positioning**: Implemented smooth lerp transitions for nodes as they dynamically shift their target positions in 3D space.

## v1.1.96 (2025-12-20)
*   **🖥️ Full-Screen Mode**: Added native full-screen support for the 3D routing map, enabling immersive real-time monitoring.
*   **💾 Docker Persistence**: Refactored database storage to a dedicated `data/` directory with persistent volume mapping, ensuring user data survives container restarts.
*   **🛡️ Permission Hardening**: Fixed database initialization errors in Docker by properly configuring directory ownership and non-root user permissions.
*   **🌐 Localization**: Integrated full-screen UI controls with multi-language support (CN/TW/EN).

## v1.1.95 (2025-12-19)
*   **🌌 3D Group Clustering & Persistence Sync (3D 群组聚类与持久化同步)**:
    *   **👥 Group Member Clustering**: Implemented automatic spatial clustering where group members gather around their respective group nodes in the 3D space, significantly improving topological clarity.
    *   **🌲 Tree-like Link Optimization**: Optimized the 3D visualization to use a hierarchical link structure (User -> Group -> Nexus -> Bot). This drastically reduces the number of overlapping lines and improves performance.
    *   **💾 Robust Persistence Sync**: Added SQLite persistence for global message counts and contact caches (friends/members). The system now automatically reloads historical data on startup.
    *   **📡 WebSocket Initial Sync**: Introduced a `SyncState` protocol for WebSocket connections. New subscribers now receive the complete historical cache (groups, friends, members, and stats) immediately upon connecting, ensuring no data loss on page refresh.
    *   **⏱️ Advanced Visual Lifecycle**: Added fade-out animations for message trails and holographic hints, preventing visual clutter while maintaining high-speed message particle effects (speed: 0.12).

## v1.1.94 (2025-12-19)
*   **🖼️ Avatar Proxy (头像代理)**:
    *   **解决跨域问题**: 新增后端头像代理接口 `/api/proxy/avatar`，解决 QQ 等第三方头像因 CORS 限制无法在 3D 画布显示的问题。
    *   **智能切换**: 前端自动识别外部 URL 并通过代理加载，确保 3D 节点头像 100% 成功渲染。

## v1.1.93 (2025-12-19)
*   **⚡ High-Speed Communication (高速通信特效)**:
    *   **🚀 Warp Particles**: Particle balls now travel at 5x speed to ensure visual synchronization with message delivery.
    *   **🕸️ Virtual Links**: Added persistent "Link Lines" between nodes during message exchange to visualize the connection path.
    *   **📍 Midpoint Anchoring**: Message text hints are now anchored at the midpoint of the link for better readability and structural clarity.
    *   **⏱️ TTL-based Persistence**: Communication visuals now stay visible for 3 seconds before gracefully fading out, independent of particle movement.

## v1.1.92 (2025-12-19)
*   **👤 User Node Enhancements (用户节点强化)**:
    *   **🖼️ Avatar Integration**: User avatars are now correctly fetched and displayed on user sprites in the 3D space.
    *   **🧲 Proximity Attraction**: Users who send messages now automatically move closer to the Nexus (Hub Center) for better visibility.
    *   **🌌 Dynamic Drifting**: Inactive users (no messages for >60s) smoothly drift back to the outer rim of the galaxy.
    *   **💎 Visual Identity**: Added unique cyan glow effects for user nodes to distinguish them from service nodes.

## v1.1.91 (2025-12-19)
*   **🌌 Visualization Optimization (可视化优化)**:
    *   **🧹 Aggressive Node Cleanup**: Offline Workers and Bots are now removed from the visualization almost immediately to prevent screen clutter.
    *   **🛡️ Fail-safe Retention**: Added a short grace period if the API fails to return any active nodes, preventing accidental total wipes.

## v1.1.90 (2025-12-19)
*   **🌌 Visualization & Visualizer Enhancements (可视化与渲染优化)**:
    *   **✨ Complete Flow Visualization**: Added missing `bot_to_user` routing event to complete the message flow visualization.
    *   **🔮 High-Clarity Holographic Hints**: Significant improvement in text clarity for holographic message hints.
        *   Increased canvas resolution by 2x.
        *   Switched to high-quality Chinese fonts (`PingFang SC`, `Microsoft YaHei`).
        *   Optimized texture filtering and disabled mipmaps for sharper rendering.
        *   Replaced blurry chromatic aberration effect with a clean drop shadow.

## v1.1.89 (2025-12-19)
*   **🌌 3D Visualization Refinement & Protocol Robustness (3D 可视化精细化与协议鲁棒性)**:
    *   **✨ Shaking Nexus Core**: Added a high-frequency shaking effect to the central Nexus node to symbolize its status as an active energy core.
    *   **🌌 Dynamic Cosmic Scaling**: Inactive user nodes now smoothly drift to a further "Outer Rim" (radius > 1400), while active users stay in the core area (radius 800).
    *   **🌠 Full Trajectory Content Sync**: Message holographic hints now follow the exact 3D arc trajectory of the message particles, ensuring content stays with the visual representation.
    *   **♾️ Permanent User Nodes**: Changed the lifecycle management to keep user nodes permanently in the scene, using distance to represent activity level instead of removing them.
    *   **🔧 OneBot 11 Echo Normalization**: Fixed a critical issue where non-string `echo` fields (e.g., integers) caused type assertion failures. All echo fields are now normalized to strings.
    *   **📡 Consistent Platform Forwarding**: Added missing `platform` field to `RoutingEvent` to ensure consistent avatar matching and message tracking across different platforms.
    *   **🛡️ Undefined Variable Fix**: Resolved a compile error in `handlers.go` where `hasEcho` was used without being defined.
    *   **👥 WxBot Group Member Fix**: Fixed `get_group_member_info` in `WxBot` to correctly identify group members by mapping IDs through `wx_client.get_client_uid`.

## v1.1.88 (2025-12-19)
*   **🌌 3D Cosmic Visualization & Load Balancing Optimization (3D宇宙可视化与负载均衡优化)**:
    *   **✨ 3D Routing Visualizer**: Refactored the forwarding path page into a high-performance Three.js 3D environment. Nodes now float organically in a cosmic starfield with smooth drifting animations and procedural star backgrounds.
    *   **🔮 Holographic Message Hints**: Introduced animated holographic message hints that float up from nodes, displaying message content and type in real-time with smooth fading effects.
    *   **⚖️ Intelligent Load Balancing**: Upgraded the worker selection algorithm to prioritize `AvgProcessTime` (message processing time) over RTT. This ensures that workers with lower computational load are selected for message handling, optimizing overall system throughput.
    *   **🖼️ High-Tech Avatar System**: Implemented matrix-inspired 3D node sprites with integrated platform avatars. Features glow effects, centered labels, and dynamic type indicators (e.g., "PROCESSOR", "MATRIX BOT").
    *   **👤 Dynamic User Clustering**: Added user nodes to the 3D space. Implemented a "drifting" logic where inactive users slowly move to the periphery of the galaxy, while active ones stay near the core.
    *   **🛡️ Automated Scene Management**: Optimized node lifecycle management to automatically remove offline nodes and cleanup resources (textures, materials) to prevent memory leaks in long-running sessions.
    *   **🔢 Normalized Resource Tracking**: Fixed an issue where CPU usage could exceed 100% in multi-core systems by normalizing process CPU usage against total core count.
    *   **📊 Real-time System Metrics**: Added a 3D overlay displaying global message counts, active node statistics, and ongoing processing tasks.
    *   **🛠️ Message Forwarding & Avatar Fixes**:
        *   Fixed an issue where passive replies from workers failed due to missing `self_id` and incorrect ID types.
        *   Enhanced avatar support for non-QQ platforms (WeChat, Tencent) and ensured avatars are correctly synchronized in the 3D view.
        *   Improved bot info retrieval to handle both `self_id` and `id` formats.
        *   Updated `getOrCreateNode` to dynamically update labels and avatars for existing nodes.

## v1.1.87 (2025-12-19)
*   **UI Refresh & Routing Resilience (UI刷新与路由弹性增强)**:
    *   **📊 5-Column Stats Grid**: Re-engineered the dashboard statistics layout with a responsive 5-column grid, improving data density and visual balance for key metrics.
    *   **📅 Enhanced Time Display**: Added full date (YYYY-MM-DD) to the system time display on the dashboard for better temporal context.
    *   **⚡ Auto-Compact View Mode**: Implemented intelligent list rendering for Bots (>12) and Workers (>8). The system now automatically switches to "Compact Mode" to maintain usability with large terminal counts, while still allowing manual overrides.
    *   **🧩 Robust Sorting & Validation**: Fixed critical JavaScript errors in Group and Friend lists (`localeCompare` of undefined, `contacts.filter` is not a function) by adding strict array validation and undefined checks.
    *   **🛡️ Self-Healing Group Routing**: Optimized `forwardWorkerRequestToBot` to automatically clear stale group-bot mappings when "Not in Group" (retcode: 1200) or "Removed" errors are detected.
    *   **🔄 Real-time Cache Sync**: Enhanced message processing to update group-bot associations on every incoming message and asynchronously trigger bot info refreshes after routing failures.
    *   **📈 Resource Visualization**: Added intuitive CPU and Memory usage progress bars to the dashboard for at-a-glance system health monitoring.
    *   **🌍 I18n Standardization**: Synchronized internationalization labels across the entire dashboard, specifically adding missing worker and runtime stats for both Chinese and English locales.
    *   **⏱️ RTT Tracking Fix**: Resolved a delimiter conflict (switching from `:` to `|`) and fixed missing timestamps to ensure accurate Round-Trip Time (RTT) monitoring for all worker connections.

## v1.1.86 (2025-12-19)
*   **Intelligent API Routing & Group Awareness (智能API路由与群组感知)**:
    *   **🎯 Context-Aware API Routing**: Significantly improved `forwardWorkerRequestToBot` to intelligently route API requests (like `get_group_member_info`) based on `self_id` or `group_id`. This resolves the "Group member does not exist" (retcode: 1200) error caused by routing requests to the wrong bot.
    *   **📂 Group-Bot Mapping Cache**: Enhanced the bot initialization process to automatically fetch and cache the entire group list (`get_group_list`) for each bot. This ensures the system knows which bot manages which group even before any messages are exchanged.
    *   **🔍 Routing Debug Transparency**: Added detailed `[ROUTING]` logs for API requests, showing the target bot, the action being performed, and the routing source (self_id, group_id cache, or fallback).
    *   **🛡️ Robust Error Handling**: Added explicit error reporting for cases where no bot can be found for a specific group, preventing silent failures and providing clear feedback to workers.

## v1.1.85 (2025-12-19)
*   **Routing & UI Optimization (路由与UI优化)**:
    *   **💾 Routing Rule Persistence**: Implemented SQLite-based persistence for routing rules. Rules are now saved to `botnexus.db` and automatically reloaded on system restart, resolving the issue where rules disappeared.
    *   **🛡️ Enhanced Routing Resilience**: Improved `forwardMessageToWorker` to handle cases where a routing rule points to a non-existent or offline worker. The system now logs a warning and automatically falls back to the load balancer.
    *   **🔍 Detailed Routing Logs**: Added explicit logging for routing decisions, including matched rules (exact/pattern), target workers, and fallback scenarios to assist in debugging.
    *   **📊 Dashboard UI Consolidation**: Merged Goroutines, active groups/users, and message statistics into a single rotating "Combined Stats" block on the dashboard, optimizing screen real estate.
    *   **🆔 Bot Info Synchronization**: Fixed a critical issue where bot IDs and nicknames were not correctly updated in the global map after fetching login info from QQ API. This ensures consistent avatar display and identification.

## v1.1.84 (2025-12-19)
*   **Persistent Configuration System (持久化配置系统)**:
    *   **⚙️ Dynamic Config Management**: Implemented a structured configuration system with persistent storage support via `config.json`.
    *   **🔌 Port Configuration UI**: Added a dedicated "Backend Configuration" section in the system settings for admin users to modify WebSocket and WebUI ports dynamically.
    *   **💾 Multi-layered Config Loading**: Established a robust config loading priority: Default Values → `config.json` → Environment Variables (highest priority).
    *   **🛡️ Secure Admin API**: Added protected REST API endpoints (`/api/admin/config`) for retrieving and updating system settings with admin-only access control.
    *   **🔧 Backward Compatibility**: Retained support for existing environment variables (`WS_PORT`, `WEBUI_PORT`, etc.) to ensure seamless upgrades and deployment flexibility.

## v1.1.83 (2025-12-19)
*   **Bot ID Display & Routing Enhancements (机器人ID显示与路由增强)**:
    *   **🆔 Bot Identification Fix**: Fixed issue where bot `self_id` was displayed as IP/Port. Bots now correctly identify themselves via handshake headers or dynamic message analysis, switching from temporary IP-based IDs to real QQ IDs automatically.
    *   **🎯 Robust Routing Rules**: Fully implemented and fixed the routing rule engine. Supported patterns include exact matches (`user_123456`, `group_789012`) and wildcard matches (`*_test`, `123*`).
    *   **⚡ Priority-Based Routing**: Implemented a strict priority system: Exact Match > Wildcard Match > RTT-based Load Balancing.
    *   **🙈 UI Privacy Update**: Hidden sensitive operating system information from the dashboard stat cards for better privacy and cleaner interface.
    *   **🛡️ Routing Resilience**: Added automatic fallback to load balancer if a rule-designated worker is offline or fails to respond.

## v1.1.82 (2025-12-19)
*   **System Dashboard & API Fixes (系统仪表盘与API修复)**:
    *   **🔧 API Response Format Fix**: Corrected `/api/bots` and `/api/workers` endpoints to return raw JSON arrays instead of wrapped objects, aligning with frontend expectations.
    *   **📊 Worker Metrics Addition**: Added `worker_count` to global system stats API (`/api/stats`) and implemented a new "Workers" metric card in the Web UI dashboard.
    *   **✅ Bot Data Enrichment**: Added missing `self_id` and `is_alive` fields to bot API responses, fixing the bot selection dropdown and online status filtering.
    *   **⚡ Real-time Stat Updates**: Updated frontend logic to correctly synchronize bot and worker counts across badges and dashboard metrics.
    *   **💾 Cache & Persistence**: Enhanced state persistence by including worker counts in local storage cache for immediate display on page load.

## v1.1.81 (2025-12-19)
*   **🔐 Unified Login System**: Redesigned a modern, full-page responsive login interface consistent across the ecosystem.
*   **💾 SQLite Database Integration**: Replaced temporary Redis user storage with persistent SQLite database, resolving the issue of losing passwords after service restarts.
*   **🛡️ Enhanced Security**: Improved admin initialization with automatic password hashing (bcrypt) and secure session management.
*   **📊 Comprehensive Dashboard**: Fixed system statistics rendering, ensuring all metrics (CPU, Memory, Goroutines, Bot Count) are displayed accurately without "undefined" values.
*   **🔗 SSO Integration**: Seamless single sign-on (SSO) between BotNexus and Overmind web UI via secure token passing.
*   **🌐 Localization Fixes**: Fixed login button functionality and multi-language support for the unified login page.
*   **📈 Real-time Monitoring**: Added missing system stats endpoints and implemented auto-refresh logic for the monitoring dashboard.
*   **📡 WebSocket Stability**: Fixed persistent WebSocket connection errors by correcting URL construction and ensuring proper endpoint mapping in the backend.

## v1.1.80 (2025-12-19)
*   **🎨 UI Modernization**: Comprehensive overhaul of the management dashboard with improved responsiveness and dark mode support.
*   **🧠 Overmind Web Integration**: Embedded Overmind's powerful management tools directly into the BotNexus dashboard.
*   **📊 Trend Visualization**: Added historical trend charts for CPU, Memory, and Message throughput.

## v1.1.70 (2025-12-18)
*   **🔄 Automatic Message Retry**: Added intelligent retry mechanism for failed bot message deliveries, ensuring reliable message transmission.
*   **⏱️ Exponential Backoff**: Implemented smart retry timing (1s, 2s, 4s intervals) to prevent system overload during recovery attempts.
*   **🎯 Max Retry Limit**: Configurable retry attempts (default: 3) with automatic cleanup of expired messages after 5 minutes.
*   **📊 Queue Management**: Enhanced `/api/queue/messages` API provides detailed retry queue status including retry count, next retry time, and error details.
*   **⚡ Background Processing**: Dedicated background worker processes retry queue every 5 seconds for efficient message recovery.
*   **🔒 Thread-Safe**: All retry operations are protected by mutex locks for concurrent access safety.

## v1.1.69 (2025-12-18)
*   **Worker-Bot Bidirectional Communication (Worker-Bot双向通信)**:
    *   **🔧 Request-Response Mapping**: Implemented complete request-response mapping system using echo field to track pending requests.
    *   **🔄 Worker→Bot Request Forwarding**: Workers can now send API requests (with echo) to bots for operations like member checks, admin verification, muting, or kicking.
    *   **📨 Bot→Worker Response Relay**: Bot responses are automatically relayed back to the originating worker using the echo identifier.
    *   **⏱️ Timeout Management**: 30-second timeout for pending requests with automatic cleanup and error response generation.
    *   **🛡️ Error Handling**: Comprehensive error handling for unavailable bots, forwarding failures, and request timeouts with appropriate error codes (1404, 1400, 1401).
    *   **🧪 Test Interface**: Added `test_worker_bot_api.html` for comprehensive testing of bidirectional communication scenarios.
    *   **🔒 Thread-Safe Operations**: All request-response operations are protected by mutex locks for concurrent access safety.

*   **Overmind Mini-Program Updates (Overmind小程序更新)**:
    *   **✅ API Address Fix**: Fixed API and WebSocket address configuration in `miniprogram_api.js` and `app.js` to use correct backend port (3001).
    *   **✅ Data Visualization Enhancement**: Improved system monitoring charts using `chart_util.js` for CPU, memory, and network metrics.
    *   **✅ WebSocket Connection Optimization**: Updated WebSocket configuration for better connection stability.
    *   **✅ Documentation Update**: Updated `PROJECT_SUMMARY.md` and mini-program `README.md` with latest feature status.

## v1.1.68 (2025-12-17)
*   **Message Retry Mechanism (消息重试机制)**:
    *   **Automatic Retry Queue**: Added automatic message retry mechanism when bot message sending fails, ensuring reliable message delivery.
    *   **Exponential Backoff**: Implemented exponential backoff strategy (1s, 2s, 4s) to prevent system overload during retries.
    *   **Max Retry Limit**: Configurable maximum retry attempts (default: 3) with automatic cleanup of expired messages after 5 minutes.
    *   **Queue Management API**: Enhanced `/api/queue/messages` endpoint to return detailed retry queue status including retry count, next retry time, and error information.
    *   **Background Processing**: Dedicated background goroutine processes retry queue every 5 seconds for efficient message recovery.
    *   **Thread-Safe Operations**: All retry queue operations are protected by mutex locks for concurrent access safety.
*   **Routing Logic Fix & Enhanced Worker Management (路由逻辑修复与Worker管理增强)**:
    *   **🎯 Corrected Routing Logic**: Fixed message routing to properly distinguish between API requests (random worker selection) and message events (routing rule application).
    *   **🔧 Worker ID Optimization**: Shortened worker IDs for better readability and management.
    *   **🔄 Duplicate ID Prevention**: Added retry mechanism to prevent duplicate worker IDs with 10-attempt retry loop.
    *   **💓 Enhanced Heartbeat**: Improved worker connection stability with ping/pong mechanism and 60-second timeout detection.
    *   **📊 Routing Test Tool**: Added `test_routing_simple.html` for easy validation of routing rule functionality.
    *   **🛡️ Load Balancing**: API requests now use proper round-robin load balancing when no target bot is available.

## v1.1.67 (2025-12-17)
*   **Worker Heartbeat Fix (Worker心跳修复)**:
    *   **Targeted Heartbeat Updates**: Fixed worker heartbeat logic to only update the specific worker that sent the heartbeat, instead of updating all workers.
    *   **Worker ID Tracking**: Added worker_id field to worker messages for proper heartbeat identification.
    *   **Connection Stability**: Improved worker connection stability by preventing false timeout disconnections.
    *   **Warning Logs**: Added warning logs when worker heartbeats are received without proper worker_id identification.
*   **Temporary Fixed Routing (临时固定路由)**:
    *   **Group/Bot Routing Rules**: Added temporary routing rules to direct specific group or bot messages to a fixed worker for testing purposes.
    *   **Admin API**: Implemented `/api/admin/routing` REST API for managing routing rules (admin only).
    *   **Priority-Based Routing**: Messages are first checked against routing rules before falling back to round-robin load balancing.
    *   **Failure Recovery**: If the fixed worker is unavailable, the system automatically falls back to round-robin distribution.
    *   **Overmind UI Integration**: Enhanced Overmind routing screen to display worker handled counts and improve dropdown selection.

## v1.1.66 (2025-12-17)
*   **Cross-Bot Message Prevention (防止跨机器人消息发送)**:
    *   **Enhanced Message Routing Security**: Completely removed fallback logic that could cause messages to be sent to incorrect bots.
    *   **Strict Target Validation**: Messages with invalid or missing self_id (including "0") are now rejected instead of being routed to random bots.
    *   **Simplified Worker Architecture**: Removed Worker-BotID binding logic as Workers are designed to be shared competing consumers.
    *   **Improved Error Logging**: Enhanced error messages to clearly indicate when messages are rejected due to invalid target bot identification.
    *   **Worker ID Tracking**: Added unique ID assignment for better Worker connection tracking and debugging.

## v1.1.65 (2025-12-17)
*   **Message Queue System Enhancement (消息队列系统增强)**:
    *   **Separated Retry Queue**: Completely separated message persistence queue from retry queue to prevent successful messages from being re-sent.
    *   **Retry Queue Isolation**: Failed messages now go into a dedicated retry queue (`RetryQueue`) instead of being mixed with the persistence queue (`MessageQueue`).
    *   **Enhanced Retry Logic**: Improved retry processing to only handle messages in the retry queue, preventing any interference with successful message deliveries.
    *   **Queue Management APIs**: Added new REST API endpoints (`/api/queue/messages` and `/api/queue/retries`) for monitoring both persistence and retry queue status.
    *   **Improved Error Handling**: Enhanced error logging and message format validation to ensure only valid retry messages are processed.

## v1.1.64 (2025-12-17)
*   **System Reliability Enhancements (系统可靠性增强)**:
    *   **Worker Disconnect Detection**: Implemented automatic heartbeat monitoring for Worker connections with 60-second timeout detection and cleanup mechanism.
    *   **Message Persistence Queue**: Added in-memory message queue to prevent message loss during Worker disconnections, with automatic message replay for new Workers.
    *   **Message Retry Mechanism**: Implemented intelligent message retry system with exponential backoff (1min, 2min, 4min) and maximum 3 retry attempts for failed message deliveries.
    *   **Bot Heartbeat Monitoring**: Added automatic heartbeat tracking for Bot connections with 5-minute timeout detection and cleanup mechanism to prevent message routing to disconnected bots.
    *   **Enhanced Message Routing**: Added debug logging and improved target bot selection logic to prevent messages from being sent to incorrect or disconnected bots.
    *   **Compile Error Fix**: Fixed missing "os" package import in WxBotGo/core/bot.go causing build failures.
    *   **Thread Safety**: Enhanced all shared resource operations with proper mutex locking for concurrent access safety.

## v1.1.63 (2025-12-17)
*   **Security Enhancements (安全增强)**:
    *   **WebSocket Authentication**: Implemented an optional token-based authentication mechanism for BotNexus WebSocket connections.
        *   **Token Injection**: `WxBot` and `WxBotGo` now automatically inject a security token (if configured) when connecting to BotNexus.
        *   **Environment Variable**: Added `MANAGER_TOKEN` support in `docker-compose.yml` to securely propagate the token across services.
        *   **Soft Check**: Currently operates in "soft check" mode (logs warnings for invalid tokens but allows connection) to ensure backward compatibility during the transition period.
*   **Internationalization (国际化)**:
    *   **Overmind Localization**: Fixed missing Chinese translations in the Overmind homepage menu (e.g., "Overmind" -> "主宰系统").
    *   **Theme Toggle**: Added translation for the theme toggle button.
    *   **Menu Items**: Ensured all sidebar menu items and tooltips are correctly localized.

## v1.1.62 (2025-12-16)
*   **Web UI Enhancements (界面增强)**:
    *   **Docker Management**: Added a new "Docker Management" menu in the sidebar. Users can now view container status, logs, and perform basic actions (start/stop) directly from the Web UI.
    *   **Overmind Integration**: Added a direct navigation link to the **Overmind** frontend, enabling seamless switching between bot management and system visualization.
    *   **Menu Organization**: Reorganized sidebar menu for better accessibility.
*   **Documentation (文档)**:
    *   **Feature Updates**: Updated README to include details about the new Docker management and Overmind integration capabilities.

## v1.1.46 (2025-12-16)
*   **Napcat & OneBot Integration (Napcat 集成)**:
    *   **Compatibility**: Added fallback support for standard OneBot implementations (like Napcat) that don't support custom count actions. It now calculates counts from `get_group_list` and `get_friend_list`.
    *   **Performance**: Optimized bot info refresh interval to 1 hour with Redis caching for approximate tracking.
*   **Dashboard UI Improvements (界面优化)**:
    *   **Smart Log Display**: Implemented log truncation (300 chars) with "Click to Expand" for better readability.
    *   **Bot Status Badges**: Added intuitive status badges for bots (Online/Offline) in the bot list.
    *   **Worker Load Indicator**: Added visual indicator for worker load based on current message processing count.
