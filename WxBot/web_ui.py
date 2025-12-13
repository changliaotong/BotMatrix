from flask import Flask, render_template_string, jsonify, request, Response, send_file
import threading
import time
import json
import logging
import psutil
import os
import platform
import io
import sys
import shutil
import subprocess
import re
from collections import deque
from werkzeug.utils import secure_filename

# 日志缓冲区，用于 WebUI 显示
log_buffer = deque(maxlen=200)
# 系统状态历史缓冲区 (时间, CPU, 内存)
stats_history = deque(maxlen=60)
# 消息流量历史 (时间, 消息数/秒)
msg_stats_history = deque(maxlen=60)
msg_count_lock = threading.Lock()
msg_count_window = 0

def record_msg():
    global msg_count_window
    with msg_count_lock:
        msg_count_window += 1

def collect_stats_loop():
    global msg_count_window
    while True:
        try:
            cpu = psutil.cpu_percent(interval=None)
            mem = psutil.virtual_memory().percent
            timestamp = time.strftime("%H:%M:%S")
            
            # Calculate message rate (msgs per 2 seconds)
            with msg_count_lock:
                msg_count = msg_count_window
                msg_count_window = 0
            
            stats_history.append({
                "time": timestamp,
                "cpu": cpu,
                "mem": mem,
                "msg_count": msg_count  # Count in last 2 seconds
            })
        except Exception:
            pass
        time.sleep(2)

# 启动状态采集线程
stats_thread = threading.Thread(target=collect_stats_loop, daemon=True)
stats_thread.start()

class StreamLogger:
    def __init__(self, original_stdout):
        self.original_stdout = original_stdout
        
    def write(self, message):
        if message.strip():
            timestamp = time.strftime("%H:%M:%S")
            log_entry = f"[{timestamp}] {message}"
            log_buffer.append(log_entry)
        self.original_stdout.write(message)
        
    def flush(self):
        self.original_stdout.flush()

# 重定向 stdout 以捕获日志
if not isinstance(sys.stdout, StreamLogger):
    sys.stdout = StreamLogger(sys.stdout)

LOGIN_TEMPLATE = r"""
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>微信扫码登录 - BotMatrix</title>
    <link href="https://cdn.bootcdn.net/ajax/libs/twitter-bootstrap/5.1.3/css/bootstrap.min.css" rel="stylesheet">
    <style>
        body { 
            background-color: #f8f9fa; 
            display: flex; 
            align-items: center; 
            justify-content: center; 
            height: 100vh; 
            margin: 0;
            font-family: system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
        }
        .login-card { 
            width: 100%; 
            max-width: 400px; 
            padding: 2.5rem; 
            border-radius: 1rem; 
            box-shadow: 0 10px 25px rgba(0,0,0,0.05); 
            background: #fff; 
            text-align: center; 
            border: 1px solid #eee;
        }
        .qr-img { 
            width: 260px; 
            height: 260px; 
            object-fit: contain; 
            margin-bottom: 1.5rem; 
            border: 1px solid #f0f0f0;
            border-radius: 8px;
            padding: 10px;
        }
        .status-text {
            color: #6c757d;
            margin-bottom: 1.5rem;
            font-size: 0.95rem;
        }
        .btn-refresh {
            background-color: #0d6efd;
            border: none;
            padding: 10px 20px;
            font-weight: 500;
            letter-spacing: 0.5px;
        }
        .btn-refresh:hover {
            background-color: #0b5ed7;
            box-shadow: 0 4px 10px rgba(13, 110, 253, 0.2);
        }
        .footer-link {
            margin-top: 1.5rem;
            display: block;
            color: #adb5bd;
            text-decoration: none;
            font-size: 0.85rem;
        }
        .footer-link:hover {
            color: #6c757d;
        }
    </style>
</head>
<body>
    <div class="login-card">
        <h4 class="mb-4 fw-bold text-dark">微信扫码登录</h4>
        <img id="qr-img" src="/api/qr_code?bot_id={{ bot_id }}&ts={{ ts }}" class="qr-img" alt="正在加载二维码...">
        <p class="status-text" id="status-text">请使用手机微信扫描二维码</p>
        
        <button class="btn btn-primary w-100 btn-refresh rounded-pill" onclick="refreshQR()">
            刷新二维码
        </button>
        
        <a href="/" class="footer-link">返回仪表盘</a>
    </div>

    <script>
        const botId = "{{ bot_id }}";
        
        function refreshQR() {
            const ts = new Date().getTime();
            document.getElementById('qr-img').src = '/api/qr_code?bot_id=' + botId + '&ts=' + ts;
        }

        // Auto check login status
        setInterval(() => {
            fetch('/api/bots').then(r => r.json()).then(data => {
                const bot = data.find(b => b.self_id == botId);
                if (bot && bot.is_alive) {
                    const statusEl = document.getElementById('status-text');
                    statusEl.innerText = "登录成功！正在跳转...";
                    statusEl.style.color = "#198754";
                    statusEl.style.fontWeight = "bold";
                    
                    setTimeout(() => window.location.href = "/", 1000);
                }
            });
        }, 2000);
    </script>
</body>
</html>
"""

DEFAULT_MANUAL_CONTENT = r"""
# 服务端使用手册 (Server Manual)

本文档详细介绍了 Python 服务端（OneBot Gateway）的功能、配置、插件指令及运维操作。

---

## 1. 系统概述

本服务端核心是一个 **OneBot 协议网关**，它负责：
1.  **连接管理**：维护与 C# 客户端 (WeChat/QQ) 的 WebSocket 连接。
2.  **消息路由**：将接收到的消息分发给插件系统和 C# 业务端。
3.  **服务插件**：提供系统监控、日志记录、广播通知等服务级功能。
4.  **Web 界面**：提供二维码登录、实时日志和状态监控看板。

---

## 2. Web 控制台

服务端启动后，内置了一个轻量级 Web 服务器。

- **访问地址**: `http://服务器IP:3001` (端口默认为 3001，可在 config.json 修改)
- **主要功能**:
    - **首页 (Dashboard)**:
        - 实时显示 CPU / 内存使用率图表。
        - 显示网关连接数、系统运行时间、消息吞吐量。
        - 实时滚动日志窗口。
    - **登录页 (/login)**:
        - 当机器人掉线需要重新扫码时，可直接访问此页面获取二维码。
        - 页面会自动刷新检测登录状态。

---

## 3. 服务端插件指令

服务端内置了一套管理指令，**仅限管理员使用**。指令必须以 `#` 开头。

### 3.1 权限验证
管理员权限通过以下两种方式验证（满足任一即可）：
1.  **配置文件**: `config.json` 中的 `admins` 列表包含用户的 WXID。
2.  **数据库**: 数据库 `User` 表中登记了用户的 WXID，且该用户在 `Member` 表中被绑定为机器人的 `AdminId`。

### 3.2 指令列表

| 模块 | 指令 | 参数示例 | 功能说明 |
| :--- | :--- | :--- | :--- |
| **系统监控** | `#status` | 无 | 查看服务器 CPU、内存占用、运行时间及网关连接数。 |
| **系统维护** | `#reload` | 无 | 热重载所有插件代码（无需重启服务）。 |
| | `#gc` | 无 | 手动触发 Python 垃圾回收，释放内存。 |
| **通知广播** | `#broadcast` | `#broadcast 今晚维护` | 向过去 3 天内活跃的所有群组发送系统广播消息。 |
| **数据库** | `#db_status` | 无 | 查看 `ChatLog`、`User`、`Member` 表的数据行数统计。 |
| | `#db_clean` | `#db_clean 30` | 清理指定天数（如 30 天）前的聊天记录，释放空间。 |
| **API 调试** | `#api` | `#api {"action": "send_msg", ...}` | 发送原始 OneBot JSON 指令，用于测试底层接口。 |
| **帮助** | `#help` | 无 | 显示帮助菜单。 |

---

## 4. 数据库功能

服务端集成了数据库日志功能，需要 SQL Server 支持。

### 4.1 自动日志 (db_logger)
- 所有接收到的消息（私聊/群聊）会自动写入 `ChatLog` 表。
- **注意**: 若 `ChatLog` 表不存在，日志功能将静默失效。

### 4.2 权限表结构
若要启用数据库鉴权，请确保数据库中存在以下表结构（建表脚本位于 `create_user_table.sql`）：

- **[User] 表**: 存储管理员信息
    - `Id`: 唯一标识
    - `WxId`: 管理员微信ID (用于匹配发送者)
- **[Member] 表**: 存储机器人信息
    - `BotUin` / `UserName`: 机器人标识
    - `AdminId`: 关联到 [User].Id

---

## 5. 配置文件说明 (config.json)

```json
{
    "admins": [
        "wxid_xxxxxx"  // 超级管理员列表 (最高权限)
    ],
    "network": {
        "ws_server": {
            "port": 3001,           // 服务监听端口
            "heartbeat_interval": 5000, // 心跳间隔 (毫秒)
            "force_push_event": true    // 强制推送事件给 C# 端
        }
    },
    "bots": [
        // 多平台机器人配置 (WeChat, WxWork, DingTalk, etc.)
    ]
}
```

---

## 6. 常见问题与运维

### Q: 机器人显示“断断续续”或 WebSocket 1006 错误？
- **原因**: 网络波动或心跳超时。
- **解决**: 我们已优化了心跳机制（5秒一次）。请检查服务器防火墙是否允许 3001 端口的长连接。

### Q: 插件修改后不生效？
- **解决**: 发送 `#reload` 指令即可热加载新代码，无需重启进程。

### Q: 数据库连接失败？
- **解决**: 检查 `SQLConn.py` 中的连接字符串或环境变量配置。使用 `#db_status` 测试连接。

### Q: 如何添加新功能？
- **开发**: 在 `plugins/` 目录下新建 `.py` 文件，实现 `handle(context)` 函数即可。
- **规范**: 返回字典 `{"reply": "...", "block": True}` 可拦截消息并回复。
"""

HTML_TEMPLATE = r"""
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>BotMatrix 管理后台</title>
    <!-- Use BootCDN for better stability in China -->
    <link href="https://cdn.bootcdn.net/ajax/libs/twitter-bootstrap/5.1.3/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.bootcdn.net/ajax/libs/bootstrap-icons/1.8.1/font/bootstrap-icons.min.css" rel="stylesheet">
    <script src="https://cdn.bootcdn.net/ajax/libs/Chart.js/3.7.1/chart.min.js"></script>
    <script src="https://cdn.bootcdn.net/ajax/libs/marked/4.0.2/marked.min.js"></script>
    <style>
        /* Light Theme (Default) */
        :root {
            --bg-body: #f8f9fa;
            --bg-sidebar: #ffffff;
            --bg-card: #ffffff;
            --text-main: #212529;
            --text-muted: #6c757d;
            --border-color: #dee2e6;
            --primary-color: #0d6efd;
            --secondary-color: #6c757d;
            --accent-glow: none;
            --font-stack: system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            
            /* Custom for Info Cards */
            --info-icon-bg: rgba(13, 110, 253, 0.1);
            --info-hover-bg: rgba(0,0,0,0.02);
        }

        /* Standard Dark Theme */
        body.dark-mode {
            --bg-body: #212529;
            --bg-sidebar: #343a40;
            --bg-card: #2c3034;
            --text-main: #f8f9fa;
            --text-muted: #adb5bd;
            --border-color: #495057;
            --primary-color: #0d6efd;
            --secondary-color: #6c757d;
            --accent-glow: none;
            --font-stack: system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            
            --info-icon-bg: rgba(13, 110, 253, 0.2);
            --info-hover-bg: rgba(255,255,255,0.05);
        }
        
        body { 
            background-color: var(--bg-body); 
            color: var(--text-main); 
            font-family: var(--font-stack); 
            transition: all 0.3s ease; 
            letter-spacing: 0.5px;
        }
        
        /* Sidebar */
        .sidebar { 
            height: 100vh; position: fixed; top: 0; left: 0; width: 250px; 
            background-color: var(--bg-sidebar); 
            border-right: 1px solid var(--border-color); 
            padding-top: 20px; transition: all 0.3s ease-in-out; z-index: 1050; 
        }
        .sidebar a { color: var(--text-muted); padding: 12px 24px; display: block; text-decoration: none; border-left: 3px solid transparent; transition: all 0.2s; font-family: var(--font-stack); }
        .sidebar a:hover { color: var(--primary-color); background-color: rgba(13, 110, 253, 0.1); }
        .sidebar a.active { color: var(--primary-color); background-color: rgba(13, 110, 253, 0.1); border-left-color: var(--primary-color); font-weight: 500; }
        .sidebar .text-muted { color: var(--text-muted) !important; }
        
        .main-content { margin-left: 250px; padding: 25px; transition: margin-left 0.3s ease-in-out; }
        
        /* Cards */
        .card { 
            background-color: var(--bg-card); 
            border: 1px solid var(--border-color); 
            margin-bottom: 24px; 
            box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
            border-radius: 0.75rem; /* More rounded like the sample */
        }
        .card:hover { box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15); }
        
        .card-header { 
            background-color: transparent; 
            border-bottom: 1px solid var(--border-color); 
            padding: 15px 20px; font-weight: 600; 
        }
        
        /* Info Rows Design */
        .info-row {
            display: flex;
            align-items: center;
            padding: 0.75rem 1rem;
            margin-bottom: 0.5rem;
            border-radius: 0.5rem;
            background-color: var(--bg-card);
            box-shadow: 0 1px 2px rgba(0,0,0,0.05);
            border: 1px solid var(--border-color);
            transition: transform 0.2s;
        }
        .info-row:hover {
            background-color: var(--info-hover-bg);
            transform: translateY(-1px);
        }
        .info-icon-wrapper {
            display: flex;
            align-items: center;
            justify-content: center;
            width: 2.5rem;
            height: 2.5rem;
            border-radius: 50%;
            margin-right: 1rem;
            background-color: var(--info-icon-bg);
            color: var(--primary-color);
        }
        .info-label {
            width: 100px;
            font-weight: 500;
            color: var(--text-main);
        }
        .info-value {
            color: var(--text-muted);
            font-family: monospace;
            margin-left: auto;
        }
        
        /* Mobile Overlay */
        .sidebar-overlay { position: fixed; top: 0; left: 0; width: 100%; height: 100%; background-color: rgba(0,0,0,0.8); z-index: 1040; display: none; backdrop-filter: blur(2px); }
        .sidebar-overlay.show { display: block; }
        
        /* Toggle Button */
        .sidebar-toggle { display: none; position: fixed; bottom: 20px; right: 20px; z-index: 1060; width: 50px; height: 50px; box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15); border-radius: 50%; }
        
        /* Metrics */
        .metric-value { font-size: 1.8rem; font-weight: 700; color: var(--primary-color); }
        .metric-label { font-size: 0.85rem; color: var(--text-muted); text-transform: uppercase; letter-spacing: 1px; }
        
        /* Logs */
        .log-box { background-color: #000; color: #f8f9fa; font-family: monospace; padding: 15px; height: calc(100vh - 220px); min-height: 400px; overflow-y: auto; border: 1px solid var(--border-color); border-radius: 0.25rem; font-size: 0.85rem; }
        
        /* Tables */
        .table { color: var(--text-main); vertical-align: middle; border-color: var(--border-color); }
        .table thead th { border-bottom: 2px solid var(--border-color); font-weight: 600; text-transform: uppercase; font-size: 0.85rem; background-color: rgba(0, 0, 0, 0.03); border-color: var(--border-color); }
        .table-hover tbody tr:hover { color: var(--text-main); background-color: rgba(0, 0, 0, 0.075); }
        .table td { border-color: var(--border-color); }
        .table-light { background-color: transparent; color: var(--text-main); }
        
        /* Badges & Buttons */
        .badge { font-weight: 500; }
        .badge-soft-success { color: #198754; background-color: rgba(25, 135, 84, 0.1); border: 1px solid transparent; }
        .badge-soft-danger { color: #dc3545; background-color: rgba(220, 53, 69, 0.1); border: 1px solid transparent; }
        
        .btn { text-transform: uppercase; letter-spacing: 1px; transition: all 0.2s; }
        .btn-primary { background-color: var(--primary-color); border-color: var(--primary-color); color: #fff; }
        .btn-primary:hover { background-color: #0b5ed7; border-color: #0a58ca; color: #fff; box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15); }
        
        .btn-outline-secondary { color: var(--secondary-color); border-color: var(--secondary-color); }
        .btn-outline-secondary:hover { background-color: var(--secondary-color); color: #fff; }
        
        .btn-close { }

        /* Form Elements */
        .form-control, .form-select { background-color: var(--bg-card); border: 1px solid var(--border-color); color: var(--text-main); }
        .form-control:focus, .form-select:focus { background-color: var(--bg-card); border-color: #86b7fe; box-shadow: 0 0 0 0.25rem rgba(13, 110, 253, 0.25); color: var(--text-main); }
        .form-control::placeholder { color: var(--text-muted); opacity: 0.8; }
        
        /* Modal */
        .modal-content { background-color: var(--bg-card); border: 1px solid var(--border-color); box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15); }
        .modal-header, .modal-footer { border-color: var(--border-color); }
        .modal-title { color: var(--text-main); }
        
        /* Sortable headers */
        th.sortable { cursor: pointer; user-select: none; }
        th.sortable:hover { color: #fff; text-shadow: 0 0 5px #fff; }
        
        /* Scrollbar */
        ::-webkit-scrollbar { width: 8px; height: 8px; }
        ::-webkit-scrollbar-track { background: var(--bg-body); }
        ::-webkit-scrollbar-thumb { background: var(--border-color); border-radius: 4px; }
        ::-webkit-scrollbar-thumb:hover { background: var(--secondary-color); }

        /* Icons Colors Override */
        .text-primary { color: var(--primary-color) !important; }
        .text-success { color: #198754 !important; }
        .text-info { color: #0dcaf0 !important; }
        .text-warning { color: #ffc107 !important; }
        .text-muted { color: var(--text-muted) !important; }
        
        /* 响应式调整 */
        @media (max-width: 768px) {
            .sidebar { transform: translateX(-100%); }
            .sidebar.show { transform: translateX(0); }
            .sidebar-overlay.show { display: block; }
            .main-content { margin-left: 0; padding: 15px; }
            .sidebar-toggle { display: flex; align-items: center; justify-content: center; }
        }
    </style>
</head>
<body>
    <div class="sidebar-overlay" onclick="toggleSidebar()"></div>
    
    <button class="btn btn-primary sidebar-toggle" onclick="toggleSidebar()">
        <i class="bi bi-list fs-4"></i>
    </button>

    <div class="sidebar">
        <div class="px-3 mb-4">
            <h4>🤖 BotMatrix</h4>
            <small class="text-muted">管理后台 Next</small>
        </div>
        <div class="px-3 mb-3">
             <label class="text-muted small mb-2">当前机器人</label>
             <select class="form-select form-select-sm bg-dark text-white border-secondary" id="bot-selector" onchange="switchBot(this.value)">
                 <option value="" disabled selected>加载中...</option>
             </select>
             <button class="btn btn-sm btn-outline-success w-100 mt-2" onclick="showAddBotModal()">+ 添加机器人</button>
        </div>
        <nav class="nav flex-column">
            <a href="#dashboard" class="nav-link active" onclick="showTab('dashboard')"><i class="bi bi-speedometer2"></i> 仪表盘</a>
            <a href="javascript:void(0)" class="nav-link" id="nav-login-link" onclick="goToLogin()"><i class="bi bi-qr-code-scan"></i> 扫码登录</a>
            <a href="#groups" class="nav-link" onclick="showTab('groups')"><i class="bi bi-people"></i> 群组管理</a>
            <a href="#msgtest" class="nav-link" onclick="showTab('msgtest')"><i class="bi bi-chat-square-dots"></i> 消息测试</a>
            <a href="#logs" class="nav-link" onclick="showTab('logs')"><i class="bi bi-journal-text"></i> 系统日志</a>
            <a href="#manual" class="nav-link" onclick="showTab('manual')"><i class="bi bi-book"></i> 使用说明</a>
            <a href="#settings" class="nav-link" onclick="showTab('settings')"><i class="bi bi-gear"></i> 设置与调试</a>
        </nav>
        <div class="mt-auto px-3 py-3 border-top border-secondary">
             <div class="d-grid gap-2 mb-2">
                <button class="btn btn-outline-secondary btn-sm" onclick="toggleDarkMode()">
                    <i class="bi bi-moon-stars"></i> <span id="dark-mode-text">黑夜模式</span>
                </button>
             </div>
             <div class="d-grid gap-2">
                <!-- <button class="btn btn-outline-light btn-sm" onclick="logoutBot()"><i class="bi bi-box-arrow-right"></i> 退出登录</button> -->
            </div>
        </div>
    </div>

    <div class="main-content">
        <!-- Dashboard Tab -->
        <div id="tab-dashboard">
            <h4 class="mb-4">系统概览</h4>
            
            <div class="row mb-4">
                <div class="col-md-3">
                    <div class="card p-3">
                        <div class="d-flex justify-content-between align-items-center">
                            <div>
                                <div class="metric-label">CPU (<span id="cpu-cores-mini">-</span>核)</div>
                                <div class="metric-value" id="cpu-usage">0%</div>
                            </div>
                            <i class="bi bi-cpu fs-1 text-primary"></i>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card p-3">
                        <div class="d-flex justify-content-between align-items-center">
                            <div>
                                <div class="metric-label">内存 (Sys: <span id="mem-sys-percent">-</span>%)</div>
                                <div class="metric-value" id="mem-usage">0 MB</div>
                            </div>
                            <i class="bi bi-memory fs-1 text-success"></i>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card p-3 cursor-pointer" style="cursor: pointer;" onclick="showTab('groups')" title="点击管理群组">
                        <div class="d-flex justify-content-between align-items-center">
                            <div>
                                <div class="metric-label">群组数量</div>
                                <div class="metric-value" id="group-count-dash">0</div>
                            </div>
                            <i class="bi bi-chat-dots fs-1 text-info"></i>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card p-3">
                        <div class="d-flex justify-content-between align-items-center">
                            <div>
                                <div class="metric-label">运行时间</div>
                                <div class="metric-value" id="uptime" style="font-size: 1.2rem;">00:00:00</div>
                            </div>
                            <i class="bi bi-clock-history fs-1 text-warning"></i>
                        </div>
                    </div>
                </div>
            </div>

            <!-- System Load Chart -->
            <div class="row mb-4">
                <div class="col-md-6">
                    <div class="card">
                        <div class="card-header d-flex justify-content-between align-items-center">
                            <span>系统负载趋势</span>
                            <small class="text-muted">实时监控 (CPU / 内存)</small>
                        </div>
                        <div class="card-body">
                            <canvas id="systemLoadChart" style="height: 250px; width: 100%;"></canvas>
                        </div>
                    </div>
                </div>
                <div class="col-md-6">
                    <div class="card">
                        <div class="card-header d-flex justify-content-between align-items-center">
                            <span>消息流量趋势</span>
                            <small class="text-muted">实时监控 (消息数 / 2秒)</small>
                        </div>
                        <div class="card-body">
                            <canvas id="msgTrafficChart" style="height: 250px; width: 100%;"></canvas>
                        </div>
                    </div>
                </div>
            </div>

            <div class="row">
                <div class="col-md-6">
                     <div class="card p-3">
                        <h5 class="mb-3 ps-1">基础信息</h5>
                        
                        <!-- WXBot Version -->
                        <div class="info-row">
                             <div class="info-icon-wrapper" style="background-color: rgba(13, 110, 253, 0.1); color: #0d6efd;">
                                <svg stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 192 512" height="1.2em" width="1.2em" xmlns="http://www.w3.org/2000/svg"><path d="M48 80a48 48 0 1 1 96 0A48 48 0 1 1 48 80zM0 224c0-17.7 14.3-32 32-32l64 0c17.7 0 32 14.3 32 32l0 224 32 0c17.7 0 32 14.3 32 32s-14.3 32-32 32L32 512c-17.7 0-32-14.3-32-32s14.3-32 32-32l32 0 0-192-32 0c-17.7 0-32-14.3-32-32z"></path></svg>
                             </div>
                             <div class="info-label">WXBot 版本</div>
                             <div class="info-value">3.0.0</div>
                             <div class="ms-2">
                                <button class="btn btn-sm btn-outline-secondary rounded-pill" style="width: 20px; height: 20px; padding: 0; display: flex; align-items: center; justify-content: center;" title="Check Update">
                                    <i class="bi bi-arrow-clockwise" style="font-size: 10px;"></i>
                                </button>
                             </div>
                        </div>

                        <!-- Protocol Version -->
                        <div class="info-row">
                             <div class="info-icon-wrapper" style="background-color: rgba(255, 193, 7, 0.1); color: #ffc107;">
                                <svg stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 448 512" height="1.2em" width="1.2em" xmlns="http://www.w3.org/2000/svg"><path d="M433.754 420.445c-11.526 1.393-44.86-52.741-44.86-52.741 0 31.345-16.136 72.247-51.051 101.786 16.842 5.192 54.843 19.167 45.803 34.421-7.316 12.343-125.51 7.881-159.632 4.037-34.122 3.844-152.316 8.306-159.632-4.037-9.045-15.25 28.918-29.214 45.783-34.415-34.92-29.539-51.059-70.445-51.059-101.792 0 0-33.334 54.134-44.859 52.741-5.37-.65-12.424-29.644 9.347-99.704 10.261-33.024 21.995-60.478 40.144-105.779C60.683 98.063 108.982.006 224 0c113.737.006 163.156 96.133 160.264 214.963 18.118 45.223 29.912 72.85 40.144 105.778 21.768 70.06 14.716 99.053 9.346 99.704z"></path></svg>
                             </div>
                             <div class="info-label">协议版本</div>
                             <div class="info-value">WeChat Web</div>
                        </div>

                        <!-- WebUI Version -->
                        <div class="info-row">
                             <div class="info-icon-wrapper" style="background-color: rgba(25, 135, 84, 0.1); color: #198754;">
                                <svg stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 512 512" height="1.2em" width="1.2em" xmlns="http://www.w3.org/2000/svg"><path d="M188.8 255.925c0 36.946 30.243 67.178 67.2 67.178s67.199-30.231 67.199-67.178c0-36.945-30.242-67.179-67.199-67.179s-67.2 30.234-67.2 67.179z"></path><path d="M476.752 217.795c-.009.005-.016.038-.024.042-1.701-9.877-4.04-19.838-6.989-28.838h-.107c2.983 9 5.352 19 7.072 29h-.002c-1.719-10-4.088-20-7.07-29h-155.39c19.044 17 31.358 40.175 31.358 67.052 0 16.796-4.484 31.284-12.314 44.724L231.044 478.452s-.009.264-.014.264l-.01.284h.015l-.005-.262c8.203.92 16.531 1.262 24.97 1.262 6.842 0 13.609-.393 20.299-1.002a223.86 223.86 0 0 0 29.777-4.733C405.68 451.525 480 362.404 480 255.941c0-12.999-1.121-25.753-3.248-38.146z"></path><path d="M256 345.496c-33.601 0-61.601-17.91-77.285-44.785L76.006 123.047l-.137-.236a223.516 223.516 0 0 0-25.903 45.123C38.407 194.945 32 224.686 32 255.925c0 62.695 25.784 119.36 67.316 160.009 29.342 28.719 66.545 49.433 108.088 58.619l.029-.051 77.683-134.604c-8.959 3.358-19.031 5.598-29.116 5.598z"></path><path d="M91.292 104.575l77.35 133.25C176.483 197.513 212.315 166 256 166h205.172c-6.921-15-15.594-30.324-25.779-43.938.039.021.078.053.117.074C445.644 135.712 454.278 151 461.172 166h.172c-6.884-15-15.514-30.38-25.668-43.99-.115-.06-.229-.168-.342-.257C394.475 67.267 329.359 32 256 32c-26.372 0-51.673 4.569-75.172 12.936-34.615 12.327-65.303 32.917-89.687 59.406l.142.243.009-.01z"></path></svg>
                             </div>
                             <div class="info-label">WebUI 版本</div>
                             <div class="info-value">Next 1.0.0</div>
                        </div>

                        <!-- System Version -->
                        <div class="info-row">
                             <div class="info-icon-wrapper" style="background-color: rgba(13, 202, 240, 0.1); color: #0dcaf0;">
                                <svg stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 24 24" height="1.2em" width="1.2em" xmlns="http://www.w3.org/2000/svg"><path d="M14 18V20L16 21V22H8L7.99639 21.0036L10 20V18H2.9918C2.44405 18 2 17.5511 2 16.9925V4.00748C2 3.45107 2.45531 3 2.9918 3H21.0082C21.556 3 22 3.44892 22 4.00748V16.9925C22 17.5489 21.5447 18 21.0082 18H14ZM4 14V16H20V14H4Z"></path></svg>
                             </div>
                             <div class="info-label">系统版本</div>
                             <div class="info-value" id="os-ver">-</div>
                        </div>

                        <!-- CPU Info -->
                        <div class="info-row align-items-start">
                             <div class="info-icon-wrapper" style="background-color: rgba(102, 16, 242, 0.1); color: #6610f2;">
                                <i class="bi bi-cpu fs-5"></i>
                             </div>
                             <div class="info-label pt-1">CPU 信息</div>
                             <div class="info-value text-end" style="font-size: 0.8rem; line-height: 1.4;">
                                 <div class="fw-bold text-truncate" style="max-width: 200px;" id="cpu-model">-</div>
                                 <div class="text-muted"><span id="cpu-cores">-</span> 核 / <span id="cpu-freq">-</span></div>
                             </div>
                        </div>

                        <!-- Memory Info -->
                        <div class="info-row align-items-start">
                             <div class="info-icon-wrapper" style="background-color: rgba(220, 53, 69, 0.1); color: #dc3545;">
                                <i class="bi bi-memory fs-5"></i>
                             </div>
                             <div class="info-label pt-1">内存信息</div>
                             <div class="info-value text-end" style="font-size: 0.8rem; line-height: 1.4;">
                                 <div>总量: <span id="mem-total-detail">-</span></div>
                                 <div class="text-muted">已用: <span id="mem-used-detail">-</span></div>
                             </div>
                        </div>
                    </div>
                </div>
                <div class="col-md-6">
                    <div class="card">
                        <div class="card-header">网络状态</div>
                        <div class="card-body">
                            <table class="table table-borderless">
                                <tbody>
                                    <tr>
                                        <td><strong>HTTP 服务器</strong></td>
                                        <td><span class="badge badge-soft-success">运行中</span> <span id="http-port" class="text-muted small">Port: 5000</span></td>
                                    </tr>
                                    <tr>
                                        <td><strong>WebSocket 服务器</strong></td>
                                        <td><span class="badge badge-soft-success">运行中</span> <span id="ws-port" class="text-muted small">Port: 3001</span></td>
                                    </tr>
                                    <tr>
                                        <td><strong>WebSocket 客户端</strong></td>
                                        <td><span class="badge badge-soft-danger">未连接</span></td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>
            
            <div class="card mt-4 border-warning" id="qr-card" style="display:none;">
                <div class="card-header bg-warning text-dark d-flex justify-content-between align-items-center">
                    <span class="fw-bold"><i class="bi bi-exclamation-triangle-fill me-2"></i>需要扫码登录</span>
                </div>
                <div class="card-body text-center py-5">
                    <div class="mb-4">
                        <i class="bi bi-qr-code-scan text-secondary" style="font-size: 4rem;"></i>
                    </div>
                    <h5 class="card-title mb-3">机器人当前未登录</h5>
                    <p class="card-text text-muted mb-4">为了安全起见，请前往专用登录页面进行扫码操作。</p>
                    <a href="javascript:void(0)" onclick="goToLogin()" class="btn btn-primary btn-lg rounded-pill px-5 shadow-sm">
                        前往扫码登录 <i class="bi bi-arrow-right ms-2"></i>
                    </a>
                </div>
            </div>
        </div>

        <!-- Docker Tab -->
        <div id="tab-docker" style="display:none;">
            <div class="d-flex justify-content-between align-items-center mb-4">
                <h4>容器管理 (Docker)</h4>
                <div>
                    <button class="btn btn-outline-secondary btn-sm me-2" onclick="showDockerSettings()">
                        <i class="bi bi-gear"></i> 设置
                    </button>
                    <button class="btn btn-primary btn-sm" onclick="loadDockerContainers()">
                        <i class="bi bi-arrow-clockwise"></i> 刷新列表
                    </button>
                </div>
            </div>
            
            <div class="card">
                <div class="card-body p-0">
                    <div class="table-responsive">
                        <table class="table table-hover align-middle mb-0">
                            <thead class="table-light">
                                <tr>
                                    <th>ID</th>
                                    <th>名称</th>
                                    <th>镜像</th>
                                    <th>状态</th>
                                    <th>端口</th>
                                    <th class="text-end">操作</th>
                                </tr>
                            </thead>
                            <tbody id="docker-list">
                                <tr><td colspan="6" class="text-center text-muted py-3">加载中...</td></tr>
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
            
            <!-- Settings Modal -->
            <div class="modal fade" id="dockerSettingsModal" tabindex="-1">
                <div class="modal-dialog">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title">Docker 设置</h5>
                            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                        </div>
                        <div class="modal-body">
                            <div class="mb-3">
                                <label class="form-label">Docker 命令前缀</label>
                                <input type="text" class="form-control" id="docker-cmd-prefix" placeholder="例如: docker 或 ssh user@host docker">
                                <div class="form-text">
                                    本地运行请保持默认 <code>docker</code>（需安装 Docker）。<br>
                                    远程管理可使用 <code>ssh user@ip docker</code>（需配置免密登录）。
                                </div>
                            </div>
                        </div>
                        <div class="modal-footer">
                            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
                            <button type="button" class="btn btn-primary" onclick="saveDockerSettings()">保存</button>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Log Modal -->
            <div class="modal fade" id="dockerLogModal" tabindex="-1">
                <div class="modal-dialog modal-lg modal-dialog-scrollable">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title">容器日志</h5>
                            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                        </div>
                        <div class="modal-body bg-dark text-white p-3">
                            <pre id="docker-log-content" style="font-size: 0.8rem; white-space: pre-wrap; margin: 0;"></pre>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Groups Tab -->
        <div id="tab-groups" style="display:none;">
            <div class="d-flex justify-content-between align-items-center mb-4">
                <h4>群组管理</h4>
                <button class="btn btn-primary btn-sm" onclick="updateGroups()"><i class="bi bi-arrow-clockwise"></i> 刷新列表</button>
            </div>
            <div class="card">
                <div class="card-body p-0">
                    <div class="table-responsive">
                        <table class="table table-hover mb-0">
                            <thead class="table-light">
                                <tr>
                                    <th class="sortable" onclick="sortGroups('name')">群名称 <i class="bi bi-arrow-down-up small text-muted"></i></th>
                                    <th class="sortable" onclick="sortGroups('owner_uid')">群主UID <i class="bi bi-arrow-down-up small text-muted"></i></th>
                                    <th class="sortable" onclick="sortGroups('member_count')" style="width: 100px;">成员数 <i class="bi bi-arrow-down-up small text-muted"></i></th>
                                    <th>操作</th>
                                </tr>
                            </thead>
                            <tbody id="group-list">
                                <!-- Group Items -->
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>

        <!-- Message Test Tab -->
        <div id="tab-msgtest" style="display:none;">
            <h4 class="mb-4">消息发送测试</h4>
            
            <div class="row">
                <div class="col-md-8">
                    <div class="card">
                        <div class="card-header">发送测试消息</div>
                        <div class="card-body">
                            <form id="msg-test-form" onsubmit="return false;">
                                <div class="mb-3">
                                    <label class="form-label">目标用户/群组 UID (UserName)</label>
                                    <div class="input-group mb-2">
                                        <select class="form-select" id="mt-group-select" onchange="onTestGroupSelectChange()">
                                            <option value="">-- 选择群组 (或手动输入 UID) --</option>
                                        </select>
                                        <button class="btn btn-outline-secondary" type="button" onclick="loadTestGroups()" title="刷新群组列表"><i class="bi bi-arrow-clockwise"></i></button>
                                    </div>
                                    <div class="input-group">
                                        <input type="text" class="form-control" id="mt-target" placeholder="@xxxx... or @@xxxx...">
                                        <button class="btn btn-outline-secondary" type="button" onclick="pasteTargetUid()">粘贴选中群组UID</button>
                                    </div>
                                    <div class="form-text">请选择群组自动填充，或手动输入 UID。</div>
                                </div>
                                
                                <div class="mb-3">
                                    <label class="form-label">消息类型</label>
                                    <select class="form-select" id="mt-type" onchange="updateMsgForm()">
                                        <option value="text">文本 (Text / At)</option>
                                        <option value="image">图片 (Image)</option>
                                        <option value="file">文件 (File)</option>
                                        <option value="voice">语音 (Voice - 实验性)</option>
                                        <option value="video">视频 (Video - 实验性)</option>
                                        <option value="music">音乐卡片 (Music)</option>
                                        <option value="share">链接卡片 (Share/Link)</option>
                                        <option value="kick">踢出群成员 (Kick)</option>
                                        <option value="tickle">拍一拍 (Tickle)</option>
                                        <option value="mod_group_name">修改群名 (Modify Name)</option>
                                        <option value="mod_group_remark">修改群备注 (Modify Remark)</option>
                                        <option value="quit_group">退群 (Quit Group)</option>
                                    </select>
                                </div>
                                
                                <!-- Text Input -->
                                <div id="mt-group-text" class="mb-3">
                                    <label class="form-label">内容</label>
                                    <textarea class="form-control" id="mt-content" rows="3" placeholder="输入文本内容... 如需艾特请直接输入 @昵称"></textarea>
                                </div>
                                
                                <!-- Group Name Input -->
                                <div id="mt-group-name" class="mb-3" style="display:none;">
                                    <label class="form-label">新群名称</label>
                                    <input type="text" class="form-control" id="mt-new-group-name" placeholder="请输入新的群名称">
                                    <div class="form-text text-warning">注意：仅群主可修改群名称。若非群主，此操作可能无效。</div>
                                </div>

                                <!-- Group Remark Input -->
                                <div id="mt-group-remark" class="mb-3" style="display:none;">
                                    <label class="form-label">新群备注</label>
                                    <input type="text" class="form-control" id="mt-new-group-remark" placeholder="请输入新的群备注 (仅自己可见)">
                                </div>

                                <!-- Kick/Tickle Member Input -->
                                <div id="mt-group-kick" class="mb-3" style="display:none;">
                                    <label class="form-label" id="mt-member-label">目标成员</label>
                                    <div class="input-group mb-2">
                                        <select class="form-select" id="mt-member-select" onchange="onMemberSelectChange()">
                                            <option value="">-- 先选择群组 --</option>
                                        </select>
                                        <button class="btn btn-outline-secondary" type="button" onclick="loadGroupMembers()" title="刷新成员列表"><i class="bi bi-arrow-clockwise"></i></button>
                                    </div>
                                    <input type="text" class="form-control" id="mt-member-id" placeholder="成员 UID (@...) - 可手动输入或从上方选择">
                                    <div class="form-text" id="mt-member-help"></div>
                                </div>
                                
                                <!-- File/Image Input -->
                                 <div id="mt-group-file" class="mb-3" style="display:none;">
                                     <label class="form-label">上传本地文件 / 图片</label>
                                     <input class="form-control mb-2" type="file" id="mt-file-upload" onchange="uploadFile()">
                                     <label class="form-label">或输入服务器文件路径</label>
                                     <input type="text" class="form-control" id="mt-filepath" placeholder="例如: D:\images\test.jpg">
                                     <div class="form-text">上传文件会自动填充下方路径。</div>
                                 </div>
                                
                                <!-- Music Input -->
                                <div id="mt-group-music" style="display:none;">
                                    <div class="mb-3">
                                        <label class="form-label">标题</label>
                                        <input type="text" class="form-control" id="mt-music-title" value="征服">
                                    </div>
                                    <div class="mb-3">
                                        <label class="form-label">描述</label>
                                        <input type="text" class="form-control" id="mt-music-desc" value="那英">
                                    </div>
                                    <div class="mb-3">
                                        <label class="form-label">跳转链接 (URL)</label>
                                        <input type="text" class="form-control" id="mt-music-url" value="https://i.y.qq.com/v8/playsong.html?hosteuin=7K6PoiSFoKn*&songid=179923&songmid=&type=0&platform=1&appsongtype=1&_wv=1&source=qq&appshare=iphone&media_mid=004LBt3k1d1J9m&ADTAG=qfshare">
                                    </div>
                                    <div class="mb-3">
                                        <label class="form-label" id="mt-music-data-label">音乐链接 (Data URL)</label>
                                        <input type="text" class="form-control" id="mt-music-data-url" placeholder="http://.../song.mp3" value="http://c6.y.qq.com/rsc/fcgi-bin/fcg_pyq_play.fcg?songid=0&songmid=003Rksq51qnUks&songtype=1&fromtag=50&uin=51437810&code=f606f">
                                    </div>
                                </div>
                                
                                <div class="d-grid">
                                    <button class="btn btn-primary" onclick="submitTestMsg()">发送消息</button>
                                </div>
                            </form>
                        </div>
                    </div>
                </div>
                
                <div class="col-md-4">
                    <div class="card">
                        <div class="card-header">说明</div>
                        <div class="card-body">
                            <ul>
                                <li><strong>文本</strong>: 支持普通文本。若要艾特某人，请直接在文本中包含 <code>@昵称</code> (需确保昵称完全匹配)。</li>
                                <li><strong>图片/文件</strong>: 需要输入服务器上的绝对文件路径。</li>
                                <li><strong>语音</strong>: 目前仅支持以文件形式发送音频文件，可能不会自动转为语音消息条。</li>
                                <li><strong>音乐</strong>: 发送类似分享卡片的消息。</li>
                            </ul>
                            <hr>
                            <div id="mt-result" class="alert alert-secondary" role="alert" style="display:none;"></div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Logs Tab -->
        <div id="tab-logs" style="display:none;">
             <div class="d-flex justify-content-between align-items-center mb-4">
                <div class="d-flex align-items-center">
                    <h4 class="mb-0 me-3">系统日志</h4>
                    <select class="form-select form-select-sm" id="log-filter" style="width: auto;" onchange="fetchLogs(true)">
                        <option value="all">全部日志</option>
                        <option value="system">系统 (System)</option>
                        <option value="bot">机器人 (Bot)</option>
                        <option value="plugin">插件 (Plugin)</option>
                        <option value="webui">网页 (WebUI)</option>
                    </select>
                </div>
                <div>
                    <span class="form-check form-switch d-inline-block me-2">
                        <input class="form-check-input" type="checkbox" id="auto-scroll" checked>
                        <label class="form-check-label" for="auto-scroll">自动滚动</label>
                    </span>
                    <button class="btn btn-outline-secondary btn-sm" onclick="clearLogs()">清空</button>
                </div>
            </div>
            <div class="log-box" id="log-container">
                <!-- Logs will appear here -->
            </div>
        </div>

        <!-- Manual Tab -->
        <div id="tab-manual" style="display:none;">
            <div class="d-flex justify-content-between align-items-center mb-4">
                <h4 class="mb-0">服务端使用说明</h4>
                <button class="btn btn-sm btn-outline-primary" onclick="loadManual()">
                    <i class="bi bi-arrow-clockwise"></i> 刷新
                </button>
            </div>
            <div class="card">
                <div class="card-body">
                    <div id="manual-content" style="white-space: pre-wrap; font-family: monospace; background-color: var(--bg-body); padding: 15px; border-radius: 5px; color: var(--text-main); max-height: 80vh; overflow-y: auto;">
                        <div class="text-center text-muted py-5">
                            <div class="spinner-border text-primary" role="status"></div>
                            <div class="mt-2">正在加载使用说明...</div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        
        <!-- Settings Tab -->
        <div id="tab-settings" style="display:none;">
            <h4 class="mb-4">设置与调试</h4>
            
            <div class="row">
                <div class="col-md-6">
                     <div class="card mb-4">
                        <div class="card-header"><i class="bi bi-robot"></i> 机器人管理</div>
                        <div class="card-body">
                            <p class="text-muted small">添加或管理当前的机器人实例。</p>
                            <div class="d-grid">
                                <button class="btn btn-outline-primary" onclick="showAddBotModal()">
                                    <i class="bi bi-plus-lg"></i> 添加新机器人
                                </button>
                            </div>
                            <div class="mt-3 small text-muted">
                                * 当前架构仅支持单实例运行。如需多实例，请启动多个进程并配置不同端口。
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="col-md-6">
                    <div class="card mb-4">
                        <div class="card-header"><i class="bi bi-code-slash"></i> 接口调试</div>
                        <div class="card-body">
                            <div class="mb-3">
                                <label class="form-label">API 测试路径</label>
                                <div class="input-group">
                                    <span class="input-group-text">/api/</span>
                                    <input type="text" class="form-control" placeholder="status" id="debug-api-path">
                                    <button class="btn btn-secondary" type="button" onclick="testApi()">GET</button>
                                </div>
                            </div>
                            <div class="bg-dark text-light p-2 rounded small" id="api-result" style="height: 100px; overflow-y: auto; font-family: monospace;">
                                // API 响应结果将显示在这里
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            
            <!-- Network Configuration -->
            <div class="card mb-4">
                <div class="card-header"><i class="bi bi-hdd-network"></i> 网络配置</div>
                <div class="card-body">
                    <ul class="nav nav-tabs" id="network-config-tabs">
                        <li class="nav-item"><a class="nav-link" data-bs-toggle="tab" href="#net-all">全部</a></li>
                        <li class="nav-item"><a class="nav-link" data-bs-toggle="tab" href="#net-http-server">HTTP服务器</a></li>
                        <li class="nav-item"><a class="nav-link" data-bs-toggle="tab" href="#net-http-client">HTTP客户端</a></li>
                        <li class="nav-item"><a class="nav-link" data-bs-toggle="tab" href="#net-http-sse">HTTP SSE服务器</a></li>
                        <li class="nav-item"><a class="nav-link active" data-bs-toggle="tab" href="#net-ws-server">Websocket服务器</a></li>
                        <li class="nav-item"><a class="nav-link" data-bs-toggle="tab" href="#net-ws-client">Websocket客户端</a></li>
                    </ul>
                    <div class="tab-content pt-3">
                        <div class="tab-pane fade" id="net-all"><p class="text-muted">暂无内容</p></div>
                        <div class="tab-pane fade" id="net-http-server"><p class="text-muted">暂无内容</p></div>
                        <div class="tab-pane fade" id="net-http-client"><p class="text-muted">暂无内容</p></div>
                        <div class="tab-pane fade" id="net-http-sse"><p class="text-muted">暂无内容</p></div>
                        
                        <div class="tab-pane fade show active" id="net-ws-server">
                            <form id="form-ws-server">
                                <div class="row mb-3">
                                    <label class="col-sm-3 col-form-label">Name</label>
                                    <div class="col-sm-9">
                                        <input type="text" class="form-control" id="ws-name" value="test">
                                    </div>
                                </div>
                                <div class="row mb-3">
                                    <label class="col-sm-3 col-form-label">Host</label>
                                    <div class="col-sm-9">
                                        <input type="text" class="form-control" id="ws-host" value="0.0.0.0">
                                    </div>
                                </div>
                                <div class="row mb-3">
                                    <label class="col-sm-3 col-form-label">Port</label>
                                    <div class="col-sm-9">
                                        <input type="number" class="form-control" id="ws-port" value="3001">
                                    </div>
                                </div>
                                <div class="row mb-3">
                                    <label class="col-sm-3 col-form-label">Heartbeat Interval (ms)</label>
                                    <div class="col-sm-9">
                                        <input type="number" class="form-control" id="ws-heartbeat" value="30000">
                                    </div>
                                </div>
                                <div class="row mb-3">
                                    <label class="col-sm-3 col-form-label">Message Format</label>
                                    <div class="col-sm-9">
                                        <select class="form-select" id="ws-format">
                                            <option value="string">string</option>
                                            <option value="json">json</option>
                                        </select>
                                    </div>
                                </div>
                                <div class="row mb-3">
                                    <label class="col-sm-3 col-form-label">Report Self Message</label>
                                    <div class="col-sm-9">
                                        <div class="form-check form-switch pt-2">
                                            <input class="form-check-input" type="checkbox" id="ws-report-self" checked>
                                        </div>
                                    </div>
                                </div>
                                <div class="row mb-3">
                                    <label class="col-sm-3 col-form-label">Force Push Event</label>
                                    <div class="col-sm-9">
                                        <div class="form-check form-switch pt-2">
                                            <input class="form-check-input" type="checkbox" id="ws-force-push" checked>
                                        </div>
                                    </div>
                                </div>
                            </form>
                        </div>
                        
                        <div class="tab-pane fade" id="net-ws-client"><p class="text-muted">暂无内容</p></div>
                    </div>
                    <div class="mt-3 text-end">
                         <div class="alert alert-warning d-inline-block me-2 mb-0 py-1 small" id="config-status" style="display:none;"></div>
                        <button class="btn btn-primary" onclick="saveNetworkConfig()">保存配置</button>
                    </div>
                </div>
            </div>
            
             <div class="card">
                <div class="card-header"><i class="bi bi-sliders"></i> 系统参数</div>
                <div class="card-body">
                    <form>
                        <div class="row mb-3">
                            <label class="col-sm-3 col-form-label">日志级别</label>
                            <div class="col-sm-9">
                                <select class="form-select" disabled>
                                    <option>INFO (默认)</option>
                                    <option>DEBUG</option>
                                    <option>WARNING</option>
                                </select>
                            </div>
                        </div>
                         <div class="row mb-3">
                            <label class="col-sm-3 col-form-label">自动回复</label>
                            <div class="col-sm-9">
                                <div class="form-check form-switch pt-2">
                                    <input class="form-check-input" type="checkbox" id="auto-reply-switch" disabled>
                                    <label class="form-check-label" for="auto-reply-switch">启用全局自动回复 (开发中)</label>
                                </div>
                            </div>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    </div>

    <!-- Add Bot Modal -->
    <div class="modal fade" id="addBotModal" tabindex="-1">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">添加新机器人</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <div class="mb-3">
                        <label class="form-label">Robot QQ/Self ID</label>
                        <input type="number" class="form-control" id="new-bot-id" placeholder="留空则自动生成">
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
                    <button type="button" class="btn btn-primary" onclick="submitAddBot()">添加并启动</button>
                </div>
            </div>
        </div>
    </div>

    <!-- Group Members Modal -->
    <div class="modal fade" id="groupMembersModal" tabindex="-1">
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">群成员列表</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                     <div class="table-responsive" style="max-height: 500px; overflow-y: auto;">
                        <table class="table table-sm table-hover">
                            <thead>
                                <tr>
                                    <th>头像</th>
                                    <th>昵称</th>
                                    <th>群名片</th>
                                    <th>UID (Masked)</th>
                                </tr>
                            </thead>
                            <tbody id="group-members-list">
                                <!-- Members -->
                            </tbody>
                        </table>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">关闭</button>
                </div>
            </div>
        </div>
    </div>

    <script src="https://cdn.bootcdn.net/ajax/libs/twitter-bootstrap/5.1.3/js/bootstrap.bundle.min.js"></script>
    <script>
        let currentTab = 'dashboard';
        let logInterval = null;
        let currentBotId = null;
        
        // Modal instance
        let addBotModal;
        
        function toggleSidebar() {
            document.querySelector('.sidebar').classList.toggle('show');
            document.querySelector('.sidebar-overlay').classList.toggle('show');
        }

        function showAddBotModal() {
            if (typeof bootstrap === 'undefined') {
                alert('Bootstrap 资源加载失败，请检查网络或刷新重试。');
                return;
            }
            if (!addBotModal) {
                addBotModal = new bootstrap.Modal(document.getElementById('addBotModal'));
            }
            addBotModal.show();
        }

        function submitAddBot() {
            const idInput = document.getElementById('new-bot-id').value;
            const payload = idInput ? { self_id: parseInt(idInput) } : {};
            
            fetch('/api/add_bot', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(payload)
            })
            .then(r => r.json())
            .then(res => {
                if (res.status === 'ok') {
                    alert('添加成功');
                    addBotModal.hide();
                    updateBotList();
                } else {
                    alert('添加失败: ' + res.msg);
                }
            });
        }
        
        function testApi() {
            const path = document.getElementById('debug-api-path').value || 'status';
            const fullPath = '/api/' + path.replace(/^\//, '') + (path.includes('?') ? '&' : '?') + 'bot_id=' + (currentBotId || '');
            
            document.getElementById('api-result').textContent = 'Loading...';
            
            fetch(fullPath)
                .then(r => r.text())
                .then(text => {
                    try {
                        const json = JSON.parse(text);
                        document.getElementById('api-result').textContent = JSON.stringify(json, null, 2);
                    } catch (e) {
                        document.getElementById('api-result').textContent = text;
                    }
                })
                .catch(err => {
                     document.getElementById('api-result').textContent = 'Error: ' + err;
                });
        }

        function updateBotList() {
            fetch('/api/bots')
                .then(r => r.json())
                .then(bots => {
                    const selector = document.getElementById('bot-selector');
                    const oldVal = selector.value || currentBotId;
                    
                    selector.innerHTML = bots.map(b => `
                        <option value="${b.self_id}">
                            ${b.nickname} (${b.self_id}) ${b.is_alive ? '🟢' : '🔴'}
                        </option>
                    `).join('');
                    
                    if (bots.length > 0) {
                        if (oldVal && bots.find(b => b.self_id === oldVal)) {
                            selector.value = oldVal;
                            currentBotId = oldVal;
                        } else {
                            selector.value = bots[0].self_id;
                            currentBotId = bots[0].self_id;
                        }
                    }
                    updateSystemStats();
                });
        }

        function switchBot(botId) {
            currentBotId = botId;
            updateSystemStats();
            updateGroups();
        }

        function loadManual() {
            const contentDiv = document.getElementById('manual-content');
            contentDiv.innerHTML = '<div class="text-center text-muted py-5"><div class="spinner-border text-primary" role="status"></div><div class="mt-2">正在加载使用说明...</div></div>';
            
            fetch('/api/manual')
                .then(r => r.json())
                .then(res => {
                    if (res.status === 'ok') {
                        contentDiv.innerHTML = marked.parse(res.content);
                    } else {
                        contentDiv.innerHTML = `<div class="alert alert-danger">加载失败: ${res.error}</div>`;
                    }
                })
                .catch(err => {
                    contentDiv.innerHTML = `<div class="alert alert-danger">请求错误: ${err}</div>`;
                });
        }

        function showTab(tabId) {
            currentTab = tabId;
            document.querySelectorAll('.main-content > div[id^="tab-"]').forEach(el => el.style.display = 'none');
            document.getElementById('tab-' + tabId).style.display = 'block';
            
            document.querySelectorAll('.sidebar .nav-link').forEach(el => el.classList.remove('active'));
            document.querySelector(`.sidebar .nav-link[href="#${tabId}"]`).classList.add('active');
            
            // Auto close sidebar on mobile
            if (window.innerWidth <= 768) {
                document.querySelector('.sidebar').classList.remove('show');
                document.querySelector('.sidebar-overlay').classList.remove('show');
            }
            
            if (tabId === 'logs') {
                startLogStream();
            } else {
                stopLogStream();
            }

            if (tabId === 'docker') {
                loadDockerContainers();
            }

            if (tabId === 'manual') {
                loadManual();
            }
            
            if (tabId === 'groups') {
                updateGroups();
            }
            
            if (tabId === 'msgtest') {
                // 如果下拉框没数据，自动加载一次
                const select = document.getElementById('mt-group-select');
                if (select && select.options.length <= 1) {
                    loadTestGroups();
                }
            }
        }

        function formatBytes(bytes, decimals = 2) {
            if (bytes === 0) return '0 Bytes';
            const k = 1024;
            const dm = decimals < 0 ? 0 : decimals;
            const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
        }

        function maskUid(uid) {
            if (!uid || uid.length < 10) return uid;
            return uid.substring(0, 4) + '***' + uid.substring(uid.length - 4);
        }

        function updateSystemStats() {
            if (!currentBotId) return;
            fetch('/api/system_stats?bot_id=' + currentBotId)
                .then(r => r.json())
                .then(data => {
                    document.getElementById('cpu-usage').textContent = data.cpu_percent + '%';
                    document.getElementById('mem-usage').textContent = formatBytes(data.memory_used);
                    document.getElementById('uptime').textContent = data.uptime;
                    
                    // Hardware info updates
                    if(document.getElementById('cpu-cores')) document.getElementById('cpu-cores').textContent = data.cpu_cores;
                    if(document.getElementById('cpu-cores-mini')) document.getElementById('cpu-cores-mini').textContent = data.cpu_cores;
                    if(document.getElementById('mem-sys-percent')) document.getElementById('mem-sys-percent').textContent = data.memory_system_percent;
                    
                    if(document.getElementById('cpu-model')) document.getElementById('cpu-model').textContent = data.cpu_model;
                    if(document.getElementById('cpu-freq')) document.getElementById('cpu-freq').textContent = data.cpu_freq;
                    if(document.getElementById('mem-total-detail')) document.getElementById('mem-total-detail').textContent = formatBytes(data.memory_system_total);
                    if(document.getElementById('mem-used-detail')) document.getElementById('mem-used-detail').textContent = formatBytes(data.memory_system_used);

                    document.getElementById('os-ver').textContent = data.system_info.os_version;
                    
                    // Tooltips
                    document.getElementById('cpu-usage').title = `核心数: ${data.cpu_cores}`;
                    document.getElementById('mem-usage').title = `进程占用: ${formatBytes(data.memory_used)}\n系统内存: ${formatBytes(data.memory_system_total)} (使用率 ${data.memory_system_percent}%)`;

                    document.getElementById('bot-nickname').textContent = data.bot_info.nickname;
                    document.getElementById('bot-uid').textContent = maskUid(data.bot_info.uid);
                    document.getElementById('py-ver').textContent = data.system_info.python_version;
                    document.getElementById('os-ver').textContent = data.system_info.os_version;
                    
                    document.getElementById('group-count-dash').textContent = data.group_count;
                    
                    if (data.bot_info.is_alive) {
                        // console.log('Bot is alive, hiding QR card');
                        if(document.getElementById('qr-card')) document.getElementById('qr-card').style.display = 'none';
                        if(document.getElementById('nav-login-link')) document.getElementById('nav-login-link').style.display = 'none';
                    } else {
                        // console.log('Bot is NOT alive, showing QR card');
                        if(document.getElementById('qr-card')) document.getElementById('qr-card').style.display = 'block';
                        if(document.getElementById('nav-login-link')) document.getElementById('nav-login-link').style.display = 'block';
                        
                        // Auto jump to login if needed
                        if (!window.isRedirectingToLogin) {
                             window.isRedirectingToLogin = true;
                             // Delay slightly to allow user to see the status or avoid race conditions
                             setTimeout(() => {
                                 goToLogin();
                             }, 1000);
                        }
                    }
                });
        }

        let groupMembersModal;
        
        // Sorting and Data State
        let currentGroupsData = [];
        let sortField = '';
        let sortAsc = true;

        function initDarkMode() {
            const isDark = localStorage.getItem('darkMode') === 'true';
            if (isDark) {
                document.body.classList.add('dark-mode');
            }
            updateDarkModeUI(isDark);
        }
        
        function toggleDarkMode() {
            document.body.classList.toggle('dark-mode');
            const isDark = document.body.classList.contains('dark-mode');
            localStorage.setItem('darkMode', isDark);
            updateDarkModeUI(isDark);
        }

        function updateDarkModeUI(isDark) {
            const textSpan = document.getElementById('dark-mode-text');
            const icon = document.querySelector('button[onclick="toggleDarkMode()"] i');
            
            if (isDark) {
                if(textSpan) textSpan.textContent = '黑夜模式';
                if(icon) icon.className = 'bi bi-moon-stars-fill';
            } else {
                if(textSpan) textSpan.textContent = '普通模式';
                if(icon) icon.className = 'bi bi-sun';
            }
        }
        
        function sortGroups(field) {
            if (sortField === field) {
                sortAsc = !sortAsc;
            } else {
                sortField = field;
                sortAsc = true;
            }
            renderGroups();
        }
        
        function renderGroups() {
            let data = [...currentGroupsData];
            if (sortField) {
                data.sort((a, b) => {
                    let va = a[sortField];
                    let vb = b[sortField];
                    
                    if (typeof va === 'string') va = va.toLowerCase();
                    if (typeof vb === 'string') vb = vb.toLowerCase();
                    
                    if (va < vb) return sortAsc ? -1 : 1;
                    if (va > vb) return sortAsc ? 1 : -1;
                    return 0;
                });
            }
            
            const tbody = document.getElementById('group-list');
            tbody.innerHTML = data.map(g => {
                const displayName = g.name && g.name.length > 10 ? g.name.substring(0, 10) + '...' : (g.name || '未知群组');
                return `
                <tr>
                    <td onclick="showGroupMembers('${g.gid}')" style="cursor:pointer; color: #0d6efd;" title="${g.name}"><strong>${displayName}</strong></td>
                    <td><span class="badge bg-light text-dark border">${maskUid(g.owner_uid)}</span></td>
                    <td>${g.member_count}</td>
                    <td>
                        <button class="btn btn-sm btn-outline-secondary" onclick="showGroupMembers('${g.gid}')"><i class="bi bi-people"></i></button>
                        <button class="btn btn-sm btn-outline-primary" onclick="sendMsg('${g.gid}')">发送</button>
                    </td>
                </tr>
            `}).join('');
            document.getElementById('group-count-dash').textContent = data.length;
        }

        function showGroupMembers(gid) {
             if (!currentBotId) return;
             
             if (typeof bootstrap === 'undefined') {
                 alert('Bootstrap 资源加载失败，请检查网络或刷新重试。');
                 return;
             }
             
             if (!groupMembersModal) {
                groupMembersModal = new bootstrap.Modal(document.getElementById('groupMembersModal'));
             }
             
             document.getElementById('group-members-list').innerHTML = '<tr><td colspan="4" class="text-center">加载中...</td></tr>';
             groupMembersModal.show();
             
             fetch(`/api/group_members?bot_id=${currentBotId}&gid=${gid}`)
                .then(r => r.json())
                .then(data => {
                    const tbody = document.getElementById('group-members-list');
                    if (data.length === 0) {
                        tbody.innerHTML = '<tr><td colspan="4" class="text-center">无成员信息或获取失败</td></tr>';
                        return;
                    }
                    
                    tbody.innerHTML = data.map(m => {
                        let imgHtml = '';
                        if (m.head_img_url) {
                            const proxyUrl = `/api/proxy_image?bot_id=${currentBotId}&url=${encodeURIComponent(m.head_img_url)}`;
                            imgHtml = `<img src="${proxyUrl}" width="30" height="30" class="rounded-circle" onerror="this.onerror=null;this.parentNode.innerHTML='<i class=\\'bi bi-person-circle fs-4\\'></i>'">`;
                        } else {
                            imgHtml = '<i class="bi bi-person-circle fs-4"></i>';
                        }
                        
                        return `
                        <tr>
                            <td>${imgHtml}</td>
                            <td>${m.nick || m.display || '未命名'}</td>
                            <td>${m.display}</td>
                            <td>${maskUid(m.uid)}</td>
                        </tr>
                        `;
                    }).join('');
                });
        }

        function updateGroups() {
            if (!currentBotId) return;
            fetch('/api/groups?bot_id=' + currentBotId)
                .then(r => r.json())
                .then(data => {
                    currentGroupsData = data;
                    renderGroups();
                });
        }
        
        function sendMsg(gid) {
            showTab('msgtest');
            document.getElementById('mt-target').value = gid;
            document.getElementById('mt-result').style.display = 'none';
        }
        
        function updateMsgForm() {
            const type = document.getElementById('mt-type').value;
            
            // Hide all
            document.getElementById('mt-group-text').style.display = 'none';
            document.getElementById('mt-group-file').style.display = 'none';
            document.getElementById('mt-group-music').style.display = 'none';
            document.getElementById('mt-group-kick').style.display = 'none';
            document.getElementById('mt-group-name').style.display = 'none';
            document.getElementById('mt-group-remark').style.display = 'none';
            
            if (type === 'text') {
                document.getElementById('mt-group-text').style.display = 'block';
            } else if (type === 'image' || type === 'file' || type === 'voice' || type === 'video') {
                document.getElementById('mt-group-file').style.display = 'block';
            } else if (type === 'music') {
                document.getElementById('mt-group-music').style.display = 'block';
                document.getElementById('mt-music-data-label').textContent = '音乐链接 (Data URL)';
                document.getElementById('mt-music-data-url').placeholder = 'http://.../song.mp3';
            } else if (type === 'share') {
                document.getElementById('mt-group-music').style.display = 'block';
                document.getElementById('mt-music-data-label').textContent = '缩略图链接 (Image URL)';
                document.getElementById('mt-music-data-url').placeholder = 'http://.../thumb.jpg';
            } else if (type === 'kick') {
                document.getElementById('mt-group-kick').style.display = 'block';
                document.getElementById('mt-member-label').textContent = '被踢成员 UID (Member to Kick)';
                document.getElementById('mt-member-help').textContent = '注意：机器人必须是群主才能踢人。';
            } else if (type === 'tickle') {
                document.getElementById('mt-group-kick').style.display = 'block';
                document.getElementById('mt-member-label').textContent = '被拍成员 UID (Member to Tickle) [可选]';
                document.getElementById('mt-member-help').textContent = '若为空则拍“你”或拍自己。Web版仅发送模拟文本。';
            } else if (type === 'mod_group_name') {
                document.getElementById('mt-group-name').style.display = 'block';
            } else if (type === 'mod_group_remark') {
                document.getElementById('mt-group-remark').style.display = 'block';
            } else if (type === 'quit_group') {
                // No extra inputs needed
            }
        }
        
        function revokeMsg(msgId, localId, toUser, btnElement) {
            if (!confirm('确认撤回？')) return;
            
            btnElement.disabled = true;
            btnElement.textContent = '撤回中...';
            
            fetch('/api/revoke_test_msg', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                    bot_id: currentBotId,
                    msg_id: msgId,
                    local_id: localId,
                    to_user: toUser
                })
            })
            .then(r => r.json())
            .then(res => {
                if (res.status === 'ok') {
                    btnElement.textContent = '已撤回';
                    btnElement.className = 'btn btn-secondary btn-sm ms-2';
                } else {
                    btnElement.disabled = false;
                    btnElement.textContent = '撤回失败: ' + res.error;
                    setTimeout(() => { btnElement.textContent = '撤回这条消息'; }, 2000);
                }
            })
            .catch(err => {
                btnElement.disabled = false;
                btnElement.textContent = '错误';
                alert('撤回请求错误: ' + err);
            });
        }

        function submitTestMsg() {
            if (!currentBotId) {
                alert('请先选择机器人');
                return;
            }
            
            const target = document.getElementById('mt-target').value;
            const type = document.getElementById('mt-type').value;
            
            if (!target) {
                alert('请输入目标 UID');
                return;
            }
            
            const payload = {
                bot_id: currentBotId,
                target_id: target,
                type: type
            };
            
            if (type === 'text') {
                payload.content = document.getElementById('mt-content').value;
            } else if (type === 'image' || type === 'file' || type === 'voice' || type === 'video') {
                payload.file_path = document.getElementById('mt-filepath').value;
            } else if (type === 'music') {
                payload.title = document.getElementById('mt-music-title').value;
                payload.desc = document.getElementById('mt-music-desc').value;
                payload.url = document.getElementById('mt-music-url').value;
                payload.music_url = document.getElementById('mt-music-data-url').value;
            } else if (type === 'share') {
                payload.title = document.getElementById('mt-music-title').value;
                payload.desc = document.getElementById('mt-music-desc').value;
                payload.url = document.getElementById('mt-music-url').value;
                payload.image_url = document.getElementById('mt-music-data-url').value;
            } else if (type === 'kick' || type === 'tickle') {
                payload.member_id = document.getElementById('mt-member-id').value;
            } else if (type === 'mod_group_name') {
                payload.group_name = document.getElementById('mt-new-group-name').value;
            } else if (type === 'mod_group_remark') {
                payload.group_remark = document.getElementById('mt-new-group-remark').value;
            } else if (type === 'quit_group') {
                if (!confirm('确定要退出当前群聊吗？此操作不可逆！')) {
                    return;
                }
            }
            
            const resultDiv = document.getElementById('mt-result');
            resultDiv.style.display = 'block';
            resultDiv.className = 'alert alert-info';
            resultDiv.textContent = '发送中...';
            
            fetch('/api/send_test_msg', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(payload)
            })
            .then(r => r.json())
            .then(res => {
                if (res.status === 'ok') {
                    resultDiv.className = 'alert alert-success';
                    resultDiv.innerHTML = '发送成功！';
                    
                    if (type === 'mod_group_name') {
                        resultDiv.innerHTML += ' <small class="d-block mt-1">注意：若群名未改变，可能是权限不足或服务器延迟，请查看后台日志。</small>';
                    }

                    if (res.msg_id) {
                        const btn = document.createElement('button');
                        btn.className = 'btn btn-danger btn-sm ms-2';
                        btn.textContent = '撤回这条消息';
                        btn.onclick = function() {
                            revokeMsg(res.msg_id, res.local_id, res.to_user, btn);
                        };
                        resultDiv.appendChild(btn);
                    }
                } else {
                    resultDiv.className = 'alert alert-danger';
                    resultDiv.textContent = '发送失败: ' + res.error;
                }
            })
            .catch(err => {
                resultDiv.className = 'alert alert-danger';
                resultDiv.textContent = '请求错误: ' + err;
            });
        }
        
        function pasteTargetUid() {
            navigator.clipboard.readText().then(text => {
                if (text) document.getElementById('mt-target').value = text;
            }).catch(err => {
                alert('无法读取剪贴板，请手动粘贴');
            });
        }
        
        function loadTestGroups() {
            if (!currentBotId) return;
            const select = document.getElementById('mt-group-select');
            select.innerHTML = '<option value="">加载中...</option>';
            
            fetch('/api/groups?bot_id=' + currentBotId)
                .then(r => r.json())
                .then(data => {
                    select.innerHTML = '<option value="">-- 选择群组 (或手动输入 UID) --</option>';
                    
                    // 排序：包含“测试”的在前，然后按人数降序
                    data.sort((a, b) => {
                        const aTest = (a.name || '').includes('测试');
                        const bTest = (b.name || '').includes('测试');
                        if (aTest && !bTest) return -1;
                        if (!aTest && bTest) return 1;
                        return b.member_count - a.member_count;
                    });
                    
                    let foundTestGroup = false;
                    
                    data.forEach(g => {
                        const isTest = (g.name || '').includes('测试');
                        const option = document.createElement('option');
                        option.value = g.gid;
                        option.textContent = (isTest ? '🧪 ' : '') + (g.name || '未知群组') + ` (${g.member_count}人)`;
                        if (isTest) {
                            option.style.fontWeight = 'bold';
                            option.style.color = 'var(--primary-color)';
                            if (!foundTestGroup) foundTestGroup = true; 
                        }
                        
                        // Default select specific group
                        if ((g.name || '').includes('【早喵】测试群')) {
                            option.selected = true;
                            document.getElementById('mt-target').value = g.gid;
                        }

                        select.appendChild(option);
                    });
                    
                    // 如果有包含“测试”的群，可以考虑提示或者...
                    // 这里不做自动选中，以免误操作，但排在最前面已经很方便了
                })
                .catch(err => {
                    select.innerHTML = '<option value="">加载失败</option>';
                    console.error(err);
                });
        }
        
        function onTestGroupSelectChange() {
            const select = document.getElementById('mt-group-select');
            const targetInput = document.getElementById('mt-target');
            if (select.value) {
                targetInput.value = select.value;
                loadGroupMembers(); // Load members for the selected group
            }
        }

        function loadGroupMembers() {
            if (!currentBotId) return;
            const groupSelect = document.getElementById('mt-group-select');
            const gid = groupSelect.value;
            const memberSelect = document.getElementById('mt-member-select');
            
            if (!gid) {
                memberSelect.innerHTML = '<option value="">-- 先选择群组 --</option>';
                return;
            }
            
            memberSelect.innerHTML = '<option value="">加载中...</option>';
            
            fetch('/api/group_members?bot_id=' + currentBotId + '&gid=' + encodeURIComponent(gid))
                .then(r => r.json())
                .then(data => {
                    memberSelect.innerHTML = '<option value="">-- 选择成员 --</option>';
                    data.forEach(m => {
                        const option = document.createElement('option');
                        option.value = m.uid;
                        option.textContent = m.name; // + ' (' + m.uid.substring(0, 6) + '...)';
                        memberSelect.appendChild(option);
                    });
                })
                .catch(err => {
                    memberSelect.innerHTML = '<option value="">加载失败</option>';
                    console.error(err);
                });
        }

        function onMemberSelectChange() {
            const select = document.getElementById('mt-member-select');
            const input = document.getElementById('mt-member-id');
            if (select.value) {
                input.value = select.value;
            }
        }
        
        function uploadFile() {
            const fileInput = document.getElementById('mt-file-upload');
            const file = fileInput.files[0];
            if (!file) return;
            
            const formData = new FormData();
            formData.append('file', file);
            
            const pathInput = document.getElementById('mt-filepath');
            const originalPlaceholder = pathInput.placeholder;
            pathInput.value = '';
            pathInput.placeholder = '上传中...';
            pathInput.disabled = true;
            fileInput.disabled = true;
            
            fetch('/api/upload_temp', {
                method: 'POST',
                body: formData
            })
            .then(r => r.json())
            .then(res => {
                pathInput.disabled = false;
                fileInput.disabled = false;
                pathInput.placeholder = originalPlaceholder;
                
                if (res.status === 'ok') {
                    pathInput.value = res.path;
                } else {
                    alert('上传失败: ' + res.error);
                    fileInput.value = ''; // clear selection
                }
            })
            .catch(err => {
                pathInput.disabled = false;
                fileInput.disabled = false;
                pathInput.placeholder = originalPlaceholder;
                alert('上传出错: ' + err);
                fileInput.value = '';
            });
        }
 
         function startLogStream() {
            if (logInterval) clearInterval(logInterval);
            fetchLogs();
            logInterval = setInterval(fetchLogs, 2000);
        }
        
        function stopLogStream() {
            if (logInterval) clearInterval(logInterval);
            logInterval = null;
        }
        
        function fetchLogs(force) {
            const autoScroll = document.getElementById('auto-scroll').checked;
            
            // 如果未开启自动滚动且非强制刷新，则暂停刷新
            if (!autoScroll && force !== true) {
                return;
            }
            
            // 如果用户正在选择文本，也暂停刷新，防止选区丢失
            const selection = window.getSelection();
            if (selection && selection.toString().length > 0) {
                return;
            }

            fetch('/api/logs')
                .then(r => r.json())
                .then(data => {
                    // Check again inside callback
                    if (!document.getElementById('auto-scroll').checked && force !== true) return;
                    if (window.getSelection().toString().length > 0) return;
                    
                    const container = document.getElementById('log-container');
                    const wasBottom = container.scrollHeight - container.clientHeight <= container.scrollTop + 50;
                    
                    // Filter Logic
                    const filter = document.getElementById('log-filter').value;
                    let filteredData = data;
                    if (filter && filter !== 'all') {
                        filteredData = data.filter(line => {
                             if (filter === 'system') return line.includes('[BotManager]') || line.includes('[main]') || line.includes('[OneBot Gateway]') || line.includes('[INFO]');
                             if (filter === 'bot') return line.includes('[WXWork]') || line.includes('[DingTalk]') || line.includes('[Feishu]') || line.includes('[Telegram]') || line.includes('[WeChat]');
                             if (filter === 'plugin') return line.includes('[PluginManager]');
                             if (filter === 'webui') return line.includes('[WebUI]') || line.includes('Flask') || line.includes('"GET /') || line.includes('"POST /') || line.includes(' * ');
                             return false;
                        });
                    }

                    container.innerHTML = filteredData.join('<br>');
                    
                    if (wasBottom || force === true) {
                        container.scrollTop = container.scrollHeight;
                    }
                });
        }

        function clearLogs() {
            if (!confirm('确定要清空所有日志吗？')) {
                return;
            }
            fetch('/api/logs', { method: 'DELETE' })
                .then(r => r.json())
                .then(res => {
                    if (res.status === 'ok') {
                        document.getElementById('log-container').innerHTML = '';
                    } else {
                        alert('清空失败');
                    }
                })
                .catch(err => {
                    alert('清空请求失败: ' + err);
                });
        }

        function goToLogin() {
            if (currentBotId) {
                window.location.href = '/login?bot_id=' + currentBotId;
            } else {
                alert('请先选择一个机器人实例');
            }
        }
        
        function logoutBot() {
            if(confirm("确定要退出登录吗？")) {
                fetch('/api/logout?bot_id=' + currentBotId, { method: 'POST' }).then(() => location.reload());
            }
        }

        function loadNetworkConfig() {
            fetch('/api/config')
                .then(r => r.json())
                .then(data => {
                    const ws = data.network?.ws_server || {};
                    document.getElementById('ws-name').value = ws.name || 'test';
                    document.getElementById('ws-host').value = ws.host || '0.0.0.0';
                    document.getElementById('ws-port').value = ws.port || 3001;
                    document.getElementById('ws-heartbeat').value = ws.heartbeat_interval || 30000;
                    document.getElementById('ws-format').value = ws.message_format || 'string';
                    document.getElementById('ws-report-self').checked = ws.report_self_message !== false;
                    document.getElementById('ws-force-push').checked = ws.force_push_event !== false;
                })
                .catch(err => console.error('Load config failed:', err));
        }

        function saveNetworkConfig() {
            const config = {
                network: {
                    ws_server: {
                        name: document.getElementById('ws-name').value,
                        host: document.getElementById('ws-host').value,
                        port: parseInt(document.getElementById('ws-port').value),
                        heartbeat_interval: parseInt(document.getElementById('ws-heartbeat').value),
                        message_format: document.getElementById('ws-format').value,
                        report_self_message: document.getElementById('ws-report-self').checked,
                        force_push_event: document.getElementById('ws-force-push').checked
                    }
                }
            };
            
            const statusDiv = document.getElementById('config-status');
            statusDiv.style.display = 'none';
            
            fetch('/api/config', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(config)
            })
            .then(r => r.json())
            .then(res => {
                if (res.status === 'ok') {
                    alert('配置已保存，重启生效');
                } else {
                    alert('保存失败: ' + res.error);
                }
            })
            .catch(err => {
                alert('请求错误: ' + err);
            });
        }

        // Docker Management
        function showDockerSettings() {
            fetch('/api/config')
                .then(r => r.json())
                .then(data => {
                    const prefix = data.docker?.command_prefix || 'docker';
                    document.getElementById('docker-cmd-prefix').value = prefix;
                    new bootstrap.Modal(document.getElementById('dockerSettingsModal')).show();
                });
        }

        function saveDockerSettings() {
            const prefix = document.getElementById('docker-cmd-prefix').value.trim();
            if (!prefix) return alert('命令前缀不能为空');
            
            fetch('/api/config')
                .then(r => r.json())
                .then(config => {
                    if (!config.docker) config.docker = {};
                    config.docker.command_prefix = prefix;
                    
                    fetch('/api/config', {
                        method: 'POST',
                        headers: {'Content-Type': 'application/json'},
                        body: JSON.stringify(config)
                    })
                    .then(r => r.json())
                    .then(res => {
                        if (res.status === 'ok') {
                            alert('设置已保存');
                            const modal = bootstrap.Modal.getInstance(document.getElementById('dockerSettingsModal'));
                            modal.hide();
                            loadDockerContainers();
                        } else {
                            alert('保存失败: ' + res.error);
                        }
                    });
                });
        }

        function loadDockerContainers() {
            const tbody = document.getElementById('docker-list');
            tbody.innerHTML = '<tr><td colspan="6" class="text-center text-muted py-3">加载中...</td></tr>';
            
            fetch('/api/docker/containers')
                .then(r => r.json())
                .then(res => {
                    if (res.status === 'ok') {
                        tbody.innerHTML = '';
                        res.data.forEach(c => {
                            const tr = document.createElement('tr');
                            tr.innerHTML = `
                                <td><span class="badge bg-secondary font-monospace">${c.id.substring(0, 12)}</span></td>
                                <td class="fw-bold text-primary">${c.name}</td>
                                <td class="text-muted small">${c.image}</td>
                                <td><span class="badge ${c.status.includes('Up') ? 'bg-success' : 'bg-danger'}">${c.status}</span></td>
                                <td class="small text-truncate" style="max-width: 150px;">${c.ports}</td>
                                <td class="text-end">
                                    <div class="btn-group btn-group-sm">
                                        <button class="btn btn-outline-info" onclick="showDockerLogs('${c.id}')" title="日志"><i class="bi bi-file-text"></i></button>
                                        <button class="btn btn-outline-warning" onclick="restartContainer('${c.id}')" title="重启"><i class="bi bi-arrow-repeat"></i></button>
                                        <button class="btn btn-outline-danger" onclick="stopContainer('${c.id}')" title="停止"><i class="bi bi-stop-circle"></i></button>
                                    </div>
                                </td>
                            `;
                            tbody.appendChild(tr);
                        });
                        if (res.data.length === 0) {
                            tbody.innerHTML = '<tr><td colspan="6" class="text-center text-muted py-3">未发现容器</td></tr>';
                        }
                    } else {
                        tbody.innerHTML = `<tr><td colspan="6" class="text-center text-danger py-3">加载失败: ${res.error}</td></tr>`;
                    }
                })
                .catch(err => {
                    tbody.innerHTML = `<tr><td colspan="6" class="text-center text-danger py-3">请求错误: ${err}</td></tr>`;
                });
        }

        function restartContainer(cid) {
            if (!confirm('确定要重启该容器吗？')) return;
            fetch('/api/docker/restart', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({container_id: cid})
            })
            .then(r => r.json())
            .then(res => {
                if (res.status === 'ok') {
                    alert('重启命令已发送');
                    loadDockerContainers();
                } else {
                    alert('重启失败: ' + res.error);
                }
            })
            .catch(err => alert('请求错误: ' + err));
        }

        function stopContainer(cid) {
            if (!confirm('确定要停止该容器吗？')) return;
            fetch('/api/docker/stop', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({container_id: cid})
            })
            .then(r => r.json())
            .then(res => {
                if (res.status === 'ok') {
                    alert('停止命令已发送');
                    loadDockerContainers();
                } else {
                    alert('停止失败: ' + res.error);
                }
            })
            .catch(err => alert('请求错误: ' + err));
        }

        function showDockerLogs(cid) {
            const content = document.getElementById('docker-log-content');
            content.textContent = '加载日志中...';
            const modal = new bootstrap.Modal(document.getElementById('dockerLogModal'));
            modal.show();
            
            fetch('/api/docker/logs?container_id=' + cid)
                .then(r => r.json())
                .then(res => {
                    if (res.status === 'ok') {
                        content.textContent = res.logs || '(空日志)';
                    } else {
                        content.textContent = '获取日志失败: ' + res.error;
                    }
                })
                .catch(err => {
                    content.textContent = '请求错误: ' + err;
                });
        }

        // Init
        document.addEventListener('DOMContentLoaded', function() {
            updateBotList();
            initDarkMode();
            loadNetworkConfig();
            initSystemChart();
            setInterval(updateSystemStats, 5000);
        });

        let sysChart = null;
        let msgChart = null;

        function initSystemChart() {
            // System Load Chart
            const ctx = document.getElementById('systemLoadChart').getContext('2d');
            sysChart = new Chart(ctx, {
                type: 'line',
                data: {
                    labels: [],
                    datasets: [{
                        label: 'CPU (%)',
                        borderColor: '#0d6efd',
                        backgroundColor: 'rgba(13, 110, 253, 0.1)',
                        data: [],
                        fill: true,
                        tension: 0.4
                    }, {
                        label: '内存 (%)',
                        borderColor: '#198754',
                        backgroundColor: 'rgba(25, 135, 84, 0.1)',
                        data: [],
                        fill: true,
                        tension: 0.4
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: { position: 'top' },
                    },
                    scales: {
                        y: {
                            beginAtZero: true,
                            max: 100
                        }
                    },
                    animation: false
                }
            });
            
            // Message Traffic Chart
            const ctx2 = document.getElementById('msgTrafficChart').getContext('2d');
            msgChart = new Chart(ctx2, {
                type: 'line',
                data: {
                    labels: [],
                    datasets: [{
                        label: '消息数 (2s)',
                        borderColor: '#fd7e14',
                        backgroundColor: 'rgba(253, 126, 20, 0.1)',
                        data: [],
                        fill: true,
                        tension: 0.4
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: { position: 'top' },
                    },
                    scales: {
                        y: {
                            beginAtZero: true,
                            suggestedMax: 10
                        }
                    },
                    animation: false
                }
            });
            
            // Start update loop
            updateSystemChart();
            setInterval(updateSystemChart, 2000);
        }

        function updateSystemChart() {
            fetch('/api/stats_history')
                .then(r => r.json())
                .then(data => {
                    if (!sysChart || !msgChart) return;
                    
                    const labels = data.map(d => d.time);
                    const cpuData = data.map(d => d.cpu);
                    const memData = data.map(d => d.mem);
                    const msgData = data.map(d => d.msg_count || 0);
                    
                    sysChart.data.labels = labels;
                    sysChart.data.datasets[0].data = cpuData;
                    sysChart.data.datasets[1].data = memData;
                    sysChart.update();
                    
                    msgChart.data.labels = labels;
                    msgChart.data.datasets[0].data = msgData;
                    msgChart.update();
                });
        }
    </script>
</body>
</html>
"""

class WebUI:
    def __init__(self, manager):
        self.manager = manager
        self.app = Flask(__name__)
        self.start_ts = time.time()
        self.setup_routes()
        
    def get_bot(self, bot_id=None):
        if not bot_id:
            return self.manager.bots[0] if self.manager.bots else None
        
        for bot in self.manager.bots:
            if str(bot.self_id) == str(bot_id):
                return bot
        return None
        
    def _get_cpu_model(self):
        try:
            if platform.system() == "Windows":
                try:
                    # Try wmic for better name
                    output = subprocess.check_output("wmic cpu get name", shell=True).decode().strip()
                    lines = [line.strip() for line in output.split('\n') if line.strip()]
                    if len(lines) > 1:
                        return lines[1]
                except:
                    pass
                return platform.processor()
            elif platform.system() == "Linux":
                command = "cat /proc/cpuinfo"
                all_info = subprocess.check_output(command, shell=True).decode().strip()
                for line in all_info.split("\n"):
                    if "model name" in line:
                        return re.sub(r".*model name.*:", "", line, 1).strip()
        except:
            pass
        return platform.machine() # Fallback

    def _get_docker_cmd(self):
        return self.manager.config.get("docker", {}).get("command_prefix", "docker")

    def setup_routes(self):
        @self.app.route('/')
        def index():
            return render_template_string(HTML_TEMPLATE)

        @self.app.route('/login')
        def login_page():
            bot_id = request.args.get("bot_id", "")
            return render_template_string(LOGIN_TEMPLATE, bot_id=bot_id, ts=int(time.time()))

        @self.app.route('/api/bots')
        def list_bots():
            bots_data = []
            for bot in self.manager.bots:
                is_alive = False
                nick = "未登录"
                
                has_account = hasattr(bot, 'my_account') and isinstance(bot.my_account, dict) and bot.my_account
                if has_account:
                    nick = bot.my_account.get("NickName", "未登录")
                    is_alive = True
                
                bots_data.append({
                    "self_id": str(bot.self_id),
                    "nickname": nick,
                    "is_alive": is_alive,
                    "group_count": len(bot.group_members) if hasattr(bot, 'group_members') else 0
                })
            return jsonify(bots_data)

        @self.app.route('/api/config', methods=['GET'])
        def get_config():
            return jsonify(self.manager.config)

        @self.app.route('/api/config', methods=['POST'])
        def update_config():
            data = request.json or {}
            if self.manager.save_config(data):
                return jsonify({"status": "ok"})
            else:
                return jsonify({"status": "error", "error": "Save failed"})

        @self.app.route('/api/add_bot', methods=['POST'])
        def add_bot_api():
            data = request.json or {}
            self_id = data.get("self_id")
            # 如果没有提供 ID，生成一个临时的
            if not self_id:
                self_id = int(time.time())
            
            try:
                self.manager.add_bot(self_id)
                return jsonify({"status": "ok", "self_id": str(self_id)})
            except Exception as e:
                return jsonify({"status": "error", "msg": str(e)})

        @self.app.route('/api/qr_code')
        def qr_code():
            bot_id = request.args.get("bot_id")
            bot = self.get_bot(bot_id)
            if not bot:
                return "Bot not found", 404
            
            qr_path = getattr(bot, 'qr_file_path', os.path.join(bot.temp_pwd, 'wxqr.png'))
            if os.path.exists(qr_path):
                return send_file(qr_path, mimetype='image/png')
            else:
                return "QR code not ready", 404

        @self.app.route('/api/stats_history')
        def stats_history_api():
            return jsonify(list(stats_history))

        @self.app.route('/api/system_stats')
        def system_stats():
            bot_id = request.args.get("bot_id")
            process = psutil.Process(os.getpid())
            mem_info = process.memory_info()
            vm = psutil.virtual_memory()
            
            # 第一次调用 interval=None 返回 0.0，后续调用返回自上次调用以来的平均值
            # 前端每 5 秒轮询一次，足以计算出间隔内的 CPU 使用率
            cpu_usage = psutil.cpu_percent(interval=None)
            if cpu_usage == 0.0:
                 # 如果是 0，尝试再获取一次，稍微阻塞一下以获得精确值，但不要太久以免卡顿接口
                 # 或者直接返回 process.cpu_percent(interval=None) 获取当前进程的使用率
                 cpu_usage = process.cpu_percent(interval=None)

            # CPU Freq
            cpu_freq_str = "0.00GHz"
            try:
                freq = psutil.cpu_freq()
                if freq:
                    curr = freq.current
                    if curr > 1000:
                        cpu_freq_str = f"{curr/1000:.2f}GHz"
                    else:
                        cpu_freq_str = f"{curr:.2f}MHz"
            except:
                pass

            uptime_seconds = int(time.time() - self.start_ts)
            m, s = divmod(uptime_seconds, 60)
            h, m = divmod(m, 60)
            uptime_str = f"{h:02d}:{m:02d}:{s:02d}"
            
            # Bot Info
            is_alive = False
            nick = "未登录"
            uid = ""
            bot = self.get_bot(bot_id)
            if bot:
                # 增强的 is_alive 检测逻辑
                # 1. 检查是否有 my_account 且非空
                has_account = hasattr(bot, 'my_account') and isinstance(bot.my_account, dict) and bot.my_account
                
                if has_account:
                    nick = bot.my_account.get("NickName", "未知")
                    uid = bot.my_account.get("UserName", "")
                    is_alive = True
                
                # Debug logging (Limited frequency could be better, but for now just print)
                # print(f"[DEBUG] Bot {bot.self_id} is_alive={is_alive} account={bool(has_account)}")
            
                group_count = len(bot.group_members) if hasattr(bot, 'group_members') else 0
                current_bot_id = str(bot.self_id)
            else:
                group_count = 0
                current_bot_id = ""

            return jsonify({
                "cpu_percent": cpu_usage,
                "cpu_cores": psutil.cpu_count(logical=True),
                "cpu_model": self._get_cpu_model(),
                "cpu_freq": cpu_freq_str,
                "memory_used": mem_info.rss,
                "memory_system_total": vm.total,
                "memory_system_used": vm.used,
                "memory_system_percent": vm.percent,
                "thread_count": process.num_threads(),
                "uptime": uptime_str,
                "bot_info": {
                    "nickname": nick,
                    "uid": uid,
                    "is_alive": is_alive,
                    "self_id": current_bot_id
                },
                "group_count": group_count,
                "system_info": {
                    "python_version": platform.python_version(),
                    "os_version": f"{platform.system()} {platform.release()} {platform.machine()}"
                }
            })

        @self.app.route('/api/groups')
        def groups():
            bot_id = request.args.get("bot_id")
            from wxgroup import wx_group
            groups_list = []
            bot = self.get_bot(bot_id)
            if bot and hasattr(bot, 'group_members'):
                for gid, members in bot.group_members.items():
                    # Get Group Name
                    name = "未知群组"
                    owner_uid = ""
                    # Try to find name in group_list
                    for g in bot.group_list:
                        if g['UserName'] == gid:
                            name = g.get('NickName', '未知群组')
                            owner_uid = g.get('ChatRoomOwner', '')
                            break
                    
                    # Get Short ID
                    try:
                        short_id = wx_group.get_group_id(gid)
                    except:
                        short_id = 0
                        
                    groups_list.append({
                        "gid": gid,
                        "name": name,
                        "short_id": short_id,
                        "owner_uid": owner_uid,
                        "member_count": len(members)
                    })
            return jsonify(groups_list)

        @self.app.route('/api/group_members')
        def group_members():
            bot_id = request.args.get("bot_id")
            gid = request.args.get("gid")
            bot = self.get_bot(bot_id)
            if not bot:
                return jsonify([])
            
            members_data = []
            if hasattr(bot, 'group_members') and gid in bot.group_members:
                members = bot.group_members[gid]
                for m in members:
                    uid = m.get('UserName', '')
                    nick = m.get('NickName', '')
                    display = m.get('DisplayName', '')
                    
                    # 优先显示群昵称
                    name = display if display else nick
                    if not name: name = "未知成员"
                    
                    members_data.append({
                        'uid': uid,
                        'name': name,
                        'nick': nick,
                        'display': display
                    })
            
            # Sort by name
            members_data.sort(key=lambda x: x['name'])
            return jsonify(members_data)

        @self.app.route('/api/logs', methods=['GET'])
        def get_logs():
            return jsonify(list(log_buffer))

        @self.app.route('/api/logs', methods=['DELETE'])
        def clear_logs():
            log_buffer.clear()
            return jsonify({'status': 'ok'})


                
        @self.app.route('/api/send_msg', methods=['POST'])
        def send_msg():
            data = request.json
            gid = data.get('gid')
            msg = data.get('msg')
            bot_id = data.get('bot_id')
            
            bot = self.get_bot(bot_id)
            if not bot:
                 return jsonify({'status': 'error', 'error': 'Bot not found'})
            
            if not gid or not msg:
                return jsonify({'status': 'error', 'error': 'Missing gid or msg'})
            
            try:
                # 尝试发送消息
                # 这里假设 bot 有 send_msg_by_uid 方法
                # gid 可能是群的 UserName (uid)
                success = bot.send_msg_by_uid(msg, gid)
                if success:
                    return jsonify({'status': 'ok'})
                else:
                    return jsonify({'status': 'error', 'error': 'Send failed'})
            except Exception as e:
                return jsonify({'status': 'error', 'error': str(e)})

        @self.app.route('/api/upload_temp', methods=['POST'])
        def upload_temp():
            if 'file' not in request.files:
                return jsonify({'status': 'error', 'error': 'No file part'})
            file = request.files['file']
            if file.filename == '':
                return jsonify({'status': 'error', 'error': 'No selected file'})
            
            if file:
                filename = secure_filename(file.filename)
                # Ensure temp directory exists
                temp_dir = os.path.join(os.getcwd(), 'temp')
                if not os.path.exists(temp_dir):
                    os.makedirs(temp_dir)
                
                # Add timestamp to filename to avoid conflict
                timestamp = int(time.time())
                name, ext = os.path.splitext(filename)
                filename = f"{name}_{timestamp}{ext}"
                
                filepath = os.path.join(temp_dir, filename)
                file.save(filepath)
                return jsonify({'status': 'ok', 'path': filepath})
            return jsonify({'status': 'error', 'error': 'Unknown error'})

        @self.app.route('/api/docker/containers')
        def list_containers():
            try:
                prefix = self._get_docker_cmd()
                # Use '___' as separator to avoid shell pipe '|' issues when using SSH
                cmd = f"{prefix} ps -a --format \"{{{{.ID}}}}___{{{{.Names}}}}___{{{{.Image}}}}___{{{{.Status}}}}___{{{{.Ports}}}}___{{{{.CreatedAt}}}}\""
                
                try:
                    # Capture stderr as well to show in error message
                    output_bytes = subprocess.check_output(cmd, shell=True, stderr=subprocess.STDOUT)
                    try:
                        output = output_bytes.decode('utf-8')
                    except UnicodeDecodeError:
                        output = output_bytes.decode('gbk', errors='ignore')
                except subprocess.CalledProcessError as e:
                    # Try to decode stderr output
                    err_out = e.output.decode('utf-8', errors='ignore') if e.output else ""
                    err_msg = f"执行失败 (Exit {e.returncode}).<br>CMD: {cmd}<br>Output: {err_out}<br>请检查 Docker 设置。若未安装 Docker，请尝试安装或配置 SSH 前缀。"
                    return jsonify({"status": "error", "error": err_msg})
                
                containers = []
                for line in output.split('\n'):
                    if not line.strip(): continue
                    parts = line.split('___')
                    if len(parts) >= 6:
                        containers.append({
                            "id": parts[0],
                            "name": parts[1],
                            "image": parts[2],
                            "status": parts[3],
                            "ports": parts[4],
                            "created": parts[5]
                        })
                return jsonify({"status": "ok", "data": containers})
            except Exception as e:
                return jsonify({"status": "error", "error": str(e)})

        @self.app.route('/api/docker/restart', methods=['POST'])
        def restart_container():
            data = request.json
            cid = data.get('container_id')
            if not cid: return jsonify({"status": "error", "error": "Missing container_id"})
            try:
                prefix = self._get_docker_cmd()
                subprocess.check_call(f"{prefix} restart {cid}", shell=True)
                return jsonify({"status": "ok"})
            except Exception as e:
                return jsonify({"status": "error", "error": str(e)})

        @self.app.route('/api/docker/stop', methods=['POST'])
        def stop_container():
            data = request.json
            cid = data.get('container_id')
            if not cid: return jsonify({"status": "error", "error": "Missing container_id"})
            try:
                prefix = self._get_docker_cmd()
                subprocess.check_call(f"{prefix} stop {cid}", shell=True)
                return jsonify({"status": "ok"})
            except Exception as e:
                return jsonify({"status": "error", "error": str(e)})

        @self.app.route('/api/docker/logs')
        def get_container_logs():
            cid = request.args.get('container_id')
            if not cid: return jsonify({"status": "error", "error": "Missing container_id"})
            try:
                prefix = self._get_docker_cmd()
                output_bytes = subprocess.check_output(f"{prefix} logs --tail 200 {cid}", shell=True, stderr=subprocess.STDOUT)
                try:
                    output = output_bytes.decode('utf-8')
                except UnicodeDecodeError:
                    output = output_bytes.decode('gbk', errors='ignore')
                return jsonify({"status": "ok", "logs": output})
            except Exception as e:
                return jsonify({"status": "error", "error": str(e)})

        @self.app.route('/api/send_test_msg', methods=['POST'])
        def send_test_msg():
            data = request.json
            bot_id = data.get('bot_id')
            target_id = data.get('target_id') # uid (UserName)
            msg_type = data.get('type')
            
            bot = self.get_bot(bot_id)
            if not bot:
                 return jsonify({'status': 'error', 'error': 'Bot not found'})
            
            if not target_id:
                return jsonify({'status': 'error', 'error': 'Missing target_id'})

            try:
                success = False
                if msg_type == 'text':
                    content = data.get('content')
                    if not content: return jsonify({'status': 'error', 'error': 'Missing content'})
                    success = bot.send_msg_by_uid(content, target_id)
                
                elif msg_type == 'image':
                    file_path = data.get('file_path')
                    if not file_path: return jsonify({'status': 'error', 'error': 'Missing file_path'})
                    # Use send_img_msg_by_uid
                    if hasattr(bot, 'send_img_msg_by_uid'):
                        success = bot.send_img_msg_by_uid(file_path, target_id)
                    else:
                        return jsonify({'status': 'error', 'error': 'Method send_img_msg_by_uid not supported'})
                
                elif msg_type == 'file':
                    file_path = data.get('file_path')
                    if not file_path: return jsonify({'status': 'error', 'error': 'Missing file_path'})
                    # Use send_file_msg_by_uid
                    if hasattr(bot, 'send_file_msg_by_uid'):
                        success = bot.send_file_msg_by_uid(file_path, target_id)
                    else:
                        return jsonify({'status': 'error', 'error': 'Method send_file_msg_by_uid not supported'})

                elif msg_type == 'voice':
                    # Experimental: Try to send as file with Type=6 (or if bot has specific voice method)
                    # Ideally voice should be sent via specific API but we try file fallback or specific implementation
                    file_path = data.get('file_path')
                    if not file_path: return jsonify({'status': 'error', 'error': 'Missing file_path'})
                    if hasattr(bot, 'send_file_msg_by_uid'):
                        # Note: This sends as file. Real voice message needs different handling.
                        # But for testing if file can be sent, this is okay.
                        # If strict voice message (playable audio) is needed, we might need to modify wxbot.py
                        success = bot.send_file_msg_by_uid(file_path, target_id)
                    else:
                        return jsonify({'status': 'error', 'error': 'Method send_file_msg_by_uid not supported'})

                elif msg_type == 'video':
                    file_path = data.get('file_path')
                    if not file_path: return jsonify({'status': 'error', 'error': 'Missing file_path'})
                    if hasattr(bot, 'send_file_msg_by_uid'):
                        success = bot.send_file_msg_by_uid(file_path, target_id)
                    else:
                        return jsonify({'status': 'error', 'error': 'Method send_file_msg_by_uid not supported'})

                elif msg_type == 'music':
                    title = data.get('title')
                    desc = data.get('desc')
                    url = data.get('url')
                    music_url = data.get('music_url')
                    # Use _send_music_card_by_uid if available
                    if hasattr(bot, '_send_music_card_by_uid'):
                        success = bot._send_music_card_by_uid(target_id, title, desc, url, music_url)
                    else:
                        return jsonify({'status': 'error', 'error': 'Method _send_music_card_by_uid not supported'})
                
                elif msg_type == 'share':
                    title = data.get('title')
                    desc = data.get('desc')
                    url = data.get('url')
                    image_url = data.get('image_url')
                    # Use _send_link_card_by_uid if available
                    if hasattr(bot, '_send_link_card_by_uid'):
                        success = bot._send_link_card_by_uid(target_id, title, desc, url, image_url)
                    else:
                        return jsonify({'status': 'error', 'error': 'Method _send_link_card_by_uid not supported'})
                
                elif msg_type == 'kick':
                    member_id = data.get('member_id')
                    if not member_id: return jsonify({'status': 'error', 'error': 'Missing member_id'})
                    # Use delete_user_from_group
                    if hasattr(bot, 'delete_user_from_group'):
                        # 仅进行提示性检查，不阻断实际调用
                        owner_warn = False
                        try:
                            if hasattr(bot, 'is_group_owner') and not bot.is_group_owner(target_id):
                                owner_warn = True
                        except Exception:
                            pass
                        success = bot.delete_user_from_group(target_id, member_id)
                    else:
                         return jsonify({'status': 'error', 'error': 'Method delete_user_from_group not supported'})
                         
                elif msg_type == 'tickle':
                    member_id = data.get('member_id')
                    if hasattr(bot, 'send_poke'):
                         success = bot.send_poke(target_id, member_id)
                    else:
                         return jsonify({'status': 'error', 'error': 'Method send_poke not supported'})

                elif msg_type == 'quit_group':
                    if hasattr(bot, 'quit_group'):
                         success = bot.quit_group(target_id)
                    else:
                         return jsonify({'status': 'error', 'error': 'Method quit_group not supported'})

                elif msg_type == 'mod_group_remark':
                    group_remark = data.get('group_remark')
                    if not group_remark:
                         return jsonify({'status': 'error', 'error': 'Missing group_remark'})
                    if hasattr(bot, 'set_group_remark'):
                         success = bot.set_group_remark(target_id, group_remark)
                    else:
                         return jsonify({'status': 'error', 'error': 'Method set_group_remark not supported'})

                elif msg_type == 'mod_group_name':
                    group_name = data.get('group_name')
                    if not group_name:
                         return jsonify({'status': 'error', 'error': 'Missing group_name'})
                    if hasattr(bot, 'set_group_name'):
                         success = bot.set_group_name(target_id, group_name)
                    else:
                         return jsonify({'status': 'error', 'error': 'Method set_group_name not supported'})

                else:
                    return jsonify({'status': 'error', 'error': 'Unknown message type'})

                if success:
                    resp_data = {'status': 'ok'}
                    try:
                        if msg_type == 'kick' and 'owner_warn' in locals() and owner_warn:
                            resp_data['warn'] = '检测到机器人可能不是群主，已尝试执行接口'
                    except Exception:
                        pass
                    if isinstance(success, dict) and 'MsgID' in success:
                         resp_data['msg_id'] = success['MsgID']
                         resp_data['local_id'] = success.get('LocalID', '')
                         resp_data['to_user'] = target_id
                    return jsonify(resp_data)
                else:
                    err = 'Send failed (returned False)'
                    try:
                        api_ret = getattr(bot, '_last_api_ret', None)
                        if api_ret and isinstance(api_ret, dict):
                            br = api_ret.get('BaseResponse') or {}
                            ret_code = br.get('Ret')
                            err_msg = br.get('ErrMsg') or ''
                            if ret_code is not None:
                                err = f'Send failed (Ret={ret_code})'
                                if err_msg:
                                    err += f' {err_msg}'
                    except Exception:
                        pass
                    return jsonify({'status': 'error', 'error': err})

            except Exception as e:
                return jsonify({'status': 'error', 'error': str(e)})

        @self.app.route('/api/revoke_test_msg', methods=['POST'])
        def revoke_test_msg():
            data = request.json
            bot_id = data.get('bot_id')
            msg_id = data.get('msg_id')
            local_id = data.get('local_id')
            to_user = data.get('to_user')
            
            bot = self.get_bot(bot_id)
            if not bot:
                 return jsonify({'status': 'error', 'error': 'Bot not found'})
            
            if hasattr(bot, 'revoke_msg'):
                if bot.revoke_msg(local_id, msg_id, to_user):
                    return jsonify({'status': 'ok'})
                else:
                    return jsonify({'status': 'error', 'error': 'Revoke failed'})
            else:
                return jsonify({'status': 'error', 'error': 'Revoke not supported'})


        @self.app.route('/api/proxy_image')
        def proxy_image():
            bot_id = request.args.get("bot_id")
            url = request.args.get("url")
            if not bot_id or not url:
                return "Missing params", 400
            
            bot = self.get_bot(bot_id)
            if not bot:
                return "Bot not found", 404
                
            # Handle relative URLs
            if not url.startswith('http'):
                # base_uri usually is like https://wx.qq.com/cgi-bin/mmwebwx-bin
                # We need domain. base_host is better.
                host = getattr(bot, 'base_host', '')
                if not host:
                    host = 'wx.qq.com'
                
                if not url.startswith('/'):
                    url = '/' + url
                
                # Assume https
                full_url = f"https://{host}{url}"
            else:
                full_url = url
            
            # print(f"Proxying image: {full_url}")
            
            try:
                # Use bot's session to fetch image
                # Need to stream response
                req = bot.session.get(full_url, stream=True, timeout=10)
                return Response(req.iter_content(chunk_size=1024), content_type=req.headers.get('content-type', 'image/jpeg'))
            except Exception as e:
                return str(e), 500

        @self.app.route('/api/manual')
        def get_manual():
            try:
                base_dir = os.path.dirname(os.path.abspath(__file__))
                # Try multiple locations
                candidates = [
                    os.path.join(base_dir, 'SERVER_MANUAL.md'),
                    os.path.join(os.getcwd(), 'SERVER_MANUAL.md'),
                    os.path.join(base_dir, '../SERVER_MANUAL.md')
                ]
                
                manual_path = None
                for p in candidates:
                    if os.path.exists(p):
                        manual_path = p
                        break
                
                if manual_path:
                    with open(manual_path, 'r', encoding='utf-8') as f:
                        content = f.read()
                    return jsonify({'status': 'ok', 'content': content})
                else:
                    # Fallback to embedded content
                    return jsonify({'status': 'ok', 'content': DEFAULT_MANUAL_CONTENT})
            except Exception as e:
                # Fallback on error too
                return jsonify({'status': 'ok', 'content': DEFAULT_MANUAL_CONTENT + f"\n\n---\n*Note: Loaded from fallback due to error: {str(e)}*"})

        @self.app.route('/api/logout', methods=['POST'])
        def logout():
            bot_id = request.args.get("bot_id")
            bot = self.get_bot(bot_id)
            # 这里简单地清除 session 文件并重启程序可能不太合适
            # 对于多实例，可能需要单独注销
            if bot:
                if hasattr(bot, 'cache_file') and os.path.exists(bot.cache_file):
                    os.remove(bot.cache_file)
                # 触发重新登录流程 (这里只是删除了 session，实际上可能需要重启 bot 线程)
                # 目前简单处理：重启整个进程 (注意：这会影响所有机器人)
                # 更好的做法是只停止该 bot 并重启它的线程
                # 但当前架构下，重启进程是最简单的重置方式
                
            # TODO: 实现单个 bot 的注销重启
            os._exit(0)
            return jsonify({'status': 'ok'})

    def run(self, port=5000):
        # Disable Flask banner
        log = logging.getLogger('werkzeug')
        log.setLevel(logging.ERROR)
        print(f"[WebUI] Server running at http://0.0.0.0:{port}/")
        self.app.run(host='0.0.0.0', port=port, debug=False, use_reloader=False)

def start_web_ui(manager, port=5000):
    ui = WebUI(manager)
    t = threading.Thread(target=ui.run, args=(port,), daemon=True)
    t.start()
