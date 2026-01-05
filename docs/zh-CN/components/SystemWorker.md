# 🧠 BotMatrix SystemWorker

[English](../../en-US/components/SystemWorker.md) | [返回项目主页](../../../README.md) | [返回文档中心](../README.md)

> **机器人网络的“皮层”**  
> *数据可视化 | 远程执行 | 全局编排*

**SystemWorker** 是 BotMatrix 生态系统的集中智能单元。如果说 **BotNexus** 是负责处理连接的高性能网关（“身体”），那么 **SystemWorker** 就是“大脑”，负责处理复杂逻辑、生成可视化数据并跨多个机器人编排行动。

---

## 🔥 核心特性

### 1. 📊 实时可视化仪表盘
告别阅读日志。**直观查看**系统状态。
- **命令**: `#sys status`
- **输出**: 动态生成的高清图像，包含：
    - **系统活力**: 实时 CPU 和内存使用率波形图。
    - **机器人矩阵**: 所有连接的机器人（QQ、微信、Telegram 等）的实时状态指示。
    - **流量预测**: AI 模拟的流量趋势（24小时）。
- **技术栈**: `Matplotlib` + `NumPy` + `Pillow`。

### 2. 💻 远程 Python 执行 (上帝模式)
无需 SSH 即可调试、修补和探索运行时环境。
- **命令**: `#sys exec <python_code>`
- **示例**: 
    ```python
    #sys exec import os; print(os.listdir('/app'))
    ```
- **安全性**: 
    - 🔒 **严格执行用户 ID 白名单**。
    - 🛡️ 通过 `contextlib` 捕获输出。

### 3. 📢 全频道广播
一条命令统治所有。立即向每个平台的每个群组推送公告。
- **命令**: `#sys broadcast <message>`
- **范围**: 微信、QQ、钉钉、飞书、Telegram、Discord。

---

## ⚙️ 配置说明

SystemWorker 通过 `docker-compose.yml` 中的环境变量进行配置。

| 变量 | 默认值 | 描述 |
|----------|---------|-------------|
| `BOT_MANAGER_URL` | `ws://bot-manager:3001` | BotNexus 网关的 WebSocket 地址。 |
| `WORKER_NAME` | `SystemWorker-Core` | 日志中显示的身份名称。 |
| `ADMIN_USER_ID` | `1098299491` | **关键**: 允许执行敏感命令的用户 ID。 |

---

## 🛠 部署指南

SystemWorker 已完全集成到 BotMatrix Docker 栈中。

### 快速开始
```bash
# 仅更新并重启 SystemWorker
python scripts/update.py --services system-worker
```

### 手动构建
```bash
cd SystemWorker
docker build -t botmatrix-system-worker .
docker run -e BOT_MANAGER_URL=ws://host.docker.internal:3001 botmatrix-system-worker
```

---

## 🧩 开发指南

SystemWorker 设计易于扩展。

### 添加新命令
编辑 `main.py` 并在 `handle_message` 中添加新条件：

```python
# 示例：添加 #ping 命令
elif raw_msg == "#ping":
    latency = (datetime.now() - start_time).total_seconds() * 1000
    await send_reply(ws, data, f"🏓 Pong! Latency: {latency:.2f}ms")
```

### 架构设计
- **语言**: Python 3.9 Slim
- **通信**: 反向 WebSocket (OneBot V11 标准)
- **并发**: `asyncio` 实现非阻塞 IO。

---

## ⚠️ 安全须知

> **警告**: `#sys exec` 命令允许执行任意代码。
> 确保 `ADMIN_USER_ID` 仅正确设置为您的用户 ID。
> 在没有严格白名单的情况下，不要在公开群组中暴露此 Worker。

---
*Powered by BotMatrix*
