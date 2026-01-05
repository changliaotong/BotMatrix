# Overmind - 移动端控制中心

[English](../../en-US/components/Overmind.md) | [返回项目主页](../../../README.md) | [返回文档中心](../README.md)

Overmind 是 BotMatrix 的中央控制单元移动端应用，提供可视化管理与监控。

## 安装与设置

由于项目结构是手动生成的，请在此目录下运行以下命令以生成平台特定文件（Android/iOS）：

```bash
flutter create .
```

## 功能特性

- **Nexus 仪表盘**: 查看所有连接的机器人及其状态。
- **日志控制台**: 来自 BotNexus 的实时日志流。
- **科幻风 UI**: 具有“主脑”美学的深色模式界面。

## 连接配置

默认情况下，应用尝试连接到：
- `ws://10.0.2.2:3005` (Android 模拟器回环地址)
- `ws://localhost:3005` (Web/桌面端)

确保 BotNexus 正在运行且 3005 端口已暴露。
