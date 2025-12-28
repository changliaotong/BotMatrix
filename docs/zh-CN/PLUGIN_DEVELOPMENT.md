# BotMatrix 插件开发文档

> [🌐 English](../en-US/PLUGIN_DEVELOPMENT.md) | [简体中文](PLUGIN_DEVELOPMENT.md)
> [⬅️ 返回文档中心](README.md) | [🏠 返回项目主页](../../README.md)

## 🎯 插件系统概述

BotMatrix插件系统是一个跨平台、稳定、可扩展的插件架构，支持多种编程语言。

### 核心特性
- **进程级插件**：每个插件作为独立进程运行，确保安全隔离
- **JSON协议**：通过标准输入输出进行JSON通信
- **跨平台**：支持Windows、Linux、macOS
- **多语言**：支持Go、Python、C#等多种语言

## 📦 插件结构

### 基本目录结构
```
src/plugins/your_plugin/
├── your_plugin.go      # Go插件
├── your_plugin.py      # Python插件
├── your_plugin.cs      # C#插件
└── plugin.json         # 插件配置文件
```

### 插件配置文件 (plugin.json)
```json
{
  "name": "echo_csharp",
  "description": "C#语言实现的回声插件",
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

## 🛠️ 使用 SDK 开发 (推荐)

虽然您可以直接处理 JSON 通信，但我们强烈建议使用官方提供的 SDK，它们封装了复杂的交互逻辑、分布式状态管理和指令路由。

- **Go SDK**: 适用于高性能插件。
- **Python SDK**: 适用于 AI 和快速原型开发。
- **C# SDK**: 适用于企业级应用。

详细使用说明请参考：**[插件 SDK 开发指南](plugins/sdk_guide.md)**。

## 📦 打包与分发 (.bmpk)

BotMatrix 使用 `.bmpk` (BotMatrix Package) 作为标准插件分发格式。它实际上是一个包含插件代码和 `plugin.json` 的压缩包。

### 使用 bm-cli 工具
1. **安装**: `go build -o bm-cli src/tools/bm-cli/main.go`
2. **打包**: `./bm-cli pack ./your_plugin_dir`
3. **安装**: 将生成的 `.bmpk` 文件上传到 BotNexus 管理后台，或放置在插件热加载目录。

## 🚀 快速开始 (原生协议)

如果您不想使用 SDK，可以参考以下原生协议实现：

### 1. Go 示例
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

### 2. Python插件示例
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
