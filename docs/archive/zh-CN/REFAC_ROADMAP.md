# BotMatrix 重构详尽路线图 (REFAC_ROADMAP.md)

## 1. 功能全清单扫描

### 1.1 核心过滤与拦截 (Middleware 候选)
- **Blacklist/Greylist**: `BlackMessage.cs`, `GreyListMessage.cs` - 全局/群黑灰名单。
- **WhiteList**: `WhiteMessage.cs` - 白名单放行。
- **Anti-Spam (Refresh)**: `RefreshMessage.cs` - 频率限制与扣分。
- **Sensitive Words**: `WarnMessage.cs` - 关键词告警与处理。
- **Power Control**: `SetupMessage.cs` - 机器人开关机、功能开启/关闭。
- **VIP/Credit Check**: `VipMessage.cs`, `CreditMessage.cs` - 权限与积分门槛。
- **Content Pre-processing**: `BotMessage.cs` (AsJianti, RemoveQqAds) - 文本清洗。

### 1.2 业务逻辑插件 (Plugin 候选)
- **Group Admin**: `KickMessage.cs`, `MuteMessage.cs`, `SetTitleMessage.cs`, `ChangeNameMessage.cs` - 踢、禁、改。
- **User System**: `SigninMessage.cs`, `CoinsMessage.cs`, `GreetingMessage.cs` - 签到、积分、欢迎语。
- **AI & Agents**: `AgentMessage.cs`, `AgentStreamMessage.cs` - 智能体交互。
- **Games & Tools**: `JielongMessage.cs`, `MusicMessage.cs`, `GroupGameMessage.cs`, `RedBlueMessage.cs` - 成语接龙、点歌、小游戏。
- **Utilities**: `ToolsMessage.cs`, `AnswerMessage.cs` - 各种工具指令。

### 1.3 核心驱动逻辑 (Core Driver - 保持在 BotMessage)
- **Context Initialization**: `BotMessage.cs` - 基础数据加载。
- **Message Sending**: `SendMessage.cs` - 消息下发。
- **Platform Adapters**: `GuildMessage.cs`, `PublicMessage.cs` - 平台差异化适配。

---

## 2. 重构阶段划分

### 阶段 1: 基础设施建设
- [ ] 创建 `Core/Pipeline/IMiddleware.cs` 接口。
- [ ] 创建 `Core/Pipeline/MessagePipeline.cs` 管道执行器。
- [ ] 在 `Program.cs` 注册管道服务。

### 阶段 2: 拦截逻辑迁移 (Middlewares)
- [ ] **LoggingMiddleware**: 迁移 `BotLog.Log` 调用。
- [ ] **BlacklistMiddleware**: 迁移黑名单拦截逻辑。
- [ ] **PowerMiddleware**: 迁移开关机检测。
- [ ] **CleanerMiddleware**: 迁移繁简转换、去广告。
- [ ] **SensitiveWordMiddleware**: 迁移关键词过滤。

### 阶段 3: 业务逻辑迁移 (Plugins)
- [ ] **InternalPluginSystem**: 确保现有的 `PluginManager` 作为一个中间件完美运行。
- [ ] **GroupAdminPlugin**: 迁移踢人、禁言逻辑。
- [ ] **SignInPlugin**: 迁移签到逻辑。
- [ ] **PointSystemPlugin**: 迁移积分计算逻辑。

### 阶段 4: 最终瘦身与验证
- [ ] 清理 `BotMessage` 中冗余的 `if-else` 分支。
- [ ] 确保所有功能点都有对应的测试或验证。

---

## 3. 安全操作规范
1. **每次变动前**：执行 `git commit`。
2. **每次迁移后**：执行冒烟测试，确保原功能未损坏。
3. **回退策略**：若出现不可预知的 Bug，立即 `git checkout`。
