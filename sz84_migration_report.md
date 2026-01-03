# sz84 功能复刻报告 (sz84 Migration Report)

本文档列明了从 sz84 机器人系统迁移到 BotMatrix 插件系统的功能对应关系。

## 核心系统 (Core Systems)

| 原功能 (sz84 Feature) | 复刻插件 (BotMatrix Plugin) | 说明 (Notes) |
| :--- | :--- | :--- |
| VIP 系统 (换群/换主人/续费/查询) | `SuperGroupManager` | 集成在 `SuperGroupPlugin` 中，支持私聊查询和群内管理。 |
| 积分系统 (积分排行榜/转账/手续费) | `PointsSystem` | 实现了 10% 转账手续费、20% 奖励手续费以及超级用户豁免逻辑。 |
| 全球积分 (群积分转全球积分) | `SuperGroupManager` | 10000 群积分 = 1-4 全球积分 (根据超级用户身份)。 |
| 签到系统 (连续签到/等级/奖励) | `SignInSystem` | 复刻了 1-10 级签到奖励、超级用户双倍奖励及成就触发。 |

## 群管理功能 (Group Management)

| 原功能 (sz84 Feature) | 复刻插件 (BotMatrix Plugin) | 说明 (Notes) |
| :--- | :--- | :--- |
| 入群申请处理 (VIP群/密码/正则) | `SuperGroupManager` | 在 `HandleRequest` 中实现 VIP 专属群校验及正则表达式匹配。 |
| 进群欢迎/退群提示 (模板化) | `SuperGroupManager` | 支持 `WelcomeTemplate` 和 `ExitTemplate`，支持自定义开关。 |
| 邀请奖励 | `SuperGroupManager` | 邀请新成员奖励积分，当奖励 > 50 时从机器人主人扣除差额。 |
| 进群自动改名/禁言 | `SuperGroupManager` | 实现了 `IsChangeEnter` 和 `IsMuteEnter` 配置逻辑。 |
| 敏感词阶梯处罚 | `SuperGroupManager` | 撤回、扣分、警告、禁言、踢出、拉黑六级处罚体系。 |
| 自动签到 (进群发言自动签到) | `SuperGroupManager` | 通过中间件调用 `SignInSystem` 的 `handle_signin` Skill 实现。 |

## 工具与娱乐 (Tools & Entertainment)

| 原功能 (sz84 Feature) | 复刻插件 (BotMatrix Plugin) | 说明 (Notes) |
| :--- | :--- | :--- |
| 音乐功能 (点歌/送歌/曲库/分享监听) | `UtilityTools` | 集成了酷我音乐 API，支持 Redis 曲库存储及音乐卡片自动入库。 |
| 天气预报 | `UtilityTools` | 对接高德地图 API。 |
| 文本翻译 | `UtilityTools` | 对接 Azure 翻译 API。 |
| 笑话/故事/鬼故事/对联 (SQL库) | `ContentTools` | 对接原 sz84 `Answer` 数据库表，实现了随机抽取及特定扣分逻辑。 |
| 抽签/解签 | `ContentTools` | 实现了基础逻辑，待进一步完善 SQL 对接。 |
| 身份证查询/简繁转换 | `ContentTools` | 基础工具类复刻。 |

## 业务逻辑一致性 (Business Logic Consistency)

- **手续费**: 严格保持 10% 转账手续费和 20% 奖励手续费。
- **豁免权**: 机器人主人、VIP、超级用户在特定操作中享受手续费豁免。
- **数据存储**: 
  - 状态数据 (VIP, 积分, 签到) 迁移至 Redis。
  - 内容数据 (笑话, 故事) 对接原有 SQL Server。
- **前缀一致**: Redis Key 使用 `sz84:` 前缀以保持兼容性。
