# 🏗️ BotMatrix Plugin System & Development Guide
> [⬅️ Back to Docs](../README.md) | [🏠 Back to Home](../../README.md)

Welcome to the BotMatrix Plugin System. This guide covers the architectural design, development with SDKs, and the standardized distribution process.

---

## 1. System Overview
The BotMatrix plugin system is a cross-platform, stable, and extensible architecture that supports multiple programming languages.

### Core Features
- **Process-Level Isolation**: Each plugin runs as an independent process, ensuring security and stability.
- **JSON Communication**: Uses a standardized JSON protocol via stdin/stdout for seamless inter-process communication.
- **Multi-Language Support**: Official SDKs provided for **Go**, **Python**, and **C#**.
- **Hot Reloading**: Supports dynamic loading and updating without restarting the core service.

---

## 2. Plugin Development with SDK (Recommended)
While you can handle JSON communication directly, we strongly recommend using the official SDKs. They encapsulate complex interaction logic, distributed state management, and command routing.

### Core SDK Features
- **Interactive Ask**: Write asynchronous conversation flows like synchronous code.
- **Smart Routing**: Support for string prefixes, regex matching, and intent-based triggering.
- **Plugin Interop**: Allows plugins to export skills and perform secure inter-plugin calls.
- **Correlation ID**: Native support for message tracking in distributed environments.

### CLI Development Tool (bm-cli)
To accelerate development, use the `bm-cli` tool:
```bash
# Install (Go)
go build -o bm-cli src/tools/bm-cli/main.go

# Initialize a new plugin
./bm-cli init MyWeatherPlugin --lang go
./bm-cli init MyAIAgent --lang python
```

---

## 3. Plugin Configuration (plugin.json)
Every plugin must include a `plugin.json` file to declare metadata, entry points, and required permissions.

```json
{
  "id": "com.example.weather",
  "name": "Weather Assistant",
  "version": "1.0.0",
  "entry": "python main.py",
  "permissions": ["send_message", "send_image"],
  "intents": [
    {
      "name": "get_weather",
      "keywords": ["weather", "temperature"],
      "priority": 10
    }
  ]
}
```

---

## 4. Nexus Core Plugin (System-Level)
The `CorePlugin` is a special system-level plugin integrated into the `BotNexus` routing layer. It handles:
- **Global Flow Control**: Maintenance modes and system-wide toggles.
- **Permission Arbitration**: Black/white lists for users, robots, and groups.
- **Content Filtering**: Sensitive word library and URL filtering.
- **Admin Commands**: `/system status`, `/system open/close`, etc.

---

## 5. Plugin Market Specification (BMPK)
To achieve a standardized distribution process, we use the **BotMatrix Plugin Package (BMPK)** format.

- **Format**: A `.bmpk` file is a signed ZIP archive containing the code and `plugin.json`.
- **Packaging**: Use `./bm-cli pack ./your_plugin_dir`.
- **Installation**: Upload via the Dashboard or place in the hot-load directory.
- **Security**: Users must confirm requested permissions during installation.

---
*Last Updated: 2026-01-13*
