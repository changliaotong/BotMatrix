# BotMatrix Plugin System

## Overview
The BotMatrix Plugin System is a cross-platform, process-level plugin architecture that allows seamless integration of plugins into BotNexus and BotWorker components.

## Architecture

### 核心组件

1. **插件管理器 (Plugin Manager)** (`src/Common/plugin/core/manager.go`)
   - 管理插件生命周期（启动、停止、重启）
   - 处理插件发现和配置
   - 实现健康检查和崩溃恢复

2. **进程管理器 (Process Manager)** (`src/Common/plugin/core/process.go`)
   - 管理插件进程
   - 通过 stdin/stdout 实现进程间通信
   - 处理热更新和灰度发布

3. **Protocol** (`src/Common/plugin/core/protocol.go`)
   - Defines JSON message structures for plugin communication
   - Implements event and action protocols

4. **Policy** (`src/Common/plugin/policy/`)
   - Defines action whitelists for different plugin types
   - Ensures security boundaries between plugins and core system

## Plugin Types

### 1. Master Plugins (总控)
- Used for system management and control
- Run on BotNexus Core
- Have elevated permissions

### 2. Feature Plugins (功能)
- Used for specific business functionality
- Run on BotWorker
- Have restricted permissions based on action whitelists

## Plugin Structure

Each plugin must have:
- `plugin.json` - Configuration file
- Executable file (exe/elf) - Main plugin code

### plugin.json Specification

```json
{
  "id": "com.botmatrix.example",
  "name": "echo",
  "description": "Echo plugin - repeats messages",
  "author": "Developer",
  "version": "1.0.0",
  "entry_point": "echo.exe",
  "run_on": ["worker"],
  "permissions": ["send_msg", "call_skill"],
  "events": ["on_message"],
  "intents": [
    {
      "name": "hello",
      "keywords": ["hello", "hi"],
      "regex": "^hi.*"
    }
  ],
  "max_restarts": 5
}
```

## Security Policy

The system implements a robust security layer for plugin actions:

1. **Action Whitelisting**: Every action requested by a plugin (via `ResponseMessage.Actions`) is validated against its `permissions` defined in `plugin.json`.
2. **Wildcard Support**: Plugins can use `*` in their permissions to allow all actions (use with caution).
3. **Global Policy**: Actions are also checked against a global worker policy to ensure they are appropriate for the environment.
4. **Validation Flow**:
   - Check if action type is in `permissions`.
   - Check if action type is allowed by global policy.
   - Log and block unauthorized attempts.


## Communication Protocol

### Event Message (Core -> Plugin)
```json
{
  "id": "event-123",
  "type": "event",
  "name": "on_message",
  "payload": {
    "text": "Hello World",
    "from": "user123",
    "group_id": "group456"
  }
}
```

### Response Message (Plugin -> Core)
```json
{
  "id": "event-123",
  "ok": true,
  "actions": [
    {
      "type": "send_message",
      "target": "user123",
      "target_id": "group456",
      "text": "Echo: Hello World"
    }
  ]
}
```

### Skill Protocol (Cross-Plugin Communication)

The Skill mechanism allows a plugin to export its functionality and another plugin to call it.

#### 1. Skill Call (Plugin A -> Core)
A plugin requests to call a skill by sending a `call_skill` action.
```json
{
  "id": "event-456",
  "actions": [
    {
      "type": "call_skill",
      "payload": {
        "plugin_id": "com.example.bank",
        "skill": "get_balance",
        "correlation_id": "skill_get_balance_1712345678",
        "additional_param": "..."
      }
    }
  ]
}
```

#### 2. Skill Execution (Core -> Plugin B)
The Core routes the request to the target plugin as a specialized event.
```json
{
  "id": "skill_get_balance_1712345678",
  "type": "event",
  "name": "skill_get_balance",
  "payload": {
    "from": "user123",
    "group_id": "group456",
    "additional_param": "..."
  }
}
```

#### 3. Skill Result (Plugin B -> Core -> Plugin A)
Plugin B returns the result, which the Core then routes back to Plugin A using the `correlation_id`.

## Usage

### Running the Test Program
```bash
go run main.go
```

### Creating a New Plugin

1. Create a new directory under `plugins/`
2. Create `plugin.json` with plugin configuration
3. Write plugin code (Go/Python/any language)
4. Compile to executable (if needed)
5. Run the plugin manager

## Security

- Plugins run as independent OS processes
- Communication limited to stdin/stdout with JSON
- Core never trusts plugin code, only plugin.json
- Action whitelists prevent unauthorized operations
- Plugin signature verification (future enhancement)

## Hot Update

The plugin system supports hot updates:
```go
pm.HotUpdatePlugin("echo", "1.0.1")
```

## Grayscale Deployment

Plugins can be deployed in grayscale to minimize downtime during updates.

## Examples

### Go Echo Plugin
```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	decoder := json.NewDecoder(os.Stdin)
	for {
		var event map[string]interface{}
		if err := decoder.Decode(&event); err != nil {
			break
		}
		
		// Handle event
		response := map[string]interface{}{
			"id": event["id"],
			"ok": true,
			"actions": []map[string]interface{}{
				{
					"type": "send_message",
					"target": event["payload"].(map[string]interface{})["from"],
					"target_id": event["payload"].(map[string]interface{})["group_id"],
					"text": "Echo: " + event["payload"].(map[string]interface{})["text"].(string),
				},
			},
		}
		
		json.NewEncoder(os.Stdout).Encode(response)
		os.Stdout.Sync()
	}
}
```

### Python Echo Plugin
```python
import json
import sys

def main():
    for line in sys.stdin:
        event = json.loads(line)
        response = {
            "id": event["id"],
            "ok": True,
            "actions": [
                {
                    "type": "send_message",
                    "target": event["payload"]["from"],
                    "target_id": event["payload"]["group_id"],
                    "text": f"Python Echo: {event['payload']['text']}"
                }
            ]
        }
        json.dump(response, sys.stdout)
        sys.stdout.write("\n")
        sys.stdout.flush()

if __name__ == "__main__":
    main()
}
```