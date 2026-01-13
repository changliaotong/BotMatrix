# 🛠️ BotMatrix 开发者全景指南 (Development Guide)

> [⬅️ 返回文档中心](README.md) | [🏠 返回项目主页](../../README.md)

本指南面向 BotMatrix 的核心开发者与贡献者，涵盖了从环境搭建、代码规范到高级 AI 能力接入的全方位技术细节。

---

## 1. 快速启动与环境搭建 (Quick Start)

### 1.1 开发环境 (Development Mode)
- **后端 (Go)**:
  - 目录: `src/BotNexus`, `src/BotWorker`, `src/Common`。
  - 运行: `go run main.go`。
- **前端 (Vue 3)**:
  - 目录: `src/WebUI`。
  - 运行: `npm run dev` (支持热更新，访问 `http://localhost:5173`)。
- **高级终端 (Flutter)**:
  - 目录: `src/Overmind`。
  - 运行: `flutter run -d chrome`。

### 1.2 生产编译 (Production Build)
- **后端**: `go build -o BotNexus.exe main.go` (在相应目录下执行)。
- **前端**: `npm run build` (产物位于 `dist/`，由 BotNexus 静态服务)。
- **Overmind**: `flutter build web` (产物位于 `build/web/`，映射至 `/overmind/` 路径)。

---

## 2. 核心开发规范 (Standard Practices)

### 2.1 国际化 (I18N) 零硬编码原则
- **严禁** 在代码中直接书写可见文本。
- **统一获取**: 使用 `t('key_name')`。
- **翻译源**: `src/WebUI/src/utils/i18n.ts`。必须同步更新 `zh-CN`, `zh-TW`, `en-US`, `ja-JP` 四个语种。
- **命名**: 使用小写蛇形命名法，如 `btn_save`, `menu_dashboard`。

### 2.2 数据库设计规范 (Database Schema)
- **统一蛇形命名**: 表名和列名必须使用 `snake_case` (如 `user_id`)。
- **核心业务表**:
  - `bot_entities`: 存储机器人实例信息 (ID, Platform, Status)。
  - `message_logs`: 记录所有原始消息流。
  - `users`: 管理系统后台用户 (RBAC)。
- **AI 核心表**:
  - `ai_providers` / `ai_models`: 配置 AI 供应商与模型参数。
  - `ai_usage_logs`: 记录 Token 消耗与响应时长。
- **ID 策略**:
  - `user_id` 和 `group_id` 统一使用 `BIGINT` (Go 中为 `int64`)。
  - 内部 ID 范围: 用户从 `980000000000` 开始，群组从 `990000000000` 开始。
- **时间字段**: 统一使用 `created_at`, `updated_at`, `deleted_at` (GORM 软删除)。

### 2.3 零耦合事件模式
- 模块间禁止直接调用，必须通过 `EventNexus` 发布/订阅事件。
- 示例: `robot.Events.PublishAsync(new SystemAuditEvent { ... })`。

---

## 3. 技术实现细节 (Technical Deep Dive)

### 3.1 平台兼容性处理
- **FlexibleInt64**: 引入自定义类型处理 JSON 中字符串与数字混用的情况，自动解析或回退。
- **QQGuild OpenID**: `users` 表包含 `target_user_id` (数字映射) 和 `user_openid` (原始字符串)，通过 `EnsureIDs` 自动维护映射。

### 3.2 并发安全
- **WebSocket 保护**: 针对 WebSocket 写入操作，必须使用 `sync.Mutex` 保护连接对象，防止高并发下的 panic。

---

## 4. AI 智能中心接入 (AI Integration)

AI 是 BotMatrix 的核心能力。开发者可以通过统一接口调用 AI 逻辑。

### 4.1 AI 解析与确认流程
1. **解析**: `POST /api/ai/parse` -> 返回 `draft_id` 及建议的操作。
2. **人工审查**: 用户在 UI 确认 AI 建议。
3. **确认**: `POST /api/ai/confirm` -> 传入 `draft_id` 执行动作。

### 4.2 系统能力清单 (Capability Manifest)
- `GET /api/system/capabilities`: 返回系统支持的所有动作（Actions）和触发器（Triggers）。
- **用途**: 作为 System Prompt 喂给 AI，使其立即理解当前系统的能力边界。

---

## 5. 测试与质量保障 (Testing)

### 5.1 测试金字塔
- **单元测试**: `src/Common` 核心逻辑覆盖率应 > 80%。使用 `testify/assert`。
- **集成测试**: 模拟 Nexus-Worker 通信及数据库交互。使用 `testcontainers-go`。
- **E2E 测试**: 使用 Playwright 模拟完整业务流。

### 5.2 持续集成 (CI/CD)
- 每次 PR 自动触发 `golangci-lint` 静态检查及自动化测试流水线。

---

## 6. 常见问题排查 (Troubleshooting)

- **日志位置不准**: 已通过 `zap.AddCallerSkip(1)` 修复，确保打印实际调用位置。
- **页面刷新 404**: 需确保 BotNexus 开启了 SPA 路由支持。
- **端口冲突**: 默认使用 5000 端口，若报错 `bind: Only one usage...` 请检查旧进程。

---
*最后更新日期：2026-01-13*
