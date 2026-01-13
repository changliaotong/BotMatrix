# BotNexus AI 智能中心使用指南
[English](../../en-US/development/AI_GUIDE.md) | [返回项目主页](../../../README.md) | [返回文档中心](../README.md)

AI 智能中心是 BotNexus 的核心特色功能，旨在通过自然语言处理（NLP）大幅降低复杂自动化配置的门槛。AI 已经深度集成到任务创建、策略管理、标签管理等多个模块。

## 1. 核心 AI 能力

### 1.1 智能任务生成 (AI Tasking)
你不再需要手动编写复杂的 JSON 任务定义，只需一句话即可生成。
- **示例输入**: "每天上午 10 点提醒大家写周报"
- **AI 动作**: 
    - 识别意图为 `create_task`
    - 提取触发规则: `cron: 0 10 * * *`
    - 提取动作: `send_message`
    - 生成结构化数据，用户一键确认即可部署。

### 1.2 自然语言策略调整 (AI Policy)
通过对话管理整个机器人矩阵的状态。
- **示例输入**: "现在系统要维护，帮我开启维护模式两小时"
- **AI 动作**:
    - 识别意图为 `adjust_policy`
    - 自动配置 `maintenance_mode` 策略
    - 设置过期时间为 2 小时后自动恢复。

### 1.3 批量标签管理 (AI Tagging)
快速对海量对象进行分类。
- **示例输入**: "把昨天最活跃的 5 个群标记为 '核心群'"
- **AI 动作**:
    - 识别意图为 `manage_tags`
    - 查询统计数据，筛选出符合条件的群组
    - 批量执行 `AddTag` 操作。

## 2. 开发者接入指南

### AI 解析接口 (Unified AI Endpoint)
`POST /api/ai/parse`

**请求参数**:
```json
{
  "input": "用户自然语言输入",
  "action_type": "可选，create_task/adjust_policy/manage_tags",
  "context": {
    "current_bot": "12345",
    "last_error": "..."
  }
}
```

**响应结果**:
```json
{
  "success": true,
  "data": {
    "draft_id": "uuid-string", // 用于确认执行的唯一标识
    "intent": "create_task",
    "summary": "创建自动化任务",
    "data": {
      "name": "AI 生成任务",
      "type": "cron",
      "action_type": "send_message",
      "action_params": "...",
      "trigger_config": "..."
    },
    "analysis": "AI 的推理建议...",
    "is_safe": true
  }
}
```

### AI 确认接口 (AI Confirm Endpoint)
`POST /api/ai/confirm`

**请求参数**:
```json
{
  "draft_id": "uuid-string"
}
```

**响应结果**:
```json
{
  "success": true,
  "message": "执行成功"
}
```

### 系统能力清单接口 (System Capabilities)
`GET /api/system/capabilities`

**核心作用**:
该接口返回 **BotNexus 功能清单 (Capability Manifest)**。它的主要作用是：
1. **喂给 AI**: 将返回的 `prompt` 字段内容作为 System Prompt 发送给大模型，让 AI 立即了解当前系统支持的所有动作（如 `send_message`, `mute_group`）和触发器（如 `cron`, `condition`）。
2. **动态更新**: 当系统增加新功能时，只需更新清单，无需修改 AI 的解析逻辑代码。
3. **前端展示**: 前端可以利用该清单动态生成功能列表或帮助界面。

## 3. 安全与人工确认流程
为了确保系统安全，AI 生成的所有指令**必须**经过以下流程：
1. **解析阶段**: AI 返回 `draft_id` 和建议的操作。
2. **人工审查**: 用户在前端 UI 审查 AI 的建议（Summary 和 Analysis）。
3. **确认阶段**: 用户点击确认，调用 `/api/ai/confirm` 接口，系统才会真正执行动作（如创建任务）。
4. **失效机制**: 所有的 `draft_id` 默认有效期为 15 分钟。

## 3. 企业版增强 AI 功能

- **意图预测**: 根据历史行为，主动建议用户创建某些任务（例如：检测到某个群广告变多，建议开启“关键词禁言”）。
- **模拟执行报告**: 在任务上线前，由 AI 生成一份详尽的“影响范围评估报告”。
- **故障自愈 AI**: 当检测到任务连续执行失败时，AI 自动分析原因并尝试修复配置。

---
*BotNexus - 让机器人管理回归自然语言。*
