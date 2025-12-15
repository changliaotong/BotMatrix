import websocket
import json
import threading
import time
import sys

def on_message(ws, message):
    data = json.loads(message)
    print(f"Received: {json.dumps(data, indent=2)}")
    if data.get("params", {}).get("message", "").startswith("【BotMatrix: Base64"):
        print("SUCCESS: Received Safe Content Warning!")
        ws.close()
    elif "CQ:image" in data.get("params", {}).get("message", ""):
        print("WARNING: Received Raw CQ Code (Filtering not applied by Worker, expected behavior)")
        # If we are simulating the Bot, we receive the raw CQ code.
        # The filtering happens inside TencentBot BEFORE calling API.
        # So receiving CQ code here is CORRECT for the Bot.
        ws.close()

def on_error(ws, error):
    print(f"Error: {error}")

def on_close(ws, close_status_code, close_msg):
    print("Closed")

def on_open(ws):
    print("Connected to Nexus")
    
    # Simulate an event from TencentBot
    event = {
        "post_type": "message",
        "message_type": "private", # C2C
        "sub_type": "friend",
        "message_id": "test_msg_123",
        "user_id": "1098299491", # Matches ADMIN_USER_ID in SystemWorker
        "message": "#sys info",
        "raw_message": "#sys info",
        "font": 0,
        "sender": {
            "user_id": "1098299491",
            "nickname": "TestUser"
        },
        "time": int(time.time()),
        "self_id": "BotTest1"
    }
    
    def send_loop(ws):
        while True:
            print("Sending #sys info...")
            try:
                ws.send(json.dumps(event))
            except Exception as e:
                print(f"Send failed: {e}")
                break
            time.sleep(5)

    # Start a thread to keep sending
    threading.Thread(target=send_loop, args=(ws,), daemon=True).start()

if __name__ == "__main__":
    # Headers to identify as a Bot
    headers = {
        "X-Self-ID": "BotTest1",
        "X-Platform": "Guild"
    }
    
    # Connect to remote Nexus
    SERVER_IP = "192.168.0.167"
    PORT = "3005"
    
    ws = websocket.WebSocketApp(f"ws://{SERVER_IP}:{PORT}",
                              on_open=on_open,
                              on_message=on_message,
                              on_error=on_error,
                              on_close=on_close,
                              header=headers)
    
    ws.run_forever()
