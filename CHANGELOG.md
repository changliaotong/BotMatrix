# Changelog | 更新日志

All notable changes to this project will be documented in this file.

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

## v1.1.16 (2025-12-15)
*   **New Integrations**:
    *   **KookBot**: Added support for Kook (Kaiheila) community platform.
    *   **EmailBot**: Added bidirectional Email support (IMAP/SMTP) bridged to OneBot.
    *   **WeComBot**: Added Enterprise WeChat support via Callback/API.
    *   **NapCat**: Integrated NapCat (NTQQ) for personal QQ account automation.
*   **Documentation**:
    *   Comprehensive deployment guide covering all 10+ bot platforms.

## v1.1.15 (2025-12-15)
*   **TencentBot Enhanced**:
    *   **Strict Separation**: Completely separated **QQ Group** and **Guild Channel** logic to align with platform concepts.
    *   **New APIs**: Added comprehensive support for Guild Channel management (Create/Delete Channels, Role Management, Member Kick, etc.).
    *   **Deployment**: Optimized targeted deployment via `deploy.ps1`.
*   **Documentation**:
    *   Updated README and Architecture diagrams to reflect the new multi-bot ecosystem.

## v1.1.14 (2025-12-15)
*   **Deployment**:
    *   **Configuration**: Simplified Docker deployment by consolidating configuration into `config.json` for TencentBot, removing duplicate environment variables.
    *   **Persistence**: Ensured configuration persistence via Docker volume mounting.
*   **Tencent Official Bot**:
    *   **Build Fix**: Resolved build errors related to token initialization and SDK compatibility.

## v1.1.13 (2025-12-15)
*   **New Bot Type**:
    *   **Tencent Official Bot**: Added support for Tencent's official QQ Bot platform (`QQOfficial`) using the official `botgo` SDK.
    *   **BotNexus Integration**: Native integration for official bots with correct platform identification and message translation to OneBot 11 standard.
*   **Data Accuracy Improvements**:
    *   **Dragon King Fix**: Excluded bot's own messages from "Top Active Users" statistics to ensure leaderboard accuracy.
    *   **Bot Status Fix**: Resolved a critical panic when accessing group counts for offline bots.
    *   **Platform Info**: Fixed platform display in the bot list to correctly show "QQOfficial" or other custom platforms instead of defaulting to "QQ".

## v1.1.12 (2025-12-15)
*   **Internationalization (i18n)**:
    *   **Complete Coverage**: Full support for **Simplified Chinese**, **Traditional Chinese**, and **English** across all UI components.
    *   **Debug Tools**: Added translations for the "Raw API" debugger and message type selectors.
*   **Core Stability**:
    *   **Data Accuracy**: Fixed daily statistics reset logic to ensure "Today's Active Users/Groups" are perfectly accurate.
    *   **OneBot Compatibility**: Enhanced ID parsing (int/string/float) for broader client compatibility (e.g., WXBot).
    *   **Group/Friend Sync**: Resolved issue where group and friend counts would display as 0 by improving synchronization logic.
*   **System Info**: Fixed Host OS and Kernel version display for accurate server monitoring.

## v1.1.11 (2025-12-14)
*   **UI/UX Overhaul**: 
    *   **Dark Mode**: Fully optimized dark theme support for Dashboard, including modals, tables, and charts.
    *   **Group Avatars**: Added visual identification for groups using QQ avatar API.
    *   **Layout Fixes**: Improved "Groups & Friends" page layout, alignment, and removed redundant headers.
    *   **System Info**: Enhanced hardware info display (Host OS, Kernel) on the dashboard.
*   **Data Accuracy**:
    *   **Real-time Stats**: Fixed "Today's Active Groups" and "Dragon King" to correctly reflect *today's* data instead of historical totals.
    *   **Consistency**: Ensured consistency between dashboard widgets and detailed statistics views.
*   **Performance**: Optimized WebSocket message handling for bot group/friend counts.
