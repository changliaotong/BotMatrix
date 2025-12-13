import asyncio
import websockets
import json
import time

async def mock_qq_bot():
    uri = "ws://localhost:3001"
    print(f"[MockQQ] Connecting to {uri}...")
    async with websockets.connect(uri) as ws:
        print("[MockQQ] Connected!")
        
        # 1. Identify
        await ws.send(json.dumps({
            "post_type": "meta_event",
            "meta_event_type": "lifecycle",
            "sub_type": "connect",
            "self_id": 123456789,
            "platform": "qq"
        }))
        
        # 2. Send a Group Message
        msg = {
            "post_type": "message",
            "message_type": "group",
            "time": int(time.time()),
            "self_id": 123456789,
            "sub_type": "normal",
            "user_id": 123456789,
            "group_id": 987654321,
            "message": "ping",
            "raw_message": "ping",
            "font": 0,
            "sender": {
                "user_id": 123456789,
                "nickname": "TestUser",
                "card": "",
                "role": "member"
            }
        }
        print(f"[MockQQ] Sending message: {msg['message']}")
        await ws.send(json.dumps(msg))
        
        # 3. Wait for API call (echo)
        while True:
            resp = await ws.recv()
            print(f"[MockQQ] Received: {resp}")

if __name__ == "__main__":
    asyncio.run(mock_qq_bot())
