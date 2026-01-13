# 积分服务 (Economy Service) 开发规范

> 状态: 草案 (Draft)  
> 最后更新: 2026-01-01

## 1. 概述
Economy Service 是 BotMatrix 的基础服务插件，负责管理全系统的虚拟资产。它通过导出标准化的 Skill 接口，供其他功能插件（如游戏、商城、签到等）调用。

## 2. 核心原则：复式记账 (Double-Entry Bookkeeping)
积分系统不再允许“凭空”加减积分。每一笔积分变动必须遵循“有借必有贷，借贷必相等”的财务原则：
- **来源账户 (From)**: 积分扣除方。
- **目标账户 (To)**: 积分接收方。
- **系统账户 (System Accounts)**: 用于发放奖励或回收积分的特殊账户（如 `SYSTEM_REWARD`, `SYSTEM_FEE`）。

---

## 3. 导出 Skill 接口 (Exported Skills)

### 3.1 `get_balance` (查询余额)
获取用户在不同层级（全局、群聊、本机）的积分余额。

*   **输入参数**: `object` (空)
*   **输出模型 (`BalanceResponse`)**:
    ```json
    {
      "global": 1000,
      "group": 50,
      "bot_local": 20
    }
    ```

### 3.2 `transfer` (积分转账)
执行积分的转账操作。

*   **输入参数 (`TransferRequest`)**:
    ```json
    {
      "from_user_id": "string (可选, 默认为 SYSTEM_REWARD 或当前调用插件)",
      "to_user_id": "string (目标用户)",
      "amount": "long (必须为正数)",
      "type": "string (global | group | bot_local)",
      "reason": "string (变动原因)"
    }
    ```
*   **输出模型**: `bool` (操作是否成功)

---

## 4. 系统预设账户 (Internal Accounts)
- `SYSTEM_REWARD`: 官方奖励池（如签到、任务发放的积分来源）。
- `SYSTEM_FEE`: 手续费回收池（如交易、打赏产生的手续费）。
- `SYSTEM_SINK`: 违规扣除或系统销毁池。

## 3. 术语定义 (Terminology)

- **全局积分 (global)**: 跨平台、跨机器人的最高级积分，绑定用户唯一标识。
- **群积分 (group)**: 仅在特定群聊中生效的积分。
- **本机积分 (bot_local)**: 绑定特定“机器人号码”的积分。若用户在多个群聊中遇到同一个机器人账号，则该积分通用。这是为了支持未来客户适配自有号码而设计的。

## 5. 安全与审计 (Security & Audit)

### 5.1 审计日志 (Audit Logging)
系统会自动记录每一笔成功的 `transfer` 操作，日志包含以下信息：
- `Id`: 唯一流水号。
- `Timestamp`: 操作时间（UTC）。
- `From`: 来源账户。
- `To`: 目标账户。
- `Amount`: 变动金额。
- `Type`: 积分类型。
- `Reason`: 业务原因。
- `CorrelationId`: 关联的消息 ID，用于追溯引发变动的原始指令。

### 5.2 财务安全建议
- **系统账户保护**: `SYSTEM_` 开头的账户只能由官方插件或 Core 逻辑调用。
- **余额负债**: 只有系统账户允许出现负余额（代表系统赤字/总支出），普通用户账户必须先校验余额。
- **原子性**: 积分变动必须在同一个分布式事务或 Session 操作中完成，防止出现“只扣不加”或“只加不扣”的情况。

---

## 6. C# 实现示例

### 6.1 模型定义
```csharp
public class TransferRequest {
    [JsonPropertyName("from_user_id")]
    public string FromUserId { get; set; } = "SYSTEM_REWARD";
    
    [JsonPropertyName("to_user_id")]
    public string ToUserId { get; set; }
    
    public long Amount { get; set; }
    public string Type { get; set; } // global, group, bot_local
    public string Reason { get; set; }
}

public class BalanceResponse {
    public long Global { get; set; }
    public long Group { get; set; }
    [JsonPropertyName("bot_local")]
    public long BotLocal { get; set; }
}
```

### 6.2 财务核心逻辑 (PerformTransfer)
```csharp
private async Task<bool> PerformTransfer(Context ctx, TransferRequest req) {
    // 1. 获取 Key
    string fromKey = GetKey(req.Type, req.FromUserId);
    string toKey = GetKey(req.Type, req.ToUserId);

    // 2. 校验来源余额 (系统账户除外)
    long fromBalance = await ctx.Session.GetAsync<long>(fromKey);
    if (fromBalance < req.Amount && !req.FromUserId.StartsWith("SYSTEM_")) return false;

    // 3. 执行变动
    await ctx.Session.SetAsync(fromKey, fromBalance - req.Amount);
    await ctx.Session.SetAsync(toKey, (await ctx.Session.GetAsync<long>(toKey)) + req.Amount);

    // 4. 记录审计
    await LogAudit(ctx, req);
    return true;
}
```
