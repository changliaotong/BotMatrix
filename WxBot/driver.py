import asyncio
import json
import os
import sys
import threading
import time
import websockets
import logging
from onebot import onebot

# Ensure WxBot root is in path
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

class NexusLogHandler(logging.Handler):
    def __init__(self, driver):
        super().__init__()
        self.driver = driver
    
    def emit(self, record):
        try:
            msg = self.format(record)
            if not msg: return
            
            # Avoid recursion if logging from within broadcast logic
            if "post_type" in msg and "log" in msg: return

            event = {
                "post_type": "log",
                "level": record.levelname,
                "message": msg,
                "time": time.strftime("%H:%M:%S"),
                "self_id": str(getattr(self.driver.bot, 'self_id', 'unknown')) if self.driver.bot else 'unknown'
            }
            
            if self.driver._loop and self.driver._loop.is_running() and self.driver._event_queue:
                 self.driver._loop.call_soon_threadsafe(self.driver._event_queue.put_nowait, event)
        except:
            self.handleError(record)

class StdoutToLogger:
    def __init__(self, logger, level):
        self.logger = logger
        self.level = level

    def write(self, message):
        if message.strip() == "": return
        self.logger.log(self.level, message.strip())

    def flush(self):
        pass

class OneBotDriver:
    """
    Universal OneBot Driver (WebSocket Client + Server)
    Supports:
    1. Reverse WebSocket (Connecting to Manager/Backend)
    2. Positive WebSocket (Listening for connections)
    """
    def __init__(self):
        # Configuration
        self.ws_urls = self._parse_ws_urls()
        self.ws_enable = os.getenv("WS_ENABLE", "true").lower() == "true"
        self.ws_host = os.getenv("WS_HOST", "0.0.0.0")
        self.ws_port = int(os.getenv("WS_PORT", "3001"))
        self.access_token = os.getenv("ACCESS_TOKEN", "")

        # State
        self._loop = None
        self._event_queue = None
        self.bot = None
        self.clients = set() # Set of active websocket connections
        self.running = True

    def _parse_ws_urls(self):
        """Parse WS_URLS or MANAGER_URL env vars"""
        urls = []
        # Support legacy/simple MANAGER_URL
        manager_url = os.getenv("MANAGER_URL")
        if manager_url:
            urls.append(manager_url)
        
        # Support WS_URLS (JSON list or comma separated)
        ws_urls_env = os.getenv("WS_URLS")
        if ws_urls_env:
            try:
                # Try JSON
                parsed = json.loads(ws_urls_env)
                if isinstance(parsed, list):
                    urls.extend(parsed)
            except:
                # Try comma separated
                urls.extend([u.strip() for u in ws_urls_env.split(",") if u.strip()])
        
        # Remove duplicates and cleanup
        return list(set(urls))

    def add_bot(self, bot, self_id):
        self.bot = bot
        print(f"[Driver] Bot {self_id} attached to Driver")

    async def start(self):
        print(f"[Driver] Starting OneBot Driver...")
        self._loop = asyncio.get_running_loop()
        self._event_queue = asyncio.Queue()
        
        tasks = []
        
        # 1. Start Event Consumer (Broadcasts events to all clients)
        tasks.append(asyncio.create_task(self._consume_events()))
        
        # 2. Start WebSocket Server (if enabled)
        if self.ws_enable:
            print(f"[Driver] Starting WebSocket Server on {self.ws_host}:{self.ws_port}")
            # Disable ping_interval to avoid disconnecting clients that don't respond to PINGs (code 1006)
            server = await websockets.serve(self._handle_connection, self.ws_host, self.ws_port, ping_interval=None)
            tasks.append(server.wait_closed())
            
        # 3. Start WebSocket Clients (Reverse WS)
        for url in self.ws_urls:
            print(f"[Driver] Adding Reverse WS Client: {url}")
            tasks.append(asyncio.create_task(self._maintain_client_connection(url)))

        if not tasks:
            print("[Driver] Warning: No server enabled and no URLs to connect to!")
            while self.running:
                await asyncio.sleep(1)
        else:
            await asyncio.gather(*tasks)

    async def _handle_connection(self, ws, path=None):
        """Handle incoming WebSocket connection (Server mode)"""
        print(f"[Driver] New connection from {ws.remote_address}")
        # Auth check could go here
        
        await self._register_client(ws)

    async def _maintain_client_connection(self, url):
        """Maintain connection to upstream (Client mode)"""
        while self.running:
            try:
                print(f"[Driver] Connecting to {url}...")
                # Add role=Universal and platform=wechat if missing
                connect_url = url
                if "?" not in connect_url:
                    connect_url += "?role=Universal"
                elif "role=" not in connect_url:
                    connect_url += "&role=Universal"
                
                if "platform=" not in connect_url:
                    connect_url += "&platform=wechat"

                async with websockets.connect(connect_url) as ws:
                    print(f"[Driver] Connected to {url}!")
                    await self._register_client(ws)
            except Exception as e:
                print(f"[Driver] Connection to {url} failed: {e}. Retry in 5s...")
                await asyncio.sleep(5)

    async def _heartbeat_loop(self, ws):
        """Send OneBot heartbeats to a client"""
        interval = 5000  # Default 5s to match lifecycle event
        while True:
            try:
                await asyncio.sleep(interval / 1000.0)
                
                event = {
                    "time": int(time.time()),
                    "self_id": self.bot.self_id if self.bot else 0,
                    "post_type": "meta_event",
                    "meta_event_type": "heartbeat",
                    "status": {
                        "online": True,
                        "good": True
                    },
                    "interval": interval
                }
                await ws.send(json.dumps(event, ensure_ascii=False))
            except asyncio.CancelledError:
                break
            except Exception as e:
                # print(f"[Driver] Heartbeat error: {e}")
                break

    async def _register_client(self, ws):
        """Register a client and handle its IO"""
        self.clients.add(ws)
        
        # Start Heartbeat Task
        heartbeat_task = asyncio.create_task(self._heartbeat_loop(ws))
        
        # Send Lifecycle Event (Connect)
        if self.bot:
            try:
                await ws.send(json.dumps({
                    "post_type": "meta_event",
                    "meta_event_type": "lifecycle",
                    "sub_type": "connect",
                    "self_id": self.bot.self_id,
                    "platform": "wechat",
                    "status": "online",
                    "interval": 5000
                }))
            except Exception as e:
                print(f"[Driver] Failed to send lifecycle event: {e}")

        try:
            await self._read_loop(ws)
        finally:
            print("[Driver] Client disconnected")
            heartbeat_task.cancel()
            self.clients.discard(ws)

    def broadcast_event(self, event):
        """Thread-safe method to push event to queue"""
        if self._loop and self._loop.is_running() and self._event_queue:
             self._loop.call_soon_threadsafe(self._event_queue.put_nowait, event)
        else:
             print("[Driver] Loop not running or queue not ready, dropping event")

    async def _consume_events(self):
        """Read events from queue and broadcast to ALL clients"""
        while True:
            event = await self._event_queue.get()
            
            # Broadcast to all connected clients
            if self.clients:
                message = json.dumps(event, ensure_ascii=False)
                # Create tasks for sending to avoid blocking
                send_tasks = []
                for ws in list(self.clients):
                    try:
                        # Log sending (truncated)
                        evt_type = event.get('post_type', 'unknown')
                        if evt_type != 'log':
                            print(f"[Driver] Push {evt_type} to {getattr(ws, 'remote_address', 'unknown')}")
                        send_tasks.append(asyncio.create_task(ws.send(message)))
                    except Exception as e:
                        print(f"[Driver] Error preparing send: {e}")
                        self.clients.discard(ws)
                
                if send_tasks:
                    await asyncio.gather(*send_tasks, return_exceptions=True)
            else:
                 # Log if no clients connected
                 print(f"[Driver] Event dropped (No clients): {event.get('post_type')}")
            
            self._event_queue.task_done()

    async def _read_loop(self, ws):
        """Read messages from a specific websocket"""
        try:
            async for msg in ws:
                try:
                    data = json.loads(msg)
                    
                    # Filter foreign events (broadcasts from other bots)
                    if "post_type" in data and "self_id" in data:
                        if self.bot and str(data.get("self_id")) != str(self.bot.self_id):
                            continue # Silently ignore events not for us

                    # Log received message
                    print(f"[Driver] Recv from {getattr(ws, 'remote_address', 'unknown')}: {str(msg)[:200]}")
                    
                    if "action" in data:
                        # Execute Action
                        if self.bot:
                            action = data.get("action")
                            params = data.get("params", {})
                            echo = data.get("echo")
                            
                            # Run action in executor to avoid blocking loop
                            res = await self._loop.run_in_executor(None, self.bot.execute_onebot_action, action, params)
                            
                            resp = {
                                "status": res.get("status"),
                                "retcode": res.get("retcode"),
                                "data": res.get("data"),
                                "echo": echo
                            }
                            await ws.send(json.dumps(resp, ensure_ascii=False))
                except Exception as e:
                    print(f"[Driver] Read error: {e}")
        except (websockets.exceptions.ConnectionClosedError, websockets.exceptions.ConnectionClosedOK):
            pass
        except Exception as e:
            print(f"[Driver] Connection error: {e}")

# Entry point
if __name__ == "__main__":
    # Load OneBot logic
    try:
        from onebot import onebot
    except ImportError:
        # Fallback if running directly
        sys.path.append(os.path.dirname(os.path.abspath(__file__)))
        from onebot import onebot
    
    # 1. Initialize Driver
    driver = OneBotDriver()

    # Setup Logging to Nexus
    logger = logging.getLogger()
    logger.setLevel(logging.INFO)
    
    # Console Handler (to see logs in terminal)
    # We must use sys.__stdout__ because we are about to redirect sys.stdout
    console_handler = logging.StreamHandler(sys.__stdout__)
    console_handler.setFormatter(logging.Formatter('%(message)s'))
    logger.addHandler(console_handler)

    # Nexus Handler (to stream logs to BotNexus)
    nexus_handler = NexusLogHandler(driver)
    nexus_handler.setFormatter(logging.Formatter('%(message)s'))
    logger.addHandler(nexus_handler)

    # Redirect print to logger
    sys.stdout = StdoutToLogger(logger, logging.INFO)
    sys.stderr = StdoutToLogger(logger, logging.ERROR)
    
    # 2. Initialize Bot
    # self_id will be auto-generated or read from env
    self_id = os.getenv("BOT_SELF_ID", "0")
    print(f"[Driver] Initializing Bot {self_id}...")
    
    # Instantiate onebot
    # This will register the bot to the driver via driver.add_bot
    bot = onebot(self_id=self_id)
    bot.set_driver(driver)
    driver.add_bot(bot, self_id)
    
    # Start WebUI
    try:
        from web_ui import start_web_ui
        class DriverAdapter:
            def __init__(self, drv):
                self.driver = drv
                self.config = {}
            @property
            def bots(self):
                return [self.driver.bot] if self.driver.bot else []
            def save_config(self, _): return True
            def add_bot(self, _): pass

        start_web_ui(DriverAdapter(driver), port=5000)
    except Exception as e:
        print(f"[Driver] Failed to start WebUI: {e}")

    # 3. Start Driver (Blocking)
    try:
        # Start WxBot in a separate thread (it has its own blocking loop)
        bot_thread = threading.Thread(target=bot.run, daemon=True)
        bot_thread.start()
        print("[Driver] WxBot thread started")

        asyncio.run(driver.start())
    except KeyboardInterrupt:
        print("[Driver] Stopped by user")
