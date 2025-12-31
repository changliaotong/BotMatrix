# BotMatrix Skill System Compatibility & Feature Toggle

> [🌐 English](SKILL_SWITCH_AND_COMPATIBILITY.md) | [简体中文](../zh-CN/SKILL_SWITCH_AND_COMPATIBILITY.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

This document records the compatibility design, feature toggle mechanism, and isolation strategies during the testing phase of the BotMatrix Skill System.

## 1. Design Goals

To ensure complete compatibility with older BotWorker clients while introducing the Redis asynchronous task queue and the Skill System, the system adopts a strategy that combines "Feature Toggle Control" and "Dynamic Capability Discovery."

- **Default Off**: All skill-related features are turned off by default in production.
- **Graceful Downgrade**: Older Workers can still process basic messages normally via WebSocket.
- **Environment Isolation**: The skill system is only activated in testing environments where `ENABLE_SKILL=true`.

## 2. Feature Toggle Mechanism

### 2.1 Global Configuration
The toggle can be controlled in three ways:

- **Web UI Dashboard**: In the "Core Config" tab of the BotNexus Config Center, you can check/uncheck "Enable Skill System" and save to restart.
- **Configuration File** (`config.json`):
  ```json
  {
    "enable_skill": false
  }
  ```
- **Environment Variables**: `ENABLE_SKILL=true` or `ENABLE_SKILL=1` can force it on.

### 2.2 BotNexus (Server) Behavior
When `ENABLE_SKILL` is `false`:
1. **Components Not Initialized**: Does not start GORM database connections, the `TaskManager` scheduler, or Redis subscription listeners.
2. **Result Reporting Interception**: Even if a `skill_result` is reported by a Worker, it will be discarded by `handleWorkerMessage` and logged.
3. **Routing Downgrade**: The system only performs traditional OneBot message forwarding logic.

### 2.3 BotWorker (Client) Behavior
When `EnableSkill` is `false`:
1. **Hide Capability Reporting**: Does not send a `capabilities` list to the `botmatrix:worker:register` channel upon startup. Nexus treats it as a basic forwarding node.
2. **Refuse Instruction Execution**: In the Redis queue listener, if a `skill_call` type message is received, it will be ignored and won't enter the execution flow.

## 3. Compatibility Routing Strategy

To support mixed-version environments (new and old Workers online simultaneously), BotNexus implements the following intelligent routing logic:

### 3.1 Skill-Aware Distribution
- **Targeted Delivery**: Before distributing a skill task, the `Dispatcher` uses `FindWorkerBySkill` to retrieve Workers that have explicitly reported that skill.
- **Load Balancing**: Performs random distribution among the set of Workers that support the skill.
- **Isolate Legacy**: Older Workers that haven't reported skill capabilities will not be included in the candidate list, avoiding unparseable `skill_call` messages.

### 3.2 Result Return Compatibility
- **Dual Channel Support**: Supports returning skill results via Redis Pub/Sub or WebSocket.
- **ID Fallback Mechanism**:
  - Priority is given to using `execution_id` to precisely match task execution records.
  - For legacy reporting that doesn't support `execution_id`, it falls back to updating the task's most recent execution status based on `task_id`.

## 4. Testing & Deployment Recommendations

1. **Testing Phase**:
   - Deploy independent test Nexus and Worker instances.
   - Set `"enable_skill": true` in the config file.
   - Validate the closed-loop flow from `skill_call` to `skill_result`.

2. **Gray Launch**:
   - Upgrade some Workers first and enable the skill toggle.
   - Enable the toggle on Nexus and observe if tasks are accurately routed to the new Workers.

3. **Official Launch**:
   - Full update of all Workers and enable the toggle uniformly in the config files.

---
*Last Updated: 2025-12-24*
