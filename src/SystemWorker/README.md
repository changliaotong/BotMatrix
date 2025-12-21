# 🧠 BotMatrix SystemWorker

> **The "Cortex" of your Bot Network.**  
> *Data Visualization | Remote Execution | Global Orchestration*

**SystemWorker** is the centralized intelligence unit for the BotMatrix ecosystem. While **BotNexus** acts as the high-performance gateway ("The Body") handling WebSocket connections, **SystemWorker** serves as "The Mind", processing complex logic, generating visualizations, and orchestrating actions across multiple bots.

---

## 🔥 Key Features

### 1. 📊 Real-Time Visual Dashboard
Stop reading logs. **See** your system status.
- **Command**: `#sys status`
- **Output**: A dynamically generated HD image containing:
    - **System Vitality**: Real-time CPU & Memory usage waveforms.
    - **Bot Matrix**: Live status indicators for all connected bots (QQ, WeChat, Telegram, etc.).
    - **Traffic Prediction**: AI-simulated traffic trends (24H).
- **Tech Stack**: `Matplotlib` + `NumPy` + `Pillow`.

### 2. 💻 Remote Python Execution (God Mode)
Debug, patch, and explore your runtime environment without SSH.
- **Command**: `#sys exec <python_code>`
- **Example**: 
    ```python
    #sys exec import os; print(os.listdir('/app'))
    ```
- **Security**: 
    - 🔒 **Strict UserID Whitelist Enforcement**.
    - 🛡️ Output capture via `contextlib`.

### 3. 📢 Omni-Channel Broadcast
One command to rule them all. Instantly push announcements to every group across every platform.
- **Command**: `#sys broadcast <message>`
- **Scope**: WeChat, QQ, DingTalk, Lark, Telegram, Discord.

---

## ⚙️ Configuration

SystemWorker is configured via environment variables in `docker-compose.yml`.

| Variable | Default | Description |
|----------|---------|-------------|
| `BOT_MANAGER_URL` | `ws://bot-manager:3001` | WebSocket address of BotNexus Gateway. |
| `WORKER_NAME` | `SystemWorker-Core` | Identity name shown in logs. |
| `ADMIN_USER_ID` | `1098299491` | **CRITICAL**: The UserID allowed to execute sensitive commands. |

---

## 🛠 Deployment

SystemWorker is fully integrated into the BotMatrix Docker stack.

### Quick Start
```bash
# Update and restart only the SystemWorker
python scripts/update.py --services system-worker
```

### Manual Build
```bash
cd SystemWorker
docker build -t botmatrix-system-worker .
docker run -e BOT_MANAGER_URL=ws://host.docker.internal:3001 botmatrix-system-worker
```

---

## 🧩 Development Guide

SystemWorker is designed to be easily extensible. 

### Adding a New Command
Edit `main.py` and add a new condition in `handle_message`:

```python
# Example: Adding a #ping command
elif raw_msg == "#ping":
    latency = (datetime.now() - start_time).total_seconds() * 1000
    await send_reply(ws, data, f"🏓 Pong! Latency: {latency:.2f}ms")
```

### Architecture
- **Language**: Python 3.9 Slim
- **Communication**: Reverse WebSocket (OneBot V11 Standard)
- **Concurrency**: `asyncio` for non-blocking IO.

---

## ⚠️ Security Notice

> **Warning**: The `#sys exec` command allows arbitrary code execution. 
> Ensure `ADMIN_USER_ID` is correctly set to YOUR UserID only. 
> Do not expose this worker to public groups without strict whitelisting.

---
*Powered by BotMatrix*
