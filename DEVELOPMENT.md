# BotMatrix 开发与部署指南

## 1. 快速启动 (Production Mode)

如果你只是想运行最新编译的版本：

1.  **关闭旧进程**：确保 5000 端口没有被占用（关闭之前的命令行窗口或结束 `BotNexus.exe`）。
2.  **运行程序**：双击运行 `src/BotNexus/BotNexus.exe`。
3.  **访问地址**：
    *   管理后台：[http://192.168.0.115:5000/](http://192.168.0.115:5000/)
    *   高级终端 (Overmind)：[http://192.168.0.115:5000/overmind/](http://192.168.0.115:5000/overmind/)

## 2. 代码修改与更新说明

### 后端 (Go)
*   **修改位置**：主要在 `src/BotNexus`, `src/BotWorker`, `src/Common`。
*   **生效方式**：**必须重新编译**。
    *   在 `src/BotNexus` 目录下运行：`go build -o BotNexus.exe main.go`
    *   编译完成后重启 `.exe` 文件。

### 前端 (WebUI - Vue)
*   **修改位置**：`src/WebUI`。
*   **生产环境生效**：
    1.  运行编译指令：`npm run build`
    2.  编译出的文件位于 `src/WebUI/dist`。
    3.  `BotNexus` 会自动读取该目录下的静态资源。
*   **开发环境 (推荐)**：
    1.  进入 `src/WebUI`。
    2.  运行：`npm run dev`
    3.  访问：`http://localhost:5173`。此模式支持**热更新**，修改代码后网页实时刷新。

### 高级终端 (Overmind - Flutter)
*   **修改位置**：`src/Overmind`。
*   **生效方式**：
    1.  运行编译指令：`flutter build web`
    2.  编译出的文件位于 `src/Overmind/build/web`。
    3.  `BotNexus` 会通过 `/overmind/` 路径服务这些文件。

## 3. 常见问题排查

*   **日志打印位置不对**：已在 `src/Common/log/log.go` 中通过 `zap.AddCallerSkip(1)` 修复，现在会打印实际调用者的位置，而不是 `log.go` 的位置。
*   **页面刷新 404**：已在 `src/BotNexus/main.go` 中添加了 SPA (Single Page Application) 路由支持，前端刷新页面不会再报错。
*   **端口冲突**：如果看到 `bind: Only one usage of each socket address` 错误，说明 5000 端口被占用，请检查并关闭旧的 `BotNexus.exe`。
