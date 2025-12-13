import asyncio
import websockets
import json
import time

async def mock_bot(bot_id):
    uri = "ws://localhost:3001"
    print(f"[MockBot-{bot_id}] Connecting...")
    async with websockets.connect(uri) as ws:
        print(f"[MockBot-{bot_id}] Connected!")
        
        # 1. Identify
        await ws.send(json.dumps({
            "post_type": "meta_event",
            "meta_event_type": "lifecycle",
            "sub_type": "connect",
            "self_id": bot_id,
            "platform": "qq"
        }))
        
        # Keep connection open and print received messages
        while True:
            try:
                msg = await ws.recv()
                print(f"[MockBot-{bot_id}] Received: {msg}")
                
                # If this is Bot 1 (10086), send a trigger message "cross" after 2 seconds
                if bot_id == 10086:
                    # Parse to see if it's just the echo of login info
                    data = json.loads(msg)
                    if data.get("action") == "get_login_info":
                        print(f"[MockBot-{bot_id}] Sending 'cross' trigger...")
                        await asyncio.sleep(2)
                        await ws.send(json.dumps({
                            "post_type": "message",
                            "message_type": "private",
                            "time": int(time.time()),
                            "self_id": bot_id,
                            "sub_type": "friend",
                            "user_id": 999,
                            "message": "cross",
                            "raw_message": "cross",
                            "sender": {
                                "user_id": 999,
                                "nickname": "Tester"
                            }
                        }))
                        # Only send once
                        bot_id = 0 # Disable trigger
            except Exception as e:
                print(f"[MockBot-{bot_id}] Error: {e}")
                break

async def main():
    # Run two bots concurrently
    await asyncio.gather(
        mock_bot(10086),
        mock_bot(10087)
    )

if __name__ == "__main__":
    asyncio.run(main())
