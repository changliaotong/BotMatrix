# 🛠️ Development Guide

> **Version**: 2.0
> **Status**: Core development workflow established
> [🌐 English](DEVELOPMENT_GUIDE.md) | [简体中文](../zh-CN/core/DEVELOPMENT_GUIDE.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

This guide provides the necessary information for developers to build and extend the BotMatrix ecosystem.

---

## 1. Development Environment

- **Languages**: Go 1.21+, Node.js 18+ (WebUI), Python 3.10+ (AI/Data).
- **Tooling**: `bm-cli` for plugin management, Docker for environment orchestration.

---

## 2. Plugin Development

BotMatrix uses a process-isolated plugin model communicating via **JSON-STDIO**.

### 2.1 Plugin Structure
A standard plugin directory contains:
- `plugin.json`: Metadata and permissions.
- Executable (compiled from Go, Python, etc.).
- Assets/Config files.

### 2.2 `plugin.json` Example
```json
{
  "id": "com.botmatrix.echo",
  "name": "Echo Plugin",
  "version": "1.0.0",
  "entry_point": "echo.exe",
  "run_on": ["worker"],
  "permissions": ["send_msg"]
}
```

---

## 3. Using `bm-cli`

The `bm-cli` tool simplifies the development lifecycle:
- **Initialize**: `./bm-cli init my_plugin --lang go`
- **Debug**: `./bm-cli debug ./my_plugin`
- **Package**: `./bm-cli pack ./my_plugin` (creates `.bmpk` file).

---

## 4. AI & Skill Development

- **MCP Tools**: Expose custom business logic as MCP tools for AI consumption.
- **Intent Mapping**: Register new intents in `AIIntentGORM` to trigger specific workflows.
- **RAG Injection**: Use the Admin API to upload domain-specific knowledge to the vector database.

---

## 5. Contribution Guidelines

1. **Coding Style**: Follow standard Go/JS linting rules.
2. **Testing**: Unit tests are required for all core logic changes.
3. **I18N**: Use the `t('key')` pattern for all user-facing strings.
