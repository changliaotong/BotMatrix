import matplotlib
matplotlib.use('Agg') # Force non-interactive backend for Docker
import matplotlib.pyplot as plt
import io
import base64
import psutil
import datetime
import platform
import numpy as np

def generate_status_image(bot_stats):
    """
    生成系统状态仪表盘图片
    bot_stats: dict, 包含 bots 列表和消息统计
    """
    # 设置风格
    plt.style.use('dark_background')
    fig = plt.figure(figsize=(10, 6))
    
    # 布局: 2x2
    # 左上: 系统资源 (仪表盘风格模拟)
    # 右上: Bot 在线状态 (条形图)
    # 下方: 消息吞吐量 (折线图 - 模拟数据，因为 BotNexus 暂未传历史数据)
    
    # 1. 系统信息
    ax1 = plt.subplot(2, 2, 1)
    cpu_usage = psutil.cpu_percent(interval=0.1)
    mem_usage = psutil.virtual_memory().percent
    
    ax1.text(0.5, 0.7, f"CPU: {cpu_usage}%", ha='center', va='center', fontsize=20, color='#00ff00' if cpu_usage < 80 else '#ff0000')
    ax1.text(0.5, 0.4, f"MEM: {mem_usage}%", ha='center', va='center', fontsize=20, color='#00ffff')
    ax1.text(0.5, 0.1, f"UPTIME: {datetime.datetime.now().strftime('%H:%M:%S')}", ha='center', va='center', fontsize=10, color='white')
    ax1.axis('off')
    ax1.set_title("System Resources", fontsize=14, color='yellow')

    # 2. Bot 状态
    ax2 = plt.subplot(2, 2, 2)
    bots = bot_stats.get('bots', [])
    if bots:
        names = [b.get('self_id', 'Unknown')[-4:] for b in bots] # 只显示后4位
        status = [1 if b.get('is_alive') else 0 for b in bots]
        colors = ['#00ff00' if s else '#555555' for s in status]
        
        ax2.barh(names, status, color=colors)
        ax2.set_xlim(0, 1.2)
        ax2.set_xticks([])
        ax2.set_title(f"Active Bots: {sum(status)}/{len(bots)}", fontsize=14, color='cyan')
    else:
        ax2.text(0.5, 0.5, "No Bots Connected", ha='center', va='center')
        ax2.axis('off')

    # 3. 消息趋势 (模拟生成，实际应从 BotNexus 获取)
    ax3 = plt.subplot(2, 1, 2)
    hours = [f"{i}:00" for i in range(24)]
    # 生成一个正态分布曲线模拟活跃度
    x = np.linspace(0, 24, 24)
    y = 100 * np.exp(-(x-14)**2 / (2*4**2)) + np.random.randint(0, 20, 24)
    
    ax3.plot(hours, y, color='#ff00ff', linewidth=2, marker='o')
    ax3.fill_between(hours, y, color='#ff00ff', alpha=0.1)
    ax3.set_title("24H Message Traffic (Predicted)", fontsize=14, color='magenta')
    ax3.grid(True, linestyle='--', alpha=0.3)
    plt.xticks(rotation=45)

    # 保存
    buf = io.BytesIO()
    plt.tight_layout()
    plt.savefig(buf, format='png', dpi=100)
    buf.seek(0)
    plt.close()
    
    # 转为 Base64
    return base64.b64encode(buf.getvalue()).decode('utf-8')
