# BotNexus Task System Documentation
[中文](../development/DOCS_PLAN.md) | [Back to Home](../../README.md) | [Back to Docs](../README.md)

## 1. Task System Architecture

The task system consists of the following core components:
- **Task (Task Definition)**: Stores task metadata, trigger rules, and action parameters.
- **Execution (Execution Instance)**: Records each execution of a task, including state transitions and execution results.
- **Scheduler**: Periodically scans tasks to be executed, generates Executions, and distributes them.
- **Dispatcher**: Responsible for actual action execution and managing the Execution state machine.
- **Tagging (Tagging System)**: Supports tagged management for groups and friends, supporting multiple tag combinations.

## 2. Data Model

### Tasks Table (tasks)
| Field | Type | Description |
| --- | --- | --- |
| id | uint | Primary Key |
| name | string | Task Name |
| type | string | Task Type (once, cron, delayed, condition) |
| action_type | string | Action Type (send_message, mute_group, unmute_group) |
| action_params | text (JSON) | Action Parameters |
| trigger_config | text (JSON) | Trigger Configuration |
| status | string | Status (pending, disabled, completed) |
| is_enterprise | bool | Whether it's an Enterprise Edition feature |
| last_run_time | datetime | Last execution time |
| next_run_time | datetime | Next estimated execution time |

### Execution Records Table (executions)
| Field | Type | Description |
| --- | --- | --- |
| id | uint | Primary Key |
| task_id | uint | Associated Task ID |
| execution_id | string | Unique Execution ID (UUID) |
| trigger_time | datetime | Theoretical trigger time |
| actual_time | datetime | Actual execution time |
| status | string | Status (pending, dispatching, running, success, failed, dead) |
| result | text (JSON) | Execution result or error message |
| retry_count | int | Number of retries performed |

## 3. JSON Schema Examples

### Scheduled Message Task (Cron)
```json
{
  "name": "Daily Morning Report",
  "type": "cron",
  "action_type": "send_message",
  "action_params": {
    "bot_id": "123456",
    "group_id": "7890",
    "message": "Good morning everyone!"
  },
  "trigger_config": {
    "cron": "0 8 * * *"
  }
}
```

### Auto-Mute Task (Condition)
```json
{
  "name": "Keyword Muting",
  "type": "condition",
  "action_type": "mute_group",
  "action_params": {
    "bot_id": "123456",
    "group_id": "7890",
    "duration": 600
  },
  "trigger_config": {
    "event": "message",
    "keyword": "advertisement"
  }
}
```

## 4. AI Generation Rule User Guide

Users can generate task rules through natural language input:
1. **Input Example**: "Help me set a task to mute the whole group every night at 11 PM"
2. **AI Parsing**: The system automatically identifies:
   - Task Name: Nighttime Auto-Mute
   - Type: cron (0 23 * * *)
   - Action: mute_group
3. **Confirmation & Execution**: The system returns the parsed JSON for user confirmation. After confirmation, the system automatically creates the Task and enters scheduling.

## 5. Version Differences (Trial vs Enterprise)

| Feature | Trial Edition | Enterprise Edition |
| --- | --- | --- |
| Task Types | once, cron | All types (including condition, delayed) |
| Tag Support | Single Tag | Multi-tag combination (AND/OR) |
| Impact Scope | Single Group/Single Friend | Batch Execution |
| Simulation Execution | Not Supported | Simulation Execution Report Supported |
| SLA Guarantee | Basic Priority | High Priority & Retry Policy |

## 6. Advanced Control Features

### Global Strategy & Interceptors
Empowers the scheduling center with "veto power" and "global control":
- **Execution Timing**: Triggered before message distribution (Dispatch).
- **Core Capabilities**:
  - **Maintenance Mode**: Configured through the `Strategy` table, one-click global silence, restricted to administrator commands.
  - **Rate Limiting**: Limit message frequency for a user or group at the Nexus level to prevent Worker overload.
  - **Security Audit**: Scan all messages for sensitive words and URL safety before distribution.

### Unified Identity System (Cross-Platform)
Achieves "one person, one ID", breaking platform isolation:
- **NexusUID**: Map IDs of the same user on QQ, WeChat, and Telegram to a unique unified ID.
- **Attribute Inheritance**: Seamless connection of user points, preferences, and other metadata across platforms.
- **Data Model**: Use the `UserIdentity` table to record mappings between multi-platform IDs and `NexusUID`.

### Intelligent Semantic Routing
Distribute tasks based on "intent" rather than "rules":
- **Intent Recognition**: Nexus automatically identifies if a message is a "question", "chat", or "command".
- **Dynamic Load**: Distribute tasks to the most suitable Worker based on intent (e.g., Knowledge Base Worker or GPT-4 Worker).
- **Timeout Downgrade**: AI parsing has a timeout mechanism (default 2s), falling back to normal forwarding paths after timeout.

### Shadow Mode & A/B Testing
Validate new rules at low cost:
- **Parallel Execution**: Send the same message to both the production Worker and the shadow Worker simultaneously.
- **Shadow Tagging**: Shadow messages carry a special `echo` identifier (format: `shadow_{timestamp}_{rand}`); the Worker records the difference upon receipt but produces no external side effects.
- **Performance Lossless**: Shadow execution occurs in a separate goroutine, not blocking real-time forwarding of production messages.

## 8. Multi-Version Parallelism & Canary Release

### Core Value
Support simultaneous operation of new and old systems, achieving seamless migration and A/B testing.

### Implementation Mechanism
- **Environment Isolation**: Workers carry `env` (prod/dev/test) tags during registration.
- **Version Routing**: The scheduling center supports precise routing based on `version`.
- **Canary Distribution**: Support routing traffic to the new version Worker by user, group, or percentage.

### Heterogeneous Worker Integration Guide
Nexus uses language-agnostic WebSocket + JSON protocol, supporting integration in any language.

#### Go Worker Integration Example
```go
// 1. Connect to Nexus
conn, _, _ := websocket.DefaultDialer.Dial("ws://nexus-address/worker", nil)

// 2. Report capabilities (with version and environment)
reg := map[string]interface{}{
    "type": "update_metadata",
    "metadata": {
      "plugins": [ ... ]
    }
    // Note: legacy type: "register_capabilities" is Deprecated
    "capabilities": []map[string]interface{}{
        {
            "name": "translate",
            "version": "2.0-go",
            "env": "prod",
            "description": "High-performance Go-based translation engine",
        },
    },
}
conn.WriteJSON(reg)

// 3. Handle commands
for {
    var cmd map[string]interface{}
    conn.ReadJSON(&cmd)
    if cmd["type"] == "skill_call" {
        // Execute business logic...
    }
}
```

### Migration Strategy (Strangler Fig Pattern)
1. **Coexistence Phase**: Old Workers handle existing features; new Go Workers integrate and report new features.
2. **Shadow Phase**: Enable `Shadow Mode`, allowing new Go Workers to process traffic in parallel; Nexus compares results but does not dispatch.
3. **Switch Phase**: Switch production traffic routing from old Workers to new Go Workers; old Workers remain as hot backups.
4. **Cleanup Phase**: After stable verification, decommission the old version Workers.

## 7. AI Semantic Understanding & Distributed Skills

### Capability Manifest
To let AI models understand system capabilities, BotNexus provides a dynamically generated manifest:
- **Core Actions**: Commands natively supported by the scheduling center (e.g., send message, group management).
- **Trigger Mechanisms**: Supported time and event trigger methods.
- **Global Rules**: System-level constraint descriptions.
- **Distributed Skills**: Business features reported by various business Workers.
- **Generation Logic**: Dynamically generate `System Prompt` when `AIParser` starts and when Worker capabilities are updated.

### Worker Skill Reporting Mechanism
Business Workers can report their capabilities to the scheduling center after connecting:
- **Reporting Interface**: Send `type: "update_metadata"` message (Legacy `register_capabilities` is Deprecated).
- **Protocol Structure**:
  ```json
  {
    "type": "update_metadata",
    "metadata": { ... } // Legacy register_capabilities is Deprecated
    "capabilities": [
      {
        "name": "checkin",
        "description": "Daily check-in to get points",
        "usage": "I want to check in",
        "params": {"user_id": "User ID"}
      }
    ]
  }
  ```
- **Dynamic Summarization**: The scheduling center automatically summarizes all online Worker capabilities and updates the AI prompt.
- **Semantic Routing**: After AI parses the user intent, if a specific skill is matched, the scheduling center distributes structured commands to Workers possessing that capability.

### Interaction Example
1. **User**: "Help me check the weather in Shanghai."
2. **AI Parsing**: Matches `skill_call` intent, target skill `weather`, parameter `city: "Shanghai"`.
3. **Scheduling Center**: Finds online Worker instances that have reported the `weather` skill.
4. **Dispatch Command**: Sends to Worker:
   ```json
   {
     "type": "skill_call",
     "skill": "weather",
     "params": {"city": "Shanghai"},
     "user_id": "12345"
   }
   ```
5. **Execution Result**: The Worker returns weather information via Passive Reply after execution.
