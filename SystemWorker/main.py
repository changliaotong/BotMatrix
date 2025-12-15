import asyncio
import json
import websockets
import os
import datetime
import platform
import sys
import io
import contextlib
import traceback
import requests
from plotter import generate_status_image

# 配置
BOT_MANAGER_URL = os.getenv("BOT_MANAGER_URL", "ws://bot-manager:3001")
BOT_MANAGER_API = os.getenv("BOT_MANAGER_API", "http://bot-manager:5000") # HTTP API for Bot List
WORKER_NAME = "SystemWorker-Core"
ADMIN_USER_ID = 1098299491 # 请替换为您的 UserID，或者实现动态鉴权

async def send_reply(ws, data, message):
    """辅助函数：发送回复"""
    reply = {
        "action": "send_msg",
        "params": {
            "user_id": data.get("user_id"),
            "message": message
        }
    }
    if data.get("message_type") == "group":
        reply["params"]["group_id"] = data.get("group_id")
    
    await ws.send(json.dumps(reply))

async def get_bot_list():
    """获取所有连接的 Bot"""
    # 临时方案：模拟数据，或者尝试调用 BotNexus HTTP API
    # 实际上 BotNexus 可能需要鉴权才能返回列表
    # 这里我们演示如何获取，如果失败则返回空
    try:
        # 假设我们有一个内部接口或者直接 blind broadcast
        # 暂时返回空，后续通过 broadcast 逻辑处理
        return []
    except:
        return []

async def handle_message(ws, data):
    """处理接收到的消息"""
    raw_msg = data.get("raw_message", "").strip()
    user_id = data.get("user_id")
    
    # 1. #sys status - 可视化仪表盘
    if raw_msg == "#sys status":
        print("Generating status image...")
        try:
            # 获取 Bot 列表信息 (模拟)
            bot_stats = {
                "bots": [
                    {"self_id": "Bot1001", "is_alive": True},
                    {"self_id": "Bot1002", "is_alive": True},
                    {"self_id": "Bot1003", "is_alive": False}
                ]
            }
            # 生成 Base64 图片
            b64_img = generate_status_image(bot_stats)
            # 构造 OneBot 格式的图片消息
            msg = f"[CQ:image,file=base64://{b64_img}]"
            await send_reply(ws, data, msg)
        except Exception as e:
            await send_reply(ws, data, f"Error generating status: {e}")

    # 2. #sys exec <code> - 远程代码执行 (危险!)
    elif raw_msg.startswith("#sys exec "):
        # 鉴权
        if user_id != ADMIN_USER_ID:
            await send_reply(ws, data, "🚫 Permission Denied")
            return

        code = raw_msg[10:].strip()
        # 捕获 stdout
        str_io = io.StringIO()
        try:
            with contextlib.redirect_stdout(str_io):
                # 包含一些常用的上下文
                exec_context = {
                    "os": os,
                    "sys": sys,
                    "datetime": datetime,
                    "data": data
                }
                exec(code, exec_context)
            output = str_io.getvalue()
            if not output:
                output = "<No Output>"
            await send_reply(ws, data, f"💻 Exec Result:\n{output}")
        except Exception as e:
            await send_reply(ws, data, f"❌ Exec Error:\n{traceback.format_exc()}")

    # 3. #sys broadcast <msg> - 全域广播
    elif raw_msg.startswith("#sys broadcast "):
        if user_id != ADMIN_USER_ID:
            await send_reply(ws, data, "🚫 Permission Denied")
            return

        broadcast_msg = raw_msg[15:].strip()
        if not broadcast_msg:
            return

        # 这里的逻辑比较 Trick：
        # 我们不知道有哪些群，所以我们利用 BotNexus 的广播机制
        # 如果 BotNexus 支持把消息转发给所有 Bot 的所有群...
        # 目前 BotNexus 仅支持把 Event 广播给 Subscriber。
        # 我们可以尝试发送一个特殊的 action 给 BotNexus？
        # 既然没有现成的 API，我们先只是回显一下，或者发给发送者自己以演示
        
        await send_reply(ws, data, f"📢 Broadcasting to ALL channels:\n{broadcast_msg}\n(Simulation: Real broadcast requires BotNexus API upgrade)")
        
        # 真正实现需要 SystemWorker 维护所有群列表，这需要数据库支持。
        # 这里演示一下向当前群发送三次以示区别
        # for i in range(3):
        #    await send_reply(ws, data, f"Broadcast {i+1}: {broadcast_msg}")

    # 4. 保留原有的 #sys info 作为纯文本备选
    elif raw_msg == "#sys info":
        sys_info = (
            f"[{WORKER_NAME}]\n"
            f"Status: Online\n"
            f"Time: {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n"
            f"Python: {platform.python_version()}\n"
            f"System: {platform.system()} {platform.release()}"
        )
        await send_reply(ws, data, sys_info)

async def main():
    connect_url = f"{BOT_MANAGER_URL}?role=worker"
    print(f"[{WORKER_NAME}] Connecting to {connect_url}...")
    
    while True:
        try:
            async with websockets.connect(connect_url) as ws:
                print(f"[{WORKER_NAME}] Connected to BotNexus!")
                
                while True:
                    try:
                        message = await ws.recv()
                        data = json.loads(message)
                        
                        # 忽略心跳和自身发送的消息
                        post_type = data.get("post_type")
                        
                        if post_type == "message":
                            await handle_message(ws, data)
                        elif post_type == "meta_event":
                            pass
                            
                    except websockets.exceptions.ConnectionClosed:
                        print(f"[{WORKER_NAME}] Connection closed by server")
                        break
                    except Exception as e:
                        print(f"[{WORKER_NAME}] Error processing message: {e}")
                        traceback.print_exc()
                        
        except Exception as e:
            print(f"[{WORKER_NAME}] Connection failed: {e}. Retrying in 5s...")
            await asyncio.sleep(5)

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print(f"[{WORKER_NAME}] Stopped.")
