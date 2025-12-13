# coding: utf-8
import os
import time
import json
import asyncio
import websockets
import threading
from onebot import onebot

class WorkerClient:
    def __init__(self, manager_url, self_id):
        self.manager_url = manager_url
        self.self_id = int(self_id)
        self.bot = None
        self.ws = None
        self._loop = asyncio.new_event_loop()
        self.running = True
        self.msg_queue = asyncio.Queue()
        
    def start(self):
        # 1. Skip OneBot start (WeChat Client)
        # print(f"[Worker] Starting OneBot Worker self_id={self.self_id}...")
        
        # 2. Start WebSocket Client to Manager as a WORKER (Consumer)
        # We need to append ?role=worker to the URL
        if "?" not in self.manager_url:
            self.manager_url += "?role=worker"
        else:
            self.manager_url += "&role=worker"
            
        t_ws = threading.Thread(target=self._run_ws_loop, daemon=True)
        t_ws.start()
        
        while self.running:
            time.sleep(1)

    def _run_ws_loop(self):
        asyncio.set_event_loop(self._loop)
        self._loop.run_until_complete(self._connect_to_manager())

    async def _connect_to_manager(self):
        while self.running:
            try:
                print(f"[Worker] Connecting to Manager at {self.manager_url}...")
                async with websockets.connect(self.manager_url) as ws:
                    self.ws = ws
                    print("[Worker] Connected to Manager.")
                    
                    # Send identification
                    await ws.send(json.dumps({
                        "post_type": "meta_event",
                        "meta_event_type": "lifecycle",
                        "sub_type": "connect",
                        "self_id": self.self_id,
                        "platform": "wechat"
                    }))
                    
                    # Create tasks for reading and writing
                    read_task = asyncio.create_task(self._read_loop(ws))
                    write_task = asyncio.create_task(self._write_loop(ws))
                    
                    done, pending = await asyncio.wait(
                        [read_task, write_task],
                        return_when=asyncio.FIRST_COMPLETED
                    )
                    
                    for task in pending:
                        task.cancel()
                            
            except Exception as e:
                print(f"[Worker] Connection lost/failed: {e}. Retrying in 5s...")
                await asyncio.sleep(5)

    async def _read_loop(self, ws):
        async for msg in ws:
            try:
                data = json.loads(msg)
                # Handle Events (Business Logic)
                if "post_type" in data:
                    await self._handle_event(data)
                elif "echo" in data:
                    print(f"[Worker] Received API response: {data}")
                elif data.get("action"):
                    # Only if we act as a plugin host
                    await self._handle_action(data)
            except Exception as e:
                print(f"[Worker] Msg handling error: {e}")

    async def _handle_event(self, event):
        print(f"[Worker] >>> Received Event: {event.get('post_type')} | {event}")
        
        # Cross-Bot Forwarding Logic
        # Scenario: Received "cross" on Bot 1 -> Send "forwarded" via Bot 2
        if event.get("post_type") == "message":
            msg = event.get("message", "")
            user_id = event.get("user_id")
            self_id = event.get("self_id") # The bot that RECEIVED the message
            
            print(f"[Worker] Logic processing message: {msg} from Bot {self_id}")
            
            if msg == "cross":
                # Assuming Bot 1 is 10086, Bot 2 is 10087
                target_bot_id = 10087
                
                print(f"[Worker] !!! Cross-Bot Triggered: Forwarding via Bot {target_bot_id} !!!")
                
                # Send API call via SPECIFIC Bot
                await self.send_api("send_private_msg", {
                    "user_id": user_id,
                    "message": f"Cross-reply from Bot {target_bot_id} (triggered by {self_id})"
                }, target_bot_id)

    async def send_api(self, action, params, target_self_id=None):
        if self.ws:
            payload = {
                "action": action,
                "params": params,
                "echo": f"echo_{int(time.time())}"
            }
            # Specify target bot if provided
            if target_self_id:
                payload["self_id"] = target_self_id
                
            print(f"[Worker] <<< Sending API to Bot {target_self_id or 'Any'}: {payload}")
            await self.ws.send(json.dumps(payload))

    async def _write_loop(self, ws):
        # We don't need this loop anymore if we send directly, 
        # or we can use it for async sending from other threads
        while True:
            # Just keep alive
            await asyncio.sleep(1)

    async def _handle_action(self, action_data):
        print(f"[Worker] Received action: {action_data}")
        
        name = action_data.get("action")
        params = action_data.get("params", {})
        echo = action_data.get("echo")
        
        result = {"status": "ok", "retcode": 0, "data": {}}

        # Intercept System Actions
        if name == "reload_plugins":
            try:
                self.pm.reload_plugins()
                result["data"] = {"message": "Plugins reloaded successfully"}
            except Exception as e:
                result = {"status": "failed", "retcode": 10002, "msg": str(e)}
        else:
            # Execute standard OneBot actions
            try:
                result = self.bot.execute_onebot_action(name, params)
            except Exception as e:
                result = {"status": "failed", "retcode": 10002, "msg": str(e)}
            
        # Add echo back if present
        if echo is not None:
            result["echo"] = echo
            
        if self.ws:
            await self.ws.send(json.dumps(result, ensure_ascii=False))

if __name__ == "__main__":
    # Get config from env or args
    MANAGER_URL = os.environ.get("MANAGER_URL", "ws://localhost:3001")
    SELF_ID = os.environ.get("BOT_SELF_ID", "123456")
    
    worker = WorkerClient(MANAGER_URL, SELF_ID)
    worker.start()
