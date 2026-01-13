# 🤖 WxBotGo 全景指南 (WxBotGo Guide)

> [⬅️ 返回组件列表](ADAPTERS_AND_COMPONENTS.md) | [🏠 返回项目主页](../../README.md)

WxBotGo 是一个基于 Go 语言开发的高性能微信机器人适配器，实现了 OneBot v11 协议标准，提供了丰富的功能接口和事件处理机制。它支持在 Windows、Linux、macOS 以及 Android (通过 Termux) 上运行。

---

## 1. 功能特性 (Features)

### 1.1 核心能力
- **OneBot v11 兼容**: 提供标准的消息上报与动作执行接口。
- **高性能**: Go 语言原生并发优势，资源占用极低。
- **多平台支持**: 一套代码，处处运行。
- **集中式配置**: 所有系统参数从中心管理平台获取，支持动态更新。

### 1.2 消息处理
- **支持类型**: 文本、图片、表情、语音、名片、文件。
- **自身消息**: 支持接收并上报机器人自身发送的消息。
- **消息管理**: 实现 `send_private_msg`, `send_group_msg`, `send_msg` 等接口。
- *注意*: 目前暂不支持 `delete_msg` (由于底层库限制)。

### 1.3 群管理
- 支持 `set_group_name` 修改群名称。
- *限制*: 暂不支持踢人、禁言、设置管理员等高级群管功能。

---

## 2. 部署与运行 (Deployment)

### 2.1 Android 手机运行 (Termux)
由于 Go 语言的跨平台编译能力，您可以直接在 Android 手机上运行 WxBotGo。

#### 编译 (在电脑上执行)
```bash
# 设置环境变量 (Windows CMD)
set GOOS=linux
set GOARCH=arm64
go build -o wxbot-android-arm64
```

#### 手机环境准备
1. 安装 **Termux** (建议从 F-Droid 下载)。
2. 获取存储权限: `termux-setup-storage`。

#### 运行方式
1. 将编译好的 `wxbot-android-arm64` 传输至手机。
2. 在 Termux 中执行:
   ```bash
   cp /sdcard/Download/wxbot-android-arm64 ~/
   chmod +x wxbot-android-arm64
   ./wxbot-android-arm64
   ```
3. **扫码登录**: 程序启动后会输出登录二维码 URL，浏览器打开扫码即可。

### 2.2 App 编译环境搭建 (Android Studio)
如果您想将 WxBot 打包成独立的 APK，需要搭建以下环境：

1. **安装 Android Studio**: 官方 SDK/NDK 环境。
2. **安装 NDK 和 CMake**:
   - 在 SDK Manager -> SDK Tools 中勾选 `NDK (Side by side)` 和 `CMake`。
3. **设置环境变量**:
   - `ANDROID_HOME`: 指向您的 Android SDK 目录。

---

## 3. 常见问题排查 (Troubleshooting)

### 3.1 Android SDK/NDK 下载失败
在中国大陆，下载 Google 服务器资源通常需要配置代理：
1. 在 Android Studio 设置中找到 **HTTP Proxy**。
2. 配置您的代理地址 (如 `127.0.0.1:7890`)。
3. 也可以使用国内镜像站 (如 [androiddevtools.cn](https://www.androiddevtools.cn/)) 下载离线包。

### 3.2 NDK 安装验证
运行 `flutter doctor`。如果 `Android toolchain` 显示绿色对勾，说明环境配置成功。

---

## 4. 更新日志 (Changelog)

### v1.0.2 (2025-12-25)
- **集中式配置**: 本地仅保存 BotNexus 连接信息，动态参数从云端获取。
- **消息管理**: 新增 `delete_msg` 动作占位（暂不可用）。

### v1.0.1 (2025-12-25)
- **自身消息支持**: 新增上报开关，支持在中心平台动态配置。
- **界面集成**: WebUI 模板更新，支持更多配置项。

### v1.0.0 (2025-12-25)
- **OneBot 标准实现**: 基础消息发送与信息查询功能上线。
- **架构定型**: 完成 core/bot.go 与 core/models.go 的重构。

---
*最后更新日期：2026-01-13*
