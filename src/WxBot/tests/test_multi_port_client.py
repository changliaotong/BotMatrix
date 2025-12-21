import asyncio
import websockets
import json

async def test_bot(port, name):
    uri = f"ws://127.0.0.1:{port}"
    print(f"[{name}] Connecting to {uri}...")
    try:
        async with websockets.connect(uri) as ws:
            print(f"[{name}] Connected!")
            
            # 1. 接收 Lifecycle 事件
            try:
                msg = await asyncio.wait_for(ws.recv(), timeout=3.0)
                print(f"[{name}] Init Event: {msg}")
            except asyncio.TimeoutError:
                print(f"[{name}] No init event received")

            # 2. 发送 get_login_info (不带 self_id，测试默认路由)
            action = {
                "action": "get_login_info",
                "echo": f"check_{port}"
            }
            await ws.send(json.dumps(action))
            
            # 3. 接收响应
            res = await ws.recv()
            print(f"[{name}] Login Info: {res}")
            
    except ConnectionRefusedError:
        print(f"[{name}] Connection Refused (Port {port} not open)")
    except Exception as e:
        print(f"[{name}] Error: {e}")

async def main():
    # 假设配置了:
    # Port 3001 -> 个人微信
    # Port 3002 -> 企业微信
    # Port 3003 -> 钉钉
    
    tasks = [
        test_bot(3001, "Personal WeChat"),
        test_bot(3002, "Enterprise WeChat"),
        test_bot(3003, "DingTalk")
    ]
    await asyncio.gather(*tasks)

if __name__ == "__main__":
    asyncio.run(main())
