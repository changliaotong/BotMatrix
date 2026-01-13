# 成就系统 (Achievement System) V1.0.0 技术文档

## 1. 系统概述
成就系统是一个全局性的基础设施模块，旨在通过追踪用户在各个插件中的行为指标（Metrics），自动解锁相应的成就勋章并给予奖励。系统采用解耦设计，业务插件仅负责上报数据，成就判定逻辑由核心系统统一管理。

## 2. 核心架构：指标驱动 (Metric-Driven)

### 2.1 运作流程
1. **上报 (Report)**: 业务插件通过 `AchievementPlugin.ReportMetricAsync` 提交用户行为数据。
2. **累积 (Accumulate)**: 系统在 `UserMetrics` 表中更新或增加该指标的数值。
3. **判定 (Judge)**: 系统扫描与该指标相关的成就定义，检查是否达到阈值。
4. **解锁 (Unlock)**: 若达到阈值且未曾解锁，则在 `UserAchievements` 表中记录解锁信息。

## 3. 数据模型

### 3.1 UserMetric (用户指标)
记录用户在特定维度上的累积数值。
- `Id`: `UserId_MetricKey` (唯一标识)
- `Value`: 当前数值
- `LastUpdateTime`: 最后更新时间

### 3.2 UserAchievement (已解锁成就)
记录用户获得的成就荣誉。
- `Id`: `UserId_AchievementId`
- `UnlockTime`: 解锁时间

### 3.3 AchievementDef (成就定义)
| 字段 | 说明 |
| :--- | :--- |
| Id | 内部唯一 ID |
| Name | 成就名称 (展示给用户) |
| MetricKey | 依赖的指标 Key |
| Threshold | 达成所需的数值阈值 |
| RewardGold | 达成后的金币奖励 |

## 4. 开发者指南：如何集成？

### 4.1 上报指标
在你的插件逻辑中，只需一行代码即可触发成就追踪：

```csharp
// 增加指标数值 (增量)
await AchievementPlugin.ReportMetricAsync(userId, "your_plugin.action_count", 1);

// 设置指标数值 (绝对值)
await AchievementPlugin.ReportMetricAsync(userId, "your_plugin.max_level", level, true);
```

### 4.2 定义新成就
在 `AchievementPlugin.cs` 的 `Definitions` 列表中添加配置：

```csharp
new AchievementDef { 
    Id = "new_act_01", 
    Name = "开拓者", 
    Description = "首次完成某项操作", 
    MetricKey = "your_plugin.action_count", 
    Threshold = 1, 
    RewardGold = 100 
}
```

## 5. 现有关联指标 (Pre-defined Metrics)

| 指标 Key | 来源模块 | 说明 |
| :--- | :--- | :--- |
| `sys.msg_count` | 统计中间件 | 累计发言次数 |
| `fishing.catch_count` | 钓鱼系统 | 累计成功钓鱼数 |
| `fishing.total_gold` | 钓鱼系统 | 卖鱼累计获得的金币 |
| `pet.adopt_count` | 宠物系统 | 领养宠物次数 |
| `pet.max_level` | 宠物系统 | 宠物的最高等级 |

## 6. 用户指令

- `我的成就`: 查看所有分类成就的解锁状态。对于未解锁成就，会显示当前指标的完成进度。

## 7. 数据库设计
系统自动维护以下两张表：
- `UserMetrics`: 存储用户各维度原始统计数据。
- `UserAchievements`: 存储用户成就勋章库。
