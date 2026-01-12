import asyncio
import websockets

async def test():
    uri = "ws://localhost:8080/ws/bots"
    headers = {"X-Self-ID": "test", "X-Platform": "qq"}
    try:
        # 尝试使用 additional_headers
        async with websockets.connect(uri, additional_headers=headers) as ws:
            print("OK")
    except Exception as e:
        print(f"Error: {type(e).__name__}: {e}")

asyncio.run(test())
