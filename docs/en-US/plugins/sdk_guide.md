# BotMatrix Plugin SDK Guide (v3.0)

> [简体中文](../../zh-CN/plugins/sdk_guide.md) | [🌐 English](sdk_guide.md)
> [⬅️ Back to Docs Center](../README.md) | [🏠 Back to Project Home](../../../README.md)

Welcome to the BotMatrix Plugin SDK. This SDK supports **Go**, **Python**, and **C#**, using a unified architectural design to provide a high-performance, interactive, and distributed-friendly plugin development experience.

---

## Core Features

- **Interactive Ask**: Write asynchronous conversation flows like synchronous code.
- **Smart Routing**: Support for string prefixes and regex matching for commands.
- **Intent-based Routing**: Triggering based on natural language intent (e.g., no command prefix required).
- **Plugin Interop**: Allows plugins to export skills and perform secure inter-plugin calls.
- **Correlation ID**: Native support for message tracking in distributed environments, solving state loss in stateless Workers.
- **Middleware System**: Support for middleware to implement cross-cutting concerns like logging and permission checks.
- **Escape Hatch**: Low-level Action call interface for seamless core protocol upgrades.
- **Dev Tooling**: `bm-cli` tool for generating multi-language plugin templates.

---

## CLI Development Tool (bm-cli)

To accelerate development, we provide the `bm-cli` tool.

### Installation (Go)
```bash
go build -o bm-cli src/tools/bm-cli/main.go
```

### Quick Start
```bash
# Create a Go plugin
./bm-cli init MyWeatherPlugin --lang go

# Create a Python plugin
./bm-cli init MyAIAgent --lang python
```

### Packaging (.bmpk)
Use the `pack` command to package a plugin directory into the standard distribution format:
```bash
./bm-cli pack ./MyWeatherPlugin
```
This will generate a `com.botmatrix.myweatherplugin_1.0.0.bmpk` file, which can be used for hot installation in BotNexus.

---

## Plugin Configuration (plugin.json)

Every plugin must include a `plugin.json` file to declare metadata, entry points, intents, and required permissions.

```json
{
  "id": "com.example.weather",
  "name": "Weather Assistant",
  "version": "1.0.0",
  "author": "BotMatrix Team",
  "description": "Provides global weather query services",
  "entry": "python main.py",
  "permissions": [
    "send_message",
    "send_image"
  ],
  "events": [
    "on_message"
  ],
  "intents": [
    {
      "name": "get_weather",
      "keywords": ["weather", "temperature", "rain"],
      "priority": 10
    }
  ]
}
```

### Intent Routing
By configuring keywords in `intents`, plugins can be triggered by natural language:
- **Python**: `@app.on_intent("get_weather")`
- **Go**: `p.OnIntent("get_weather", handler)`

---

## Cross-Plugin Skill Calls

Plugin A can call skills exported by Plugin B. This is implemented via IPC at the Core layer.

### Exporting a Skill (Plugin B)
- **Go**: `p.ExportSkill("check_stock", handler)`
- **Python**: `@app.export_skill("check_stock")`

### Calling a Skill (Plugin A)
- **Go**: `ctx.CallSkill("plugin_b_id", "check_stock", payload)`
- **Python**: `await ctx.call_skill("plugin_b_id", "check_stock", payload)`

---

## UI Components

Plugins can inject custom UI components into the BotMatrix WebUI by declaring the `ui` field.

### Configuration Example
```json
{
  "ui": [
    {
      "type": "panel",
      "position": "sidebar",
      "title": "Weather Panel",
      "icon": "cloud",
      "entry": "./web/index.html"
    }
  ]
}
```

### Supported Types & Positions
- **Type**: `panel` (sidebar panel), `button` (button), `tab` (tab page)
- **Position**: `sidebar` (global sidebar), `dashboard` (dashboard), `chat_action` (chat window toolbar)

---

## Permission Security Mechanism

1. **Whitelist Mode**: Only Actions declared in `permissions` are allowed to be executed by the SDK.
2. **SDK-side Interception**: If you try to call an undeclared Action (e.g., `ctx.KickUser()`), the SDK will block it and report an error immediately.
3. **Core-side Verification**: The core system also performs a second verification based on this configuration.

---

## Distributed Session Store

In distributed (stateless) environments, plugins can use the `SessionStore` provided by the SDK to store persistent data.

### Use Cases
- User login state and conversation context caching.
- Data synchronization across BotWorker instances.
- Intermediate state recording for complex workflows.

### Example (Python)
```python
# Set data
await ctx.session.set("user_level", 10, expire=3600)

# Get data
level = await ctx.session.get("user_level")
```

### Example (Go)
```go
// Set data
ctx.Session.Set("last_query", time.Now(), 24 * time.Hour)

// Get data
var lastTime time.Time
ctx.Session.Get("last_query", &lastTime)
```

---

## Code Examples

Please refer to the [Plugin Development Guide](../PLUGIN_DEVELOPMENT.md) for more detailed code examples.

---

## Key API Reference

### Context

Each Handler receives a `Context` object containing all event information and reply methods.

| Method/Property | Description |
| :--- | :--- |
| `ctx.Reply(text)` | Quickly reply with a text message to the sender/group. |
| `ctx.Ask(prompt, timeout)` | **Core**: Send a prompt and block-wait for the user's next reply. |
| `ctx.Args` | Array of arguments following the command (split by space). |
| `ctx.CorrelationId` | Unique identifier for the current session, used for distributed tracing. |

---
*Last Updated: 2025-12-28*
