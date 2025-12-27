# BotMatrix Plugin Development

> [🌐 English](PLUGIN_DEVELOPMENT.md) | [简体中文](../zh-CN/PLUGIN_DEVELOPMENT.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

## 🎯 Plugin System Overview

The BotMatrix plugin system is a cross-platform, stable, and extensible architecture that supports multiple programming languages.

### Core Features
- **Process-Level Plugins**: Each plugin runs as an independent process, ensuring secure isolation.
- **JSON Protocol**: JSON communication via standard input/output.
- **Cross-Platform**: Supports Windows, Linux, and macOS.
- **Multi-Language**: Supports various languages including Go, Python, and C#.

## 📦 Plugin Structure

### Basic Directory Structure
```
src/plugins/your_plugin/
├── your_plugin.go      # Go plugin
├── your_plugin.py      # Python plugin
├── your_plugin.cs      # C# plugin
└── plugin.json         # Plugin configuration file
```

### Plugin Configuration (plugin.json)
```json
{
  "name": "echo_csharp",
  "description": "Echo plugin implemented in C#",
  "api_version": "1.0.0",
  "version": "1.0.0",
  "entry_point": "echo_csharp.exe",
  "run_on": ["worker"],
  "capabilities": ["echo"],
  "actions": ["send_message"],
  "timeout_ms": 5000,
  "max_concurrency": 1,
  "max_restarts": 3,
  "signature": "",
  "plugin_level": "feature",
  "source": "internal"
}
```

## 🚀 Quick Start

### 1. Go Plugin Example
```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)
	
	for {
		var msg map[string]interface{}
		decoder.Decode(&msg)
		
		response := map[string]interface{}{
			"id": msg["id"],
			"ok": true,
			"actions": []map[string]interface{}{
				{
					"type": "send_message",
					"text": "Go Echo: " + msg["payload"].(map[string]interface{})["text"].(string),
				},
			},
		}
		
		encoder.Encode(response)
	}
}
```

### 2. Python Plugin Example
```python
import json
import sys

def main():
    for line in sys.stdin:
        msg = json.loads(line)
        response = {
            "id": msg["id"],
            "ok": True,
            "actions": [
                {
                    "type": "send_message",
                    "text": f"Python Echo: {msg['payload']['text']}"
                }
            ]
        }
        print(json.dumps(response))
        sys.stdout.flush()
```
