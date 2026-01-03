# BotMatrix Skill 系统架构与规划指南

> 版本: v1.0 (2026-01-01)  
> 状态: 核心规范

## 1. 设计初衷 (Motivation)

在传统的机器人开发中，插件往往是孤立的。例如，“天气插件”无法直接调用“积分插件”来扣除费用。  
**Skill 系统** 的目标是打破这种孤岛，将每个插件的能力原子化、接口化，从而实现：
- **跨插件复用**: 一个功能（如支付、权限检查）只需开发一次，全平台通用。
- **跨语言协同**: C# 编写的积分系统可以无缝被 Go 编写的游戏插件调用。
- **财务合规性**: 通过统一的 Skill 入口强制执行复式记账和审计日志。

---

## 2. 核心概念 (Core Concepts)

### 2.1 Skill Export (能力导出)
插件通过 `ExportSkill<TIn, TOut>` 声明自己具备的能力。这相当于定义了一个 RPC 接口。
- **TIn**: 输入参数模型（强类型）。
- **TOut**: 输出结果模型（强类型）。

### 2.2 Skill Call (能力调用)
调用方通过 `CallSkillAsync<TOut>` 发起请求。请求由 Core 进行路由，并通过 `CorrelationId` 实现异步转同步。

---

## 3. 协议层定义 (Protocol)

Skill 的交互基于 JSON-RPC 风格的异步消息：

1.  **Call (请求)**:
    ```json
    {
      "type": "call_skill",
      "payload": {
        "plugin_id": "com.example.bank",
        "skill": "transfer",
        "correlation_id": "unique-id-123",
        "params": { ... }
      }
    }
    ```
2.  **Result (响应)**:
    ```json
    {
      "type": "skill_result",
      "correlation_id": "unique-id-123",
      "ok": true,
      "data": { ... }
    }
    ```

---

## 4. 已实现的 Skill 规范示例 (以积分服务为例)

为了确保全系统的一致性，所有基础服务必须遵循预定义的接口契约。

### Economy Service (`com.botmatrix.official.bank`)

| Skill 名称 | 功能描述 | 输入参数 | 返回值 |
| :--- | :--- | :--- | :--- |
| `get_balance` | 查询用户余额 | `{}` | `BalanceResponse` (包含 Global/Group/BotLocal) |
| `transfer` | 财务转账 (有进有出) | `TransferRequest` (包含 From/To/Amount/Type) | `bool` (是否成功) |

---

## 5. 开发最佳实践 (Best Practices)

1.  **财务合规**: 任何涉及资产变动的 Skill 必须在内部调用 `PerformTransfer` 逻辑，确保审计日志的生成。
2.  **精简回复**: Skill 的 Handler 内部通常不直接使用 `ctx.Reply()`，而是将结果返回给调用方，由调用方决定如何展示。
3.  **幂等性**: 调用方应负责 `correlation_id` 的生成，Skill 提供方未来将支持基于此 ID 的幂等校验。
4.  **超时处理**: 建议所有跨插件调用设置 10-30 秒的超时时间。

---

## 7. 混合路由机制 (Hybrid Routing)

为了支持复杂的插件生态，系统实现了混合路由机制：

1.  **进程间路由 (Process-to-Process)**:
    - 核心检测目标 `plugin_id` 是否为已运行的外部进程插件。
    - 如果是，则将 `call_skill` 封装为 `event` 发送到目标进程的 `stdin`。
2.  **内部路由 (Internal Execution)**:
    - 如果目标不是进程插件，核心会尝试在内部 `SkillRegistry` 中查找对应的 Skill。
    - 这支持了外部插件调用内部 Go 插件（如官方积分系统）的能力。
3.  **自动报备 (Capability Discovery)**:
    - `BotWorker` 会定期扫描所有内部插件和外部插件包装器，提取其 `GetSkills()` 定义，并同步至 `BotNexus` 中心，实现全局可见性。

