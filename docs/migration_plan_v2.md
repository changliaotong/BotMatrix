# BotWorker Migration Plan: sz84 Functionality

## 1. Overview
The goal is to migrate the remaining functionality from the legacy C# `sz84` bot to the new Go-based `BotWorker`. The migration focuses on separating core system concerns (security, settings) from feature logic (plugins), utilizing a middleware-based architecture for cross-cutting concerns.

## 2. Architecture

### 2.1 Core System (Main Program)
- **Settings Management**: Global feature switches (Enable/Disable functions) and system configuration.
- **Security**: Blacklist/Whitelist management, Sensitive Word system.
- **Middleware Chain**:
  1.  **SecurityMiddleware**: Handles Blacklist checks.
  2.  **SensitiveWordMiddleware**: Handles sensitive word detection and actions (Recall, Mute, Kick).
  3.  **FriendlyMessageMiddleware**: Handles URL filtering and message formatting (Group VIP bypass).
- **Execution**: `CombinedServer` implements `ActionCaller` interface to execute OneBot actions triggered by middleware.

### 2.2 Plugin System (Non-Core Features)
- **Plugin Interface**: Standardized `PluginModule` interface for easy extension.
- **Entertainment Plugin**: Dice, Jokes, etc.
- **Economy Plugin**: Points, Sign-in, Transfer.
- **Future Plugins**: Custom business logic.

## 3. Implementation Status

### 3.1 Data Models (`src/Common/models/sz84_models.go`)
- [x] `UserInfo`, `GroupInfo`: Core user and group data.
- [x] `SensitiveWord`: Sensitive word definitions with Action (Recall/Mute/Kick) and Duration.
- [x] `SystemSetting`: Global key-value settings.
- [x] `BlackList`, `WhiteList`: Security lists.

### 3.2 Service Layer (`src/BotWorker/internal/services/`)
- [x] `SettingService`: Manages global settings and group info.
- [x] `SensitiveWordService`: Caches regex patterns and performs text matching.
- [x] `GroupService`: Handles group and member data retrieval.
- [x] `EconomyService`: (Integrated into EconomyPlugin) Points management.

### 3.3 Middleware (`src/BotWorker/internal/server/`)
- [x] `SensitiveWordMiddleware`:
    - Checks message content against cached patterns.
    - Executes actions: Recall, Mute (with duration), Kick.
    - Uses `ActionCaller` to communicate with OneBot.
- [x] `FriendlyMessageMiddleware`:
    - Filters non-whitelisted URLs in group messages.
    - **VIP Bypass**: Checks `GroupInfo.IsWhite` to allow trusted groups to send any URL.
    - Formats messages (escape character handling).

### 3.4 Plugins (`src/BotWorker/plugins/`)
- [x] `EntertainmentPlugin`: Dice rolling, jokes.
- [x] `EconomyPlugin`:
    - Check Points (`积分`, `查询积分`)
    - Daily Sign-in (`签到`, `打卡`)
    - Points Transfer (`转账`, `打赏`)

### 3.5 Infrastructure
- [x] `CombinedServer`: Implements `ActionCaller` (SendMessage, DeleteMessage, SetGroupKick, SetGroupBan).
- [x] `plugin_integration.go`: Loads and initializes plugins.

## 4. Remaining Tasks & Verification

### 4.1 Verification
- [ ] **Runtime Testing**:
    - Test Sensitive Word action (send a word with Action=2 and verify Mute).
    - Test URL filtering in normal group vs. VIP group.
    - Test Economy plugin commands.
- [ ] **Data Migration**: Ensure existing data in SQL Server/PostgreSQL is compatible with GORM models.

### 4.2 Future Enhancements
- [ ] **Web UI**: Create a management interface for System Settings and Sensitive Words.
- [ ] **Hot Reload**: Allow reloading sensitive words/settings without restarting the bot (partially supported via Service caching).

## 5. Execution Plan
1.  **Deploy**: Build and deploy the latest `BotWorker` to the test environment.
2.  **Config**: Populate `SystemSetting` table with initial switches (`Func_SensitiveWord`, `Func_FriendlyMessage` = "true").
3.  **Test**: Perform functional testing on all migrated features.
4.  **Monitor**: Watch logs for any execution errors in Middleware.
