# Changelog | 更新日志

All notable changes to this project will be documented in this file.

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
    *   **Auto-Pause**: Log refreshing now automatically pauses when a user expands a log entry or selects text (e.g., for copying).
*   **TencentBot**:
    *   **Upload Logic**: Enhanced avatar uploading with automatic fallback from URL to multipart upload to bypass link blocking.
*   **Message Routing (消息路由)**:
    *   **API Response Broadcast**: Fixed an issue where API responses (Echo) were being load-balanced to random workers. They are now broadcast to all connected workers to ensure the requester receives the response.
    *   **Self-ID Injection**: Ensured `self_id` is always present in broadcasted messages to guarantee correct permission filtering and routing.
    *   **Group Member Check**: Added a new "Check Member" tool in the Group Actions tab, allowing admins to quickly verify if a specific user ID exists in the group (useful for checking status of users not in the cached list).
    *   **Log Center i18n**: Added missing internationalization support for the Log Center, including log filtering options, titles, and expand/collapse actions.
    *   **Float ID Routing**: Fixed a bug where `self_id` passed as a float (scientific notation) in API requests caused message routing failures (sending to wrong/random bot).

## v1.1.45 (2025-12-15)
*   **Centralized Log Management (集中式日志管理)**:
    *   **BotNexus Dashboard**: Added a new real-time log viewer with per-bot filtering capabilities.
    *   **Universal Streaming**: Bot clients (TencentBot, DingTalkBot, WxBot) now stream their console logs directly to BotNexus for centralized monitoring.
    *   **Architecture**: Implemented a scalable log aggregation protocol via WebSocket "log" events.

## v1.1.18 (2025-12-15)
*   **Feature Highlights**:
    *   **Burn After Reading (阅后即焚)**: Renamed and enhanced the Auto-Recall feature. Users can now set messages to self-destruct after 0-120s directly from the dashboard.
    *   **Smart Robot Collaboration (机器人智能协作)**: Officially documented the "Smart Send" mechanism that allows ordinary bots to "wake up" guild bots, bypassing platform restrictions.

## [Unreleased]

### Fixed
- Fixed WxBot displaying "unknown" for users and groups by prioritizing `RemarkName` (remark name) correctly in name resolution.

## v1.1.17 (2025-12-15)
*   **WxBot Core Improvements**:
    *   **System Message Parsing**: Fixed parsing errors for system events (Invite, Group Rename, Tickle) to prevent junk data in `wx_client`.
    *   **OneBot 11 Compliance**: Added native support for `group_increase` (Join/Invite), `group_update` (Name Change), and `poke` (Tickle) events.
    *   **Data Integrity**: Implemented strict guard clauses and ID validation to ensure only valid User IDs are stored.
    *   **Codebase Optimization**: Merged `wxclientv2` logic into the main `wxclient` to unify user matching algorithms and removed redundant files.
