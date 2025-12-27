# BotNexus 任务系统文档
[English](../../en-US/development/DOCS_PLAN.md) | [返回项目主页](../../../README.md) | [返回文档中心](../README.md)

## 1. 任务系统架构

任务系统由以下核心组件组成：
- **Task (任务定义)**: 存储任务的元数据、触发规则和动作参数。
- **Execution (执行实例)**: 记录每一次任务的执行情况，包含状态流转和执行结果。
- **Scheduler (调度器)**: 定期扫描待执行任务，生成 Execution 并分发。
- **Dispatcher (分发器)**: 负责实际执行动作，管理 Execution 的状态机。
- **Tagging (标签系统)**: 支持对群组和好友进行标签化管理，支持多标签组合。

## 2. 数据模型

### 任务表 (tasks)
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uint | 主键 |
| name | string | 任务名称 |
| type | string | 任务类型 (once, cron, delayed, condition) |
| action_type | string | 动作类型 (send_message, mute_group, unmute_group) |
| action_params | text (JSON) | 动作参数 |
| trigger_config | text (JSON) | 触发配置 |
| status | string | 状态 (pending, disabled, completed) |
| is_enterprise | bool | 是否为企业版功能 |
| last_run_time | datetime | 最后执行时间 |
| next_run_time | datetime | 下次预计执行时间 |

### 执行记录表 (executions)
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uint | 主键 |
| task_id | uint | 关联任务ID |
| execution_id | string | 唯一执行ID (UUID) |
| trigger_time | datetime | 理论触发时间 |
| actual_time | datetime | 实际执行时间 |
| status | string | 状态 (pending, dispatching, running, success, failed, dead) |
| result | text (JSON) | 执行结果或错误信息 |
| retry_count | int | 已重试次数 |

## 3. JSON Schema 示例

### 定时消息任务 (Cron)
```json
{
  "name": "每日早报",
  "type": "cron",
  "action_type": "send_message",
  "action_params": {
    "bot_id": "123456",
    "group_id": "7890",
    "message": "大家早上好！"
  },
  "trigger_config": {
    "cron": "0 8 * * *"
  }
}
```

### 自动禁言任务 (Condition)
```json
{
  "name": "关键词禁言",
  "type": "condition",
  "action_type": "mute_group",
  "action_params": {
    "bot_id": "123456",
    "group_id": "7890",
    "duration": 600
  },
  "trigger_config": {
    "event": "message",
    "keyword": "广告"
  }
}
```

## 4. AI 生成规则使用指南

用户可以通过自然语言输入来生成任务规则：
1. **输入示例**: "帮我设置一个每天晚上11点全群禁言的任务"
2. **AI 解析**: 系统会自动识别出：
   - 任务名称: 夜间自动禁言
   - 类型: cron (0 23 * * *)
   - 动作: mute_group
3. **确认与执行**: 系统会返回解析出的 JSON 供用户确认。用户确认后，系统自动创建 Task 并进入调度。

## 5. 版本差异 (Trial vs Enterprise)

| 功能 | 试用版 | 企业版 |
| --- | --- | --- |
| 任务类型 | once, cron | 所有类型 (含 condition, delayed) |
| 标签支持 | 单标签 | 多标签组合 (AND/OR) |
| 影响范围 | 单群/单好友 | 批量执行 |
| 模拟执行 | 不支持 | 支持模拟执行报告 |
| SLA 保证 | 基础优先级 | 高优先级 & 重试策略 |

## 6. 高级管控功能

### 全局策略与拦截器 (Global Strategy & Interceptors)
赋予调度中心“一票否决权”和“全局管控力”：
- **执行时机**: 在消息分发 (Dispatch) 前置触发。
- **核心能力**:
  - **维护模式**: 通过 `Strategy` 表配置，一键进入全局静默，仅限管理员指令。
  - **频率控制 (Rate Limiting)**: 在 Nexus 层级限制某个用户或群组的消息频率，保护 Worker 不被击穿。
  - **安全审计**: 分发前对所有消息进行敏感词、URL 安全扫描。

### 跨平台身份统一映射 (Unified Identity System)
实现“一人一号”，打破平台隔离：
- **NexusUID**: 将同一个用户在 QQ、微信、Telegram 的 ID 映射到唯一的统一 ID。
- **属性继承**: 用户跨平台的积分、偏好等元数据无缝衔接。
- **数据模型**: 使用 `UserIdentity` 表记录多平台 ID 与 `NexusUID` 的映射关系。

### 智能语义路由 (Intelligent Semantic Routing)
根据“意图”而非“规则”分发任务：
- **意图识别**: Nexus 自动识别消息是“提问”、“闲聊”还是“指令”。
- **动态负载**: 根据意图将任务分发给最合适的 Worker (如知识库 Worker 或 GPT-4 Worker)。
- **超时降级**: AI 解析设有超时机制（默认 2s），超时后回退至普通转发路径。

### 影子执行与 A/B 测试 (Shadow Mode)
低成本验证新规则：
- **平行执行**: 同一条消息同时发送给正式 Worker 和影子 Worker。
- **影子标记**: 影子消息带上特殊的 `echo` 标识（格式：`shadow_{timestamp}_{rand}`），Worker 收到后记录差异但不产生外部副作用。
- **性能无损**: 影子执行在独立协程中进行，不阻塞正式消息的实时转发。

## 8. 多版本并行与灰度发布

### 核心价值
支持新旧系统并行运行，实现无感迁移与 A/B 测试。

### 实现机制
- **环境隔离**: Worker 报备时携带 `env` (prod/dev/test) 标识。
- **版本路由**: 调度中心支持根据 `version` 进行精准路由。
- **灰度分流**: 支持按用户、按群组或按比例将流量导向新版 Worker。

### 异构 Worker 接入指南
Nexus 采用语言无关的 WebSocket + JSON 协议，支持任何语言接入。

#### Go Worker 接入示例
```go
// 1. 连接 Nexus
conn, _, _ := websocket.DefaultDialer.Dial("ws://nexus-address/worker", nil)

// 2. 报备能力 (携带版本和环境)
reg := map[string]interface{}{
    "type": "register_capabilities",
    "capabilities": []map[string]interface{}{
        {
            "name": "translate",
            "version": "2.0-go",
            "env": "prod",
            "description": "高性能 Go 版翻译引擎",
        },
    },
}
conn.WriteJSON(reg)

// 3. 处理指令
for {
    var cmd map[string]interface{}
    conn.ReadJSON(&cmd)
    if cmd["type"] == "skill_call" {
        // 执行业务逻辑...
    }
}
```

### 迁移策略 (絞杀者模式)
1. **共存阶段**: 旧 Worker 处理存量功能，新 Go Worker 接入并报备新功能。
2. **影子阶段**: 开启 `Shadow Mode`，让新 Go Worker 并行处理流量，Nexus 对比结果但不下发。
3. **切换阶段**: 将生产流量路由从旧 Worker 切换至新 Go Worker，旧 Worker 保持热备。
4. **清理阶段**: 验证稳定后，下线旧版 Worker。

## 7. AI 语义理解与分布式技能

### 系统功能清单 (Capability Manifest)
为了让 AI 大模型理解系统能力，BotNexus 提供动态生成的清单：
- **核心动作 (Actions)**: 调度中心原生支持的指令（如：发消息、群管理）。
- **触发机制 (Triggers)**: 支持的时间和事件触发方式。
- **全局规则 (Rules)**: 系统级的约束说明。
- **分布式技能 (Skills)**: 由各业务 Worker 报备的业务功能。
- **生成逻辑**: `AIParser` 启动时及 Worker 能力更新时，动态生成 `System Prompt`。

### Worker 技能报备机制
业务 Worker 在连接后可向调度中心报备其具备的能力：
- **报备接口**: 发送 `type: "register_capabilities"` 消息。
- **协议结构**:
  ```json
  {
    "type": "register_capabilities",
    "capabilities": [
      {
        "name": "checkin",
        "description": "每日签到获取积分",
        "usage": "我要签到",
        "params": {"user_id": "用户ID"}
      }
    ]
  }
  ```
- **动态汇总**: 调度中心自动汇总所有在线 Worker 的能力，并更新 AI 提示词。
- **语义路由**: AI 解析用户意图后，如果匹配到特定技能，调度中心会将结构化指令分发给具备该能力的 Worker。

### 交互示例
1. **用户**: "帮我查一下上海的天气。"
2. **AI 解析**: 匹配到 `skill_call` 意图，目标技能 `weather`，参数 `city: "上海"`。
3. **调度中心**: 查找在线 Worker 中报备过 `weather` 技能的实例。
4. **分发指令**: 向 Worker 发送：
   ```json
   {
     "type": "skill_call",
     "skill": "weather",
     "params": {"city": "上海"},
     "user_id": "12345"
   }
   ```
5. **执行结果**: Worker 执行后通过 Passive Reply 返回天气信息。
