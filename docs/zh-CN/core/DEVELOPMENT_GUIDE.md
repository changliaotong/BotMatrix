# BotMatrix 开发全书 (Development Guide)

> [⬅️ 返回文档中心](../README.md) | [🏠 返回项目主页](../../README.md)

本指南涵盖了 BotMatrix 的开发环境搭建、代码规范、测试规划以及 CI/CD 流程，旨在为开发者提供全方位的技术支持。

## 7. 插件开发 (Plugin Development)

BotMatrix 插件系统采用跨平台、解耦的架构，支持多种语言开发。

### 7.1 核心特性
- **进程级隔离**：每个插件作为独立进程运行。
- **JSON 通信**：通过标准输入输出 (STDIO) 进行 JSON 交互。
- **打包格式**：使用 `.bmpk` (BotMatrix Package) 标准格式进行分发。

### 7.2 插件配置 (`plugin.json`)
```json
{
  "id": "com.botmatrix.example",
  "name": "echo_plugin",
  "entry_point": "echo.exe",
  "run_on": ["worker"],
  "permissions": ["send_msg", "call_skill"]
}
```

### 7.3 开发工具 (`bm-cli`)
- **初始化**：`./bm-cli init my_plugin --lang go`
- **本地调试**：`./bm-cli debug ./my_plugin` (模拟核心环境进行交互测试)
- **打包**：`./bm-cli pack ./my_plugin`

---

## 1. 快速启动 (Development Setup)

### 1.1 后端开发 (Go)
*   **核心模块**：`src/BotNexus`, `src/BotWorker`, `src/Common`。
*   **编译运行**：
    ```bash
    cd src/BotNexus
    go build -o BotNexus.exe main.go
    ./BotNexus.exe
    ```
*   **ID 设计**：
    - `user_id` 从 `980000000000` 开始自增。
    - `group_id` 从 `990000000000` 开始自增。
    - 使用 `FlexibleInt64` 处理 JSON 中的数字/字符串兼容性。

### 1.2 前端开发 (WebUI - Vue)
*   **模块位置**：`src/WebUI`。
*   **开发模式**：`npm run dev` (支持热更新，访问 `http://localhost:5173`)。
*   **生产构建**：`npm run build` (输出至 `src/WebUI/dist`)。

### 1.3 终端开发 (Overmind - Flutter)
*   **模块位置**：`src/Overmind`。
*   **构建指令**：`flutter build web` (输出至 `src/Overmind/build/web`)。

---

## 2. 国际化 (I18N) 规范

为了确保全球化支持，本项目遵循“零硬编码”原则。

- **严禁** 在代码中直接书写可见文本，必须使用 `t('key_name')`。
- **翻译同步**：新增 Key 时必须同时在 `zh-CN`, `zh-TW`, `en-US`, `ja-JP` 中添加。
- **Key 命名**：使用小写蛇形命名法，如 `menu_dashboard`, `btn_save`。
- **自动化审计**：运行 `node scripts/audit-i18n.js` 检查 Key 是否对齐。

---

## 3. AI 智能中心开发

AI 已经深度集成到任务创建、策略管理等多个模块。

- **核心接口**：
    - `POST /api/ai/parse`: 自然语言意图解析。
    - `POST /api/ai/confirm`: 确认并执行 AI 生成的任务。
    - `GET /api/system/capabilities`: 获取系统能力清单，作为 AI 的 System Prompt。
- **安全流程**：所有 AI 生成的指令必须经过人工确认 (Draft ID 机制)。

---

## 4. 自动化测试规划

我们采用测试金字塔模型：

- **单元测试** (Base): 验证 `src/Common`, `src/BotNexus` 等核心模块逻辑。使用 `go test`。
- **集成测试** (Middle): 验证 Nexus-Worker 通信、数据库交互。使用 `httptest` 和临时数据库容器。
- **端到端测试** (Top): 验证完整业务流程。使用 Playwright 模拟 WebUI 操作。

---

## 5. CI/CD 流程

### 5.1 持续集成 (CI)
- **触发**：每次 PR 和 Push。
- **任务**：Linting (`golangci-lint`)、单元测试、安全性扫描 (`gosec`)。

### 5.2 持续交付 (CD)
- **镜像管理**：推送到 GitHub Packages (`ghcr.io`)。
- **部署策略**：
    - **云端**：Docker Compose 自动拉取更新。
    - **私有化**：CI 编译跨平台二进制文件并发布 GitHub Release。

---

## 6. 常见问题排查

- **日志位置**：已通过 `zap.AddCallerSkip(1)` 修复，打印实际调用位置。
- **刷新 404**：Nexus 已内置 SPA 路由支持。
- **端口冲突**：确保 5000 端口未被占用。
- **并发安全**：WebSocket 写入已使用 `sync.Mutex` 保护。

---
