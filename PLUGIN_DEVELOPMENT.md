# BotMatrix 插件开发文档

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

## 🚀 快速开始

### 1. Go插件示例
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

if __name__ == "__main__":
    main()
```

### 3. C#插件示例
```csharp
using System;
using System.Text.Json;

namespace EchoPlugin
{
    public class EventMessage
    {
        public string id { get; set; }
        public string type { get; set; }
        public string name { get; set; }
        public JsonElement payload { get; set; }
    }

    class Program
    {
        static void Main(string[] args)
        {
            while (true)
            {
                string line = Console.ReadLine();
                var msg = JsonSerializer.Deserialize<EventMessage>(line);
                
                var response = new {
                    id = msg.id,
                    ok = true,
                    actions = new[] {
                        new {
                            type = "send_message",
                            text = $"C# Echo: {msg.payload.GetProperty(\"text\").GetString()}"
                        }
                    }
                };
                
                Console.WriteLine(JsonSerializer.Serialize(response));
                Console.Out.Flush();
            }
        }
    }
}
```

## 🧪 测试插件

### 1. 手动测试
```bash
echo '{"id":"test1","type":"event","name":"on_message","payload":{"text":"hello"}}' | ./your_plugin
```

### 2. 自动化测试
```bash
python test_plugin.py
```

### 3. 集成测试
```bash
go run src/robot_test_framework.go your_plugin --test-file test_cases.json
```

## 📦 发布插件

### 1. Go编译
```bash
go build -o your_plugin.exe src/plugins/your_plugin/your_plugin.go
```

### 2. Python打包
```bash
pyinstaller --onefile src/plugins/your_plugin/your_plugin.py
```

### 3. C#编译
```bash
dotnet publish -c Release -r win-x64 --self-contained true
```

### 4. 多平台发布
```bash
bash publish_multiplatform.sh
```

## 🎨 最佳实践

### 1. 错误处理
```go
if err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    continue
}
```

### 2. 性能优化
```go
encoder.SetEscapeHTML(false) // 提高JSON序列化性能
```

### 3. 安全注意事项
- 不要泄露敏感信息
- 验证所有输入数据
- 限制插件权限

## 📚 参考资料

### 协议文档
- [插件通信协议](src/plugin/core/protocol.go)
- [插件管理系统](src/plugin/core/manager.go)

### 示例插件
- [Go回声插件](src/plugins/echo/echo.go)
- [Python回声插件](src/plugins/echo_python/echo.py)
- [C#回声插件](src/plugins/echo_csharp/Program.cs)

## 🤝 贡献指南

### 开发流程
1. Fork仓库
2. 创建功能分支
3. 提交代码
4. 发起Pull Request

### 代码规范
- 遵循Go/Python/C#代码规范
- 添加必要的注释
- 编写单元测试

## 📞 支持

### 问题反馈
- [GitHub Issues](https://github.com/BotMatrix/BotMatrix/issues)
- [Discord社区](https://discord.gg/botmatrix)
- [文档](https://botmatrix.github.io/docs)

---

**BotMatrix Team** | 2024
