# 数据库双轨运行与 sz84.com 功能迁移详细计划

## 1. 数据库演进路线 (Database Evolution)

### 1.1 双轨并行阶段 (当前)
- **主数据库 (PostgreSQL)**: 存储所有新功能数据（如 AI 意图编排、MCP 技能配置、新版用户体系）。
- **旧数据库 (SQL Server)**: 维持现有 BotWorker 运行所需的 legacy 数据（如旧版积分、历史订单、原有机器人设置）。
- **连接方式**: BotNexus (Go) 通过 `GORMManager` 同时持有 `DB` (PG) 和 `LegacyDB` (MSSQL) 两个连接池。

### 1.2 迁移策略
- **读写分离**: 
    - **写操作**: 优先写入 PG。如果涉及旧业务，通过消息队列同步更新 MSSQL。
    - **读操作**: 优先查 PG，若无数据则回退查 MSSQL（Lazy Migration 模式）。
- **最终目标**: 当所有 BotWorker 逻辑重构完成，废弃 MSSQL，统一使用 PostgreSQL。

## 2. sz84.com 功能迁移范围

针对原有网站后台 (sz84.com)，我们不进行搬家式迁移，而是“取其精华”：

### 2.1 必须迁移的核心功能 (Go 重写)
1.  **全局配置管理**:
    - 系统回复话术 (如：`RetryMsgTooFast`, `OwnerOnlyMsg`)。
    - 官方机器人白名单 ([AppConfig.cs](file:///c:/Users/彭光辉/projects/BotMatrix/src/BotWorker/Common/Common.cs#L20) 中的 `OfficalBots`)。
    - 功能开关 (如：积分系统开关、AI 模式切换)。
2.  **用户与资产**:
    - 用户基础资料、余额、积分 (Credits)。
    - 会员等级 (Vip/YearOnly)。
3.  **机器人实例管理**:
    - 机器人的 Token、运行参数、分配的平台。

### 2.2 暂时保留或延后的功能
- 极其复杂的历史订单报表（可继续在旧后台查看）。
- 尚未确定的边缘业务插件配置。

## 3. 技术细节讨论

### 3.1 跨数据库的一致性
- **问题**: 如何保证 PG 和 MSSQL 的数据同步？
- **方案**: 在 BotNexus 中封装 `DataFacade` 层，当修改关键数据（如用户积分）时，同时触发两边更新，或者使用 Redis 作为中间缓存。

### 3.2 sz84 API Key 的处理
- **现状**: BotWorker 通过 `apiKey` 访问 sz84.com 接口。
- **重构**: 
    1. 在 BotNexus 中模拟 sz84.com 的部分接口。
    2. 修改 BotWorker 的 `url` 指向 BotNexus。
    3. 这样无需大规模改动 BotWorker 代码，即可实现平滑迁移。

### 3.3 数据表映射计划 (示例)

| 旧表 (MSSQL) | 新模型 (Go GORM) | 说明 |
| :--- | :--- | :--- |
| `UserInfo` | `models.User` | 增加 `legacy_id` 字段关联旧数据 |
| `BotInfo` | `models.Bot` | 重新设计，支持多平台扩展 |
| `SystemSetting` | `models.Config` | 采用 Key-Value 格式存储，方便动态扩展 |

---

**请确认上述计划。如果有特定的“重要功能”需要优先考虑，请告知。**
