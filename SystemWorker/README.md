# 🧠 BotMatrix SystemWorker

> The "Cortex" of your Bot Network.

**SystemWorker** is the high-level logic processing unit for BotMatrix. While BotNexus handles the "body" (connections), SystemWorker handles the "mind".

## 🔥 Features (The "Wow" Factor)

### 1. 📊 Visual Dashboard
Forget text logs. Get a real-time generated dashboard of your bot network.
- **Command**: `#sys status`
- **Output**: A generated image showing CPU/Memory waves, Bot online status, and traffic predictions.

### 2. 💻 Remote Python Exec (God Mode)
Execute arbitrary Python code directly from your chat window.
- **Command**: `#sys exec <python_code>`
- **Example**: `#sys exec import os; print(os.listdir('/'))`
- **Security**: Strict UserID check enabled.

### 3. 📢 Global Broadcast
One command to rule them all. Send announcements to all channels across all bots.
- **Command**: `#sys broadcast <message>`

## 🛠 Deployment

SystemWorker is automatically deployed via Docker Compose in the BotMatrix stack.

```bash
# Manual update
python scripts/update.py --services system-worker
```

## 🧩 Architecture

- **Language**: Python 3.9
- **Libraries**: `websockets` (IO), `matplotlib` (Viz), `pandas` (Data)
- **Role**: Connected as a `worker` to BotNexus.

---
*Part of the BotMatrix Project.*
