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
  "id": "com.botmatrix.example",
  "name": "echo_csharp",
  "description": "C#语言实现的回声插件",
  "author": "Developer",
  "version": "1.0.0",
  "entry_point": "echo_csharp.exe",
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

## 🛠️ 使用 SDK 开发 (推荐)

虽然您可以直接处理 JSON 通信，但我们强烈建议使用官方提供的 SDK，它们封装了复杂的交互逻辑、分布式状态管理和指令路由。

- **Go SDK**: 适用于高性能插件。
- **Python SDK**: 适用于 AI 和快速原型开发。
- **C# SDK**: 适用于企业级应用。

详细使用说明请参考：**[插件 SDK 开发指南](plugins/sdk_guide.md)**。

## 📦 打包与分发 (.bmpk)

BotMatrix 使用 `.bmpk` (BotMatrix Package) 作为标准插件分发格式。

### 使用 bm-cli 工具
1. **安装**: `go build -o bm-cli src/tools/bm-cli/main.go`
2. **初始化**: `./bm-cli init my_plugin --lang go` (自动生成模版代码和规范的 `plugin.json`)
3. **本地调试**: `./bm-cli debug ./my_plugin` (无需安装，直接在本地模拟核心环境进行交互测试)
4. **自动化测试**: `./bm-cli test ./my_plugin` (运行 `tests.json` 中定义的自动化测试用例)
5. **打包**: `./bm-cli pack ./my_plugin`
6. **安装**: 将生成的 `.bmpk` 文件上传到 BotNexus 管理后台。

## 🔍 调试插件

为了方便开发者调试，`bm-cli` 提供了交互式的调试环境：

```bash
./bm-cli debug ./your_plugin_dir
```

### 调试命令
- `msg <text>`: 模拟发送一条文本消息。插件会收到 `on_message` 事件。
- `event <name> <json_payload>`: 模拟发送自定义事件。
- `exit`: 退出调试会话。

### 调试特性
- **实时日志**: 插件输出到 `stderr` 的日志会实时显示在控制台中。
- **动作捕获**: 插件尝试执行的所有 `Action`（如发送消息、调用技能）都会被拦截并打印在控制台，方便验证逻辑。
- **独立运行**: 调试环境完全模拟了核心协议，无需运行完整的 BotNexus 或 BotWorker。


## 🧪 自动化测试

`bm-cli` 支持基于 JSON 的自动化回归测试。在插件目录下创建 `tests.json` 文件：

```json
[
  {
    "name": "基础 Ping 测试",
    "input": {
      "type": "on_message",
      "payload": { "text": "ping" }
    },
    "expect": [
      { "type": "send_text", "text": "pong!" }
    ]
  }
]
```

### 运行测试
```bash
./bm-cli test ./your_plugin_dir
```

该工具会：
1. 启动插件。
2. 发送 `input` 中定义的事件。
3. 捕获插件的响应。
4. 验证响应中的 `actions` 是否与 `expect` 一致。
5. 输出测试结果报告。


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
